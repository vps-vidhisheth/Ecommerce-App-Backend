package cart

import (
	"time"

	"ecommerce/errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Cart struct {
	ID          uuid.UUID      `gorm:"type:char(36);primaryKey" json:"id"`
	UserID      uuid.UUID      `gorm:"type:char(36);not null" json:"user_id"`
	ProductID   uuid.UUID      `gorm:"type:char(36);not null" json:"product_id"`
	Quantity    int            `gorm:"not null;default:1" json:"quantity"`
	TotalAmount float64        `gorm:"type:decimal(10,2);default:0" json:"total_amount"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (c *Cart) Validate(isUpdate bool) error {
	if c.UserID == uuid.Nil {
		return errors.NewValidationError("Cart must have a valid user ID")
	}
	if c.ProductID == uuid.Nil {
		return errors.NewValidationError("Cart must have a valid product ID")
	}
	if c.Quantity <= 0 {
		c.Quantity = 1
	}
	return nil
}

type DTO struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	ProductID   uuid.UUID `json:"product_id"`
	Quantity    int       `json:"quantity"`
	TotalAmount float64   `json:"total_amount"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (DTO) TableName() string {
	return "carts"
}
