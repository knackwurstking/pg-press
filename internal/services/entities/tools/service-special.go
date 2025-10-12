package tools

import (
	"fmt"

	"github.com/knackwurstking/pgpress/internal/services/shared/validation"
	"github.com/knackwurstking/pgpress/pkg/models"
)

func (t *Service) AddWithNotes(tool *models.Tool, user *models.User, notes ...*models.Note) (*models.ToolWithNotes, error) {
	if err := ValidateTool(tool); err != nil {
		return nil, err
	}

	if err := validation.ValidateNotNil(user, "user"); err != nil {
		return nil, err
	}

	t.Log.Debug("Adding tool with notes by %s: notes_count: %d",
		user.String(), len(notes))

	var createdNotes []*models.Note
	for _, note := range notes {
		noteID, err := t.notes.Add(note)
		if err != nil {
			// Cleanup previously created notes on failure
			for _, cn := range createdNotes {
				if deleteErr := t.notes.Delete(cn.ID, user); deleteErr != nil {
					t.Log.Error("Failed to cleanup note %d: %v", cn.ID, deleteErr)
				}
			}
			return nil, fmt.Errorf("failed to create note: %v", err)
		}
		note.ID = noteID
		createdNotes = append(createdNotes, note)
	}

	toolID, err := t.Add(tool, user)
	if err != nil {
		return nil, err
	}

	tool.ID = toolID
	return &models.ToolWithNotes{Tool: tool, LoadedNotes: createdNotes}, nil
}

func (t *Service) GetActiveToolsForPress(pressNumber models.PressNumber) []*models.Tool {
	t.Log.Debug("Getting active tools for press: %d", pressNumber)

	query := fmt.Sprintf(`
		SELECT %s
		FROM tools WHERE regenerating = 0 AND is_dead = 0 AND press = ?
	`, ToolQuerySelect)
	rows, err := t.DB.Query(query, pressNumber)
	if err != nil {
		t.Log.Error("Failed to query active tools: %v", err)
		return nil
	}
	defer rows.Close()

	tools, err := ScanToolsFromRows(rows)
	if err != nil {
		t.Log.Error("Failed to scan active tools: %v", err)
		return nil
	}

	return tools
}

func (t *Service) GetByPress(pressNumber *models.PressNumber) ([]*models.Tool, error) {
	if pressNumber != nil && !models.IsValidPressNumber(pressNumber) {
		return nil, fmt.Errorf("invalid press number: %d (must be 0-5)", *pressNumber)
	}

	t.Log.Debug("Getting tools by press: %v", pressNumber)

	query := fmt.Sprintf(`
		SELECT %s
		FROM tools WHERE press = ? AND regenerating = 0 AND is_dead = 0
	`, ToolQuerySelect)

	rows, err := t.DB.Query(query, pressNumber)
	if err != nil {
		return nil, t.HandleSelectError(err, "tools")
	}
	defer rows.Close()

	tools, err := ScanToolsFromRows(rows)
	if err != nil {
		return nil, err
	}

	return tools, nil
}

func (t *Service) GetPressUtilization() ([]models.PressUtilization, error) {
	t.Log.Debug("Getting press utilization")

	var utilization []models.PressUtilization

	// Valid press numbers: 0, 2, 3, 4, 5
	validPresses := []models.PressNumber{0, 2, 3, 4, 5}

	for _, pressNum := range validPresses {
		tools := t.GetActiveToolsForPress(pressNum)
		count := len(tools)

		utilization = append(utilization, models.PressUtilization{
			PressNumber: pressNum,
			Tools:       tools,
			Count:       count,
			Available:   count == 0,
		})
	}

	return utilization, nil
}

func (t *Service) GetWithNotes(id int64) (*models.ToolWithNotes, error) {
	if err := validation.ValidateID(id, "tool"); err != nil {
		return nil, err
	}

	t.Log.Debug("Getting tool with notes: %d", id)

	tool, err := t.Get(id)
	if err != nil {
		return nil, err
	}

	notes, err := t.notes.GetByTool(id)
	if err != nil {
		return nil, fmt.Errorf("failed to load notes for tool")
	}

	return &models.ToolWithNotes{Tool: tool, LoadedNotes: notes}, nil
}

func (t *Service) ListToolsNotDead() ([]*models.Tool, error) {
	t.Log.Debug("Listing active tools")

	query := fmt.Sprintf(`
		SELECT
			%s
		FROM
			tools
		WHERE
			is_dead = 0
		ORDER BY format ASC, code ASC
	`, ToolQuerySelect)

	rows, err := t.DB.Query(query)
	if err != nil {
		return nil, t.HandleSelectError(err, "tools")
	}
	defer rows.Close()

	tools, err := ScanToolsFromRows(rows)
	if err != nil {
		return nil, err
	}

	return tools, nil
}

func (t *Service) ListDeadTools() ([]*models.Tool, error) {
	t.Log.Debug("Listing dead tools")

	query := fmt.Sprintf(`
		SELECT
			%s
		FROM
			tools
		WHERE
			is_dead = 1
		ORDER BY format ASC, code ASC
	`, ToolQuerySelect)

	rows, err := t.DB.Query(query)
	if err != nil {
		return nil, t.HandleSelectError(err, "tools")
	}
	defer rows.Close()

	tools, err := ScanToolsFromRows(rows)
	if err != nil {
		return nil, err
	}

	return tools, nil
}

func (t *Service) ListWithNotes() ([]*models.ToolWithNotes, error) {
	t.Log.Debug("Listing tools with notes")

	tools, err := t.List()
	if err != nil {
		return nil, err
	}

	var result []*models.ToolWithNotes
	for _, tool := range tools {
		notes, err := t.notes.GetByTool(tool.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load notes for tool %d", tool.ID)
		}
		result = append(result, &models.ToolWithNotes{Tool: tool, LoadedNotes: notes})
	}

	return result, nil
}

func (t *Service) Update(tool *models.Tool, user *models.User) error {
	if err := ValidateTool(tool); err != nil {
		return err
	}

	if err := validation.ValidateID(tool.ID, "tool"); err != nil {
		return err
	}

	if err := validation.ValidateNotNil(user, "user"); err != nil {
		return err
	}

	t.Log.Debug("Updating tool by %s: id: %d", user.String(), tool.ID)

	if err := t.validateToolUniqueness(tool, tool.ID); err != nil {
		return err
	}

	formatBytes, err := marshalFormat(tool.Format)
	if err != nil {
		return err
	}

	query := fmt.Sprintf(`
		UPDATE
			tools
		SET
			%s
		WHERE
			id = ?
	`, ToolQueryUpdate)

	result, err := t.DB.Exec(query,
		tool.Position,
		formatBytes,
		tool.Type,
		tool.Code,
		tool.Regenerating,
		tool.IsDead,
		tool.Press,
		tool.Binding,
		tool.ID,
	)
	if err != nil {
		return t.HandleUpdateError(err, "tools")
	}

	if err := t.CheckRowsAffected(result, "tool", tool.ID); err != nil {
		return err
	}

	return nil
}

func (t *Service) UpdatePress(toolID int64, press *models.PressNumber, user *models.User) error {
	if err := validation.ValidateID(toolID, "tool"); err != nil {
		return err
	}

	if err := validation.ValidateNotNil(user, "user"); err != nil {
		return err
	}

	t.Log.Debug("Updating tool press by %s: id: %d, press: %v", user.String(), toolID, press)

	tool, err := t.Get(toolID)
	if err != nil {
		return fmt.Errorf("failed to get tool for press update: %v", err)
	}

	if models.IsEqualPressNumbers(tool.Press, press) {
		return nil
	}

	if err := tool.SetPress(press); err != nil {
		return fmt.Errorf("failed to set press for tool %d: %v", toolID, err)
	}

	query := `UPDATE tools SET press = ? WHERE id = ?`
	result, err := t.DB.Exec(query, press, toolID)
	if err != nil {
		return t.HandleUpdateError(err, "tools")
	}

	if err := t.CheckRowsAffected(result, "tool", toolID); err != nil {
		return err
	}

	return nil
}

func (t *Service) UpdateRegenerating(toolID int64, regenerating bool, user *models.User) error {
	if err := validation.ValidateID(toolID, "tool"); err != nil {
		return err
	}

	if err := validation.ValidateNotNil(user, "user"); err != nil {
		return err
	}

	// Get the current tool to check if the regeneration status is actually changing
	currentTool, err := t.Get(toolID)
	if err != nil {
		return fmt.Errorf("failed to get current tool state: %v", err)
	}

	if currentTool.Regenerating == regenerating {
		return nil
	}

	t.Log.Debug("Updating tool regenerating status by %s: id: %d, regenerating: %v", user.String(), toolID, regenerating)

	query := `UPDATE tools SET regenerating = ? WHERE id = ?`
	result, err := t.DB.Exec(query, regenerating, toolID)
	if err != nil {
		return t.HandleUpdateError(err, "tools")
	}

	if err := t.CheckRowsAffected(result, "tool", toolID); err != nil {
		return err
	}

	return nil
}

func (t *Service) MarkAsDead(toolID int64, user *models.User) error {
	if err := validation.ValidateID(toolID, "tool"); err != nil {
		return err
	}

	if err := validation.ValidateNotNil(user, "user"); err != nil {
		return err
	}

	t.Log.Debug("Marking tool as dead by %s: id: %d", user.String(), toolID)

	// Mark as dead and clear press assignment
	query := `UPDATE tools SET is_dead = 1, press = NULL WHERE id = ?`
	result, err := t.DB.Exec(query, toolID)
	if err != nil {
		return t.HandleUpdateError(err, "tools")
	}

	if err := t.CheckRowsAffected(result, "tool", toolID); err != nil {
		return err
	}

	return nil
}

func (t *Service) ReviveTool(toolID int64, user *models.User) error {
	if err := validation.ValidateID(toolID, "tool"); err != nil {
		return err
	}

	if err := validation.ValidateNotNil(user, "user"); err != nil {
		return err
	}

	t.Log.Debug("Reviving dead tool by %s: id: %d", user.String(), toolID)

	// Mark as alive (not dead)
	query := `UPDATE tools SET is_dead = 0 WHERE id = ?`
	result, err := t.DB.Exec(query, toolID)
	if err != nil {
		return t.HandleUpdateError(err, "tools")
	}

	if err := t.CheckRowsAffected(result, "tool", toolID); err != nil {
		return err
	}

	return nil
}
