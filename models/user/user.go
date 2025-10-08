package user

import (
	"ecommerce/errors"
	"ecommerce/util"
	"time"

	"ecommerce/models/baseStruct"

	"github.com/google/uuid"
)

type User struct {
	baseStruct.Base
	Name                string     `gorm:"type:varchar(100);not null" json:"name"`
	Email               string     `gorm:"type:varchar(100);not null" json:"email"`
	Password            string     `gorm:"type:varchar(255);not null" json:"password"`
	Role                string     `gorm:"type:varchar(50);not null" json:"role"`
	ProfilePic          []byte     `gorm:"type:LONGBLOB" json:"profilePic,omitempty"`
	IsActive            bool       `gorm:"default:true" json:"isActive"`
	LastDeactivatedAt   *time.Time `json:"lastDeactivatedAt,omitempty"`
	ResetToken          string     `gorm:"type:varchar(255);index" json:"resetToken,omitempty"`
	ResetTokenExpiresAt *time.Time `json:"resetTokenExpiresAt,omitempty"`
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
	ProfilePic []byte    `json:"profilePic"`
	IsActive   bool      `json:"isActive"`
	CreatedAt  time.Time `json:"createdAT"`
	UpdatedAt  time.Time `json:"updatedAT"`
}

func (DTO) TableName() string {
	return "users"
}
