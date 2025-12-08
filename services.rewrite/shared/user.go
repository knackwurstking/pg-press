package shared

import (
	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/utils"
)

const (
	UserNameMinLength = 1
	UserNameMaxLength = 100
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
	if !utils.ValidateAPIKey(e.ApiKey) {
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

var _ Entity[*User] = (*User)(nil)
