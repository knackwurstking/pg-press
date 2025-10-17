package tool

import (
	"errors"
	"fmt"

	"github.com/knackwurstking/pgpress/internal/web/features/tool/templates"
	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/labstack/echo/v4"
)

type RegenerationEditFormData struct {
	Reason string
}

func (h *Handler) HTMXGetEditRegeneration(c echo.Context) error {
	regenerationID, err := h.ParseInt64Query(c, "id")
	if err != nil {
		return h.RenderBadRequest(c, err.Error())
	}

	regeneration, err := h.DB.ToolRegenerations.Get(regenerationID)
	if err != nil {
		return h.HandleError(c, err, "get regeneration failed")
	}

	resolvedRegeneration, err := h.resolveRegeneration(c, regeneration)
	if err != nil {
		return err
	}

	dialog := templates.DialogEditRegeneration(resolvedRegeneration)

	if err := dialog.Render(c.Request().Context(), c.Response()); err != nil {
		return h.HandleError(c, err, "render dialog failed")
	}

	return errors.New("under construction")
}

func (h *Handler) HTMXPutEditRegeneration(c echo.Context) error {
	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.RenderBadRequest(c, "failed to get user from context: "+err.Error())
	}

	regenerationID, err := h.ParseInt64Query(c, "id")
	if err != nil {
		return h.RenderBadRequest(c, "failed to get the regeneration id from url query: "+err.Error())
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
	regenerationID, err := h.ParseInt64Query(c, "id")
	if err != nil {
		return h.RenderBadRequest(c, "failed to get the regeneration id from url query: "+err.Error())
	}

	if err := h.DB.ToolRegenerations.Delete(regenerationID); err != nil {
		return h.HandleError(c, err, "failed to delete regeneration")
	}

	return nil
}

func (h *Handler) parseRegenerationEditFormData(c echo.Context) *RegenerationEditFormData {
	return &RegenerationEditFormData{
		Reason: c.FormValue("reason"),
	}
}

func (h *Handler) resolveRegeneration(c echo.Context, r *models.Regeneration) (*models.ResolvedRegeneration, error) {
	tool, err := h.DB.Tools.Get(r.ToolID)
	if err != nil {
		return nil, h.HandleError(c, err,
			fmt.Sprintf("failed to get tool %d for regeneration %d",
				r.ToolID, r.ID))
	}

	cycle, err := h.DB.PressCycles.Get(r.CycleID)
	if err != nil {
		return nil, h.HandleError(c, err,
			fmt.Sprintf("failed to get press cycle %d for regeneration %d",
				r.CycleID, r.ID))
	}

	user, err := h.DB.Users.Get(*r.PerformedBy)
	if err != nil {
		return nil, h.HandleError(c, err,
			fmt.Sprintf("failed to get user %d for regeneration %d",
				*r.PerformedBy, r.ID))
	}

	return models.NewResolvedRegeneration(r, tool, cycle, user), nil
}
