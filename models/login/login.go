package login

import (
	"time"

	"github.com/google/uuid"
)

type LoginAttempt struct {
	UserID      uuid.UUID  `json:"user_id"`
	FailedCount int        `json:"failed_count"`
	LastAttempt time.Time  `json:"last_attempt"`
	LockedUntil *time.Time `json:"locked_until,omitempty"`
}
