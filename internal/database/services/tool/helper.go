package tool

import (
	"encoding/json"
	"fmt"

	"github.com/knackwurstking/pgpress/internal/database/dberror"
	feedmodels "github.com/knackwurstking/pgpress/internal/database/models/feed"
	notemodels "github.com/knackwurstking/pgpress/internal/database/models/note"
	pressmodels "github.com/knackwurstking/pgpress/internal/database/models/press"
	toolmodels "github.com/knackwurstking/pgpress/internal/database/models/tool"
	usermodels "github.com/knackwurstking/pgpress/internal/database/models/user"
	"github.com/knackwurstking/pgpress/internal/database/services/note"
	"github.com/knackwurstking/pgpress/internal/database/services/presscycle"
	"github.com/knackwurstking/pgpress/internal/logger"
)

type Helper struct {
	tools       *Service
	notes       *note.Service
	pressCycles *presscycle.Service
}

// NewHelper creates a new ToolsHelper instance.
func NewHelper(
	tools *Service,
	notes *note.Service,
	pressCycles *presscycle.Service,
) *Helper {
	return &Helper{
		tools:       tools,
		notes:       notes,
		pressCycles: pressCycles,
	}
}

// GetWithNotes retrieves a tool by its ID and loads its associated notes.
func (h *Helper) GetWithNotes(id int64) (*toolmodels.ToolWithNotes, error) {
	logger.DBToolsHelper().Debug(
		"Getting tools with notes, id: %d", id)

	// Get the tool
	tool, err := h.tools.Get(id)
	if err != nil {
		return nil, err
	}

	// Load notes
	notes, err := h.notes.GetByIDs(tool.LinkedNotes)
	if err != nil {
		return nil, dberror.WrapError(err, "failed to load attachments for trouble report")
	}

	return &toolmodels.ToolWithNotes{
		Tool:        tool,
		LoadedNotes: notes,
	}, nil
}

// ListWithNotes retrieves all tools and loads their associated notes.
func (h *Helper) ListWithNotes() ([]*toolmodels.ToolWithNotes, error) {
	logger.DBToolsHelper().Debug("Listing tools with notes")

	// Get all tools
	tools, err := h.tools.List()
	if err != nil {
		return nil, err
	}

	var result []*toolmodels.ToolWithNotes

	for _, tool := range tools {
		// Load notes for each tool
		notes, err := h.notes.GetByIDs(tool.LinkedNotes)
		if err != nil {
			return nil, dberror.WrapError(err,
				fmt.Sprintf("failed to load notes for tool %d", tool.ID))
		}

		result = append(result, &toolmodels.ToolWithNotes{
			Tool:        tool,
			LoadedNotes: notes,
		})
	}

	return result, nil
}

// AddWithNotes creates a new tool and its associated notes in a single transaction.
func (h *Helper) AddWithNotes(tool *toolmodels.Tool, user *usermodels.User, notes ...*notemodels.Note) (*toolmodels.ToolWithNotes, error) {
	logger.DBToolsHelper().Debug("Adding tool with notes")

	// First, add all notes and collect their IDs
	var noteIDs []int64
	for _, note := range notes {
		noteID, err := h.notes.Add(note, user)
		if err != nil {
			return nil, dberror.WrapError(err, "failed to add note")
		}
		noteIDs = append(noteIDs, noteID)
	}

	// Set the note IDs in the tool
	tool.LinkedNotes = noteIDs

	// Add the tool
	toolID, err := h.tools.Add(tool, user)
	if err != nil {
		if err == dberror.ErrAlreadyExists {
			return nil, err
		}
		return nil, dberror.WrapError(err, "failed to add tool")
	}

	// Set the tool ID
	tool.ID = toolID

	// Return the tool with loaded notes
	return &toolmodels.ToolWithNotes{
		Tool:        tool,
		LoadedNotes: notes,
	}, nil
}

// GetByPress returns all active tools for a specific press (0-5).
func (h *Helper) GetByPress(pressNumber *pressmodels.PressNumber) ([]*toolmodels.Tool, error) {
	if !pressmodels.IsValidPressNumber(pressNumber) {
		return nil, fmt.Errorf("invalid press number: %d (must be 0-5)", *pressNumber)
	}

	if pressNumber == nil {
		logger.DBTools().Info("Getting inactive tools")
	} else {
		logger.DBTools().Info("Getting active tools for press: %d", *pressNumber)
	}

	const query = `
		SELECT id, position, format, type, code, regenerating, press, notes, mods FROM tools WHERE press = $1 AND regenerating = 0;
	`
	rows, err := h.tools.db.Query(query, pressNumber)
	if err != nil {
		return nil, dberror.NewDatabaseError("select", "tools",
			fmt.Sprintf("failed to query tools for press %v", pressNumber), err)
	}
	defer rows.Close()

	var tools []*toolmodels.Tool

	for rows.Next() {
		tool, err := h.tools.scanTool(rows)
		if err != nil {
			return nil, dberror.WrapError(err, "failed to scan tool")
		}
		tools = append(tools, tool)
	}

	if err := rows.Err(); err != nil {
		return nil, dberror.NewDatabaseError("select", "tools",
			"error iterating over rows", err)
	}

	return tools, nil
}

// UpdateRegenerating updates only the regenerating field of a tool.
func (h *Helper) UpdateRegenerating(toolID int64, regenerating bool, user *usermodels.User) error {
	logger.DBTools().Info("Updating tool regenerating status: %d to %v", toolID, regenerating)

	// Get current tool to track changes
	tool, err := h.tools.Get(toolID)
	if err != nil {
		return fmt.Errorf("failed to get tool for regenerating status update: %w", err)
	}

	if tool.Regenerating == regenerating {
		return nil
	}

	// Update tool
	tool.Regenerating = regenerating

	// Update mods
	h.tools.updateMods(user, tool)

	// Marshal mods for database update
	modsBytes, err := json.Marshal(tool.Mods)
	if err != nil {
		return dberror.NewDatabaseError("update", "tools",
			"failed to marshal mods", err)
	}

	const query = `UPDATE tools SET regenerating = ?, mods = ? WHERE id = ?`
	_, err = h.tools.db.Exec(query, tool.Regenerating, modsBytes, tool.ID)
	if err != nil {
		return dberror.NewDatabaseError("update", "tools",
			fmt.Sprintf("failed to update regenerating for tool %d", tool.ID), err)
	}

	// Trigger feed update
	if h.tools.feeds != nil {
		h.tools.feeds.Add(feedmodels.New(
			feedmodels.TypeToolUpdate,
			&feedmodels.ToolUpdate{
				ID:         tool.ID,
				Tool:       tool.String(),
				ModifiedBy: user,
			},
		))
	}

	return nil
}

// UpdatePress updates only the press field of a tool.
func (h *Helper) UpdatePress(toolID int64, press *pressmodels.PressNumber, user *usermodels.User) error {
	logger.DBTools().Info("Updating tool press: %d", toolID)

	// Get current tool to track changes
	tool, err := h.tools.Get(toolID)
	if err != nil {
		return fmt.Errorf("failed to get tool for press update: %w", err)
	}

	if equalPressNumbers(tool.Press, press) {
		return nil
	}

	// Update tool
	if err := tool.SetPress(press); err != nil {
		return fmt.Errorf("failed to set press for tool %d: %w", toolID, err)
	}

	// Update mods
	h.tools.updateMods(user, tool)

	// Marshal mods for database update
	modsBytes, err := json.Marshal(tool.Mods)
	if err != nil {
		return dberror.NewDatabaseError("update", "tools",
			"failed to marshal mods", err)
	}

	const query = `UPDATE tools SET press = ?, mods = ? WHERE id = ?`
	_, err = h.tools.db.Exec(query, press, modsBytes, toolID)
	if err != nil {
		return dberror.NewDatabaseError("update", "tools",
			fmt.Sprintf("failed to update press for tool %d", toolID), err)
	}

	// Trigger feed update
	if h.tools.feeds != nil {
		tool.Press = press // Update press for correct display
		h.tools.feeds.Add(feedmodels.New(
			feedmodels.TypeToolUpdate,
			&feedmodels.ToolUpdate{
				ID:         toolID,
				Tool:       tool.String(),
				ModifiedBy: user,
			},
		))
	}

	return nil
}

// equalPressNumbers compares two press number pointers for equality
func equalPressNumbers(a, b *pressmodels.PressNumber) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
