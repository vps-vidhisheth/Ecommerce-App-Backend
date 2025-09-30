package modules

import (
	"ecommerce/app"
	"ecommerce/components/order/controller"
	"ecommerce/components/order/service"
	"ecommerce/repository"
)

func registerOrderRoutes(appObj *app.App, repository repository.EcommerceRepository) {

	orderService := service.NewOrderService(appObj.DB, repository, []string{})

	orderController := controller.NewOrderController(orderService, appObj.Log)

	appObj.ResigterControllerRoutes([]app.Controller{orderController})
}
