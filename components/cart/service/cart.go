package service

import (
	"time"

	"ecommerce/errors"
	"ecommerce/models/cart"
	"ecommerce/models/products"
	"ecommerce/repository"

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

func (s *CartService) calculateTotalAmount(c *cart.Cart) error {
	uow := repository.NewUnitOfWork(s.db, true)
	defer uow.RollBack()

	total := 0.0
	for i := range c.Products {
		cp := &c.Products[i]

		var product products.Products
		err := s.repository.GetRecord(uow, &product,
			repository.Filter("id = ?", cp.ProductID),
			repository.NotDeleted(),
		)
		if err != nil {
			return err
		}

		if cp.Quantity <= 0 {
			cp.Quantity = 1
		}

		total += product.Price * float64(cp.Quantity)
	}

	c.TotalAmount = total
	uow.Commit()
	return nil
}

func (s *CartService) CreateCart(newCart *cart.Cart) (*cart.Cart, error) {
	uow := repository.NewUnitOfWork(s.db, false)
	defer uow.RollBack()

	for i := range newCart.Products {
		var product products.Products
		err := s.repository.GetRecord(uow, &product,
			repository.Filter("id = ? AND deleted_at IS NULL AND is_active = ?", newCart.Products[i].ProductID, true))
		if err != nil {
			return nil, errors.NewValidationError("invalid or inactive product")
		}
	}

	var existingCarts []cart.Cart
	err := s.repository.GetAll(uow, &existingCarts,
		repository.Filter("user_id = ?", newCart.UserID),
		repository.NotDeleted(),
	)
	if err != nil {
		return nil, err
	}

	if len(existingCarts) > 0 {
		existing := existingCarts[0]

		for i := range newCart.Products {
			newCart.Products[i].CartID = existing.ID
			if err := s.repository.Add(uow, &newCart.Products[i]); err != nil {
				return nil, err
			}
		}
		if err := s.calculateTotalAmount(&existing); err != nil {
			return nil, err
		}

		if err := s.repository.UpdateWithMap(uow, &existing, map[string]interface{}{
			"total_amount": existing.TotalAmount,
			"updated_at":   time.Now(),
		}); err != nil {
			return nil, err
		}

		uow.Commit()
		return &existing, nil
	}

	if err := s.calculateTotalAmount(newCart); err != nil {
		return nil, err
	}

	if err := s.repository.Add(uow, newCart); err != nil {
		return nil, err
	}

	uow.Commit()
	return newCart, nil
}

func (s *CartService) UpdateCartProductQuantity(cartID, userID uuid.UUID, productID uuid.UUID, newQty int) (*cart.Cart, error) {
	uow := repository.NewUnitOfWork(s.db, false)
	defer uow.RollBack()

	var existingCart cart.Cart
	err := s.repository.GetRecord(uow, &existingCart,
		repository.Filter("id = ? AND user_id = ?", cartID, userID),
		repository.Preload("Products"),
		repository.NotDeleted(),
	)
	if err != nil {
		return nil, errors.NewValidationError("cart not found")
	}

	var cpToUpdate *cart.CartProduct
	for i := range existingCart.Products {
		if existingCart.Products[i].ProductID == productID {
			cpToUpdate = &existingCart.Products[i]
			break
		}
	}
	if cpToUpdate == nil {
		return nil, errors.NewValidationError("product not found in cart")
	}

	cpToUpdate.Quantity = newQty
	cpToUpdate.UpdatedAt = time.Now()

	err = s.repository.UpdateWithMap(uow, &cart.CartProduct{}, map[string]interface{}{
		"quantity":   newQty,
		"updated_at": cpToUpdate.UpdatedAt,
	},
		repository.Filter("cart_id = ? AND product_id = ?", cartID, productID),
	)
	if err != nil {
		return nil, err
	}

	if err := s.calculateTotalAmount(&existingCart); err != nil {
		return nil, err
	}
	existingCart.UpdatedAt = time.Now()
	err = s.repository.UpdateWithMap(uow, &cart.Cart{}, map[string]interface{}{
		"total_amount": existingCart.TotalAmount,
		"updated_at":   existingCart.UpdatedAt,
	},
		repository.Filter("id = ?", existingCart.ID),
	)
	if err != nil {
		return nil, err
	}

	uow.Commit()
	return &existingCart, nil
}

func (s *CartService) GetCartByUserID(userID uuid.UUID) ([]cart.Cart, error) {
	uow := repository.NewUnitOfWork(s.db, true)
	defer uow.RollBack()

	var carts []cart.Cart
	err := s.repository.GetAll(uow, &carts,
		repository.Filter("user_id = ?", userID), repository.NotDeleted(), repository.Preload("Products"))
	if err != nil {
		return nil, err
	}

	for ci := range carts {
		_ = s.calculateTotalAmount(&carts[ci])
	}

	uow.Commit()
	return carts, nil
}

func (s *CartService) DeleteProductFromCart(cartID, userID, productID uuid.UUID) (*cart.Cart, error) {
	uow := repository.NewUnitOfWork(s.db, false)
	defer uow.RollBack()

	var c cart.Cart
	err := s.repository.GetRecord(uow, &c,
		repository.Filter("id = ? AND user_id = ?", cartID, userID),
		repository.Preload("Products"),
		repository.NotDeleted(),
	)
	if err != nil {
		return nil, errors.NewValidationError("cart not found or not owned by user")
	}

	productExists := false
	remainingProducts := []cart.CartProduct{}
	for _, p := range c.Products {
		if p.ProductID == productID {
			productExists = true
			continue
		}
		remainingProducts = append(remainingProducts, p)
	}
	if !productExists {
		return nil, errors.NewValidationError("product not found in cart")
	}

	err = s.repository.UpdateWithMap(uow, &cart.CartProduct{}, map[string]interface{}{
		"deleted_at": time.Now(),
	},
		repository.Filter("cart_id = ? AND product_id = ?", cartID, productID),
	)
	if err != nil {
		return nil, err
	}

	c.Products = remainingProducts

	if err := s.calculateTotalAmount(&c); err != nil {
		return nil, err
	}

	c.UpdatedAt = time.Now()
	err = s.repository.UpdateWithMap(uow, &cart.Cart{}, map[string]interface{}{
		"total_amount": c.TotalAmount,
		"updated_at":   c.UpdatedAt,
	},
		repository.Filter("id = ?", c.ID),
	)
	if err != nil {
		return nil, err
	}

	uow.Commit()
	return &c, nil
}

func (s *CartService) GetTotalAmountByCartID(cartID uuid.UUID) (float64, error) {
	cartObj, err := s.GetCartByID(cartID)
	if err != nil {
		return 0, err
	}
	return cartObj.TotalAmount, nil
}

func (s *CartService) GetCartByID(id uuid.UUID) (*cart.Cart, error) {
	uow := repository.NewUnitOfWork(s.db, true)
	defer uow.RollBack()

	var c cart.Cart
	err := s.repository.GetRecord(uow, &c, repository.Filter("id = ?", id), repository.NotDeleted(), repository.Preload("Products"))
	if err != nil {
		return nil, err
	}

	_ = s.calculateTotalAmount(&c)
	uow.Commit()
	return &c, nil
}
