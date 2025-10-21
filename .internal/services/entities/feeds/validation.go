package feeds

import (
	"github.com/knackwurstking/pgpress/internal/services/shared/validation"
	"github.com/knackwurstking/pgpress/pkg/models"
)

func validateFeed(feed *models.Feed) error {
	if err := validation.ValidateNotNil(feed, "feed"); err != nil {
		return err
	}

	if err := validation.ValidateNotEmpty(feed.Title, "title"); err != nil {
		return err
	}

	if err := validation.ValidateNotEmpty(feed.Content, "content"); err != nil {
		return err
	}

	if err := validation.ValidatePositiveInt64(feed.UserID, "user_id"); err != nil {
		return err
	}

	return validation.ValidatePositiveInt64(feed.CreatedAt, "created_at")
}
