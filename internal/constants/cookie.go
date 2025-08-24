// Package constants provides shared constants for the application.
package constants

import "time"

// Cookie configuration
const (
	// CookieName is the name of the cookie used to store the API key
	CookieName = "pgpress-api-key"

	// CookieExpirationDuration is the duration for which the cookie remains valid (6 months)
	CookieExpirationDuration = time.Hour * 24 * 31 * 6
)
