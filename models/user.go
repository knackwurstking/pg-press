package models

import (
	"fmt"
	"slices"
	"strings"

	"github.com/knackwurstking/pg-press/env"
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
func (u *User) Validate() bool {
	if u.TelegramID <= 0 {
		return false
	}

	if u.Name == "" {
		return false
	}
	if len(u.Name) < env.UserNameMinLength {
		return false
	}
	if len(u.Name) > env.UserNameMaxLength {
		return false
	}

	if u.ApiKey == "" {
		return false
	}
	if len(u.ApiKey) < env.MinAPIKeyLength {
		return false
	}

	if u.LastFeed < 0 {
		return false
	}

	return true
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
