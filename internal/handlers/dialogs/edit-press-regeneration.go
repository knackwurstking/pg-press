package dialogs

import (
	"fmt"
	"log/slog"

	"github.com/knackwurstking/pg-press/internal/env"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/dialogs/templates"
	"github.com/knackwurstking/pg-press/models"
	"github.com/knackwurstking/pg-press/utils"
	"github.com/labstack/echo/v4"
)

func (h *Handler) GetEditPressRegeneration(c echo.Context) *echo.HTTPError {
	id, merr := utils.ParseQueryInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}

	r, merr := h.registry.PressRegenerations.Get(models.PressRegenerationID(id))
	if merr != nil {
		return merr.Echo()
	}

	d := templates.EditPressRegenerationDialog(r)
	err := d.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "EditPressRegenerationDialog")
	}

	return nil
}

func (h *Handler) PutEditPressRegeneration(c echo.Context) *echo.HTTPError {
	slog.Info("Updating press regeneration entry")

	user, merr := utils.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	id, merr := utils.ParseQueryInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}

	r, merr := h.registry.PressRegenerations.Get(models.PressRegenerationID(id))
	if merr != nil {
		return merr.Echo()
	}

	r.Reason = c.FormValue("reason")
	merr = h.registry.PressRegenerations.Update(r)
	if merr != nil {
		return merr.Echo()
	}

	feedTitle := fmt.Sprintf("\"Regenerierung\" für Presse %d aktualisiert", r.PressNumber)
	feedContent := fmt.Sprintf("Presse: %d\n", r.PressNumber)
	feedContent += fmt.Sprintf("Von: %s, Bis: %s\n", r.StartedAt.Format(env.DateTimeFormat), r.CompletedAt.Format(env.DateTimeFormat))
	feedContent += fmt.Sprintf("Bemerkung: %s", r.Reason)

	merr = h.registry.Feeds.Add(feedTitle, feedContent, user.TelegramID)
	if merr != nil {
		slog.Warn("Add feed", "title", feedTitle, "error", merr)
	}

	utils.SetHXTrigger(c, env.HXGlobalTrigger)

	return nil
}
