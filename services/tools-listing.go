package services

import (
	"fmt"
	"net/http"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
)

func (t *Tools) List() ([]*models.Tool, *errors.MasterError) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		ORDER BY format ASC, code ASC`,
		ToolQuerySelect, TableNameTools)

	rows, err := t.DB.Query(query)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	defer rows.Close()

	return ScanRows(rows, ScanTool)
}

func (t *Tools) ListToolsNotDead() ([]*models.Tool, *errors.MasterError) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE is_dead = 0
		ORDER BY format ASC, code ASC`,
		ToolQuerySelect, TableNameTools)

	rows, err := t.DB.Query(query)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	defer rows.Close()

	return ScanRows(rows, ScanTool)
}

func (t *Tools) ListDeadTools() ([]*models.Tool, *errors.MasterError) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE is_dead = 1
		ORDER BY format ASC, code ASC`,
		ToolQuerySelect, TableNameTools)

	rows, err := t.DB.Query(query)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	defer rows.Close()

	return ScanRows(rows, ScanTool)
}

func (t *Tools) ListByPress(pressNumber *models.PressNumber) ([]*models.Tool, *errors.MasterError) {
	if pressNumber != nil && !models.IsValidPressNumber(pressNumber) {
		return nil, errors.NewMasterError(fmt.Errorf("invalid press number: %d", pressNumber), http.StatusBadRequest)
	}

	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE press = ? AND regenerating = 0 AND is_dead = 0`,
		ToolQuerySelect, TableNameTools)

	rows, err := t.DB.Query(query, pressNumber)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	defer rows.Close()

	return ScanRows(rows, ScanTool)
}

func (t *Tools) ListActiveToolsForPress(pressNumber models.PressNumber) ([]*models.Tool, *errors.MasterError) {
	if !models.IsValidPressNumber(&pressNumber) {
		return nil, errors.NewMasterError(fmt.Errorf("invalid press number: %d", pressNumber), http.StatusBadRequest)
	}

	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE regenerating = 0 AND is_dead = 0 AND press = ?`,
		ToolQuerySelect, TableNameTools)

	rows, err := t.DB.Query(query, pressNumber)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	defer rows.Close()

	tools, dberr := ScanRows(rows, ScanTool)
	if dberr != nil {
		return nil, dberr
	}

	return tools, nil
}

func (t *Tools) PressUtilization() ([]models.PressUtilization, *errors.MasterError) {
	// Valid press numbers: 0, 2, 3, 4, 5
	validPresses := []models.PressNumber{0, 2, 3, 4, 5}
	utilization := make([]models.PressUtilization, 0, len(validPresses))

	for _, pressNum := range validPresses {
		tools, err := t.ListActiveToolsForPress(pressNum)
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
