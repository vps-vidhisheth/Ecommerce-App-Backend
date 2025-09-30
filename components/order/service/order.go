package service

import (
	"ecommerce/errors"
	"ecommerce/models/cart"
	"ecommerce/models/order"
	"ecommerce/repository"
	"ecommerce/util"
	"net/url"
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

func (s *OrderService) CreateOrder(newOrder *order.Order) error {
	var c cart.Cart
	if err := s.db.Where("id = ? AND user_id = ? AND deleted_at IS NULL", newOrder.CartID, newOrder.UserID).First(&c).Error; err != nil {
		return errors.NewValidationError("Invalid cart for this user")
	}

	newOrder.TotalAmount = c.TotalAmount
	newOrder.CreatedAt = time.Now()
	newOrder.UpdatedAt = time.Now()

	uow := repository.NewUnitOfWork(s.db, false)
	defer uow.RollBack()

	newOrder.ID = uuid.New()

	if err := s.repository.Add(uow, newOrder); err != nil {
		return err
	}

	uow.Commit()
	return nil
}

func (s *OrderService) UpdateOrder(orderToUpdate *order.Order) error {
	existing, err := s.doesOrderExist(orderToUpdate.ID)
	if err != nil {
		return err
	}

	uow := repository.NewUnitOfWork(s.db, false)
	defer uow.RollBack()

	orderToUpdate.CreatedAt = existing.CreatedAt
	orderToUpdate.UpdatedAt = time.Now()

	var c cart.Cart
	if err := s.db.Where("id = ? AND user_id = ? AND deleted_at IS NULL", orderToUpdate.CartID, orderToUpdate.UserID).First(&c).Error; err != nil {
		return errors.NewValidationError("Invalid cart for this user")
	}

	orderToUpdate.TotalAmount = c.TotalAmount

	if err := s.repository.Update(uow, orderToUpdate); err != nil {
		return err
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

	now := time.Now()
	updateMap := map[string]interface{}{"DeletedAt": now}
	if err := s.repository.UpdateWithMap(uow, existing, updateMap, repository.Filter("id = ?", orderToDelete.ID)); err != nil {
		return err
	}

	uow.Commit()
	return nil
}

func (s *OrderService) GetOrderByID(id uuid.UUID) (*order.Order, error) {
	uow := repository.NewUnitOfWork(s.db, true)
	defer uow.RollBack()

	var o order.Order
	if err := s.repository.GetRecord(uow, &o, repository.Filter("id = ?", id)); err != nil {
		return nil, err
	}

	uow.Commit()
	return &o, nil
}

func (s *OrderService) GetAllOrders(allOrders *[]order.Order, limit, offset int, totalCount *int, requestForm url.Values) error {
	uow := repository.NewUnitOfWork(s.db, true)
	defer uow.RollBack()

	var queryProcessors []repository.QueryProcessor

	queryProcessors = append(queryProcessors, func(db *gorm.DB, out interface{}) (*gorm.DB, error) {
		return db.Where("deleted_at IS NULL"), nil
	})
	queryProcessors = append(queryProcessors, repository.Paginate(limit, offset, totalCount))

	if err := s.repository.GetAll(uow, allOrders, queryProcessors...); err != nil {
		return err
	}

	uow.Commit()
	return nil
}

func (s *OrderService) doesOrderExist(ID uuid.UUID) (*order.Order, error) {
	var o order.Order
	uow := repository.NewUnitOfWork(s.db, true)
	defer uow.RollBack()

	err := s.repository.GetRecord(uow, &o, repository.Filter("id = ?", ID))
	if err != nil {
		return nil, errors.NewValidationError("Order ID is invalid")
	}

	return &o, nil
}

func (s *OrderService) GetOrdersByUserID(userID uuid.UUID) ([]order.Order, error) {
	var orders []order.Order
	if err := s.db.Where("user_id = ? AND deleted_at IS NULL", userID).Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func (s *OrderService) buildSearchQuery(requestForm url.Values) repository.QueryProcessor {
	if len(requestForm) == 0 {
		return nil
	}

	var columnNames []string
	var conditions []string
	var operators []string
	var values []interface{}

	if searchTerm := requestForm.Get("search"); searchTerm != "" {
		columns := []string{"`id`", "`cart_id`", "`user_id`"}
		for _, col := range columns {
			util.AddToSlice(col, "LIKE ?", "OR", "%"+searchTerm+"%", &columnNames, &conditions, &operators, &values)
		}
	}

	return repository.FilterWithOperator(columnNames, conditions, operators, values)
}

func (s *OrderService) GetTotalAmountByUserID(userID uuid.UUID) (float64, error) {
	var orders []order.Order
	if err := s.db.Where("user_id = ? AND deleted_at IS NULL", userID).Find(&orders).Error; err != nil {
		return 0, err
	}

	var total float64
	for _, o := range orders {
		total += o.TotalAmount
	}

	return total, nil
}
