package pressregenerations

import (
	"time"

	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/utils"

	"github.com/labstack/echo/v4"
)

func ParseForm(c echo.Context, press shared.PressNumber) (*shared.PressRegeneration, *errors.MasterError) {
	startStr := c.FormValue("started")
	if startStr == "" {
		return nil, errors.NewValidationError("started date is required").HTTPError()
	}

	stopStr := c.FormValue("completed")
	if stopStr == "" {
		return nil, errors.NewValidationError("completed date is required").HTTPError()
	}

	start, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		return nil, errors.NewValidationError("invalid started date format: %s", err.Error()).HTTPError()
	}

	stop, err := time.Parse("2006-01-02", stopStr)
	if err != nil {
		return nil, errors.NewValidationError("invalid completed date format: %s", err.Error()).HTTPError()
	}

	reason := c.FormValue("reason")

	data := &shared.PressRegeneration{
		PressNumber: press,
		Start:       shared.UnixMilli(start.UnixMilli()),
		Stop:        shared.UnixMilli(stop.UnixMilli()),
	}

	if err := data.Validate(); err != nil {
		return nil, errors.NewValidationError(err.Error()).HTTPError()
	}

	return data, nil
}

// RegenerationsFormData holds the form data for press regeneration
type RegenerationsFormData struct {
	Started   time.Time
	Completed time.Time
}

// ParseForm parses the form data from the press regenerations page
// and validates it to create a PressRegeneration model
func ParseForm(c echo.Context, press models.PressNumber) (
	*models.PressRegeneration, *errors.MasterError,
) {

	reason := c.FormValue("reason") // This is optional for now

	startedAt, merr := utils.ParseFormValueTime(c, "started")
	if merr != nil {
		return nil, merr
	}

	completedAt, merr := utils.ParseFormValueTime(c, "completed")
	if merr != nil {
		return nil, merr
	}

	r := &models.PressRegeneration{
		PressNumber: press,
		StartedAt:   startedAt,
		CompletedAt: completedAt,
		Reason:      reason,
	}

	verr := r.Validate()
	if verr != nil {
		return r, verr.MasterError()
	}

	return r, nil
}
