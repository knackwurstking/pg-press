package utils

import (
	"fmt"

	"github.com/knackwurstking/pg-press/env"
)

// ValidateAPIKey validates an API key according to the minimum length requirement
func ValidateAPIKey(apiKey string) error {
	if apiKey == "" {
		return fmt.Errorf("api_key is required")
	}

	if len(apiKey) < env.MinAPIKeyLength {
		return fmt.Errorf("api key must be at least %d characters", env.MinAPIKeyLength)
	}

	return nil
}
