package pressregenerations

import (
	"time"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
	"github.com/knackwurstking/pg-press/utils"
	"github.com/labstack/echo/v4"
)

// RegenerationsFormData holds the form data for press regeneration
type RegenerationsFormData struct {
	Started   time.Time
	Completed time.Time
}

// ParseFormRegenerationsPage parses the form data from the press regenerations page
// and validates it to create a PressRegeneration model
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
		return nil, errors.NewBadRequestError(err, "parsing form value \"started\"")
	}

	completedAt, err = utils.ParseFormValueTime(c, "completed")
	if err != nil {
		return nil, errors.NewBadRequestError(err, "parsing form value \"completed\"")
	}

	r := &models.PressRegeneration{
		PressNumber: press,
		StartedAt:   startedAt,
		CompletedAt: completedAt,
		Reason:      reason,
	}

	if err = r.Validate(); err != nil {
		return r, errors.NewBadRequestError(err, "invalid press regeneration data")
	}

	return r, nil
}

func (h *Handler) parseParamPress(c echo.Context) (models.PressNumber, *echo.HTTPError) {
	pressNum, err := utils.ParseParamInt8(c, "press")
	if err != nil {
		return -1, errors.NewBadRequestError(err, "invalid or missing press parameter")
	}

	press := models.PressNumber(pressNum)
	if !models.IsValidPressNumber(&press) {
		return -1, errors.NewBadRequestError(err, "invalid press number")
	}

	return press, nil
}
