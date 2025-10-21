package cookies

import (
	"github.com/knackwurstking/pgpress/internal/services/shared/validation"
	"github.com/knackwurstking/pgpress/pkg/models"
)

func validateCookie(cookie *models.Cookie) error {
	if err := validation.ValidateNotNil(cookie, "cookie"); err != nil {
		return err
	}

	if err := validation.ValidateNotEmpty(cookie.Value, "value"); err != nil {
		return err
	}

	if err := validation.ValidateAPIKey(cookie.ApiKey); err != nil {
		return err
	}

	return validation.ValidatePositiveInt64(cookie.LastLogin, "last_login")
}
