package payment

import (
	"ecommerce/errors"
	"ecommerce/models/baseStruct"

	"github.com/google/uuid"
)

type Payment struct {
	baseStruct.Base
	UserID  uuid.UUID `gorm:"type:char(36);not null" json:"userID"`
	CartID  uuid.UUID `gorm:"type:char(36);not null" json:"cartID"`
	Amount  float64   `gorm:"type:decimal(10,2);not null" json:"amount"`
	OrderID uuid.UUID `gorm:"type:char(36);not null" json:"orderID"`
}

func (p *Payment) Validate(isUpdate bool) error {
	if p.UserID == uuid.Nil {
		return errors.NewValidationError("User ID must be specified")
	}
	if p.CartID == uuid.Nil {
		return errors.NewValidationError("Cart ID must be specified")
	}
	if p.Amount <= 0 {
		return errors.NewValidationError("Amount must be greater than zero")
	}
	return nil
}
