package models

import (
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/knackwurstking/pgpress/internal/database/errors"
	"github.com/knackwurstking/pgpress/internal/database/utils"
)

const (
	MinUserNameLength = 1
	MaxUserNameLength = 100
)

// User represents a system user with Telegram integration
type User struct {
	TelegramID int64  `json:"telegram_id"`
	UserName   string `json:"user_name"`
	ApiKey     string `json:"api_key"`
	LastFeed   int64  `json:"last_feed"`
}

// NewUser creates a new user with the provided details
func NewUser(telegramID int64, userName, apiKey string) *User {
	return &User{
		TelegramID: telegramID,
		UserName:   strings.TrimSpace(userName),
		ApiKey:     strings.TrimSpace(apiKey),
		LastFeed:   0,
	}
}

func NewUserFromInterfaceMap(modified map[string]any) *User {
	user := &User{
		TelegramID: int64(modified["telegram_id"].(float64)),
		UserName:   modified["user_name"].(string),
		ApiKey:     modified["api_key"].(string),
		LastFeed:   int64(modified["last_feed"].(float64)),
	}

	return user
}

// Validate checks if the user has valid data
func (u *User) Validate() error {
	if u.TelegramID <= 0 {
		return dberror.NewValidationError("telegram_id", "must be positive", u.TelegramID)
	}

	if u.UserName == "" {
		return dberror.NewValidationError("user_name", "cannot be empty", u.UserName)
	}
	if len(u.UserName) < MinUserNameLength {
		return dberror.NewValidationError("user_name", "too short", len(u.UserName))
	}
	if len(u.UserName) > MaxUserNameLength {
		return dberror.NewValidationError("user_name", "too long", len(u.UserName))
	}

	if u.ApiKey == "" {
		return dberror.NewValidationError("api_key", "cannot be empty", u.ApiKey)
	}
	if len(u.ApiKey) < dbutils.MinAPIKeyLength {
		return dberror.NewValidationError(
			"api_key",
			fmt.Sprintf("too short for security, must be at least %d characters", dbutils.MinAPIKeyLength),
			len(u.ApiKey),
		)
	}

	if u.LastFeed < 0 {
		return dberror.NewValidationError("last_feed", "cannot be negative", u.LastFeed)
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
		"user_name":   u.UserName,
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
		u.TelegramID, u.UserName, adminStatus)
}

// Clone creates a deep copy of the user
func (u *User) Clone() *User {
	return &User{
		TelegramID: u.TelegramID,
		UserName:   u.UserName,
		ApiKey:     u.ApiKey,
		LastFeed:   u.LastFeed,
	}
}

// Sanitize creates a version of the user with sensitive data removed (for logging).
func (u *User) Sanitize() *User {
	sanitized := u.Clone()
	sanitized.ApiKey = dbutils.MaskString(sanitized.ApiKey)
	return sanitized
}

// Equals checks if two users are equal
func (u *User) Equals(other *User) bool {
	if other == nil {
		return false
	}

	return u.TelegramID == other.TelegramID &&
		u.UserName == other.UserName &&
		u.ApiKey == other.ApiKey &&
		u.LastFeed == other.LastFeed
}

// EqualsBasic checks if two users have the same basic identity (ID and name)
func (u *User) EqualsBasic(other *User) bool {
	if other == nil {
		return false
	}

	return u.TelegramID == other.TelegramID &&
		u.UserName == other.UserName
}
