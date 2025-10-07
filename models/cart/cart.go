package cart

import (
	"ecommerce/errors"
	"ecommerce/models/baseStruct"
	"ecommerce/models/products"
	"time"

	"github.com/google/uuid"
)

type CartProduct struct {
	CartID    uuid.UUID         `gorm:"type:char(36);primaryKey;not null" json:"cartID"`
	ProductID uuid.UUID         `gorm:"type:char(36);primaryKey;not null" json:"productID"`
	Product   products.Products `gorm:"foreignKey:ProductID;references:ID" json:"products"`
	Quantity  int               `gorm:"type:int;default:1" json:"quantity"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	DeletedAt *time.Time        `gorm:"index" json:"deleted_at,omitempty"`
}

type Cart struct {
	baseStruct.Base
	UserID      uuid.UUID     `gorm:"type:char(36);not null" json:"userID"`
	Products    []CartProduct `gorm:"foreignKey:CartID" json:"products"`
	TotalAmount float64       `gorm:"type:decimal(10,2);default:0" json:"totalAmount"`
}

func (c *Cart) Validate(isUpdate bool) error {
	if c.UserID == uuid.Nil {
		return errors.NewValidationError("Cart must have a valid user ID")
	}
	return nil
}

type DTO struct {
	ID          uuid.UUID     `json:"id"`
	UserID      uuid.UUID     `json:"userID"`
	Products    []CartProduct `json:"products"`
	TotalAmount float64       `json:"totalAmount"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

func (DTO) TableName() string {
	return "carts"
}
