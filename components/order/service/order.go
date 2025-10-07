package service

import (
	"ecommerce/errors"
	"ecommerce/models/cart"
	"ecommerce/models/order"
	"ecommerce/models/products"
	"ecommerce/repository"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

const CancelationWindow = 30 * time.Minute

type OrderService struct {
	db           *gorm.DB
	repository   repository.EcommerceRepository
	associations []string
}

func NewOrderService(db *gorm.DB, repo repository.EcommerceRepository, associations []string) *OrderService {
	return &OrderService{
		db:           db,
		repository:   repo,
		associations: associations,
	}
}
func (s *OrderService) CreateOrder(newOrder *order.Order, totalAmount float64) error {
	uow := repository.NewUnitOfWork(s.db, false)
	defer uow.RollBack()

	var c cart.Cart
	if err := s.repository.GetRecord(
		uow,
		&c,
		repository.Filter("id = ?", newOrder.CartID),
		repository.NotDeleted(),
		repository.Preload("Products.Product"),
	); err != nil {
		return errors.NewValidationError("Cart not found")
	}

	newOrder.UserID = c.UserID
	newOrder.TotalAmount = totalAmount

	orderProducts := make([]products.Products, 0, len(c.Products))
	for _, cp := range c.Products {
		orderProducts = append(orderProducts, cp.Product)
	}
	newOrder.Products = orderProducts

	if err := s.repository.Add(uow, newOrder); err != nil {
		return err
	}

	for _, item := range c.Products {
		_ = s.repository.UpdateWithMap(
			uow,
			&cart.CartProduct{},
			map[string]interface{}{"deleted_at": time.Now()},
			repository.Filter("cart_id = ? AND product_id = ?", c.ID, item.ProductID),
		)
	}

	uow.Commit()
	return nil
}

func (s *OrderService) DeleteOrder(orderToDelete *order.Order) error {
	existing, err := s.doesOrderExist(orderToDelete.ID)
	if err != nil {
		return err
	}

	if time.Since(existing.CreatedAt) > CancelationWindow {
		return errors.NewValidationError("Order can no longer be cancelled — cancelation window has passed")
	}

	uow := repository.NewUnitOfWork(s.db, false)
	defer uow.RollBack()

	updateMap := map[string]interface{}{
		"status":     "canceled",
		"updated_at": time.Now(),
	}

	if err := s.repository.UpdateWithMap(
		uow,
		existing,
		updateMap,
		repository.Filter("id = ?", orderToDelete.ID),
	); err != nil {
		return err
	}

	uow.Commit()
	return nil
}

func (s *OrderService) GetOrdersByUserID(userID uuid.UUID) ([]order.Order, error) {
	uow := repository.NewUnitOfWork(s.db, true)
	defer uow.RollBack()

	var orders []order.Order
	if err := s.repository.GetAll(
		uow,
		&orders,
		repository.Filter("user_id = ?", userID),
		repository.NotDeleted(),
	); err != nil {
		return nil, err
	}

	uow.Commit()
	return orders, nil
}

func (s *OrderService) GetAllOrders(allOrders *[]order.Order, limit, offset int, totalCount *int) error {
	uow := repository.NewUnitOfWork(s.db, true)
	defer uow.RollBack()

	queryProcessors := []repository.QueryProcessor{
		repository.NotDeleted(),
		repository.Paginate(limit, offset, totalCount),
	}

	if err := s.repository.GetAll(uow, allOrders, queryProcessors...); err != nil {
		return err
	}

	uow.Commit()
	return nil
}

func (s *OrderService) doesOrderExist(ID uuid.UUID) (*order.Order, error) {
	uow := repository.NewUnitOfWork(s.db, true)
	defer uow.RollBack()

	var o order.Order
	if err := s.repository.GetRecord(
		uow,
		&o,
		repository.Filter("id = ?", ID),
		repository.NotDeleted(),
	); err != nil {
		return nil, errors.NewValidationError("Order ID is invalid")
	}

	return &o, nil
}
