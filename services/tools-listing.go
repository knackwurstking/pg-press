package services

import (
	"fmt"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
)

func (t *Tools) List() ([]*models.Tool, error) {
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

func (t *Tools) GetActiveToolsForPress(pressNumber models.PressNumber) ([]*models.Tool, error) {
	if !models.IsValidPressNumber(&pressNumber) {
		return nil, errors.NewValidationError(fmt.Sprintf("invalid press number: %d", pressNumber))
	}

	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE regenerating = 0 AND is_dead = 0 AND press = ?`,
		ToolQuerySelect, TableNameTools)

	rows, err := t.DB.Query(query, pressNumber)
	if err != nil {
		return nil, t.GetSelectError(err)
	}
	defer rows.Close()

	tools, err := ScanRows(rows, scanTool)
	if err != nil {
		return nil, t.GetSelectError(err)
	}

	return tools, nil
}

func (t *Tools) GetPressUtilization() ([]models.PressUtilization, error) {
	// Valid press numbers: 0, 2, 3, 4, 5
	validPresses := []models.PressNumber{0, 2, 3, 4, 5}
	utilization := make([]models.PressUtilization, 0, len(validPresses))

	for _, pressNum := range validPresses {
		tools, err := t.GetActiveToolsForPress(pressNum)
		if err != nil {
			return nil, err
		}
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
