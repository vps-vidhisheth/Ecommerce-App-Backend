package modules

import (
	"ecommerce/app"
	"ecommerce/components/cart/controller"
	"ecommerce/components/cart/service"
	userService "ecommerce/components/user/service"
	"ecommerce/repository"
)

func registerCartRoutes(appObj *app.App, repository repository.EcommerceRepository, userSvc *userService.UserService) {
	cartService := service.NewCartService(appObj.DB, repository, []string{})

	cartController := controller.NewCartController(cartService, appObj.Log, userSvc)

	appObj.ResigterControllerRoutes([]app.Controller{cartController})
}
