package modules

import (
	"ecommerce/app"
	"ecommerce/components/products/controller"
	"ecommerce/components/products/service"
	userService "ecommerce/components/user/service"
	"ecommerce/repository"
)

func registerProductRoutes(appObj *app.App, repository repository.EcommerceRepository, userSvc *userService.UserService) {

	productService := service.NewProductService(appObj.DB, repository, []string{})

	productController := controller.NewProductController(productService, appObj.Log)

	// Pass userSvc explicitly to RegisterRoutes
	productController.RegisterRoutes(appObj.Router, userSvc)
}
