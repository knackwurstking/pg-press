package shared

import (
	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/utils"
)

const (
	UserNameMinLength = 1
	UserNameMaxLength = 100
)

func (e *User) Validate() *errors.ValidationError {
	if e.Name == "" {
		return errors.NewValidationError("name cannot be empty")
	}
	if len(e.Name) < UserNameMinLength || len(e.Name) > UserNameMaxLength {
		return errors.NewValidationError(
			"name length must be between %d and %d characters",
			UserNameMinLength, UserNameMaxLength,
		)
	}
	if !utils.ValidateAPIKey(e.ApiKey) {
		return errors.NewValidationError("api_key is not valid")
	}
	return nil
}

func (e *Cookie) Validate() *errors.ValidationError

func (e *Session) Validate() *errors.ValidationError
