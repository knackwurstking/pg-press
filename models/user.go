// TODO: Remove useless stuff
package models

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/knackwurstking/pgpress/env"
	"github.com/knackwurstking/pgpress/errors"
)

// TODO: Maybe move to env/constants.go
const (
	MinNameLength = 1
	MaxNameLength = 100
)

type TelegramID int64

// User represents a system user with Telegram integration
type User struct {
	TelegramID TelegramID `json:"telegram_id"`
	Name       string     `json:"user_name"`
	ApiKey     string     `json:"api_key"`
	LastFeed   FeedID     `json:"last_feed"`
}

// NewUser creates a new user with the provided details
func NewUser(telegramID TelegramID, userName, apiKey string) *User {
	return &User{
		TelegramID: telegramID,
		Name:       strings.TrimSpace(userName),
		ApiKey:     strings.TrimSpace(apiKey),
		LastFeed:   0,
	}
}

func NewUserFromInterfaceMap(modified map[string]any) *User {
	user := &User{
		TelegramID: TelegramID(modified["telegram_id"].(float64)),
		Name:       modified["user_name"].(string),
		ApiKey:     modified["api_key"].(string),
		LastFeed:   FeedID(modified["last_feed"].(float64)),
	}

	return user
}

// Validate checks if the user has valid data
func (u *User) Validate() error {
	if u.TelegramID <= 0 {
		return errors.NewValidationError("telegram_id: must be positive")
	}

	if u.Name == "" {
		return errors.NewValidationError("user_name: cannot be empty")
	}
	if len(u.Name) < MinNameLength {
		return errors.NewValidationError("user_name: too short")
	}
	if len(u.Name) > MaxNameLength {
		return errors.NewValidationError("user_name: too long")
	}

	if u.ApiKey == "" {
		return errors.NewValidationError("api_key: cannot be empty")
	}
	if len(u.ApiKey) < env.MinAPIKeyLength {
		return errors.NewValidationError(
			fmt.Sprintf(
				"api_key: too short for security, must be at least %d characters",
				env.MinAPIKeyLength,
			),
		)
	}

	if u.LastFeed < 0 {
		return errors.NewValidationError("last_feed: cannot be negative")
	}

	return nil
}

// IsAdmin checks if the user is an administrator
func (u *User) IsAdmin() bool {
	adminsEnv := os.Getenv("ADMINS")
	if adminsEnv == "" {
		return false
	}

	adminIDs := strings.Split(adminsEnv, ",")
	userIDStr := fmt.Sprintf("%d", u.TelegramID)

	return slices.Contains(adminIDs, userIDStr)
}

// IsValidAPIKey checks if the provided API key matches the user's API key
func (u *User) IsValidAPIKey(apiKey string) bool {
	return u.ApiKey == apiKey
}

// GetDisplayInfo returns safe user information for display (without API key)
func (u *User) GetDisplayInfo() map[string]any {
	return map[string]any{
		"telegram_id": u.TelegramID,
		"name":        u.Name,
		"has_api_key": true,
		"is_admin":    u.IsAdmin(),
		"last_feed":   u.LastFeed,
	}
}

// String returns a string representation of the user (without sensitive data)
func (u *User) String() string {
	adminStatus := ""
	if u.IsAdmin() {
		adminStatus = " (admin)"
	}

	return fmt.Sprintf("User{ID: %d, Name: %s%s [has API key]}",
		u.TelegramID, u.Name, adminStatus)
}

// Clone creates a deep copy of the user
func (u *User) Clone() *User {
	return &User{
		TelegramID: u.TelegramID,
		Name:       u.Name,
		ApiKey:     u.ApiKey,
		LastFeed:   u.LastFeed,
	}
}

// Equals checks if two users are equal
func (u *User) Equals(other *User) bool {
	if other == nil {
		return false
	}

	return u.TelegramID == other.TelegramID &&
		u.Name == other.Name &&
		u.ApiKey == other.ApiKey &&
		u.LastFeed == other.LastFeed
}

// EqualsBasic checks if two users have the same basic identity (ID and name)
func (u *User) EqualsBasic(other *User) bool {
	if other == nil {
		return false
	}

	return u.TelegramID == other.TelegramID &&
		u.Name == other.Name
}
