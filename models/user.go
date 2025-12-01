package models

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/knackwurstking/pg-press/env"
	"github.com/knackwurstking/pg-press/errors"
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

// Validate checks if the user has valid data
func (u *User) Validate() error {
	if u.TelegramID <= 0 {
		return errors.NewValidationError("telegram_id: must be positive")
	}

	if u.Name == "" {
		return errors.NewValidationError("user_name: cannot be empty")
	}
	if len(u.Name) < env.UserNameMinLength {
		return errors.NewValidationError("user_name: too short")
	}
	if len(u.Name) > env.UserNameMaxLength {
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
	if env.Admins == "" {
		return false
	}

	adminIDs := strings.Split(env.Admins, ",")
	userIDStr := fmt.Sprintf("%d", u.TelegramID)

	return slices.Contains(adminIDs, userIDStr)
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
