package modules

import (
	"ecommerce/app"
	"ecommerce/models/cart"
	"ecommerce/models/order"
	payment "ecommerce/models/payments"
	"ecommerce/models/products"
	"ecommerce/models/user"
)

func Configure(appObj *app.App) {
	modules := []app.ModuleConfig{
		user.NewUserModuleConfig(appObj.DB),
		products.NewProductModuleConfig(appObj.DB),
		cart.NewCartModuleConfig(appObj.DB),
		order.NewOrderModuleConfig(appObj.DB),
		payment.NewPaymentModuleConfig(appObj.DB),
	}
	appObj.MigrateTables(modules)
}
