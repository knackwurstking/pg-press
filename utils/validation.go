package utils

import (
	"github.com/knackwurstking/pg-press/env"
)

// ValidateAPIKey validates an API key according to the minimum length requirement
func ValidateAPIKey(apiKey string) bool {
	return len(apiKey) == env.MinAPIKeyLength
}
