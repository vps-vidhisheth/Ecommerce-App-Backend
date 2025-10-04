package modules

import (
	"ecommerce/app"
	cartService "ecommerce/components/cart/service"
	"ecommerce/components/order/service"
	paymentController "ecommerce/components/payments/controller"
	paymentService "ecommerce/components/payments/service"
	"ecommerce/repository"
)

func registerpaymentsRoutes(appObj *app.App, repository repository.EcommerceRepository) {

	paymentsService := paymentService.NewPaymentService(appObj.DB, repository)

	orderSvc := service.NewOrderService(appObj.DB, repository, []string{})
	cartsSvc := cartService.NewCartService(appObj.DB, repository, []string{})

	paymentsController := paymentController.NewPaymentController(paymentsService, orderSvc, cartsSvc, appObj.Log)

	appObj.ResigterControllerRoutes([]app.Controller{paymentsController})
}
