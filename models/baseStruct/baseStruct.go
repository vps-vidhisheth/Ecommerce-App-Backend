package baseStruct

import (
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

type Base struct {
	ID        uuid.UUID  `gorm:"type:char(36);primary_key" json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}

func (b *Base) BeforeCreate(scope *gorm.Scope) error {
	if b.ID == uuid.Nil {
		if err := scope.SetColumn("ID", uuid.New()); err != nil {
			return err
		}
	}
	return nil
}
