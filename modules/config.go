package modules

import (
	"ecommerce/app"
	"ecommerce/models/cart"

	"ecommerce/models/order"
	"ecommerce/models/products"
	"ecommerce/models/user"
)

func Configure(appObj *app.App) {

	userModule := user.NewUserModuleConfig(appObj.DB)
	appObj.MigrateTables([]app.ModuleConfig{userModule})

	productModule := products.NewProductModuleConfig(appObj.DB)
	appObj.MigrateTables([]app.ModuleConfig{productModule})

	cartModule := cart.NewCartModuleConfig(appObj.DB)
	appObj.MigrateTables([]app.ModuleConfig{cartModule})

	orderModule := order.NewOrderModuleConfig(appObj.DB)
	appObj.MigrateTables([]app.ModuleConfig{orderModule})
}
