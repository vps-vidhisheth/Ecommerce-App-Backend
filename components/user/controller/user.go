package controller

import (
	"ecommerce/components/log"
	"ecommerce/components/user/service"
	"ecommerce/errors"
	"ecommerce/models/baseStruct"
	"ecommerce/models/user"
	authmiddleware "ecommerce/security/authMiddleWare"
	"ecommerce/web"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type UserController struct {
	log     log.Log
	service *service.UserService
}

func NewUserController(userService *service.UserService, logger log.Log) *UserController {
	return &UserController{
		service: userService,
		log:     logger,
	}
}

func (c *UserController) RegisterRoutes(router *mux.Router) {
	userRouter := router.PathPrefix("/users").Subrouter()

	userRouter.HandleFunc("/register", c.RegisterUser).Methods(http.MethodPost)
	userRouter.HandleFunc("/request-password-reset", c.RequestPasswordReset).Methods(http.MethodPost)
	userRouter.HandleFunc("/reset-password", c.ResetPasswordWithToken).Methods(http.MethodPost)

	authRouter := userRouter.NewRoute().Subrouter()
	authRouter.Use(func(next http.Handler) http.Handler {
		return authmiddleware.AuthMiddleware(c.service, next)
	})
	authRouter.HandleFunc("/me", c.GetMyProfile).Methods(http.MethodGet)
	authRouter.HandleFunc("/me", c.UpdateMyProfile).Methods(http.MethodPut)
	authRouter.HandleFunc("/{id}", c.DeleteUser).Methods(http.MethodDelete)

	adminRouter := userRouter.NewRoute().Subrouter()
	adminRouter.Use(
		func(next http.Handler) http.Handler {
			return authmiddleware.AuthMiddleware(c.service, next)
		},
		authmiddleware.AdminMiddleware,
	)
	adminRouter.HandleFunc("", c.GetAllUsers).Methods(http.MethodGet)
	adminRouter.HandleFunc("/{id}/status", c.UpdateUserStatus).Methods(http.MethodPut)
	adminRouter.HandleFunc("/{id}", c.DeleteUser).Methods(http.MethodDelete)
	adminRouter.HandleFunc("/{id}/role", c.UpdateUserRole).Methods(http.MethodPut)

	c.log.Print("======== User Routes Registered =========")
}

func (c *UserController) RegisterUser(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		web.RespondError(w, err)
		return
	}

	newUser := user.User{
		Name:     r.FormValue("name"),
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
		Role:     "customer",
	}

	if err := newUser.Validate(false); err != nil {
		web.RespondError(w, err)
		return
	}

	if file, _, err := r.FormFile("profilePic"); err == nil {
		defer file.Close()
		if fileBytes, err := io.ReadAll(file); err == nil {
			newUser.ProfilePic = fileBytes
		} else {
			web.RespondError(w, err)
			return
		}
	}

	if err := c.service.CreateUser(&newUser); err != nil {
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusCreated, newUser)
}

func (c *UserController) UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	role, err := authmiddleware.GetUserRoleFromContext(r.Context())
	if err != nil {
		web.RespondError(w, errors.NewValidationError("Unauthorized: invalid token"))
		return
	}

	if !strings.EqualFold(role, "ADMIN") {
		web.RespondError(w, errors.NewValidationError("Unauthorized: only admin can update user roles"))
		return
	}

	idParam := mux.Vars(r)["id"]
	userID, parseErr := uuid.Parse(idParam)
	if parseErr != nil {
		web.RespondError(w, errors.NewValidationError("Invalid user ID"))
		return
	}

	var req struct {
		Role string `json:"role"`
	}
	if err := web.UnmarshalJSON(r, &req); err != nil {
		web.RespondError(w, errors.NewValidationError("Invalid request body"))
		return
	}

	roleLower := strings.ToLower(req.Role)
	if roleLower != "admin" && roleLower != "customer" {
		web.RespondError(w, errors.NewValidationError("Invalid role: only 'admin' or 'customer' allowed"))
		return
	}

	if err := c.service.UpdateUserRole(userID, req.Role); err != nil {
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusOK, map[string]string{
		"message": "User role updated successfully",
	})
}

func (c *UserController) UpdateMyProfile(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		web.RespondError(w, err)
		return
	}

	userID, httpErr := authmiddleware.GetUserIDFromContext(r.Context())
	if httpErr != nil {
		web.RespondError(w, httpErr)
		return
	}

	userToUpdate := user.User{
		Base:  baseStruct.Base{ID: userID},
		Name:  r.FormValue("name"),
		Email: r.FormValue("email"),
	}

	if password := r.FormValue("password"); password != "" {
		userToUpdate.Password = password
	}

	if file, _, err := r.FormFile("profilePic"); err == nil {
		defer file.Close()
		if fileBytes, err := io.ReadAll(file); err == nil {
			userToUpdate.ProfilePic = fileBytes
		} else {
			web.RespondError(w, err)
			return
		}
	}

	if userToUpdate.Name == "" || userToUpdate.Email == "" {
		web.RespondError(w, errors.NewValidationError("Name and Email are required"))
		return
	}

	if err := c.service.UpdateUserProfile(&userToUpdate); err != nil {
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusOK, userToUpdate)
}

func (c *UserController) GetMyProfile(w http.ResponseWriter, r *http.Request) {
	userID, httpErr := authmiddleware.GetUserIDFromContext(r.Context())
	if httpErr != nil {
		web.RespondError(w, httpErr)
		return
	}

	u, err := c.service.GetUserByID(userID)
	if err != nil {
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusOK, u)
}

func (c *UserController) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		web.RespondError(w, err)
		return
	}

	u := user.User{Base: baseStruct.Base{ID: id}}
	if err := c.service.DeleteUser(&u); err != nil {
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusOK, map[string]string{"message": "User deleted successfully"})
}

func (c *UserController) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	userID, httpErr := authmiddleware.GetUserIDFromContext(r.Context())
	if httpErr != nil {
		web.RespondError(w, httpErr)
		return
	}

	allUsers := []user.DTO{}
	var totalCount int
	parser := web.NewParser(r)
	limit, offset := parser.ParseLimitAndOffset()
	requestForm := r.URL.Query()

	if err := c.service.GetAllUsers(&allUsers, limit, offset, &totalCount, requestForm, userID); err != nil {
		web.RespondError(w, err)
		return
	}

	web.RespondJSONWithXTotalCount(w, http.StatusOK, totalCount, allUsers)
}

func (c *UserController) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := web.UnmarshalJSON(r, &req); err != nil {
		web.RespondError(w, err)
		return
	}

	token, err := c.service.GenerateResetToken(req.Email, 15)
	if err != nil {
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusOK, map[string]string{
		"message": "Password reset token generated. Check your email.",
		"token":   token,
	})
}

func (c *UserController) ResetPasswordWithToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email       string `json:"email"`
		Token       string `json:"token"`
		NewPassword string `json:"newPassword"`
	}

	if err := web.UnmarshalJSON(r, &req); err != nil {
		web.RespondError(w, err)
		return
	}

	if err := c.service.ResetPasswordWithToken(req.Email, req.Token, req.NewPassword); err != nil {
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusOK, map[string]string{"message": "Password reset successfully"})
}

func (c *UserController) UpdateUserStatus(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		web.RespondError(w, err)
		return
	}

	var req struct {
		IsActive bool `json:"isActive"`
	}
	if err := web.UnmarshalJSON(r, &req); err != nil {
		web.RespondError(w, err)
		return
	}

	if err := c.service.UpdateUserStatus(id, req.IsActive); err != nil {
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"message":  "User status updated successfully",
		"userID":   id,
		"isActive": req.IsActive,
	})
}
