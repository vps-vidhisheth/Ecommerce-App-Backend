package controller

import (
	"ecommerce/components/cart/service"
	"ecommerce/components/log"
	"ecommerce/errors"
	"ecommerce/models/cart"
	authmiddleware "ecommerce/security/authMiddleWare"
	"ecommerce/web"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type CartController struct {
	log     log.Log
	service *service.CartService
}

func NewCartController(cartService *service.CartService, logger log.Log) *CartController {
	return &CartController{
		log:     logger,
		service: cartService,
	}
}

func (c *CartController) RegisterRoutes(router *mux.Router) {
	cartRouter := router.PathPrefix("/carts").Subrouter()
	cartRouter.Use(authmiddleware.AuthMiddleware)

	cartRouter.Handle("/{id}", http.HandlerFunc(c.GetCartByID)).Methods(http.MethodGet)
	cartRouter.Handle("/user/{user_id}", http.HandlerFunc(c.GetCartByUserID)).Methods(http.MethodGet)
	cartRouter.Handle("/user/{user_id}/total", http.HandlerFunc(c.GetTotalAmountByUserID)).Methods(http.MethodGet)
	cartRouter.Handle("", http.HandlerFunc(c.CreateCart)).Methods(http.MethodPost)
	cartRouter.Handle("/{id}/products/{product_id}", http.HandlerFunc(c.DeleteProductFromCart)).Methods(http.MethodDelete)
	cartRouter.Handle("/{cartID}/products/{productID}/quantity", http.HandlerFunc(c.UpdateCartProductQuantity)).Methods(http.MethodPut)
	c.log.Print("======== Cart Routes Registered =========")
}

func (c *CartController) CreateCart(w http.ResponseWriter, r *http.Request) {
	var newCart cart.Cart
	if err := web.UnmarshalJSON(r, &newCart); err != nil {
		c.log.Print(err)
		web.RespondError(w, errors.NewHTTPError(err.Error(), http.StatusBadRequest))
		return
	}

	userID, httpErr := authmiddleware.GetUserIDFromContext(r.Context())
	if httpErr != nil {
		web.RespondError(w, httpErr)
		return
	}
	newCart.UserID = userID

	if err := newCart.Validate(false); err != nil {
		c.log.Print(err.Error())
		web.RespondError(w, errors.NewHTTPError(err.Error(), http.StatusBadRequest))
		return
	}

	createdCart, err := c.service.CreateCart(&newCart)
	if err != nil {
		c.log.Print(err.Error())
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusCreated, createdCart)
}

func (c *CartController) UpdateCartProductQuantity(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	cartID, err := uuid.Parse(vars["cartID"])
	if err != nil {
		web.RespondError(w, errors.NewHTTPError("invalid cart id", http.StatusBadRequest))
		return
	}

	productID, err := uuid.Parse(vars["productID"])
	if err != nil {
		web.RespondError(w, errors.NewHTTPError("invalid product id", http.StatusBadRequest))
		return
	}

	userID, httpErr := authmiddleware.GetUserIDFromContext(r.Context())
	if httpErr != nil {
		web.RespondError(w, httpErr)
		return
	}

	var req struct {
		Quantity int `json:"quantity"`
	}
	if err := web.UnmarshalJSON(r, &req); err != nil {
		c.log.Print(err)
		web.RespondError(w, errors.NewHTTPError("invalid request body", http.StatusBadRequest))
		return
	}
	if req.Quantity <= 0 {
		web.RespondError(w, errors.NewHTTPError("quantity must be greater than 0", http.StatusBadRequest))
		return
	}
	updatedCart, err := c.service.UpdateCartProductQuantity(cartID, userID, productID, req.Quantity)
	if err != nil {
		c.log.Print(err.Error())
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusOK, updatedCart)
}

func (c *CartController) GetCartByID(w http.ResponseWriter, r *http.Request) {
	parser := web.NewParser(r)
	idStr := parser.GetParameter("id")
	cartID, err := uuid.Parse(idStr)
	if err != nil {
		web.RespondError(w, errors.NewValidationError("invalid cart id"))
		return
	}

	cartObj, err := c.service.GetCartByID(cartID)
	if err != nil {
		c.log.Print(err.Error())
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusOK, cartObj)
}

func (c *CartController) GetCartByUserID(w http.ResponseWriter, r *http.Request) {
	userID, httpErr := authmiddleware.GetUserIDFromContext(r.Context())
	if httpErr != nil {
		web.RespondError(w, httpErr)
		return
	}

	carts, err := c.service.GetCartByUserID(userID)
	if err != nil {
		c.log.Print(err.Error())
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusOK, carts)
}

func (c *CartController) GetTotalAmountByUserID(w http.ResponseWriter, r *http.Request) {
	parser := web.NewParser(r)
	userIDStr := parser.GetParameter("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		web.RespondError(w, errors.NewValidationError("invalid user id"))
		return
	}

	total, err := c.service.GetTotalAmountByCartID(userID)
	if err != nil {
		c.log.Print(err.Error())
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusOK, map[string]float64{"total_amount": total})
}

func (c *CartController) DeleteProductFromCart(w http.ResponseWriter, r *http.Request) {
	parser := web.NewParser(r)
	cartIDStr := parser.GetParameter("id")
	productIDStr := parser.GetParameter("product_id")

	cartID, err := uuid.Parse(cartIDStr)
	if err != nil {
		web.RespondError(w, errors.NewValidationError("invalid cart id"))
		return
	}

	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		web.RespondError(w, errors.NewValidationError("invalid product id"))
		return
	}

	userID, httpErr := authmiddleware.GetUserIDFromContext(r.Context())
	if httpErr != nil {
		web.RespondError(w, httpErr)
		return
	}

	updatedCart, err := c.service.DeleteProductFromCart(cartID, userID, productID)
	if err != nil {
		c.log.Print(err.Error())
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusOK, updatedCart)
}
