package baseStruct

import (
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

type Base struct {
	ID        uuid.UUID  `gorm:"type:char(36);primary_key" json:"id"`
	CreatedAt time.Time  `json:"createdAT"`
	UpdatedAt time.Time  `json:"updatedAT"`
	DeletedAt *time.Time `gorm:"index" json:"deletedAT,omitempty"`
}

func (b *Base) BeforeCreate(scope *gorm.Scope) error {
	if b.ID == uuid.Nil {
		if err := scope.SetColumn("ID", uuid.New()); err != nil {
			return err
		}
	}
	return nil
}
