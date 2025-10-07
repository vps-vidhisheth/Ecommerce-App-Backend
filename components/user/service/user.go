package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
	"time"

	"ecommerce/errors"
	"ecommerce/models/user"
	"ecommerce/repository"
	"ecommerce/util"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	db           *gorm.DB
	repository   repository.EcommerceRepository
	associations []string
}

func NewUserService(db *gorm.DB, repo repository.EcommerceRepository, associations []string) *UserService {
	return &UserService{
		db:           db,
		repository:   repo,
		associations: associations,
	}
}

func (s *UserService) hashPassword(password string) (string, error) {
	if password == "" {
		return "", nil
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func (s *UserService) doesUserExist(ID uuid.UUID) (*user.User, error) {
	uow := repository.NewUnitOfWork(s.db, true)
	defer uow.RollBack()

	var existing user.User
	if err := s.repository.GetRecordForUser(uow, ID, &existing, "id"); err != nil {
		return nil, errors.NewValidationError("User ID is invalid")
	}

	uow.Commit()
	return &existing, nil
}

func (s *UserService) CreateUser(newUser *user.User) error {
	uow := repository.NewUnitOfWork(s.db, false)
	defer uow.RollBack()

	hashed, err := s.hashPassword(newUser.Password)
	if err != nil {
		return err
	}
	newUser.Password = hashed
	newUser.IsActive = true

	if err := s.repository.Add(uow, newUser); err != nil {
		return err
	}

	uow.Commit()
	return nil
}

func (s *UserService) UpdateUserRole(userID uuid.UUID, newRole string) error {
	uow := repository.NewUnitOfWork(s.db, false)
	defer uow.RollBack()

	var targetUser user.User
	if err := s.repository.GetRecordForUser(uow, userID, &targetUser, "id"); err != nil {
		return errors.NewValidationError("User not found")
	}

	if newRole != "admin" && newRole != "customer" {
		return errors.NewValidationError("Invalid role: must be 'admin' or 'customer'")
	}

	if targetUser.Role == newRole {
		return errors.NewValidationError(fmt.Sprintf("User already has role '%s'", newRole))
	}

	updateMap := map[string]interface{}{
		"Role":      newRole,
		"UpdatedAt": time.Now(),
	}

	if err := s.repository.UpdateWithMap(uow, &targetUser, updateMap, repository.Filter("id = ?", userID)); err != nil {
		return err
	}

	uow.Commit()
	return nil
}

func (s *UserService) UpdateUserProfile(userToUpdate *user.User) error {
	existing, err := s.doesUserExist(userToUpdate.ID)
	if err != nil {
		return err
	}

	uow := repository.NewUnitOfWork(s.db, false)
	defer uow.RollBack()

	userToUpdate.CreatedAt = existing.CreatedAt
	userToUpdate.UpdatedAt = time.Now()

	if userToUpdate.Password != "" {
		hashed, err := s.hashPassword(userToUpdate.Password)
		if err != nil {
			return err
		}
		userToUpdate.Password = hashed
	} else {
		userToUpdate.Password = existing.Password
	}

	userToUpdate.Role = existing.Role

	if err := s.repository.Update(uow, userToUpdate); err != nil {
		return err
	}

	uow.Commit()
	return nil
}

func (s *UserService) DeleteUser(userToDelete *user.User) error {
	existing, err := s.doesUserExist(userToDelete.ID)
	if err != nil {
		return err
	}

	uow := repository.NewUnitOfWork(s.db, false)
	defer uow.RollBack()

	now := time.Now()
	updateMap := map[string]interface{}{
		"DeletedAt":           now,
		"IsActive":            false,
		"ResetToken":          "",
		"ResetTokenExpiresAt": nil,
	}

	if err := s.repository.UpdateWithMap(uow, existing, updateMap, repository.Filter("id = ?", existing.ID)); err != nil {
		return err
	}

	uow.Commit()
	return nil
}

func (s *UserService) GetUserByID(id uuid.UUID) (*user.User, error) {
	uow := repository.NewUnitOfWork(s.db, true)
	defer uow.RollBack()

	var u user.User
	if err := s.repository.GetRecordForUser(uow, id, &u, "id"); err != nil {
		return nil, err
	}

	uow.Commit()
	return &u, nil
}

func (s *UserService) GetAllUsers(allUsers *[]user.DTO, limit, offset int, totalCount *int, requestForm url.Values, currentUserID uuid.UUID) error {
	uow := repository.NewUnitOfWork(s.db, true)
	defer uow.RollBack()

	var queryProcessors []repository.QueryProcessor

	searchQuery := s.buildSearchQuery(requestForm)
	if searchQuery != nil {
		queryProcessors = append(queryProcessors, searchQuery)
	}

	queryProcessors = append(queryProcessors, repository.Filter("id != ?", currentUserID))

	queryProcessors = append(queryProcessors, repository.NotDeleted())
	queryProcessors = append(queryProcessors, repository.Paginate(limit, offset, totalCount))

	if err := s.repository.GetAll(uow, allUsers, queryProcessors...); err != nil {
		return err
	}

	uow.Commit()
	return nil
}

func (s *UserService) buildSearchQuery(requestForm url.Values) repository.QueryProcessor {
	if len(requestForm) == 0 {
		return nil
	}

	var columnNames []string
	var conditions []string
	var operators []string
	var values []interface{}

	if searchTerm := requestForm.Get("search"); searchTerm != "" {
		columns := []string{"`name`", "`email`", "`role`"}

		for _, col := range columns {
			util.AddToSlice(col, "LIKE ?", "OR", "%"+searchTerm+"%", &columnNames, &conditions, &operators, &values)
		}
	}

	return repository.FilterWithOperator(columnNames, conditions, operators, values)
}

func (s *UserService) GenerateResetToken(email string, validityMinutes int) (string, error) {
	uow := repository.NewUnitOfWork(s.db, false)
	defer uow.RollBack()

	var u user.User
	if err := s.repository.GetRecord(uow, &u, repository.Filter("email = ?", email)); err != nil {
		return "", errors.NewValidationError("Email not found")
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	token := hex.EncodeToString(tokenBytes)
	expiry := time.Now().Add(time.Duration(validityMinutes) * time.Minute)

	updateMap := map[string]interface{}{
		"ResetToken":          token,
		"ResetTokenExpiresAt": &expiry,
	}
	if err := s.repository.UpdateWithMap(uow, &u, updateMap, repository.Filter("id = ?", u.ID)); err != nil {
		return "", err
	}
	uow.Commit()

	subject := "Password Reset Request"
	body := fmt.Sprintf("Hello %s,\n\nYou requested a password reset.\nYour reset token is: %s\nIt expires in %d minutes.\n\nIf you didn't request this, ignore this email.", u.Name, token, validityMinutes)

	if err := util.SendEmail(u.Email, subject, body); err != nil {
		return "", fmt.Errorf("failed to send email: %v", err)
	}

	return token, nil
}

func (s *UserService) EnsureActiveUser(userID uuid.UUID, tokenIAT int64) (*user.User, error) {
	u, err := s.doesUserExist(userID)
	if err != nil {
		return nil, err
	}

	if !u.IsActive {
		return nil, errors.NewValidationError("User is inactive. Please login again.")
	}

	if u.LastDeactivatedAt != nil && tokenIAT < u.LastDeactivatedAt.Unix() {
		return nil, errors.NewValidationError("Token issued before deactivation. Please login again.")
	}

	return u, nil
}

func (s *UserService) ResetPasswordWithToken(email, token, newPassword string) error {
	uow := repository.NewUnitOfWork(s.db, false)
	defer uow.RollBack()

	var u user.User
	if err := s.repository.GetRecord(uow, &u, repository.Filter("email = ?", email)); err != nil {
		return errors.NewValidationError("Email not found")
	}

	if u.ResetToken == "" || u.ResetToken != token {
		return errors.NewValidationError("Invalid reset token")
	}
	if u.ResetTokenExpiresAt == nil || u.ResetTokenExpiresAt.Before(time.Now()) {
		return errors.NewValidationError("Reset token expired")
	}

	hashed, err := s.hashPassword(newPassword)
	if err != nil {
		return err
	}
	u.Password = hashed
	u.ResetToken = ""
	u.ResetTokenExpiresAt = nil
	u.UpdatedAt = time.Now()

	if err := s.repository.Update(uow, &u); err != nil {
		return err
	}

	uow.Commit()
	return nil
}

func (s *UserService) UpdateUserStatus(userID uuid.UUID, isActive bool) error {
	existing, err := s.doesUserExist(userID)
	if err != nil {
		return err
	}

	uow := repository.NewUnitOfWork(s.db, false)
	defer uow.RollBack()

	updateMap := map[string]interface{}{
		"IsActive":  isActive,
		"UpdatedAt": time.Now(),
	}

	if !isActive {
		now := time.Now()
		updateMap["LastDeactivatedAt"] = &now
	}

	if err := s.repository.UpdateWithMap(uow, existing, updateMap, repository.Filter("id = ?", existing.ID)); err != nil {
		return err
	}

	uow.Commit()
	return nil
}
