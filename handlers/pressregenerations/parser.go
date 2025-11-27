package pressregenerations

import (
	"time"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
	"github.com/knackwurstking/pg-press/utils"
	"github.com/labstack/echo/v4"
)

type RegenerationsFormData struct {
	Started   time.Time
	Completed time.Time
}

func ParseFormRegenerationsPage(c echo.Context, press models.PressNumber) (*models.PressRegeneration, *echo.HTTPError) {
	var (
		reason      string
		startedAt   time.Time
		completedAt time.Time
		err         error
	)

	reason = c.FormValue("reason") // This is optional for now

	startedAt, err = utils.ParseFormValueTime(c, "started")
	if err != nil {
		return nil, errors.BadRequest(err, "parsing form value \"started\"")
	}

	completedAt, err = utils.ParseFormValueTime(c, "completed")
	if err != nil {
		return nil, errors.BadRequest(err, "parsing form value \"completed\"")
	}

	return &models.PressRegeneration{
		PressNumber: press,
		StartedAt:   startedAt,
		CompletedAt: completedAt,
		Reason:      reason,
	}, nil
}
