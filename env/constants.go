package env

import "time"

const (
	MinAPIKeyLength = 32

	CookieExpirationDuration = time.Hour * 24 * 31 * 6
	CookieName               = "pgpress-api-key"

	DateFormat     = "02.01.2006"
	TimeFormat     = "15:04"
	DateTimeFormat = DateFormat + " " + TimeFormat

	MaxFeedsPerPage = 50

	UserNameMinLength = 1
	UserNameMaxLength = 100
)
