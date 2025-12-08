package shared

import "github.com/knackwurstking/pg-press/errors"

// Cookie represents a user session with authentication information.
type Cookie struct {
	UserAgent string     `json:"user_agent"` // User agent string of the client
	Value     string     `json:"value"`      // Unique UUID cookie value
	UserID    TelegramID `json:"user_id"`    // Associated Telegram ID
	LastLogin UnixMilly  `json:"last_login"` // Last login timestamp in milliseconds
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

var _ Entity[*Cookie] = (*Cookie)(nil)
