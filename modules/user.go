package modules

import (
	"ecommerce/app"
	"ecommerce/components/user/controller"
	"ecommerce/components/user/service"
	"ecommerce/repository"
)

func registerUserRoutes(appObj *app.App, repository repository.EcommerceRepository) {

	userService := service.NewUserService(appObj.DB, repository, []string{})

	userController := controller.NewUserController(userService, appObj.Log)

	appObj.ResigterControllerRoutes([]app.Controller{userController})
}
