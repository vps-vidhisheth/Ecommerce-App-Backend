package controller

import (
	"ecommerce/components/log"
	"ecommerce/components/order/service"
	"ecommerce/errors"
	"ecommerce/models/order"
	authmiddleware "ecommerce/security/authMiddleWare"
	"ecommerce/web"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type OrderController struct {
	log     log.Log
	service *service.OrderService
}

func NewOrderController(orderService *service.OrderService, logger log.Log) *OrderController {
	return &OrderController{
		log:     logger,
		service: orderService,
	}
}

func (c *OrderController) RegisterRoutes(router *mux.Router) {
	orderRouter := router.PathPrefix("/orders").Subrouter()

	orderRouter.Handle("", authmiddleware.AuthMiddleware(http.HandlerFunc(c.CreateOrder))).Methods(http.MethodPost)
	orderRouter.Handle("/{id}", authmiddleware.AuthMiddleware(http.HandlerFunc(c.GetOrderByID))).Methods(http.MethodGet)
	orderRouter.Handle("/user/{userID}", authmiddleware.AuthMiddleware(http.HandlerFunc(c.GetOrdersByUserID))).Methods(http.MethodGet)
	orderRouter.Handle("/{id}", authmiddleware.AuthMiddleware(http.HandlerFunc(c.UpdateOrder))).Methods(http.MethodPut)
	orderRouter.Handle("/{id}", authmiddleware.AuthMiddleware(http.HandlerFunc(c.DeleteOrder))).Methods(http.MethodDelete)
	orderRouter.Handle("/user/{user_id}/total", authmiddleware.AuthMiddleware(http.HandlerFunc(c.GetTotalAmountByUserID))).Methods(http.MethodGet)

	adminRouter := router.PathPrefix("/orders").Subrouter()
	adminRouter.Handle("", authmiddleware.AuthMiddleware(authmiddleware.AdminMiddleware(http.HandlerFunc(c.GetAllOrders)))).Methods(http.MethodGet)

	c.log.Print("======== Order Routes Registered =========")
}

func (c *OrderController) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var newOrder order.Order
	if err := web.UnmarshalJSON(r, &newOrder); err != nil {
		web.RespondError(w, err)
		return
	}

	if err := newOrder.Validate(false); err != nil {
		web.RespondError(w, err)
		return
	}

	if err := c.service.CreateOrder(&newOrder); err != nil {
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusCreated, newOrder)
}

func (c *OrderController) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := uuid.Parse(idStr)
	if err != nil {
		web.RespondError(w, err)
		return
	}

	var orderToUpdate order.Order
	if err := web.UnmarshalJSON(r, &orderToUpdate); err != nil {
		web.RespondError(w, err)
		return
	}
	orderToUpdate.ID = id

	if err := orderToUpdate.Validate(true); err != nil {
		web.RespondError(w, err)
		return
	}

	if err := c.service.UpdateOrder(&orderToUpdate); err != nil {
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusOK, orderToUpdate)
}

func (c *OrderController) DeleteOrder(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := uuid.Parse(idStr)
	if err != nil {
		web.RespondError(w, err)
		return
	}

	o := order.Order{ID: id} // order  struct is there with only id field
	if err := c.service.DeleteOrder(&o); err != nil {
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusOK, map[string]string{"message": "Order cancelled successfully (within allowed time frame)"})
}

func (c *OrderController) GetOrderByID(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := uuid.Parse(idStr)
	if err != nil {
		web.RespondError(w, err)
		return
	}

	o, err := c.service.GetOrderByID(id)
	if err != nil {
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusOK, o)
}

func (c *OrderController) GetAllOrders(w http.ResponseWriter, r *http.Request) {
	allOrders := []order.Order{}
	var totalCount int
	requestForm := r.URL.Query()

	limit, offset := web.NewParser(r).ParseLimitAndOffset()

	if err := c.service.GetAllOrders(&allOrders, limit, offset, &totalCount, requestForm); err != nil {
		web.RespondError(w, err)
		return
	}

	web.RespondJSONWithXTotalCount(w, http.StatusOK, totalCount, allOrders)
}

func (c *OrderController) GetOrdersByUserID(w http.ResponseWriter, r *http.Request) {
	userIDStr := mux.Vars(r)["userID"]
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		web.RespondError(w, err)
		return
	}

	orders, err := c.service.GetOrdersByUserID(userID)
	if err != nil {
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusOK, orders)
}

func (c *OrderController) GetTotalAmountByUserID(w http.ResponseWriter, r *http.Request) {
	parser := web.NewParser(r)
	userIDStr := parser.GetParameter("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		web.RespondError(w, errors.NewValidationError("Invalid User ID"))
		return
	}

	total, err := c.service.GetTotalAmountByUserID(userID)
	if err != nil {
		c.log.Print(err.Error())
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusOK, map[string]float64{"total_amount": total})
}
