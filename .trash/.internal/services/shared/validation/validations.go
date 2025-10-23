package validation

import (
	"fmt"

	"github.com/knackwurstking/pgpress/pkg/utils"
)

func ValidateNotEmpty(value, fieldName string) error {
	if value == "" {
		return utils.NewValidationError(fmt.Sprintf("%s cannot be empty", fieldName))
	}
	return nil
}

func ValidatePositiveInt64(value int64, fieldName string) error {
	if value <= 0 {
		return utils.NewValidationError(fmt.Sprintf("%s must be positive", fieldName))
	}
	return nil
}

func ValidateNotNil(entity any, entityName string) error {
	if entity == nil {
		return utils.NewValidationError(fmt.Sprintf("%s cannot be nil", entityName))
	}
	return nil
}

func ValidateMinLength(value, fieldName string, minLength int) error {
	if len(value) < minLength {
		return utils.NewValidationError(
			fmt.Sprintf("%s must be at least %d characters", fieldName, minLength),
		)
	}
	return nil
}

// ValidateID checks if an ID is valid (positive)
func ValidateID(id int64, entityName string) error {
	return ValidatePositiveInt64(id, fmt.Sprintf("%s_id", entityName))
}

// ValidateTimestamp checks if a timestamp is valid (positive)
func ValidateTimestamp(timestamp int64, fieldName string) error {
	return ValidatePositiveInt64(timestamp, fieldName)
}
