package pressregenerations

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/knackwurstking/pg-press/env"
	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/handlers/pressregenerations/templates"
	"github.com/knackwurstking/pg-press/models"
	"github.com/knackwurstking/pg-press/services"
	"github.com/knackwurstking/pg-press/utils"
	"github.com/knackwurstking/ui"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	registry *services.Registry
}

func NewHandler(r *services.Registry) *Handler {
	return &Handler{
		registry: r,
	}
}

func (h *Handler) RegisterRoutes(e *echo.Echo, path string) {
	ui.RegisterEchoRoutes(e, env.ServerPathPrefix, []*ui.EchoRoute{
		// Press regeneration page
		ui.NewEchoRoute(
			http.MethodGet,
			path+"/:press",
			h.GetRegenerationPage,
		),

		ui.NewEchoRoute(
			http.MethodPost,
			path+"/:press",
			h.HxAddRegeneration,
		),

		ui.NewEchoRoute(
			http.MethodDelete,
			path+"/:press/delete",
			h.HxDeleteRegeneration,
		),
	})
}

func (h *Handler) GetRegenerationPage(c echo.Context) error {
	press, eerr := h.parseParamPress(c)
	if eerr != nil {
		return eerr
	}

	if err := templates.Page(templates.PageProps{
		Press: press,
	}).Render(c.Request().Context(), c.Response()); err != nil {
		return errors.Handler(err, "render press page template")
	}

	return nil
}

func (h *Handler) HxAddRegeneration(c echo.Context) error {
	slog.Info("Add a new press regeneration entry")

	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}

	press, eerr := h.parseParamPress(c)
	if eerr != nil {
		return eerr
	}

	data, eerr := ParseFormRegenerationsPage(c, press)
	if eerr != nil {
		return eerr
	}

	if _, err := h.registry.PressRegenerations.Add(data); err != nil {
		return errors.Handler(err, "add press regeneration")
	}

	h.createFeed(fmt.Sprintf("\"Regenerierung\" für Presse %d Hinzugefügt", press), data, user)
	utils.SetHXRedirect(c, utils.UrlPress(press).Page)

	return nil
}

func (h *Handler) HxDeleteRegeneration(c echo.Context) (err error) {
	slog.Info("Remove an press regeneration entry")

	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}

	id, err := utils.ParseQueryInt64(c, "id") // PressRegenerationID
	if err != nil {
		return errors.BadRequest(err, "missing id query")
	}

	rid := models.PressRegenerationID(id)
	r, _ := h.registry.PressRegenerations.Get(rid) // Need this for the feed
	if err := h.registry.PressRegenerations.Delete(rid); err != nil {
		return errors.Handler(err, "delete press regeneration")
	}

	h.createFeed(fmt.Sprintf("\"Regenerierung\" für Presse %d entfernt", r.PressNumber), r, user)
	utils.SetHXTrigger(c, env.HXGlobalTrigger)

	return nil
}

func (h *Handler) createFeed(feedTitle string, r *models.PressRegeneration, user *models.User) {
	feedContent := fmt.Sprintf("Presse: %d\n", r.PressNumber)
	feedContent += fmt.Sprintf("Von %s bis %s\n", r.StartedAt.Format(env.DateFormat), r.CompletedAt.Format(env.DateFormat))
	feedContent += fmt.Sprintf("Reason: %s", r.Reason)
	if _, err := h.registry.Feeds.AddSimple(feedTitle, feedContent, user.TelegramID); err != nil {
		slog.Warn("Failed to create feed for deleting a press regeneration", "error", err)
	}
}
