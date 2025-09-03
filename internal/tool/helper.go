package tool

import (
	"encoding/json"
	"fmt"

	"github.com/knackwurstking/pgpress/internal/dberror"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/models"
	"github.com/knackwurstking/pgpress/internal/note"
	"github.com/knackwurstking/pgpress/internal/presscycle"
)

// ToolWithNotes represents a tool with its related notes loaded.
type ToolWithNotes struct {
	*models.Tool
	LoadedNotes []*models.Note `json:"loaded_notes"`
}

type ToolsHelper struct {
	tools       *Service
	notes       *note.Service
	pressCycles *presscycle.Service
}

// NewToolsHelper creates a new ToolsHelper instance.
func NewToolsHelper(
	tools *Service,
	notes *note.Service,
	pressCycles *presscycle.Service,
) *ToolsHelper {
	return &ToolsHelper{
		tools:       tools,
		notes:       notes,
		pressCycles: pressCycles,
	}
}

// GetWithNotes retrieves a tool by its ID and loads its associated notes.
func (th *ToolsHelper) GetWithNotes(id int64) (*ToolWithNotes, error) {
	logger.DBToolsHelper().Debug(
		"Getting tools with notes, id: %d", id)

	// Get the tool
	tool, err := th.tools.Get(id)
	if err != nil {
		return nil, err
	}

	// Load notes
	notes, err := th.notes.GetByIDs(tool.LinkedNotes)
	if err != nil {
		return nil, dberror.WrapError(err, "failed to load attachments for trouble report")
	}

	return &ToolWithNotes{
		Tool:        tool,
		LoadedNotes: notes,
	}, nil
}

// ListWithNotes retrieves all tools and loads their associated notes.
func (th *ToolsHelper) ListWithNotes() ([]*ToolWithNotes, error) {
	logger.DBToolsHelper().Debug("Listing tools with notes")

	// Get all tools
	tools, err := th.tools.List()
	if err != nil {
		return nil, err
	}

	var result []*ToolWithNotes

	for _, tool := range tools {
		// Load notes for each tool
		notes, err := th.notes.GetByIDs(tool.LinkedNotes)
		if err != nil {
			return nil, dberror.WrapError(err,
				fmt.Sprintf("failed to load notes for tool %d", tool.ID))
		}

		result = append(result, &ToolWithNotes{
			Tool:        tool,
			LoadedNotes: notes,
		})
	}

	return result, nil
}

// AddWithNotes creates a new tool and its associated notes in a single transaction.
func (th *ToolsHelper) AddWithNotes(tool *models.Tool, user *models.User, notes ...*models.Note) (*ToolWithNotes, error) {
	logger.DBToolsHelper().Debug("Adding tool with notes")

	// First, add all notes and collect their IDs
	var noteIDs []int64
	for _, note := range notes {
		noteID, err := th.notes.Add(note, user)
		if err != nil {
			return nil, dberror.WrapError(err, "failed to add note")
		}
		noteIDs = append(noteIDs, noteID)
	}

	// Set the note IDs in the tool
	tool.LinkedNotes = noteIDs

	// Add the tool
	toolID, err := th.tools.Add(tool, user)
	if err != nil {
		return nil, dberror.WrapError(err, "failed to add tool")
	}

	// Set the tool ID
	tool.ID = toolID

	// Return the tool with loaded notes
	return &ToolWithNotes{
		Tool:        tool,
		LoadedNotes: notes,
	}, nil
}

// GetByPress returns all active tools for a specific press (0-5).
func (th *ToolsHelper) GetByPress(pressNumber *models.PressNumber) ([]*models.Tool, error) {
	if pressNumber != nil && !(*pressNumber).IsValid() {
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
	rows, err := th.tools.db.Query(query, pressNumber)
	if err != nil {
		return nil, dberror.NewDatabaseError("select", "tools",
			fmt.Sprintf("failed to query tools for press %v", pressNumber), err)
	}
	defer rows.Close()

	var tools []*models.Tool

	for rows.Next() {
		tool, err := th.tools.scanTool(rows)
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
func (th *ToolsHelper) UpdateRegenerating(toolID int64, regenerating bool, user *models.User) error {
	logger.DBTools().Info("Updating tool regenerating status: %d to %v", toolID, regenerating)

	// Get current tool to track changes
	tool, err := th.tools.Get(toolID)
	if err != nil {
		return fmt.Errorf("failed to get tool for regenerating status update: %w", err)
	}

	if tool.Regenerating == regenerating {
		return nil
	}

	// Update tool
	tool.Regenerating = regenerating

	// Update mods
	th.tools.updateMods(user, tool)

	// Marshal mods for database update
	modsBytes, err := json.Marshal(tool.Mods)
	if err != nil {
		return dberror.NewDatabaseError("update", "tools",
			"failed to marshal mods", err)
	}

	const query = `UPDATE tools SET regenerating = ?, mods = ? WHERE id = ?`
	_, err = th.tools.db.Exec(query, tool.Regenerating, modsBytes, tool.ID)
	if err != nil {
		return dberror.NewDatabaseError("update", "tools",
			fmt.Sprintf("failed to update regenerating for tool %d", tool.ID), err)
	}

	// Trigger feed update
	if th.tools.feeds != nil {
		th.tools.feeds.Add(models.NewFeed(
			models.FeedTypeToolUpdate,
			&models.FeedToolUpdate{
				ID:         tool.ID,
				Tool:       tool.String(),
				ModifiedBy: user,
			},
		))
	}

	return nil
}

// UpdatePress updates only the press field of a tool.
func (th *ToolsHelper) UpdatePress(toolID int64, press *models.PressNumber, user *models.User) error {
	logger.DBTools().Info("Updating tool press: %d", toolID)

	// Get current tool to track changes
	tool, err := th.tools.Get(toolID)
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
	th.tools.updateMods(user, tool)

	// Marshal mods for database update
	modsBytes, err := json.Marshal(tool.Mods)
	if err != nil {
		return dberror.NewDatabaseError("update", "tools",
			"failed to marshal mods", err)
	}

	const query = `UPDATE tools SET press = ?, mods = ? WHERE id = ?`
	_, err = th.tools.db.Exec(query, press, modsBytes, toolID)
	if err != nil {
		return dberror.NewDatabaseError("update", "tools",
			fmt.Sprintf("failed to update press for tool %d", toolID), err)
	}

	// Trigger feed update
	if th.tools.feeds != nil {
		tool.Press = press // Update press for correct display
		th.tools.feeds.Add(models.NewFeed(
			models.FeedTypeToolUpdate,
			&models.FeedToolUpdate{
				ID:         toolID,
				Tool:       tool.String(),
				ModifiedBy: user,
			},
		))
	}

	return nil
}

// equalPressNumbers compares two press number pointers for equality
func equalPressNumbers(a, b *models.PressNumber) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
