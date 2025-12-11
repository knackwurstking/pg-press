package notes

import (
	"log/slog"
	"net/http"

	"github.com/knackwurstking/pg-press/env"
	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/handlers/notes/templates"
	"github.com/knackwurstking/pg-press/models"
	"github.com/knackwurstking/pg-press/services"
	"github.com/knackwurstking/pg-press/utils"
	ui "github.com/knackwurstking/ui/ui-templ"
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
		// Notes page
		ui.NewEchoRoute(http.MethodGet, path, h.GetNotesPage),

		// HTMX routes for notes deletion
		ui.NewEchoRoute(http.MethodDelete, path+"/delete", h.HTMXDeleteNote),

		// Render Notes Grid
		ui.NewEchoRoute(http.MethodGet, path+"/grid", h.HTMXGetNotesGrid),
	})
}

// GetNotesPage serves the main notes page
func (h *Handler) GetNotesPage(c echo.Context) error {
	// Get all notes with defensive error handling
	notes, merr := h.registry.Notes.List()
	if merr != nil {
		return merr.Echo()
	}

	// Handle case where notes might be nil
	if notes == nil {
		notes = []*models.Note{}
	}

	// Get all tools to show relationships
	tools, merr := h.registry.Tools.List()
	if merr != nil {
		return merr.Echo()
	}

	// Handle case where tools might be nil
	if tools == nil {
		tools = []*models.Tool{}
	}

	t := templates.Page(notes, tools)
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "Notes Page")
	}

	return nil
}

// HTMXDeleteNote deletes a note and unlinks it from all tools
func (h *Handler) HTMXDeleteNote(c echo.Context) error {
	slog.Info("Deleting note")

	user, merr := utils.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	idq, merr := utils.ParseQueryInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	noteID := models.NoteID(idq)

	// Delete the note
	merr = h.registry.Notes.Delete(noteID)
	if merr != nil {
		return merr.Echo()
	}

	// Create feed entry
	merr = h.registry.Feeds.Add("Notiz gelöscht", "Eine Notiz wurde gelöscht", user.TelegramID)
	if merr != nil {
		slog.Warn("Failed to create feed for note deletion", "error", merr)
	}

	// Trigger reload of notes sections
	utils.SetHXTrigger(c, env.HXGlobalTrigger)

	return nil
}

func (h *Handler) HTMXGetNotesGrid(c echo.Context) error {
	notes, merr := h.registry.Notes.List()
	if merr != nil {
		return merr.Echo()
	}

	tools, merr := h.registry.Tools.List()
	if merr != nil {
		return merr.Echo()
	}

	ng := templates.NotesGrid(notes, tools)
	err := ng.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "NotesGrid")
	}
	return nil
}
