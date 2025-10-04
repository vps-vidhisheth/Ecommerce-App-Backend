package cart

import (
	"time"

	"ecommerce/errors"
	"ecommerce/models/baseStruct"

	"github.com/google/uuid"
)

type CartProduct struct {
	CartID    uuid.UUID  `gorm:"type:char(36);primaryKey;not null" json:"cart_id"`
	ProductID uuid.UUID  `gorm:"type:char(36);primaryKey;not null" json:"product_id"`
	Quantity  int        `gorm:"type:int;default:1" json:"quantity"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}

type Cart struct {
	baseStruct.Base
	UserID      uuid.UUID     `gorm:"type:char(36);not null" json:"user_id"`
	Products    []CartProduct `gorm:"foreignKey:CartID" json:"products"` // Each product has quantity
	TotalAmount float64       `gorm:"type:decimal(10,2);default:0" json:"total_amount"`
}

func (c *Cart) Validate(isUpdate bool) error {
	if c.UserID == uuid.Nil {
		return errors.NewValidationError("Cart must have a valid user ID")
	}
	return nil
}

type DTO struct {
	ID          uuid.UUID     `json:"id"`
	UserID      uuid.UUID     `json:"user_id"`
	Products    []CartProduct `json:"products"`
	TotalAmount float64       `json:"total_amount"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

func (DTO) TableName() string {
	return "carts"
}
