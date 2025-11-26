package services

import (
	"fmt"
	"log/slog"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
)

func (t *Tools) List() ([]*models.Tool, error) {
	slog.Debug("Listing tools")

	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		ORDER BY format ASC, code ASC`,
		ToolQuerySelect, TableNameTools)

	rows, err := t.DB.Query(query)
	if err != nil {
		return nil, t.GetSelectError(err)
	}
	defer rows.Close()

	return ScanRows(rows, scanTool)
}

func (t *Tools) ListToolsNotDead() ([]*models.Tool, error) {
	slog.Debug("Listing active tools")

	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE is_dead = 0
		ORDER BY format ASC, code ASC`,
		ToolQuerySelect, TableNameTools)

	rows, err := t.DB.Query(query)
	if err != nil {
		return nil, t.GetSelectError(err)
	}
	defer rows.Close()

	return ScanRows(rows, scanTool)
}

func (t *Tools) ListDeadTools() ([]*models.Tool, error) {
	slog.Debug("Listing dead tools")

	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE is_dead = 1
		ORDER BY format ASC, code ASC`,
		ToolQuerySelect, TableNameTools)

	rows, err := t.DB.Query(query)
	if err != nil {
		return nil, t.GetSelectError(err)
	}
	defer rows.Close()

	return ScanRows(rows, scanTool)
}

func (t *Tools) GetByPress(pressNumber *models.PressNumber) ([]*models.Tool, error) {
	slog.Debug("Getting tools by press", "press", pressNumber)

	if pressNumber != nil && !models.IsValidPressNumber(pressNumber) {
		return nil, errors.NewValidationError(
			fmt.Sprintf("invalid press number: %d (must be 0-5)", *pressNumber),
		)
	}

	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE press = ? AND regenerating = 0 AND is_dead = 0`,
		ToolQuerySelect, TableNameTools)

	rows, err := t.DB.Query(query, pressNumber)
	if err != nil {
		return nil, t.GetSelectError(err)
	}
	defer rows.Close()

	return ScanRows(rows, scanTool)
}

func (t *Tools) GetActiveToolsForPress(pressNumber models.PressNumber) []*models.Tool {
	slog.Debug("Getting active tools for press", "press", pressNumber)

	if !models.IsValidPressNumber(&pressNumber) {
		slog.Error("Invalid press number (must be 0-5)", "press", pressNumber)
		return nil
	}

	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE regenerating = 0 AND is_dead = 0 AND press = ?`,
		ToolQuerySelect, TableNameTools)

	rows, err := t.DB.Query(query, pressNumber)
	if err != nil {
		slog.Error("Failed to query active tools", "error", err)
		return nil
	}
	defer rows.Close()

	tools, err := ScanRows(rows, scanTool)
	if err != nil {
		slog.Error("Failed to scan active tools", "error", err)
		return nil
	}

	return tools
}

func (t *Tools) GetPressUtilization() ([]models.PressUtilization, error) {
	slog.Debug("Getting press utilization")

	// Valid press numbers: 0, 2, 3, 4, 5
	validPresses := []models.PressNumber{0, 2, 3, 4, 5}
	utilization := make([]models.PressUtilization, 0, len(validPresses))

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
