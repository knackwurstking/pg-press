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
	notes, err := h.registry.Notes.List()
	if err != nil {
		return errors.Handler(err, "get notes from database")
	}

	// Handle case where notes might be nil
	if notes == nil {
		notes = []*models.Note{}
	}

	// Get all tools to show relationships
	tools, err := h.registry.Tools.List()
	if err != nil {
		return errors.Handler(err, "get tools from database")
	}

	// Handle case where tools might be nil
	if tools == nil {
		tools = []*models.Tool{}
	}

	page := templates.Page(notes, tools)
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.Handler(err, "render notes page")
	}

	return nil
}

// HTMXDeleteNote deletes a note and unlinks it from all tools
func (h *Handler) HTMXDeleteNote(c echo.Context) error {
	slog.Info("Deleting note")

	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}

	idq, err := utils.ParseQueryInt64(c, "id")
	if err != nil {
		return errors.BadRequest(err, "parse note ID")
	}
	noteID := models.NoteID(idq)

	// Delete the note
	if err := h.registry.Notes.Delete(noteID); err != nil {
		return errors.Handler(err, "delete note")
	}

	// Create feed entry
	if _, err := h.registry.Feeds.AddSimple("Notiz gelöscht", "Eine Notiz wurde gelöscht", user.TelegramID); err != nil {
		slog.Warn("Failed to create feed for note deletion", "error", err)
	}

	// Trigger reload of notes sections
	utils.SetHXTrigger(c, env.HXGlobalTrigger)

	return nil
}

func (h *Handler) HTMXGetNotesGrid(c echo.Context) error {
	notes, err := h.registry.Notes.List()
	if err != nil {
		return errors.Handler(err, "list notes")
	}

	tools, err := h.registry.Tools.List()
	if err != nil {
		return errors.Handler(err, "list tools")
	}

	ng := templates.NotesGrid(notes, tools)
	if err := ng.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.Handler(err, "render notes grid")
	}
	return nil
}
