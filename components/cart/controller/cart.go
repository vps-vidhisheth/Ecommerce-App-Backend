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

func NewCartController(CartService *service.CartService, logger log.Log) *CartController {
	return &CartController{
		log:     logger,
		service: CartService,
	}
}

func (controller *CartController) RegisterRoutes(router *mux.Router) {
	cartRouter := router.PathPrefix("/carts").Subrouter()

	cartRouter.Handle("", authmiddleware.AuthMiddleware(http.HandlerFunc(controller.GetAllCarts))).Methods(http.MethodGet)
	cartRouter.Handle("/{id}", authmiddleware.AuthMiddleware(http.HandlerFunc(controller.GetCartByID))).Methods(http.MethodGet)
	cartRouter.Handle("/user/{user_id}", authmiddleware.AuthMiddleware(http.HandlerFunc(controller.GetCartByUserID))).Methods(http.MethodGet)
	cartRouter.Handle("/user/{user_id}/total", authmiddleware.AuthMiddleware(http.HandlerFunc(controller.GetTotalAmountByUserID))).Methods(http.MethodGet)

	adminRouter := router.PathPrefix("/carts").Subrouter()
	adminRouter.Handle("", authmiddleware.AuthMiddleware(authmiddleware.AdminMiddleware(http.HandlerFunc(controller.CreateCart)))).Methods(http.MethodPost)
	adminRouter.Handle("/{id}", authmiddleware.AuthMiddleware(authmiddleware.AdminMiddleware(http.HandlerFunc(controller.UpdateCart)))).Methods(http.MethodPut)
	adminRouter.Handle("/{id}", authmiddleware.AuthMiddleware(authmiddleware.AdminMiddleware(http.HandlerFunc(controller.DeleteCart)))).Methods(http.MethodDelete)

	controller.log.Print("======== Cart Routes Registered =========")
}

func (c *CartController) CreateCart(w http.ResponseWriter, r *http.Request) {
	var newCart cart.Cart
	if err := web.UnmarshalJSON(r, &newCart); err != nil {
		c.log.Print(err)
		web.RespondError(w, err)
		return
	}

	if err := newCart.Validate(false); err != nil {
		c.log.Print(err.Error())
		web.RespondError(w, errors.NewHTTPError(err.Error(), http.StatusBadRequest))
		return
	}

	if err := c.service.CreateCart(&newCart); err != nil {
		c.log.Print(err.Error())
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusCreated, newCart)
}

func (c *CartController) GetAllCarts(w http.ResponseWriter, r *http.Request) {
	allCarts := []cart.DTO{}
	var totalCount int
	parser := web.NewParser(r)
	limit, offset := parser.ParseLimitAndOffset()

	if err := c.service.GetAllCarts(&allCarts, limit, offset, &totalCount, parser.Form); err != nil {
		c.log.Print(err.Error())
		web.RespondError(w, err)
		return
	}

	web.RespondJSONWithXTotalCount(w, http.StatusOK, totalCount, allCarts)
}

func (c *CartController) GetCartByID(w http.ResponseWriter, r *http.Request) {
	parser := web.NewParser(r)
	idStr := parser.GetParameter("id")

	cartID, err := uuid.Parse(idStr)
	if err != nil {
		web.RespondError(w, errors.NewValidationError("invalid Cart id"))
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

func (c *CartController) UpdateCart(w http.ResponseWriter, r *http.Request) {
	cartToUpdate := cart.Cart{}

	if err := web.UnmarshalJSON(r, &cartToUpdate); err != nil {
		c.log.Print(err.Error())
		web.RespondError(w, errors.NewHTTPError(err.Error(), http.StatusBadRequest))
		return
	}

	if err := cartToUpdate.Validate(true); err != nil {
		c.log.Print(err.Error())
		web.RespondError(w, errors.NewHTTPError(err.Error(), http.StatusBadRequest))
		return
	}

	parser := web.NewParser(r)
	idStr := parser.GetParameter("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		web.RespondError(w, errors.NewValidationError("invalid cart id"))
		return
	}
	cartToUpdate.ID = id

	if err := c.service.UpdateCart(&cartToUpdate); err != nil {
		c.log.Print(err.Error())
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusOK, cartToUpdate)
}

func (c *CartController) DeleteCart(w http.ResponseWriter, r *http.Request) {
	cartToDelete := cart.Cart{}
	parser := web.NewParser(r)
	idStr := parser.GetParameter("id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		web.RespondError(w, errors.NewValidationError("invalid cart id"))
		return
	}

	cartToDelete.ID = id

	if err := c.service.DeleteCart(&cartToDelete); err != nil {
		c.log.Print(err.Error())
		web.RespondError(w, err)
		return
	}

	web.RespondJSON(w, http.StatusOK, "Cart Deleted Successfully")
}

func (c *CartController) GetCartByUserID(w http.ResponseWriter, r *http.Request) {
	parser := web.NewParser(r)
	userIDStr := parser.GetParameter("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		web.RespondError(w, errors.NewValidationError("invalid User id"))
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
		web.RespondError(w, errors.NewValidationError("invalid User id"))
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
