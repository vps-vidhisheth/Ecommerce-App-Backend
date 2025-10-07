package modules

import (
	"ecommerce/app"
	cartService "ecommerce/components/cart/service"
	"ecommerce/components/order/service"
	paymentController "ecommerce/components/payments/controller"
	paymentService "ecommerce/components/payments/service"
	userService "ecommerce/components/user/service"
	"ecommerce/repository"
)

func registerPaymentsRoutes(appObj *app.App, repository repository.EcommerceRepository, userSvc *userService.UserService) {

	paymentsService := paymentService.NewPaymentService(appObj.DB, repository)

	orderSvc := service.NewOrderService(appObj.DB, repository, []string{})
	cartsSvc := cartService.NewCartService(appObj.DB, repository, []string{})

	paymentsController := paymentController.NewPaymentController(paymentsService, orderSvc, cartsSvc, appObj.Log)

	// Pass userSvc explicitly to RegisterRoutes
	paymentsController.RegisterRoutes(appObj.Router, userSvc)
}
