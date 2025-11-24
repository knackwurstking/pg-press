package models

import (
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
	return c.GetLastLoginTime().Format(env.DateTimeFormat)
}

func (c *Cookie) Age() time.Duration {
	return time.Since(c.GetLastLoginTime())
}

// IsExpired checks if the cookie has expired based on the default expiration time.
func (c *Cookie) IsExpired() bool {
	return c.Age() > DefaultExpiration
}

func (c *Cookie) Expires() time.Time {
	return time.UnixMilli(c.LastLogin).Add(DefaultExpiration)
}

// UpdateLastLogin updates the last login timestamp to the current time.
func (c *Cookie) UpdateLastLogin() {
	c.LastLogin = time.Now().UnixMilli()
}

// String returns a string representation of the cookie (without sensitive data).
func (c *Cookie) String() string {
	return fmt.Sprintf("Cookie{UserAgent: %s, LastLogin: %s, Age: %v}",
		c.UserAgent, c.TimeString(), c.Age())
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
