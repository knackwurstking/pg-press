package tools

import (
	"database/sql"
	"fmt"

	"github.com/knackwurstking/pgpress/internal/services/shared/validation"
	"github.com/knackwurstking/pgpress/pkg/models"
)

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

	query := `UPDATE tools SET press = ? WHERE id = ?`
	result, err := t.DB.Exec(query, press, toolID)
	if err != nil {
		return t.HandleUpdateError(err, "tools")
	}

	// Handle binding
	if tool.Binding != nil {
		// Update press for bound tool
		query = `UPDATE tools SET press = ? WHERE id = ?`
		result, err = t.DB.Exec(query, press, *tool.Binding)
		if err != nil {
			return t.HandleUpdateError(err, "tools")
		}
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

func (s *Service) Bind(cassette, target int64) error {
	if err := s.validateBindingTools(cassette, target); err != nil {
		return err
	}

	// Get press from the target tool
	var press *models.PressNumber
	if t, err := s.Get(target); err != nil {
		return err
	} else {
		press = t.Press
	}

	// Now the actual binding logic
	queries := []string{
		// Bindings
		`UPDATE tools SET binding = :target WHERE id = :cassette;`,
		`UPDATE tools SET binding = :cassette WHERE id = :target;`,
		// Unbind press
		`UPDATE tools SET press = NULL WHERE press = :press AND position = "cassette top";`,
		// Bind press from target to cassette
		`UPDATE tools SET press = :press WHERE id = :cassette;`,
	}

	for _, query := range queries {
		if _, err := s.DB.Exec(query,
			sql.Named("target", target),
			sql.Named("cassette", cassette),
			sql.Named("press", press),
		); err != nil {
			return s.HandleUpdateError(err, "tools")
		}
	}

	return nil
}

func (s *Service) UnBind(toolID int64) error {
	if err := validation.ValidateID(toolID, "tool"); err != nil {
		return err
	}

	// Get the tool to find its binding
	tool, err := s.Get(toolID)
	if err != nil {
		return err
	}

	// If no binding exists, nothing to unbind
	if tool.Binding == nil {
		return nil
	}

	// Clear the binding by setting binding to NULL for both tools
	query := `
		UPDATE
			tools
		SET
			binding = NULL
		WHERE
			id = :toolID OR id = :binding;
	`
	if _, err := s.DB.Exec(query,
		sql.Named("toolID", toolID),
		sql.Named("binding", *tool.Binding),
	); err != nil {
		return s.HandleUpdateError(err, "tools")
	}

	return nil
}
