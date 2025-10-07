package modules

import (
	"ecommerce/app"
	"ecommerce/components/order/controller"
	"ecommerce/components/order/service"
	userService "ecommerce/components/user/service"
	"ecommerce/repository"
)

func registerOrderRoutes(appObj *app.App, repository repository.EcommerceRepository, userSvc *userService.UserService) {
	orderService := service.NewOrderService(appObj.DB, repository, []string{})

	orderController := controller.NewOrderController(orderService, appObj.Log, userSvc)

	appObj.ResigterControllerRoutes([]app.Controller{orderController})
}
