package modules

import (
	"ecommerce/app"
	"ecommerce/components/cart/controller"
	"ecommerce/components/cart/service"
	"ecommerce/repository"
)

func registerCartRoutes(appObj *app.App, repository repository.EcommerceRepository) {

	cartService := service.NewCartService(appObj.DB, repository, []string{})

	cartController := controller.NewCartController(cartService, appObj.Log)

	appObj.ResigterControllerRoutes([]app.Controller{cartController})
}
