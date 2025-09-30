package order

import (
	"ecommerce/errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Order struct {
	ID          uuid.UUID      `gorm:"type:char(36);primaryKey" json:"id"`
	UserID      uuid.UUID      `gorm:"type:char(36);not null" json:"user_id"`
	CartID      uuid.UUID      `gorm:"type:char(36);not null" json:"cart_id"`
	TotalAmount float64        `gorm:"type:decimal(10,2);default:0" json:"total_amount"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (o *Order) Validate(isUpdate bool) error {
	if o.UserID == uuid.Nil {
		return errors.NewValidationError("User ID must be specified")
	}
	if o.CartID == uuid.Nil {
		return errors.NewValidationError("Cart ID must be specified")
	}
	if o.TotalAmount < 0 {
		return errors.NewValidationError("Total amount cannot be negative")
	}
	return nil
}

type DTO struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	CartID      uuid.UUID `json:"cart_id"`
	TotalAmount float64   `json:"total_amount"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (DTO) TableName() string {
	return "orders"
}
