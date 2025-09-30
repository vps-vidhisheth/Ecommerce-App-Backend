package service

import (
	"ecommerce/errors"
	"ecommerce/models/cart"
	"ecommerce/models/products"
	"ecommerce/repository"
	"ecommerce/util"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

type CartService struct {
	db           *gorm.DB
	repository   repository.EcommerceRepository
	associations []string
}

func NewCartService(db *gorm.DB, repo repository.EcommerceRepository, associations []string) *CartService {
	return &CartService{
		db:           db,
		repository:   repo,
		associations: associations,
	}
}

func (s *CartService) calculateTotalAmount(cartItem *cart.Cart) error {
	var prod products.Products
	if err := s.db.First(&prod, "id = ?", cartItem.ProductID).Error; err != nil {
		return errors.NewValidationError("Invalid product ID")
	}
	cartItem.TotalAmount = float64(cartItem.Quantity) * prod.Price
	return nil
}

func (s *CartService) CreateCart(newCart *cart.Cart) error {
	if err := s.calculateTotalAmount(newCart); err != nil {
		return err
	}

	uow := repository.NewUnitOfWork(s.db, false)
	defer uow.RollBack()

	newCart.ID = uuid.New()
	newCart.CreatedAt = time.Now()
	newCart.UpdatedAt = time.Now()

	if err := s.repository.Add(uow, newCart); err != nil {
		return err
	}

	uow.Commit()
	return nil
}

func (s *CartService) UpdateCart(cartToUpdate *cart.Cart) error {
	existing, err := s.doesCartExist(cartToUpdate.ID)
	if err != nil {
		return err
	}

	uow := repository.NewUnitOfWork(s.db, false)
	defer uow.RollBack()

	cartToUpdate.CreatedAt = existing.CreatedAt
	cartToUpdate.UpdatedAt = time.Now()

	if err := s.calculateTotalAmount(cartToUpdate); err != nil {
		return err
	}

	if err := s.repository.Update(uow, cartToUpdate); err != nil {
		return err
	}

	uow.Commit()
	return nil
}

func (s *CartService) doesCartExist(ID uuid.UUID) (*cart.Cart, error) {
	uow := repository.NewUnitOfWork(s.db, true)
	defer uow.RollBack()

	var c cart.Cart
	if err := s.repository.GetRecord(uow, &c, repository.Filter("id = ?", ID)); err != nil {
		return nil, errors.NewValidationError("Cart ID is invalid")
	}

	return &c, nil
}

func (s *CartService) DeleteCart(cartToDelete *cart.Cart) error {
	existing, err := s.doesCartExist(cartToDelete.ID)
	if err != nil {
		return err
	}

	uow := repository.NewUnitOfWork(s.db, false)
	defer uow.RollBack()

	now := time.Now()
	updateMap := map[string]interface{}{"DeletedAt": now}
	if err := s.repository.UpdateWithMap(uow, existing, updateMap, repository.Filter("id=?", cartToDelete.ID)); err != nil {
		return err
	}

	uow.Commit()
	return nil
}

func (s *CartService) GetCartByID(id uuid.UUID) (*cart.Cart, error) {
	uow := repository.NewUnitOfWork(s.db, true)
	defer uow.RollBack()

	var c cart.Cart
	if err := s.repository.GetRecord(uow, &c, repository.Filter("id = ?", id)); err != nil {
		return nil, err
	}

	uow.Commit()
	return &c, nil
}

func (s *CartService) GetAllCarts(allCarts *[]cart.DTO, limit, offset int, totalCount *int, requestForm url.Values) error {
	uow := repository.NewUnitOfWork(s.db, true)
	defer uow.RollBack()

	var queryProcessors []repository.QueryProcessor

	// searchQuery := s.buildSearchQuery(requestForm)
	// if searchQuery != nil {
	// 	queryProcessors = append(queryProcessors, searchQuery)
	// }

	queryProcessors = append(queryProcessors, func(db *gorm.DB, out interface{}) (*gorm.DB, error) {
		return db.Where("deleted_at IS NULL"), nil
	})
	queryProcessors = append(queryProcessors, repository.Paginate(limit, offset, totalCount))

	if err := s.repository.GetAll(uow, allCarts, queryProcessors...); err != nil {
		return err
	}

	uow.Commit()
	return nil
}

func (s *CartService) GetCartByUserID(userID uuid.UUID) ([]cart.Cart, error) {
	var carts []cart.Cart
	if err := s.db.Where("user_id = ? AND deleted_at IS NULL", userID).Find(&carts).Error; err != nil {
		return nil, err
	}

	for i := range carts {
		if err := s.calculateTotalAmount(&carts[i]); err != nil {
			return nil, err
		}
	}

	return carts, nil
}

func (s *CartService) GetTotalAmountByUserID(userID uuid.UUID) (float64, error) {
	carts, err := s.GetCartByUserID(userID)
	if err != nil {
		return 0, err
	}

	var total float64
	for _, c := range carts {
		total += c.TotalAmount
	}

	return total, nil
}

func (s *CartService) buildSearchQuery(requestForm url.Values) repository.QueryProcessor {
	if len(requestForm) == 0 {
		return nil
	}

	var columnNames []string
	var conditions []string
	var operators []string
	var values []interface{}

	if searchTerm := requestForm.Get("search"); searchTerm != "" {
		columns := []string{"`id`", "`user_id`", "`product_id`"}
		for _, col := range columns {
			util.AddToSlice(col, "LIKE ?", "OR", "%"+searchTerm+"%", &columnNames, &conditions, &operators, &values)
		}
	}

	return repository.FilterWithOperator(columnNames, conditions, operators, values)
}
