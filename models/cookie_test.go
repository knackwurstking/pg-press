package models

import (
	"testing"
	"time"

	"github.com/knackwurstking/pg-press/env"
)

func TestCookieIsExpired(t *testing.T) {
	// Test with a fresh cookie (should not be expired)
	cookie := NewCookie("test-agent", "test-value", "test-api-key")
	if cookie.IsExpired() {
		t.Error("Fresh cookie should not be expired")
	}

	// Test with a cookie that has exceeded the expiration time
	cookie.LastLogin = time.Now().Add(-env.DefaultExpiration - time.Hour).UnixMilli()
	if !cookie.IsExpired() {
		t.Error("Cookie with expired time should be expired")
	}

	// Test with a cookie that is just under the expiration limit
	cookie.LastLogin = time.Now().Add(-env.DefaultExpiration + time.Minute).UnixMilli()
	if cookie.IsExpired() {
		t.Error("Cookie just under expiration limit should not be expired")
	}
}

func TestCookieExpires(t *testing.T) {
	cookie := NewCookie("test-agent", "test-value", "test-api-key")
	expectedExpiry := time.UnixMilli(cookie.LastLogin).Add(env.DefaultExpiration)

	actualExpiry := cookie.Expires()
	if expectedExpiry.UnixMilli() != actualExpiry.UnixMilli() {
		t.Error("Cookie expiry time does not match")
	}
}

func TestCookieAge(t *testing.T) {
	cookie := NewCookie("test-agent", "test-value", "test-api-key")

	// Should be very close to zero age for a fresh cookie
	age := cookie.Age()
	if age >= time.Second {
		t.Error("Fresh cookie should have very small age")
	}
}
