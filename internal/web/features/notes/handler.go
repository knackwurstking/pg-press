package notes

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/web/features/notes/templates"
	"github.com/knackwurstking/pgpress/internal/web/shared/handlers"
	"github.com/knackwurstking/pgpress/pkg/logger"
	"github.com/knackwurstking/pgpress/pkg/models"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	*handlers.BaseHandler
}

func NewHandler(db *database.DB) *Handler {
	return &Handler{
		BaseHandler: handlers.NewBaseHandler(
			db,
			logger.NewComponentLogger("Notes"),
		),
	}
}

// HTMXGetEditNoteDialog renders the edit note dialog
func (h *Handler) HTMXGetEditNoteDialog(c echo.Context) error {
	props := &templates.DialogEditNoteProps{
		Note:         &models.Note{}, // Default empty note for creation
		LinkToTables: []string{},
	}

	// Parse linked tables from query parameter
	if linkToTables := c.QueryParam("link_to_tables"); linkToTables != "" {
		props.LinkToTables = strings.Split(linkToTables, ",")
	}

	// Check if we're editing an existing note
	if noteID, _ := h.ParseInt64Query(c, "id"); noteID > 0 {
		h.LogDebug("Opening edit dialog for note %d", noteID)

		note, err := h.DB.Notes.Get(noteID)
		if err != nil {
			return h.HandleError(c, err, "failed to get note from database")
		}
		props.Note = note
	} else {
		h.LogDebug("Opening create dialog for new note")
	}

	dialog := templates.DialogEditNote(*props)
	if err := dialog.Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c, "failed to render edit note dialog: "+err.Error())
	}

	return nil
}

// HTMXPostEditNoteDialog creates a new note
func (h *Handler) HTMXPostEditNoteDialog(c echo.Context) error {
	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed to get user from context")
	}

	h.LogDebug("User %s creating new note", user.Name)

	note, linkedTables, err := h.parseNoteFromForm(c)
	if err != nil {
		return h.RenderBadRequest(c, "failed to parse note form data: "+err.Error())
	}

	// Create the note
	noteID, err := h.DB.Notes.Add(note)
	if err != nil {
		return h.HandleError(c, err, "failed to create note")
	}

	h.LogInfo("User %s created note %d", user.Name, noteID)

	// Link tables to this note
	if err := h.linkTablesToNote(noteID, linkedTables); err != nil {
		h.LogError("Failed to link tables to note %d: %v", noteID, err)
		// Don't fail the entire operation if linking fails
	}

	// Create feed entry
	title := "Neue Notiz erstellt"
	content := fmt.Sprintf("Eine neue Notiz wurde erstellt: %s", note.Content)

	// Add linked tables info if any
	if len(linkedTables) > 0 {
		content += fmt.Sprintf("\nVerknüpfte Tabellen: %s", strings.Join(linkedTables, ", "))
	}

	feed := models.NewFeed(title, content, user.TelegramID)
	if err := h.DB.Feeds.Add(feed); err != nil {
		h.LogError("Failed to create feed for cycle creation: %v", err)
	}

	// Trigger reload of notes sections
	c.Response().Header().Set("HX-Trigger", "noteCreated, pageLoaded")
	return nil
}

// HTMXPutEditNoteDialog updates an existing note
func (h *Handler) HTMXPutEditNoteDialog(c echo.Context) error {
	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed to get user from context")
	}

	noteID, err := h.ParseInt64Query(c, "id")
	if err != nil {
		return h.RenderBadRequest(c, "failed to parse note ID: "+err.Error())
	}

	h.LogDebug("User %s updating note %d", user.Name, noteID)

	note, linkedTables, err := h.parseNoteFromForm(c)
	if err != nil {
		return h.RenderBadRequest(c, "failed to parse note form data: "+err.Error())
	}

	// Set the ID for update
	note.ID = noteID

	// Update the note
	if err := h.DB.Notes.Update(note); err != nil {
		return h.HandleError(c, err, "failed to update note")
	}

	h.LogInfo("User %s updated note %d", user.Name, noteID)

	// Update table links for this note
	if err := h.updateTableLinksForNote(noteID, linkedTables); err != nil {
		h.LogError("Failed to update table links for note %d: %v", noteID, err)
		// Don't fail the entire operation if linking fails
	}

	// Create feed entry
	title := "Notiz aktualisiert"
	content := fmt.Sprintf("Eine Notiz wurde aktualisiert: %s", note.Content)

	// Add linked tables info if any
	if len(linkedTables) > 0 {
		content += fmt.Sprintf("\nVerknüpfte Tabellen: %s", strings.Join(linkedTables, ", "))
	}

	feed := models.NewFeed(title, content, user.TelegramID)
	if err := h.DB.Feeds.Add(feed); err != nil {
		h.LogError("Failed to create feed for cycle creation: %v", err)
	}

	// Trigger reload of notes sections
	c.Response().Header().Set("HX-Trigger", "noteUpdated, pageLoaded")
	return nil
}

// HTMXDeleteNote deletes a note and unlinks it from all tools
func (h *Handler) HTMXDeleteNote(c echo.Context) error {
	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed to get user from context")
	}

	noteID, err := h.ParseInt64Query(c, "id")
	if err != nil {
		return h.RenderBadRequest(c, "failed to parse note ID: "+err.Error())
	}

	h.LogDebug("User %s deleting note %d", user.Name, noteID)

	// First unlink the note from all tools that reference it
	tools, err := h.DB.Tools.List()
	if err != nil {
		h.LogError("Failed to get tools list for note cleanup: %v", err)
		// Continue with deletion even if we can't clean up tool references
	} else {
		for _, tool := range tools {
			for _, linkedNoteID := range tool.LinkedNotes {
				if linkedNoteID == noteID {
					if err := h.unlinkNoteFromTool(noteID, tool.ID); err != nil {
						h.LogError("Failed to unlink note %d from tool %d: %v", noteID, tool.ID, err)
					}
					break
				}
			}
		}
	}

	// Delete the note
	if err := h.DB.Notes.Delete(noteID, user); err != nil {
		return h.HandleError(c, err, "failed to delete note")
	}

	h.LogInfo("User %s deleted note %d", user.Name, noteID)

	// Create feed entry
	feed := models.NewFeed("Notiz gelöscht", "Eine Notiz wurde gelöscht", user.TelegramID)
	if err := h.DB.Feeds.Add(feed); err != nil {
		h.LogError("Failed to create feed for note deletion: %v", err)
	}

	// Trigger reload of notes sections
	c.Response().Header().Set("HX-Trigger", "noteDeleted, pageLoaded")
	return nil
}

// GetNotesPage serves the main notes page
func (h *Handler) GetNotesPage(c echo.Context) error {
	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed to get user from context")
	}

	// Get all notes
	notes, err := h.DB.Notes.List()
	if err != nil {
		return h.HandleError(c, err, "failed to get notes from database")
	}

	// Get all tools to show relationships
	tools, err := h.DB.Tools.List()
	if err != nil {
		return h.HandleError(c, err, "failed to get tools from database")
	}

	page := templates.Page(&templates.PageProps{
		User:  user,
		Notes: notes,
		Tools: tools,
	})

	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c, "failed to render notes page: "+err.Error())
	}

	return nil
}

// parseNoteFromForm parses note data from form submission
func (h *Handler) parseNoteFromForm(c echo.Context) (note *models.Note, linkedTables []string, err error) {
	note = &models.Note{}

	// Parse level
	levelStr := c.FormValue("level")
	if levelStr == "" {
		return nil, nil, fmt.Errorf("level is required")
	}

	levelInt, err := strconv.Atoi(levelStr)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid level format: %v", err)
	}

	// Validate level is within valid range (0=INFO, 1=ATTENTION, 2=BROKEN)
	if levelInt < 0 || levelInt > 2 {
		return nil, nil, fmt.Errorf("invalid level value: %d (must be 0, 1, or 2)", levelInt)
	}

	note.Level = models.Level(levelInt)

	// Parse content
	note.Content = strings.TrimSpace(c.FormValue("content"))
	if note.Content == "" {
		return nil, nil, fmt.Errorf("content is required")
	}

	// Handle linked_tables
	linkedTables = c.Request().Form["linked_tables"]

	return note, linkedTables, nil
}

// linkTablesToNote links a note to the specified tables
func (h *Handler) linkTablesToNote(noteID int64, linkedTables []string) error {
	for _, tableRef := range linkedTables {
		if err := h.linkNoteToTable(noteID, tableRef); err != nil {
			return fmt.Errorf("failed to link note to table %s: %v", tableRef, err)
		}
	}
	return nil
}

// updateTableLinksForNote updates the table links for a note (unlinks old, links new)
func (h *Handler) updateTableLinksForNote(noteID int64, newLinkedTables []string) error {
	// For updates, we need to handle unlinking from tools that no longer have this note
	// First, get all tools and check which ones currently have this note
	tools, err := h.DB.Tools.List()
	if err != nil {
		return fmt.Errorf("failed to get tools list: %v", err)
	}

	// Unlink from tools that currently have this note but shouldn't anymore
	for _, tool := range tools {
		hasNote := false
		for _, linkedNoteID := range tool.LinkedNotes {
			if linkedNoteID == noteID {
				hasNote = true
				break
			}
		}

		if hasNote {
			shouldKeepNote := false
			for _, tableRef := range newLinkedTables {
				if tableRef == fmt.Sprintf("tool_%d", tool.ID) {
					shouldKeepNote = true
					break
				}
				// Check if this tool is on a press that should have this note
				if tool.Press != nil {
					if tableRef == fmt.Sprintf("press_%d", *tool.Press) {
						shouldKeepNote = true
						break
					}
				}
			}

			if !shouldKeepNote {
				if err := h.unlinkNoteFromTool(noteID, tool.ID); err != nil {
					h.LogError("Failed to unlink note %d from tool %d: %v", noteID, tool.ID, err)
				}
			}
		}
	}

	// Now link to new tables
	return h.linkTablesToNote(noteID, newLinkedTables)
}

// linkNoteToTable links a note to a specific table (tool_ID or press_ID)
func (h *Handler) linkNoteToTable(noteID int64, tableRef string) error {
	parts := strings.Split(tableRef, "_")
	if len(parts) != 2 {
		return fmt.Errorf("invalid table reference format: %s", tableRef)
	}

	tableType := parts[0]
	tableIDStr := parts[1]

	switch tableType {
	case "tool":
		toolID, err := strconv.ParseInt(tableIDStr, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid tool ID in table reference %s: %v", tableRef, err)
		}
		return h.linkNoteToTool(noteID, toolID)

	case "press":
		pressNum, err := strconv.ParseInt(tableIDStr, 10, 8)
		if err != nil {
			return fmt.Errorf("invalid press number in table reference %s: %v", tableRef, err)
		}
		return h.linkNoteToPress(noteID, int8(pressNum))

	default:
		return fmt.Errorf("unsupported table type: %s", tableType)
	}
}

// linkNoteToTool adds a note ID to a tool's LinkedNotes
func (h *Handler) linkNoteToTool(noteID, toolID int64) error {
	tool, err := h.DB.Tools.Get(toolID)
	if err != nil {
		return fmt.Errorf("failed to get tool %d: %v", toolID, err)
	}

	// Check if note is already linked
	for _, linkedNoteID := range tool.LinkedNotes {
		if linkedNoteID == noteID {
			h.LogDebug("Note %d already linked to tool %d", noteID, toolID)
			return nil
		}
	}

	// Add note ID to tool's LinkedNotes
	tool.LinkedNotes = append(tool.LinkedNotes, noteID)

	// Create a dummy user for the update (we need this for the API but it's not used for note linking)
	// TODO: This should be improved to get the actual user from context
	dummyUser := &models.User{TelegramID: 1, Name: "system"}

	if err := h.DB.Tools.Update(tool, dummyUser); err != nil {
		return fmt.Errorf("failed to update tool %d with note link: %v", toolID, err)
	}

	h.LogDebug("Successfully linked note %d to tool %d", noteID, toolID)
	return nil
}

// linkNoteToPress links a note to all tools on a specific press
func (h *Handler) linkNoteToPress(noteID int64, pressNum int8) error {
	press := models.PressNumber(pressNum)
	if !models.IsValidPressNumber(&press) {
		return fmt.Errorf("invalid press number: %d", pressNum)
	}

	// Get all tools on this press
	tools, err := h.DB.Tools.GetByPress(&press)
	if err != nil {
		return fmt.Errorf("failed to get tools for press %d: %v", pressNum, err)
	}

	// Link note to each tool on the press
	for _, tool := range tools {
		if err := h.linkNoteToTool(noteID, tool.ID); err != nil {
			h.LogError("Failed to link note %d to tool %d on press %d: %v", noteID, tool.ID, pressNum, err)
			// Continue with other tools even if one fails
		}
	}

	h.LogDebug("Successfully linked note %d to press %d (%d tools)", noteID, pressNum, len(tools))
	return nil
}

// unlinkNoteFromTool removes a note ID from a tool's LinkedNotes
func (h *Handler) unlinkNoteFromTool(noteID, toolID int64) error {
	tool, err := h.DB.Tools.Get(toolID)
	if err != nil {
		return fmt.Errorf("failed to get tool %d: %v", toolID, err)
	}

	// Remove note ID from tool's LinkedNotes
	var newLinkedNotes []int64
	for _, linkedNoteID := range tool.LinkedNotes {
		if linkedNoteID != noteID {
			newLinkedNotes = append(newLinkedNotes, linkedNoteID)
		}
	}

	tool.LinkedNotes = newLinkedNotes

	// Create a dummy user for the update
	dummyUser := &models.User{TelegramID: 1, Name: "system"}

	if err := h.DB.Tools.Update(tool, dummyUser); err != nil {
		return fmt.Errorf("failed to update tool %d to unlink note: %v", toolID, err)
	}

	h.LogDebug("Successfully unlinked note %d from tool %d", noteID, toolID)
	return nil
}
