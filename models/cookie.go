// TODO: Remove useless stuff
package models

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/knackwurstking/pg-press/env"
	"github.com/knackwurstking/pg-press/errors"
)

const (
	DefaultExpiration  = 6 * 30 * 24 * time.Hour
	MinValueLength     = 16
	MaxUserAgentLength = 1000
)

// Cookie represents a user session with authentication information.
type Cookie struct {
	UserAgent string `json:"user_agent"`
	Value     string `json:"value"`
	ApiKey    string `json:"api_key"`
	LastLogin int64  `json:"last_login"`
}

// New creates a new cookie with the current timestamp.
func NewCookie(userAgent, value, apiKey string) *Cookie {
	return &Cookie{
		UserAgent: strings.TrimSpace(userAgent),
		Value:     strings.TrimSpace(value),
		ApiKey:    strings.TrimSpace(apiKey),
		LastLogin: time.Now().UnixMilli(),
	}
}

// Validate checks if the cookie has valid data.
func (c *Cookie) Validate() error {
	if c.UserAgent == "" {
		return errors.NewValidationError("user_agent: cannot be empty")
	}
	if len(c.UserAgent) > MaxUserAgentLength {
		return errors.NewValidationError("user_agent: too long")
	}

	if c.Value == "" {
		return errors.NewValidationError("value: cannot be empty")
	}
	if len(c.Value) < MinValueLength {
		return errors.NewValidationError("value: too short for security")
	}

	if c.ApiKey == "" {
		return errors.NewValidationError("api_key: cannot be empty")
	}
	if len(c.ApiKey) < env.MinAPIKeyLength {
		return errors.NewValidationError("api_key: too short for security")
	}

	if c.LastLogin <= 0 {
		return errors.NewValidationError("last_login: must be positive")
	}

	return nil
}

// GetLastLoginTime returns the last login time as a Go time.Time.
func (c *Cookie) GetLastLoginTime() time.Time {
	return time.UnixMilli(c.LastLogin)
}

// TimeString returns a formatted string representation of the last login time.
func (c *Cookie) TimeString() string {
	return c.GetLastLoginTime().Format("2006/01/02 15:04:05")
}

// TimeStringISO returns the last login time in ISO 8601 format.
func (c *Cookie) TimeStringISO() string {
	return c.GetLastLoginTime().Format(time.RFC3339)
}

// Age returns the duration since the last login.
func (c *Cookie) Age() time.Duration {
	return time.Since(c.GetLastLoginTime())
}

// IsExpired checks if the cookie has expired based on the default expiration time.
func (c *Cookie) IsExpired() bool {
	return c.IsExpiredAfter(DefaultExpiration)
}

// IsExpiredAfter checks if the cookie has expired after a specific duration.
func (c *Cookie) IsExpiredAfter(duration time.Duration) bool {
	return c.Age() > duration
}

// IsActive checks if the cookie is still active (not expired).
func (c *Cookie) IsActive() bool {
	return !c.IsExpired()
}

// UpdateLastLogin updates the last login timestamp to the current time.
func (c *Cookie) UpdateLastLogin() {
	c.LastLogin = time.Now().UnixMilli()
}

// RefreshToken generates a new secure token for the cookie value.
func (c *Cookie) RefreshToken() error {
	newValue, err := GenerateSecureToken(32)
	if err != nil {
		return fmt.Errorf("failed to refresh token: %v", err)
	}

	c.Value = newValue
	c.UpdateLastLogin()
	return nil
}

// RefreshAPIKey generates a new secure API key.
func (c *Cookie) RefreshAPIKey() error {
	newAPIKey, err := GenerateSecureToken(env.MinAPIKeyLength)
	if err != nil {
		return fmt.Errorf("failed to refresh API key: %v", err)
	}

	c.ApiKey = newAPIKey
	c.UpdateLastLogin()
	return nil
}

// MatchesUserAgent checks if the provided user agent matches the cookie's user agent.
func (c *Cookie) MatchesUserAgent(userAgent string) bool {
	return c.UserAgent == userAgent
}

// String returns a string representation of the cookie (without sensitive data).
func (c *Cookie) String() string {
	return fmt.Sprintf("Cookie{UserAgent: %s, LastLogin: %s, Age: %v}",
		c.UserAgent, c.TimeString(), c.Age())
}

// Clone creates a deep copy of the cookie.
func (c *Cookie) Clone() *Cookie {
	return &Cookie{
		UserAgent: c.UserAgent,
		Value:     c.Value,
		ApiKey:    c.ApiKey,
		LastLogin: c.LastLogin,
	}
}

// Equals checks if two cookies are equal.
func (c *Cookie) Equals(other *Cookie) bool {
	if other == nil {
		return false
	}

	return c.UserAgent == other.UserAgent &&
		c.Value == other.Value &&
		c.ApiKey == other.ApiKey &&
		c.LastLogin == other.LastLogin
}

// SortCookies sorts a slice of cookies by last login time in descending order.
func SortCookies(cookies []*Cookie) []*Cookie {
	if len(cookies) <= 1 {
		return cookies
	}

	cookiesSorted := make([]*Cookie, 0, len(cookies))

outer:
	for _, cookie := range cookies {
		for i, sortedCookie := range cookiesSorted {
			if cookie.LastLogin > sortedCookie.LastLogin {
				cookiesSorted = slices.Insert(cookiesSorted, i, cookie)
				continue outer
			}
		}
		cookiesSorted = append(cookiesSorted, cookie)
	}

	return cookiesSorted
}

func GenerateSecureToken(length int) (string, error) {
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
