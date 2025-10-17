package tool

import (
	"errors"

	"github.com/labstack/echo/v4"
)

type RegenerationEditFormData struct {
	Reason string
}

func (h *Handler) HTMXGetEditRegeneration(c echo.Context) error {
	// TODO: ...

	return errors.New("under construction")
}

func (h *Handler) HTMXPutEditRegeneration(c echo.Context) error {
	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.RenderBadRequest(c, "failed to get user from context: "+err.Error())
	}

	regenerationID, err := h.ParseInt64Query(c, "id")
	if err != nil {
		return h.RenderBadRequest(c, "failed to get tool id from url param: "+err.Error())
	}

	formData := h.parseRegenerationEditFormData(c)

	regeneration, err := h.DB.ToolRegenerations.Get(regenerationID)
	if err != nil {
		return h.HandleError(c, err, "failed to get regeneration")
	}
	regeneration.Reason = formData.Reason

	err = h.DB.ToolRegenerations.Update(regeneration, user)
	if err != nil {
		return h.HandleError(c, err, "failed to update regeneration")
	}

	return nil
}

func (h *Handler) HTMXDeleteRegeneration(c echo.Context) error {
	// TODO: ...

	return errors.New("under construction")
}

func (h *Handler) parseRegenerationEditFormData(c echo.Context) *RegenerationEditFormData {
	return &RegenerationEditFormData{
		Reason: c.FormValue("reason"),
	}
}
