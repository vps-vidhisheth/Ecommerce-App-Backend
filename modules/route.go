package modules

import (
	"ecommerce/app"
	"ecommerce/repository"
)

func RegisterModuleRoutes(app *app.App, repo repository.EcommerceRepository) {
	log := app.Log
	log.Print("======== RegisterModuleRoutes.go ========")

	registerUserRoutes(app, repo)
	registerLoginRoutes(app, repo)
	registerProductRoutes(app, repo)
	registerCartRoutes(app, repo)
	registerOrderRoutes(app, repo)
	registerpaymentsRoutes(app, repo)

	log.Print("[ All module routes registered successfully ]")
}
