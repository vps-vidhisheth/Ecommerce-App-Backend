package controller

import (
	"ecommerce/components/log"
	"ecommerce/components/order/service"
	Userservice "ecommerce/components/user/service"
	"ecommerce/errors"
	"ecommerce/models/baseStruct"
	"ecommerce/models/order"
	authmiddleware "ecommerce/security/authMiddleWare"
	"ecommerce/web"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type OrderController struct {
	log         log.Log
	service     *service.OrderService
	userService *Userservice.UserService
}

func NewOrderController(orderService *service.OrderService, logger log.Log, userService *Userservice.UserService) *OrderController {
	return &OrderController{
		log:         logger,
		service:     orderService,
		userService: userService,
	}
}

func (c *OrderController) RegisterRoutes(router *mux.Router) {
	orderRouter := router.PathPrefix("/orders").Subrouter()

	authMiddleware := func(next http.Handler) http.Handler {
		return authmiddleware.AuthMiddleware(c.userService, next)
	}
	orderRouter.Use(authMiddleware)

	orderRouter.HandleFunc("", c.CreateOrder).Methods(http.MethodPost)
	orderRouter.HandleFunc("/user/{userID}", c.GetOrdersByUserID).Methods(http.MethodGet)
	orderRouter.HandleFunc("/{id}", c.DeleteOrder).Methods(http.MethodDelete)

	adminRouter := router.PathPrefix("/orders").Subrouter()
	adminRouter.Use(authMiddleware)
	adminRouter.Use(authmiddleware.AdminMiddleware)

	adminRouter.HandleFunc("", c.GetAllOrders).Methods(http.MethodGet)

	c.log.Print("======== Order Routes Registered =========")
}

func (c *OrderController) CreateOrder(w http.ResponseWriter, r *http.Request) {
	web.RespondError(w, errors.NewHTTPError(
		"Orders must be created after payment", http.StatusForbidden))
}

func (c *OrderController) DeleteOrder(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := uuid.Parse(idStr)
	if err != nil {
		web.RespondError(w, err)
		return
	}

	o := order.Order{Base: baseStruct.Base{ID: id}}

	if err := c.service.DeleteOrder(&o); err != nil {
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusOK, map[string]string{
		"message": "Order cancelled successfully (within allowed time frame)",
	})
}

func (c *OrderController) GetOrdersByUserID(w http.ResponseWriter, r *http.Request) {
	userIDStr := mux.Vars(r)["userID"]
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		web.RespondError(w, errors.NewValidationError("Invalid User ID"))
		return
	}

	orders, err := c.service.GetOrdersByUserID(userID)
	if err != nil {
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusOK, orders)
}

func (c *OrderController) GetAllOrders(w http.ResponseWriter, r *http.Request) {
	allOrders := []order.Order{}
	var totalCount int

	limit, offset := web.NewParser(r).ParseLimitAndOffset()

	if err := c.service.GetAllOrders(&allOrders, limit, offset, &totalCount); err != nil {
		web.RespondError(w, err)
		return
	}

	web.RespondJSONWithXTotalCount(w, http.StatusOK, totalCount, allOrders)
}
