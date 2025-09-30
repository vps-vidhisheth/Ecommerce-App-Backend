package user

import (
	"ecommerce/errors"
	"ecommerce/util"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID                  uuid.UUID      `gorm:"type:char(36);primaryKey" json:"id"`
	Name                string         `gorm:"type:varchar(100);not null" json:"name"`
	Email               string         `gorm:"type:varchar(100);unique;not null" json:"email"`
	Password            string         `gorm:"type:varchar(255);not null" json:"password"`
	Role                string         `gorm:"type:varchar(50);not null" json:"role"`
	ProfilePic          []byte         `gorm:"type:blob" json:"profile_pic,omitempty"`
	IsActive            bool           `gorm:"default:true" json:"is_active"`
	ResetToken          string         `gorm:"type:varchar(255);index" json:"reset_token,omitempty"`
	ResetTokenExpiresAt *time.Time     `json:"reset_token_expires_at,omitempty"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (u *User) Validate(isUpdate bool) error {
	if util.IsEmpty(u.Name) {
		return errors.NewValidationError("Name must be specified")
	}
	if util.IsEmpty(u.Email) {
		return errors.NewValidationError("Email must be specified")
	}

	if !isUpdate {
		if len(u.Password) < 6 || len(u.Password) > 50 {
			return errors.NewValidationError("Password length should be between 6 and 50 characters")
		}
	} else {
		if u.Password != "" && (len(u.Password) < 6 || len(u.Password) > 50) {
			return errors.NewValidationError("Password length should be between 6 and 50 characters")
		}
	}

	if util.IsEmpty(u.Role) {
		return errors.NewValidationError("Role must be specified (ADMIN or CUSTOMER)")
	}

	return nil
}

type DTO struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	Email      string    `json:"email"`
	Role       string    `json:"role"`
	ProfilePic []byte    `json:"profile_pic"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (DTO) TableName() string {
	return "users"
}
