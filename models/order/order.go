package order

import (
	"ecommerce/errors"
	"ecommerce/models/baseStruct"
	"ecommerce/models/products"
	"time"

	"github.com/google/uuid"
)

type Order struct {
	baseStruct.Base
	UserID      uuid.UUID           `gorm:"type:char(36);not null" json:"userID"`
	CartID      uuid.UUID           `gorm:"type:char(36);not null" json:"cartID"`
	Status      string              `gorm:"type:varchar(20);not null;default:'confirmed'" json:"status"`
	TotalAmount float64             `gorm:"type:decimal(10,2);default:0" json:"totalAmount"`
	Products    []products.Products `gorm:"many2many:orderProducts;" json:"products"`
}

func (o *Order) Validate(isUpdate bool) error {
	if o.UserID == uuid.Nil {
		return errors.NewValidationError("User ID must be specified")
	}
	if o.CartID == uuid.Nil {
		return errors.NewValidationError("Cart ID must be specified")
	}
	if o.Status == "" {
		o.Status = "confirmed"
	}
	return nil
}

type DTO struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"userID"`
	CartID      uuid.UUID `json:"cartID"`
	Status      string    `json:"status"`
	TotalAmount float64   `json:"totalAmount"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (DTO) TableName() string {
	return "orders"
}
