package modules

import (
	"ecommerce/app"
	"ecommerce/components/products/controller"
	"ecommerce/components/products/service"
	"ecommerce/repository"
)

func registerProductRoutes(appObj *app.App, repository repository.EcommerceRepository) {

	productService := service.NewProductService(appObj.DB, repository, []string{})

	productController := controller.NewProductController(productService, appObj.Log)

	appObj.ResigterControllerRoutes([]app.Controller{productController})
}
