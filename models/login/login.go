package login

import (
	"time"

	"github.com/google/uuid"
)

type LoginAttempt struct {
	UserID      uuid.UUID  `json:"userID"`
	FailedCount int        `json:"failedCount"`
	LastAttempt time.Time  `json:"lastAttempt"`
	LockedUntil *time.Time `json:"lockedUntil,omitempty"`
}
