package products

import (
	"ecommerce/errors"
	"ecommerce/util"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Products struct {
	ID          uuid.UUID      `gorm:"type:char(36);primaryKey" json:"id"`
	Name        string         `gorm:"type:varchar(255);not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	Images      datatypes.JSON `gorm:"type:longblob" json:"images"` //`gorm:"type:json" json:"images"`
	Price       float64        `gorm:"type:decimal(10,2);not null" json:"price"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (p *Products) Validate(isUpdate bool) error {
	if util.IsEmpty(p.Name) {
		return errors.NewValidationError("Product name must be specified")
	}
	if util.IsEmpty(p.Description) {
		return errors.NewValidationError("Product description must be specified")
	}
	if p.Price <= 0 {
		return errors.NewValidationError("Product price must be greater than zero")
	}
	return nil
}

type DTO struct {
	ID          uuid.UUID      `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Images      datatypes.JSON `json:"images"`
	Price       float64        `json:"price"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

func (DTO) TableName() string {
	return "products"
}
