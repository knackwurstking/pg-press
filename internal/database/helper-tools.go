package database

import (
	"fmt"

	"github.com/knackwurstking/pgpress/internal/logger"
)

type ToolWithNotes struct {
	*Tool
	LoadedNotes []*Note `json:"loaded_notes"`
}

type ToolsHelper struct {
	tools *Tools
	notes *Notes
}

func NewToolsHelper(
	tools *Tools,
	notes *Notes,
) *ToolsHelper {
	return &ToolsHelper{
		tools: tools,
		notes: notes,
	}
}

func (th *ToolsHelper) GetWithNotes(id int64) (*ToolWithNotes, error) {
	logger.Tools().Debug(
		"Getting tools with notes, id: %d", id)

	// Get the trouble report
	tool, err := th.tools.Get(id)
	if err != nil {
		return nil, err
	}

	// Load attachments
	notes, err := th.notes.GetByIDs(tool.LinkedNotes)
	if err != nil {
		return nil, WrapError(err, "failed to load attachments for trouble report")
	}

	return &ToolWithNotes{
		Tool:        tool,
		LoadedNotes: notes,
	}, nil
}

func (th *ToolsHelper) ListWithNotes() ([]*ToolWithNotes, error) {
	logger.Tools().Debug("Listing tools with notes")

	// Get all trouble reports
	tools, err := th.tools.List()
	if err != nil {
		return nil, err
	}

	var result []*ToolWithNotes

	for _, tool := range tools {
		// Load attachments for each report
		notes, err := th.notes.GetByIDs(tool.LinkedNotes)
		if err != nil {
			return nil, WrapError(err,
				fmt.Sprintf("failed to load notes for tool %d", tool.ID))
		}

		result = append(result, &ToolWithNotes{
			Tool:        tool,
			LoadedNotes: notes,
		})
	}

	return result, nil
}

func (th *ToolsHelper) AddWithNotes(tool *Tool, user *User, notes ...*Note) (*ToolWithNotes, error) {
	logger.Tools().Debug("Adding tool with notes")

	// Ensure tool has initial mod entry if it doesn't have one
	if len(tool.Mods) == 0 {
		initialMod := NewModified(user, ToolMod{
			Position:    tool.Position,
			Format:      tool.Format,
			Type:        tool.Type,
			Code:        tool.Code,
			Status:      tool.Status,
			Press:       tool.Press,
			LinkedNotes: tool.LinkedNotes,
		})
		tool.Mods = []*Modified[ToolMod]{initialMod}
	}

	// First, add all notes and collect their IDs
	var noteIDs []int64
	for _, note := range notes {
		noteID, err := th.notes.Add(note)
		if err != nil {
			return nil, WrapError(err, "failed to add note")
		}
		noteIDs = append(noteIDs, noteID)
	}

	// Set the note IDs in the tool
	tool.LinkedNotes = noteIDs

	// Add the tool
	toolID, err := th.tools.Add(tool, user)
	if err != nil {
		return nil, WrapError(err, "failed to add tool")
	}

	// Set the tool ID
	tool.ID = toolID

	// Return the tool with loaded notes
	return &ToolWithNotes{
		Tool:        tool,
		LoadedNotes: notes,
	}, nil
}
