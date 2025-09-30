package controller

import (
	"ecommerce/components/log"
	"ecommerce/components/login/service"
	"ecommerce/errors"
	"ecommerce/models/credentials"
	"ecommerce/security/token"
	"ecommerce/web"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

type LoginController struct {
	log     log.Log
	service *service.LoginService
}

func NewLoginController(loginService *service.LoginService, log log.Log) *LoginController {
	return &LoginController{
		service: loginService,
		log:     log,
	}
}

func (c *LoginController) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/login", c.UserLogin).Methods(http.MethodPost)
	fmt.Println("Login Route Registered")
}

func (c *LoginController) UserLogin(w http.ResponseWriter, r *http.Request) {
	var userCredentials credentials.Credentials

	err := web.UnmarshalJSON(r, &userCredentials)
	if err != nil {
		c.log.Print(err)
		web.RespondError(w, errors.NewHTTPError(err.Error(), http.StatusBadRequest))
		return
	}

	targetUser, err := c.service.ConfirmUserCredentials(&userCredentials)
	if err != nil {
		c.log.Print(err.Error())
		if err.Error()[:12] == "account lock" {
			web.RespondError(w, errors.NewHTTPError(err.Error(), http.StatusForbidden))
		} else {
			web.RespondError(w, errors.NewHTTPError("invalid email or password", http.StatusUnauthorized))
		}
		return
	}

	authToken, err := token.GenerateAuthToken(targetUser.ID, targetUser.Name, targetUser.Role)
	if err != nil {
		c.log.Print(err.Error())
		web.RespondError(w, errors.NewHTTPError("failed to generate token", http.StatusInternalServerError))
		return
	}

	web.RespondJSON(w, http.StatusOK, map[string]string{
		"token": authToken,
		"role":  targetUser.Role,
	})
}
