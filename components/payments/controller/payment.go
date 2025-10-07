package controller

import (
	cartService "ecommerce/components/cart/service"
	"ecommerce/components/log"
	orderService "ecommerce/components/order/service"
	paymentService "ecommerce/components/payments/service"
	Userservice "ecommerce/components/user/service"
	orderModel "ecommerce/models/order"
	paymentModel "ecommerce/models/payments"
	authmiddleware "ecommerce/security/authMiddleWare"
	"ecommerce/web"
	"net/http"

	"github.com/gorilla/mux"
)

type PaymentController struct {
	log          log.Log
	service      *paymentService.PaymentService
	orderService *orderService.OrderService
	cartService  *cartService.CartService
}

func NewPaymentController(paymentService *paymentService.PaymentService, orderService *orderService.OrderService, cartService *cartService.CartService, logger log.Log) *PaymentController {
	return &PaymentController{
		log:          logger,
		service:      paymentService,
		orderService: orderService,
		cartService:  cartService,
	}
}

func (c *PaymentController) RegisterRoutes(router *mux.Router, userService *Userservice.UserService) {
	paymentRouter := router.PathPrefix("/payments").Subrouter()

	authMiddleware := func(next http.Handler) http.Handler {
		return authmiddleware.AuthMiddleware(userService, next)
	}
	paymentRouter.Use(authMiddleware)

	paymentRouter.HandleFunc("", c.CreatePayment).Methods(http.MethodPost)

	adminRouter := router.PathPrefix("/payments").Subrouter()
	adminRouter.Use(authMiddleware)
	adminRouter.Use(authmiddleware.AdminMiddleware)

	adminRouter.HandleFunc("", c.GetAllPayments).Methods(http.MethodGet)

	c.log.Print("======== Payment Routes Registered =========")
}

func (c *PaymentController) CreatePayment(w http.ResponseWriter, r *http.Request) {
	var newPayment paymentModel.Payment
	if err := web.UnmarshalJSON(r, &newPayment); err != nil {
		web.RespondError(w, err)
		return
	}

	userID, httpErr := authmiddleware.GetUserIDFromContext(r.Context())
	if httpErr != nil {
		web.RespondError(w, httpErr)
		return
	}

	totalAmount, err := c.cartService.GetTotalAmountByCartID(newPayment.CartID)
	if err != nil {
		c.log.Print("Failed to calculate cart total: ", err)
		web.RespondError(w, err)
		return
	}

	newPayment.UserID = userID
	newPayment.Amount = totalAmount

	newOrder := orderModel.Order{
		UserID: userID,
		CartID: newPayment.CartID,
	}

	if err := c.orderService.CreateOrder(&newOrder, totalAmount); err != nil {
		c.log.Print("Order creation failed: ", err)
		web.RespondError(w, err)
		return
	}

	newPayment.OrderID = newOrder.ID
	if err := newPayment.Validate(false); err != nil {
		web.RespondError(w, err)
		return
	}

	if err := c.service.CreatePayment(&newPayment); err != nil {
		c.log.Print("Payment creation failed: ", err)
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusCreated, map[string]interface{}{
		"payment": newPayment,
		"order":   newOrder,
	})
}

func (c *PaymentController) GetAllPayments(w http.ResponseWriter, r *http.Request) {
	allPayments := []paymentModel.Payment{}

	if err := c.service.GetAllPayments(&allPayments); err != nil {
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusOK, allPayments)
}
