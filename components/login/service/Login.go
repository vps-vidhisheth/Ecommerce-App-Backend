package service

import (
	"fmt"
	"time"

	"ecommerce/models/credentials"
	"ecommerce/models/login"
	"ecommerce/models/user"
	"ecommerce/repository"
	security "ecommerce/security/captcha"
	"ecommerce/security/token"

	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

type LoginService struct {
	db            *gorm.DB
	repository    repository.EcommerceRepository
	loginAttempts map[string]*login.LoginAttempt
}

func NewLoginService(db *gorm.DB, repo repository.EcommerceRepository) *LoginService {
	return &LoginService{
		db:            db,
		repository:    repo,
		loginAttempts: make(map[string]*login.LoginAttempt),
	}
}

func (s *LoginService) ConfirmUserCredentials(userCreds *credentials.Credentials) (user.User, string, error) {
	if err := security.VerifyRecaptcha(userCreds.Captcha); err != nil {
		return user.User{}, "", fmt.Errorf("captcha verification failed: %v", err)
	}

	uow := repository.NewUnitOfWork(s.db, false)
	defer uow.RollBack()

	var tempUser user.User
	err := s.repository.GetRecord(uow, &tempUser, repository.Filter("`email` = ?", userCreds.Email))
	if err != nil {
		return tempUser, "", err
	}

	if !tempUser.IsActive {
		return tempUser, "", fmt.Errorf("account is inactive, please contact admin")
	}

	if userCreds.Role != "" && tempUser.Role != userCreds.Role {
		return tempUser, "", fmt.Errorf("invalid role for this account")
	}

	attempt, exists := s.loginAttempts[tempUser.Email]
	if !exists {
		attempt = &login.LoginAttempt{UserID: tempUser.ID}
		s.loginAttempts[tempUser.Email] = attempt
	}

	if attempt.LockedUntil != nil && attempt.LockedUntil.After(time.Now()) {
		return tempUser, "", fmt.Errorf("account locked until %v", attempt.LockedUntil)
	}

	err = bcrypt.CompareHashAndPassword([]byte(tempUser.Password), []byte(userCreds.Password))
	if err != nil {
		attempt.FailedCount++
		attempt.LastAttempt = time.Now()

		if attempt.FailedCount >= 3 {
			lockTime := time.Now().Add(15 * time.Minute)
			attempt.LockedUntil = &lockTime
			attempt.FailedCount = 0
		}

		return tempUser, "", fmt.Errorf("invalid email or password or role ")
	}

	attempt.FailedCount = 0
	attempt.LockedUntil = nil

	authToken, err := token.GenerateAuthToken(tempUser.ID, tempUser.Name, tempUser.Role)
	if err != nil {
		return tempUser, "", fmt.Errorf("failed to generate token: %v", err)
	}

	return tempUser, authToken, nil
}
