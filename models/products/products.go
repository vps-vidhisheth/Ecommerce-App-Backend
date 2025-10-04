package products

import (
	"ecommerce/errors"
	"ecommerce/models/baseStruct"
	"ecommerce/util"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

type Products struct {
	baseStruct.Base
	Name        string         `gorm:"type:varchar(255);not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	Price       float64        `gorm:"type:decimal(10,2);not null" json:"price"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	Images      []ProductImage `gorm:"foreignKey:ProductID" json:"images"`
}

type ProductImage struct {
	baseStruct.Base
	ProductID uuid.UUID `gorm:"type:char(36);index" json:"product_id"`
	Image     []byte    `gorm:"type:longblob" json:"image"`
}

func (p *ProductImage) BeforeCreate(scope *gorm.Scope) error {
	if p.ID == uuid.Nil {
		return scope.SetColumn("ID", uuid.New())
	}
	return nil
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
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Images      [][]byte  `json:"images"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (DTO) TableName() string {
	return "products"
}
