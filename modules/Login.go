package modules

import (
	"ecommerce/app"
	"ecommerce/components/login/controller"
	"ecommerce/components/login/service"
	"ecommerce/repository"
)

func registerLoginRoutes(appObj *app.App, repository repository.EcommerceRepository) {

	CustomerService := service.NewLoginService(appObj.DB, repository)
	CustomerController := controller.NewLoginController(CustomerService, appObj.Log)
	appObj.ResigterControllerRoutes([]app.Controller{CustomerController})
}
