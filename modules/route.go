package modules

import (
	"ecommerce/app"
	userService "ecommerce/components/user/service"
	"ecommerce/repository"
)

func RegisterModuleRoutes(app *app.App, repo repository.EcommerceRepository) {
	// Create UserService internally
	userSvc := userService.NewUserService(app.DB, repo, []string{})

	registerUserRoutes(app, repo)
	registerLoginRoutes(app, repo)
	registerProductRoutes(app, repo, userSvc)
	registerCartRoutes(app, repo, userSvc)
	registerOrderRoutes(app, repo, userSvc)
	registerPaymentsRoutes(app, repo, userSvc)
}
