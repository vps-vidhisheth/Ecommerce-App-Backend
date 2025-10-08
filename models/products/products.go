package products

import (
	"ecommerce/errors"
	"ecommerce/models/baseStruct"
	"ecommerce/util"
	"time"

	"github.com/google/uuid"
)

type Products struct {
	baseStruct.Base
	Name        string         `gorm:"type:varchar(255);not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	Price       float64        `gorm:"type:decimal(10,2);not null" json:"price"`
	IsActive    bool           `gorm:"default:true" json:"isActive"`
	Images      []ProductImage `gorm:"foreignKey:ProductID" json:"images"`
}

type ProductImage struct {
	baseStruct.Base
	ProductID uuid.UUID `gorm:"type:char(36);index" json:"productID"`
	Image     []byte    `gorm:"type:longblob" json:"image"`
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

func ToDTO(p *Products) DTO {
	images := [][]byte{}
	imageIDs := []uuid.UUID{}

	for _, img := range p.Images {
		images = append(images, img.Image)
		imageIDs = append(imageIDs, img.ID)
	}

	return DTO{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		Images:      images,
		ImageIDs:    imageIDs,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

type DTO struct {
	ID          uuid.UUID   `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Price       float64     `json:"price"`
	Images      [][]byte    `json:"images"`
	ImageIDs    []uuid.UUID `json:"imageIds"`
	CreatedAt   time.Time   `json:"createdAT"`
	UpdatedAt   time.Time   `json:"updatedAT"`
}

func (DTO) TableName() string {
	return "products"
}
