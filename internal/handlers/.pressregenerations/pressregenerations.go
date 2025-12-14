package pressregenerations

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/knackwurstking/pg-press/internal/common"
	"github.com/knackwurstking/pg-press/internal/env"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/pressregenerations/templates"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/shared"

	ui "github.com/knackwurstking/ui/ui-templ"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	registry *common.DB
}

func NewHandler(r *common.DB) *Handler {
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
	press, merr := h.parseParamPress(c)
	if merr != nil {
		return merr.Echo()
	}

	err := templates.Page(templates.PageProps{
		Press: press,
	}).Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "Press Regenration Page")
	}

	return nil
}

func (h *Handler) HxAddRegeneration(c echo.Context) error {
	slog.Info("Adding new press regeneration entry")

	user, merr := utils.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	press, merr := h.parseParamPress(c)
	if merr != nil {
		return merr.Echo()
	}

	data, merr := ParseForm(c, press)
	if merr != nil {
		return merr.Echo()
	}

	_, merr = h.registry.PressRegenerations.Add(data)
	if merr != nil {
		return merr.Echo()
	}

	h.createFeed(fmt.Sprintf("\"Regenerierung\" für Presse %d Hinzugefügt", press), data, user)
	utils.SetHXRedirect(c, utils.UrlPress(press).Page)

	return nil
}

func (h *Handler) HxDeleteRegeneration(c echo.Context) error {
	slog.Info("Removing press regeneration entry")

	user, merr := utils.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	id, merr := utils.ParseQueryInt64(c, "id") // PressRegenerationID
	if merr != nil {
		return merr.Echo()
	}

	rid := models.PressRegenerationID(id)
	r, _ := h.registry.PressRegenerations.Get(rid) // Need this for the feed

	merr = h.registry.PressRegenerations.Delete(rid)
	if merr != nil {
		return merr.Echo()
	}

	h.createFeed(fmt.Sprintf("\"Regenerierung\" für Presse %d entfernt", r.PressNumber), r, user)
	utils.SetHXTrigger(c, env.HXGlobalTrigger)

	return nil
}

func (h *Handler) createFeed(feedTitle string, r *models.PressRegeneration, user *models.User) {
	feedContent := fmt.Sprintf("Presse: %d\n", r.PressNumber)
	feedContent += fmt.Sprintf("Von %s bis %s\n", r.StartedAt.Format(env.DateFormat), r.CompletedAt.Format(env.DateFormat))
	feedContent += fmt.Sprintf("Reason: %s", r.Reason)
	merr := h.registry.Feeds.Add(feedTitle, feedContent, user.TelegramID)
	if merr != nil {
		slog.Warn("Failed to create feed for deleting a press regeneration", "error", merr)
	}
}
