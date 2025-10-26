package services

import (
	"fmt"

	"github.com/knackwurstking/pg-press/env"
	"github.com/knackwurstking/pg-press/errors"
)

func ValidateAPIKey(apiKey string) error {
	if apiKey == "" {
		return errors.NewValidationError("api_key is required")
	}

	if len(apiKey) < env.MinAPIKeyLength {
		return errors.NewValidationError(
			fmt.Sprintf("api key must be at least %d characters", env.MinAPIKeyLength),
		)
	}

	return nil
}
