package users

import (
	"github.com/knackwurstking/pgpress/internal/services/shared/validation"
	"github.com/knackwurstking/pgpress/pkg/models"
)

func validateUser(user *models.User) error {
	if err := validation.ValidateNotNil(user, "user"); err != nil {
		return err
	}

	if err := validation.ValidateNotEmpty(user.Name, "user_name"); err != nil {
		return err
	}

	return validation.ValidateAPIKey(user.ApiKey)
}
