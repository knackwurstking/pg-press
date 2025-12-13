package shared

import (
	"github.com/knackwurstking/pg-press/internal/errors"
)

const (
	UserNameMinLength = 1
	UserNameMaxLength = 100
	MinAPIKeyLength   = 32
)

// User represents a user entity with relevant information.
type User struct {
	ID       TelegramID `json:"id"`        // Unique Telegram ID for the user
	Name     string     `json:"name"`      // User's display name
	ApiKey   string     `json:"api_key"`   // Unique API key for the user
	LastFeed EntityID   `json:"last_feed"` // ID of the last feed accessed by the user
}

func (e *User) Validate() *errors.ValidationError {
	if e.Name == "" {
		return errors.NewValidationError("name cannot be empty")
	}
	if len(e.Name) < UserNameMinLength || len(e.Name) > UserNameMaxLength {
		return errors.NewValidationError(
			"name length must be between %d and %d characters",
			UserNameMinLength, UserNameMaxLength,
		)
	}
	if !ValidateAPIKey(e.ApiKey) {
		return errors.NewValidationError("api_key is not valid")
	}
	return nil
}

func (e *User) Clone() *User {
	return &User{
		ID:       e.ID,
		Name:     e.Name,
		ApiKey:   e.ApiKey,
		LastFeed: e.LastFeed,
	}
}

func (e *User) String() string {
	return "User[ID=" + e.ID.String() + ", Name=" + e.Name + ", ApiKey=" + MaskString(e.ApiKey) + "]"
}

var _ Entity[*User] = (*User)(nil)

// ValidateAPIKey validates an API key according to the minimum length requirement
func ValidateAPIKey(apiKey string) bool {
	return len(apiKey) >= MinAPIKeyLength
}
