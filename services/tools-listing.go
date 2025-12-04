package services

import (
	"fmt"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
)

func (t *Tools) List() ([]*models.Tool, *errors.DBError) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		ORDER BY format ASC, code ASC`,
		ToolQuerySelect, TableNameTools)

	rows, err := t.DB.Query(query)
	if err != nil {
		return nil, errors.NewDBError(err, errors.DBTypeSelect)
	}
	defer rows.Close()

	return ScanRows(rows, ScanTool)
}

func (t *Tools) ListToolsNotDead() ([]*models.Tool, *errors.DBError) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE is_dead = 0
		ORDER BY format ASC, code ASC`,
		ToolQuerySelect, TableNameTools)

	rows, err := t.DB.Query(query)
	if err != nil {
		return nil, errors.NewDBError(err, errors.DBTypeSelect)
	}
	defer rows.Close()

	return ScanRows(rows, ScanTool)
}

func (t *Tools) ListDeadTools() ([]*models.Tool, *errors.DBError) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE is_dead = 1
		ORDER BY format ASC, code ASC`,
		ToolQuerySelect, TableNameTools)

	rows, err := t.DB.Query(query)
	if err != nil {
		return nil, errors.NewDBError(err, errors.DBTypeSelect)
	}
	defer rows.Close()

	return ScanRows(rows, ScanTool)
}

func (t *Tools) GetByPress(pressNumber *models.PressNumber) ([]*models.Tool, *errors.DBError) {
	if pressNumber != nil && !models.IsValidPressNumber(pressNumber) {
		return nil, errors.NewDBError(
			fmt.Errorf("invalid press number: %d (must be 0-5)", *pressNumber),
			errors.DBTypeValidation,
		)
	}

	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE press = ? AND regenerating = 0 AND is_dead = 0`,
		ToolQuerySelect, TableNameTools)

	rows, err := t.DB.Query(query, pressNumber)
	if err != nil {
		return nil, errors.NewDBError(err, errors.DBTypeSelect)
	}
	defer rows.Close()

	return ScanRows(rows, ScanTool)
}

func (t *Tools) GetActiveToolsForPress(pressNumber models.PressNumber) ([]*models.Tool, *errors.DBError) {
	if !models.IsValidPressNumber(&pressNumber) {
		return nil, errors.NewDBError(
			fmt.Errorf("invalid press number: %d", pressNumber),
			errors.DBTypeValidation,
		)
	}

	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE regenerating = 0 AND is_dead = 0 AND press = ?`,
		ToolQuerySelect, TableNameTools)

	rows, err := t.DB.Query(query, pressNumber)
	if err != nil {
		return nil, errors.NewDBError(err, errors.DBTypeSelect)
	}
	defer rows.Close()

	tools, dberr := ScanRows(rows, ScanTool)
	if dberr != nil {
		return nil, dberr
	}

	return tools, nil
}

func (t *Tools) GetPressUtilization() ([]models.PressUtilization, *errors.DBError) {
	// Valid press numbers: 0, 2, 3, 4, 5
	validPresses := []models.PressNumber{0, 2, 3, 4, 5}
	utilization := make([]models.PressUtilization, 0, len(validPresses))

	for _, pressNum := range validPresses {
		tools, dberr := t.GetActiveToolsForPress(pressNum)
		if dberr != nil {
			return nil, dberr
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
