// ai: Orgnaize
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
)

// Cookie represents a user session cookie with authentication information
type Cookie struct {
	UserAgent string `json:"user_agent"` // Browser user agent string
	Value     string `json:"value"`      // Cookie value/token
	ApiKey    string `json:"api_key"`    // API key for authentication
	LastLogin int64  `json:"last_login"` // Last login timestamp (UNIX milliseconds)
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
		return nil, fmt.Errorf("failed to generate cookie value: %w", err)
	}

	apiKey, err := generateSecureToken(MinAPIKeyLength)
	if err != nil {
		return nil, fmt.Errorf("failed to generate API key: %w", err)
	}

	return NewCookie(userAgent, value, apiKey), nil
}

// Validate checks if the cookie has valid data
func (c *Cookie) Validate() error {
	if c.UserAgent == "" {
		return fmt.Errorf("user agent cannot be empty")
	}

	if c.Value == "" {
		return fmt.Errorf("cookie value cannot be empty")
	}

	if c.ApiKey == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	if len(c.ApiKey) < MinAPIKeyLength {
		return fmt.Errorf("API key must be at least %d characters long", MinAPIKeyLength)
	}

	if c.LastLogin <= 0 {
		return fmt.Errorf("last login timestamp must be positive")
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
		return fmt.Errorf("failed to refresh token: %w", err)
	}

	c.Value = newValue
	c.UpdateLastLogin()
	return nil
}

// RefreshAPIKey generates a new secure API key
func (c *Cookie) RefreshAPIKey() error {
	newAPIKey, err := generateSecureToken(MinAPIKeyLength)
	if err != nil {
		return fmt.Errorf("failed to refresh API key: %w", err)
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
