package pressregenerations

import (
	"fmt"
	"net/http"
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
func ParseFormRegenerationsPage(c echo.Context, press models.PressNumber) (
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

	if !r.Validate() {
		return r, errors.NewMasterError(
			fmt.Errorf("invalid regeneration data for press %d", press),
			http.StatusBadRequest,
		)
	}

	return r, nil
}

func (h *Handler) parseParamPress(c echo.Context) (models.PressNumber, *errors.MasterError) {
	pressNum, merr := utils.ParseParamInt8(c, "press")
	if merr != nil {
		return -1, merr
	}

	press := models.PressNumber(pressNum)
	if !models.IsValidPressNumber(&press) {
		return -1, merr
	}

	return press, nil
}
