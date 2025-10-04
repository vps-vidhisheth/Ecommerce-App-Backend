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
	authRouter.Use(authmiddleware.AuthMiddleware)
	authRouter.HandleFunc("/me", c.GetMyProfile).Methods(http.MethodGet)
	authRouter.HandleFunc("/me", c.UpdateMyProfile).Methods(http.MethodPut)
	authRouter.HandleFunc("/{id}", c.DeleteUser).Methods(http.MethodDelete)

	adminRouter := userRouter.NewRoute().Subrouter()
	adminRouter.Use(authmiddleware.AuthMiddleware, authmiddleware.AdminMiddleware)
	adminRouter.HandleFunc("", c.GetAllUsers).Methods(http.MethodGet)
	adminRouter.HandleFunc("/{id}/status", c.UpdateUserStatus).Methods(http.MethodPut)
	adminRouter.HandleFunc("/{id}", c.DeleteUser).Methods(http.MethodDelete)

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
		Role:     r.FormValue("role"),
	}

	if err := newUser.Validate(false); err != nil {
		web.RespondError(w, err)
		return
	}

	if file, _, err := r.FormFile("profile_pic"); err == nil {
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

	if file, _, err := r.FormFile("profile_pic"); err == nil {
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
	allUsers := []user.DTO{}
	var totalCount int
	parser := web.NewParser(r)
	limit, offset := parser.ParseLimitAndOffset()
	requestForm := r.URL.Query()

	if err := c.service.GetAllUsers(&allUsers, limit, offset, &totalCount, requestForm); err != nil {
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
		NewPassword string `json:"new_password"`
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
		IsActive bool `json:"is_active"`
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
		"message":   "User status updated successfully",
		"user_id":   id,
		"is_active": req.IsActive,
	})
}
