// ai: Organize
package pgvis

import (
	"os"
	"slices"
	"strconv"
	"strings"
)

const (
	// User validation constants
	MinUserNameLength = 1
	MaxUserNameLength = 100
)

// User represents a system user with Telegram integration
type User struct {
	TelegramID int64  `json:"telegram_id"` // Telegram user ID
	UserName   string `json:"user_name"`   // Display name
	ApiKey     string `json:"api_key"`     // API key for authentication (optional)
	LastFeed   int    `json:"last_feed"`   // Last viewed feed ID
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

// NewUserWithLastFeed creates a new user with a specific last feed ID
func NewUserWithLastFeed(telegramID int64, userName, apiKey string, lastFeed int) *User {
	user := NewUser(telegramID, userName, apiKey)
	user.LastFeed = lastFeed
	return user
}

// NewBasicUser creates a new user with minimal information
func NewBasicUser(telegramID int64, userName string) *User {
	return NewUser(telegramID, userName, "")
}

// Validate checks if the user has valid data
func (u *User) Validate() error {
	if u.TelegramID <= 0 {
		return NewValidationError("telegram_id", "must be positive", u.TelegramID)
	}

	if u.UserName == "" {
		return NewValidationError("user_name", "cannot be empty", u.UserName)
	}
	if len(u.UserName) < MinUserNameLength {
		return NewValidationError("user_name", "too short", len(u.UserName))
	}
	if len(u.UserName) > MaxUserNameLength {
		return NewValidationError("user_name", "too long", len(u.UserName))
	}

	if u.ApiKey != "" && len(u.ApiKey) < MinAPIKeyLength {
		return NewValidationError("api_key", "too short for security", len(u.ApiKey))
	}

	if u.LastFeed < 0 {
		return NewValidationError("last_feed", "cannot be negative", u.LastFeed)
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

// HasAPIKey checks if the user has an API key configured
func (u *User) HasAPIKey() bool {
	return u.ApiKey != ""
}

// IsValidAPIKey checks if the provided API key matches the user's API key
func (u *User) IsValidAPIKey(apiKey string) bool {
	if !u.HasAPIKey() {
		return false
	}
	return u.ApiKey == apiKey
}

// UpdateUserName updates the user's display name
func (u *User) UpdateUserName(newUserName string) error {
	newUserName = strings.TrimSpace(newUserName)

	if newUserName == "" {
		return NewValidationError("user_name", "cannot be empty", newUserName)
	}

	if len(newUserName) < MinUserNameLength {
		return NewValidationError("user_name", "too short", len(newUserName))
	}

	if len(newUserName) > MaxUserNameLength {
		return NewValidationError("user_name", "too long", len(newUserName))
	}

	u.UserName = newUserName
	return nil
}

// UpdateAPIKey updates the user's API key
func (u *User) UpdateAPIKey(newAPIKey string) error {
	newAPIKey = strings.TrimSpace(newAPIKey)

	if newAPIKey != "" && len(newAPIKey) < MinAPIKeyLength {
		return NewValidationError("api_key", "too short for security", len(newAPIKey))
	}

	u.ApiKey = newAPIKey
	return nil
}

// UpdateLastFeed updates the last viewed feed ID
func (u *User) UpdateLastFeed(feedID int) error {
	if feedID < 0 {
		return NewValidationError("feed_id", "cannot be negative", feedID)
	}

	u.LastFeed = feedID
	return nil
}

// ClearAPIKey removes the user's API key
func (u *User) ClearAPIKey() {
	u.ApiKey = ""
}

// GetDisplayInfo returns safe user information for display (without API key)
func (u *User) GetDisplayInfo() map[string]any {
	return map[string]any{
		"telegram_id": u.TelegramID,
		"user_name":   u.UserName,
		"has_api_key": u.HasAPIKey(),
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

	apiKeyStatus := ""
	if u.HasAPIKey() {
		apiKeyStatus = " [has API key]"
	}

	return "User{ID: " + strconv.FormatInt(u.TelegramID, 10) +
		", Name: " + u.UserName + adminStatus + apiKeyStatus + "}"
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

// Sanitize creates a version of the user with sensitive data removed (for logging)
func (u *User) Sanitize() *User {
	sanitized := u.Clone()
	if sanitized.ApiKey != "" {
		sanitized.ApiKey = maskString(sanitized.ApiKey)
	}
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
