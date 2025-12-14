package dialogs

import (
	"fmt"
	"log/slog"

	"github.com/knackwurstking/pg-press/internal/env"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/dialogs/templates"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/services"
	"github.com/knackwurstking/pg-press/internal/shared"

	"github.com/labstack/echo/v4"
)

func (h *Handler) GetEditToolRegeneration(c echo.Context) *echo.HTTPError {
	rid, merr := utils.ParseQueryInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	regenerationID := models.ToolRegenerationID(rid)

	regeneration, merr := h.registry.ToolRegenerations.Get(regenerationID)
	if merr != nil {
		return merr.Echo()
	}

	resolvedRegeneration, merr := services.ResolveToolRegeneration(h.registry, regeneration)
	if merr != nil {
		return merr.Echo()
	}

	dialog := templates.EditToolRegenerationDialog(resolvedRegeneration)
	err := dialog.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "EditToolRegenerationDialog")
	}

	return nil
}

func (h *Handler) PutEditToolRegeneration(c echo.Context) *echo.HTTPError {
	slog.Info("Updating tool regeneration entry")

	user, merr := utils.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	var regenerationID models.ToolRegenerationID
	id, merr := utils.ParseQueryInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	} else {
		regenerationID = models.ToolRegenerationID(id)
	}

	r, merr := h.registry.ToolRegenerations.Get(regenerationID)
	if merr != nil {
		return merr.Echo()
	}

	regeneration, merr := services.ResolveToolRegeneration(h.registry, r)
	if merr != nil {
		return merr.Echo()
	}

	formData := GetEditToolRegenerationFormData(c)
	regeneration.Reason = formData.Reason

	merr = h.registry.ToolRegenerations.Update(regeneration.ToolRegeneration, user)
	if merr != nil {
		return merr.Echo()
	}

	// Create Feed
	title := "Werkzeug Regenerierung aktualisiert"
	content := fmt.Sprintf(
		"Tool: %s\nGebundener Zyklus: %s (Teil Zyklen: %d)",
		regeneration.GetTool().String(),
		regeneration.GetCycle().Date.Format(env.DateFormat),
		regeneration.GetCycle().PartialCycles,
	)

	if regeneration.Reason != "" {
		content += fmt.Sprintf("\nReason: %s", regeneration.Reason)
	}

	merr = h.registry.Feeds.Add(title, content, user.TelegramID)
	if merr != nil {
		slog.Warn("Failed to create feed for cycle creation", "error", merr)
	}

	utils.SetHXTrigger(c, env.HXGlobalTrigger)

	return nil
}

type EditToolRegenerationFormData struct {
	Reason string
}

func GetEditToolRegenerationFormData(c echo.Context) *EditToolRegenerationFormData {
	return &EditToolRegenerationFormData{
		Reason: c.FormValue("reason"),
	}
}
