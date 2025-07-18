// Package pgvis cookie models.
//
// This file defines the Cookie data structure and its associated
// validation and utility methods. Cookies represent user sessions
// and authentication tokens in the system.
package pgvis

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

const (
	// Default cookie expiration time (6 months)
	DefaultCookieExpiration = 6 * 30 * 24 * time.Hour

	// Validation constants for cookies
	MinCookieValueLength = 16
	MaxUserAgentLength   = 1000
)

// Cookie represents a user session cookie with authentication information.
// It contains session data and authentication tokens for user sessions.
type Cookie struct {
	// UserAgent is the browser user agent string
	UserAgent string `json:"user_agent"`
	// Value is the unique cookie value/token
	Value string `json:"value"`
	// ApiKey is the API key for authentication
	ApiKey string `json:"api_key"`
	// LastLogin is the last login timestamp (UNIX milliseconds)
	LastLogin int64 `json:"last_login"`
}

// NewCookie creates a new cookie with the current timestamp
func NewCookie(userAgent, value, apiKey string) *Cookie {
	return &Cookie{
		UserAgent: strings.TrimSpace(userAgent),
		Value:     strings.TrimSpace(value),
		ApiKey:    strings.TrimSpace(apiKey),
		LastLogin: time.Now().UnixMilli(),
	}
}

// NewCookieWithTime creates a new cookie with a specific timestamp
func NewCookieWithTime(userAgent, value, apiKey string, timestamp int64) *Cookie {
	return &Cookie{
		UserAgent: strings.TrimSpace(userAgent),
		Value:     strings.TrimSpace(value),
		ApiKey:    strings.TrimSpace(apiKey),
		LastLogin: timestamp,
	}
}

// GenerateSecureCookie creates a new cookie with generated secure values
func GenerateSecureCookie(userAgent string) (*Cookie, error) {
	value, err := generateSecureToken(32)
	if err != nil {
		return nil, WrapError(err, "failed to generate cookie value")
	}

	apiKey, err := generateSecureToken(MinAPIKeyLength)
	if err != nil {
		return nil, WrapError(err, "failed to generate API key")
	}

	return NewCookie(userAgent, value, apiKey), nil
}

// Validate checks if the cookie has valid data.
//
// Returns:
//   - error: ValidationError for the first validation failure, or nil if valid
func (c *Cookie) Validate() error {
	// Validate user agent
	if c.UserAgent == "" {
		return NewValidationError("user_agent", "cannot be empty", c.UserAgent)
	}
	if len(c.UserAgent) > MaxUserAgentLength {
		return NewValidationError("user_agent", "too long", len(c.UserAgent))
	}

	// Validate cookie value
	if c.Value == "" {
		return NewValidationError("value", "cannot be empty", c.Value)
	}
	if len(c.Value) < MinCookieValueLength {
		return NewValidationError("value", "too short for security", len(c.Value))
	}

	// Validate API key
	if c.ApiKey == "" {
		return NewValidationError("api_key", "cannot be empty", c.ApiKey)
	}
	if len(c.ApiKey) < MinAPIKeyLength {
		return NewValidationError("api_key", "too short for security", len(c.ApiKey))
	}

	// Validate timestamp
	if c.LastLogin <= 0 {
		return NewValidationError("last_login", "must be positive", c.LastLogin)
	}

	return nil
}

// GetLastLoginTime returns the last login time as a Go time.Time
func (c *Cookie) GetLastLoginTime() time.Time {
	return time.UnixMilli(c.LastLogin)
}

// TimeString returns a formatted string representation of the last login time
func (c *Cookie) TimeString() string {
	t := c.GetLastLoginTime()
	return t.Format("2006/01/02 15:04:05")
}

// TimeStringISO returns the last login time in ISO 8601 format
func (c *Cookie) TimeStringISO() string {
	t := c.GetLastLoginTime()
	return t.Format(time.RFC3339)
}

// Age returns the duration since the last login
func (c *Cookie) Age() time.Duration {
	return time.Since(c.GetLastLoginTime())
}

// IsExpired checks if the cookie has expired based on the default expiration time
func (c *Cookie) IsExpired() bool {
	return c.IsExpiredAfter(DefaultCookieExpiration)
}

// IsExpiredAfter checks if the cookie has expired after a specific duration
func (c *Cookie) IsExpiredAfter(duration time.Duration) bool {
	return c.Age() > duration
}

// IsActive checks if the cookie is still active (not expired)
func (c *Cookie) IsActive() bool {
	return !c.IsExpired()
}

// UpdateLastLogin updates the last login timestamp to the current time
func (c *Cookie) UpdateLastLogin() {
	c.LastLogin = time.Now().UnixMilli()
}

// RefreshToken generates a new secure token for the cookie value
func (c *Cookie) RefreshToken() error {
	newValue, err := generateSecureToken(32)
	if err != nil {
		return WrapError(err, "failed to refresh token")
	}

	c.Value = newValue
	c.UpdateLastLogin()
	return nil
}

// RefreshAPIKey generates a new secure API key
func (c *Cookie) RefreshAPIKey() error {
	newAPIKey, err := generateSecureToken(MinAPIKeyLength)
	if err != nil {
		return WrapError(err, "failed to refresh API key")
	}

	c.ApiKey = newAPIKey
	c.UpdateLastLogin()
	return nil
}

// MatchesUserAgent checks if the provided user agent matches the cookie's user agent
func (c *Cookie) MatchesUserAgent(userAgent string) bool {
	return c.UserAgent == userAgent
}

// String returns a string representation of the cookie (without sensitive data)
func (c *Cookie) String() string {
	return fmt.Sprintf("Cookie{UserAgent: %s, LastLogin: %s, Age: %v}",
		c.UserAgent, c.TimeString(), c.Age())
}

// Clone creates a deep copy of the cookie
func (c *Cookie) Clone() *Cookie {
	return &Cookie{
		UserAgent: c.UserAgent,
		Value:     c.Value,
		ApiKey:    c.ApiKey,
		LastLogin: c.LastLogin,
	}
}

// Sanitize creates a version of the cookie with sensitive data removed (for logging)
func (c *Cookie) Sanitize() *Cookie {
	sanitized := c.Clone()
	sanitized.Value = maskString(c.Value)
	sanitized.ApiKey = maskString(c.ApiKey)
	return sanitized
}

// Equals checks if two cookies are equal
func (c *Cookie) Equals(other *Cookie) bool {
	if other == nil {
		return false
	}

	return c.UserAgent == other.UserAgent &&
		c.Value == other.Value &&
		c.ApiKey == other.ApiKey &&
		c.LastLogin == other.LastLogin
}

// Helper function to generate secure random tokens
func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
