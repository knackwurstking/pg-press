package env

import (
	"time"
)

const (
	MinAPIKeyLength = 32

	CookieExpirationDuration = time.Hour * 24 * 31 * 6
	CookieName               = "pgpress-api-key"

	DateFormat     = "02.01.2006"
	TimeFormat     = "15:04"
	DateTimeFormat = DateFormat + " " + TimeFormat

	MaxFeedsPerPage = 50

	HXGlobalTrigger        = "pageLoaded"
	HXPageTool_ToolUpdated = "toolUpdated"

	// Attachment constants
	MinIDLength = 1
	MaxIDLength = 255
	MaxDataSize = 10 * 1024 * 1024 // 10MB

	// Cookie constants
	DefaultExpiration  = 6 * 30 * 24 * time.Hour
	MinValueLength     = 16
	MaxUserAgentLength = 1000

	// Tool constants
	ToolCycleWarning int64 = 800000  // Orange
	ToolCycleError   int64 = 1000000 // Red

	// TroubleReport constants
	MinTitleLength   = 1
	MaxTitleLength   = 500
	MinContentLength = 1
	MaxContentLength = 50000
)
