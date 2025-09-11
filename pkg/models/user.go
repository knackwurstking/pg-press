package models

import (
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/knackwurstking/pgpress/pkg/constants"
	"github.com/knackwurstking/pgpress/pkg/utils"
)

const (
	MinNameLength   = 1
	MaxNameLength   = 100
	MinAPIKeyLength = constants.MinAPIKeyLength
)

// User represents a system user with Telegram integration
type User struct {
	TelegramID int64  `json:"telegram_id"`
	Name       string `json:"user_name"`
	ApiKey     string `json:"api_key"`
	LastFeed   int64  `json:"last_feed"`
}

// NewUser creates a new user with the provided details
func NewUser(telegramID int64, userName, apiKey string) *User {
	return &User{
		TelegramID: telegramID,
		Name:       strings.TrimSpace(userName),
		ApiKey:     strings.TrimSpace(apiKey),
		LastFeed:   0,
	}
}

func NewUserFromInterfaceMap(modified map[string]any) *User {
	user := &User{
		TelegramID: int64(modified["telegram_id"].(float64)),
		Name:       modified["user_name"].(string),
		ApiKey:     modified["api_key"].(string),
		LastFeed:   int64(modified["last_feed"].(float64)),
	}

	return user
}

// Validate checks if the user has valid data
func (u *User) Validate() error {
	if u.TelegramID <= 0 {
		return utils.NewValidationError("telegram_id: must be positive")
	}

	if u.Name == "" {
		return utils.NewValidationError("user_name: cannot be empty")
	}
	if len(u.Name) < MinNameLength {
		return utils.NewValidationError("user_name: too short")
	}
	if len(u.Name) > MaxNameLength {
		return utils.NewValidationError("user_name: too long")
	}

	if u.ApiKey == "" {
		return utils.NewValidationError("api_key: cannot be empty")
	}
	if len(u.ApiKey) < MinAPIKeyLength {
		return utils.NewValidationError(
			fmt.Sprintf(
				"api_key: too short for security, must be at least %d characters", MinAPIKeyLength,
			),
		)
	}

	if u.LastFeed < 0 {
		return utils.NewValidationError("last_feed: cannot be negative")
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
	userIDStr := strconv.FormatInt(u.TelegramID, 10)

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

// Sanitize creates a version of the user with sensitive data removed (for logging).
func (u *User) Sanitize() *User {
	sanitized := u.Clone()
	sanitized.ApiKey = utils.MaskString(sanitized.ApiKey)
	return sanitized
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
