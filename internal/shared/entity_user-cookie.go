package shared

import (
	"time"

	"github.com/knackwurstking/pg-press/internal/errors"
)

var (
	// CookieExpirationDuration is the duration after which cookies expire (6 months)
	CookieExpirationDuration int64 = (time.Hour * 24 * 31 * 6).Milliseconds() // 6 months in milliseconds
)

// Cookie represents a user session with authentication information.
type Cookie struct {
	UserAgent string     `json:"user_agent"` // User agent string of the client
	Value     string     `json:"value"`      // Unique UUID cookie value
	UserID    TelegramID `json:"user_id"`    // Associated Telegram ID
	LastLogin UnixMilli  `json:"last_login"` // Last login timestamp in milliseconds
}

func (e *Cookie) Validate() *errors.ValidationError {
	if e.Value == "" {
		return errors.NewValidationError("cookie value is required")
	}
	if e.UserID == 0 {
		return errors.NewValidationError("cookie user_id is required")
	}
	return nil
}

func (e *Cookie) Clone() *Cookie {
	return &Cookie{
		UserAgent: e.UserAgent,
		Value:     e.Value,
		UserID:    e.UserID,
		LastLogin: e.LastLogin,
	}
}

func (e *Cookie) IsExpired() bool {
	return time.Now().UnixMilli()-int64(e.LastLogin) > CookieExpirationDuration
}

func (e *Cookie) ExipredAt() int64 {
	return int64(e.LastLogin) + CookieExpirationDuration
}

func (e *Cookie) ExpiredAtTime() time.Time {
	return time.UnixMilli(e.ExipredAt())
}

func (e *Cookie) String() string {
	return "Cookie{UserID:" + e.UserID.String() + ", Value:" + MaskString(e.Value) + "}"
}
