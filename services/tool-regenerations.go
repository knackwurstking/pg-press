package services

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
)

const TableNameToolRegenerations = "tool_regenerations"

type ToolRegenerations struct {
	*Base
}

func NewToolRegenerations(r *Registry) *ToolRegenerations {
	base := NewBase(r)

	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %[1]s (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tool_id INTEGER NOT NULL,
			cycle_id INTEGER NOT NULL,
			reason TEXT,
			performed_by INTEGER NOT NULL
		);
	`, TableNameToolRegenerations, TableNameTools)

	if err := base.CreateTable(query, TableNameToolRegenerations); err != nil {
		panic(err)
	}

	return &ToolRegenerations{
		Base: base,
	}
}

func (s *ToolRegenerations) Get(id models.RegenerationID) (*models.Regeneration, error) {
	slog.Debug("Getting tool regeneration by ID", "id", id)

	query := fmt.Sprintf(`SELECT * FROM %s WHERE id = ?`, TableNameToolRegenerations)
	row := s.DB.QueryRow(query, id)

	regeneration, err := ScanSingleRow(row, scanToolRegeneration)
	if err != nil {
		return nil, s.GetSelectError(err)
	}

	return regeneration, nil
}

func (s *ToolRegenerations) Add(toolID models.ToolID, cycleID models.CycleID, reason string, user *models.User) (models.RegenerationID, error) {
	slog.Debug("Adding tool regeneration",
		"user_name", user.Name, "tool_id", toolID, "cycle_id", cycleID, "reason", reason)

	if err := user.Validate(); err != nil {
		return 0, err
	}

	r := models.NewRegeneration(toolID, cycleID, reason, &user.TelegramID)
	if err := r.Validate(); err != nil {
		return 0, err
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (tool_id, cycle_id, reason, performed_by)
		VALUES (?, ?, ?, ?)
	`, TableNameToolRegenerations)

	result, err := s.DB.Exec(query, r.ToolID, r.CycleID, r.Reason, user.TelegramID)
	if err != nil {
		return 0, s.GetInsertError(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get last insert ID: %v", err)
	}

	return models.RegenerationID(id), nil
}

func (s *ToolRegenerations) Update(r *models.Regeneration, user *models.User) error {
	slog.Debug("Updating tool regeneration", "user_name", user.Name, "regeneration_id", r.ID)

	if err := user.Validate(); err != nil {
		return err
	}

	if err := r.Validate(); err != nil {
		return err
	}

	query := fmt.Sprintf(`
		UPDATE %s
		SET cycle_id = ?, reason = ?, performed_by = ?
		WHERE id = ?
	`, TableNameToolRegenerations)

	_, err := s.DB.Exec(query, r.CycleID, r.Reason, user.TelegramID, r.ID)
	if err != nil {
		return s.GetUpdateError(err)
	}

	return nil
}

func (s *ToolRegenerations) Delete(id models.RegenerationID) error {
	slog.Debug("Deleting tool regeneration", "id", id)

	query := fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, TableNameToolRegenerations)
	_, err := s.DB.Exec(query, id)
	if err != nil {
		return s.GetDeleteError(err)
	}

	return nil
}

func (s *ToolRegenerations) StartToolRegeneration(toolID models.ToolID, reason string, user *models.User) (models.RegenerationID, error) {
	slog.Debug("Starting tool regeneration", "user_name", user.Name, "tool_id", toolID)

	if err := user.Validate(); err != nil {
		return 0, err
	}

	cycle, err := s.Registry.PressCycles.GetLastToolCycle(toolID)
	if err != nil {
		return 0, err
	}

	if err := s.Registry.Tools.UpdateRegenerating(toolID, true, user); err != nil {
		return 0, err
	}

	regenerationID, err := s.Add(toolID, cycle.ID, reason, user)
	if err != nil {
		if undoErr := s.Registry.Tools.UpdateRegenerating(toolID, false, user); undoErr != nil {
			return 0, fmt.Errorf(
				"failed to create regeneration record: %v, failed to undo tool regeneration status: %v",
				err, undoErr,
			)
		}
		return 0, err
	}

	return regenerationID, nil
}

func (s *ToolRegenerations) StopToolRegeneration(toolID models.ToolID, user *models.User) error {
	slog.Debug("Stopping tool regeneration", "user_name", user.Name, "tool_id", toolID)

	if err := user.Validate(); err != nil {
		return err
	}

	if err := s.Registry.Tools.UpdateRegenerating(toolID, false, user); err != nil {
		return fmt.Errorf("update tool regeneration status: %v", err)
	}

	return nil
}

func (s *ToolRegenerations) AbortToolRegeneration(toolID models.ToolID, user *models.User) error {
	slog.Debug("Aborting tool regeneration", "user_name", user.Name, "tool_id", toolID)

	if err := user.Validate(); err != nil {
		return err
	}

	lastRegeneration, err := s.GetLastRegeneration(toolID)
	if err != nil {
		if !errors.IsNotFoundError(err) {
			return fmt.Errorf("get last regeneration record: %v", err)
		}
		slog.Debug("No regeneration record found to abort: tool", "tool", toolID)
	} else {
		tool, err := s.Registry.Tools.Get(toolID)
		if err != nil {
			return fmt.Errorf("get tool: %v", err)
		}
		if !tool.Regenerating {
			return fmt.Errorf("tool is not regenerating")
		}

		slog.Debug("Deleting regeneration record", "user_name", user.Name, "last_regeneration_id", lastRegeneration.ID)
		if err := s.Delete(lastRegeneration.ID); err != nil {
			return fmt.Errorf("delete regeneration record: %v", err)
		}
	}

	slog.Debug("Setting tool to non-regenerating status: tool", "tool", toolID)
	if err := s.Registry.Tools.UpdateRegenerating(toolID, false, user); err != nil {
		return fmt.Errorf("update tool regeneration status: %v", err)
	}

	return nil
}

func (s *ToolRegenerations) GetLastRegeneration(toolID models.ToolID) (*models.Regeneration, error) {
	query := fmt.Sprintf(`
		SELECT id, tool_id, cycle_id, reason, performed_by
		FROM %s
		WHERE tool_id = ?
		ORDER BY id DESC
		LIMIT 1
	`, TableNameToolRegenerations)

	row := s.DB.QueryRow(query, toolID)
	regeneration, err := ScanSingleRow(row, scanToolRegeneration)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError(
				fmt.Sprintf("tool regeneration for tool_id: %d", toolID),
			)
		}
		return nil, err
	}

	slog.Debug("Got last regeneration for tool", "tool", toolID, "regeneration", regeneration)
	return regeneration, nil
}

func (s *ToolRegenerations) HasRegenerationsForCycle(cycleID models.CycleID) (bool, error) {
	slog.Debug("Checking if cycle has regenerations", "cycle", cycleID)

	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE cycle_id = ?`, TableNameToolRegenerations)

	var count int
	err := s.DB.QueryRow(query, cycleID).Scan(&count)
	if err != nil {
		return false, s.GetSelectError(err)
	}

	return count > 0, nil
}

func (s *ToolRegenerations) GetRegenerationHistory(toolID models.ToolID) ([]*models.Regeneration, error) {
	slog.Debug("Getting regeneration history for tool", "tool", toolID)

	query := fmt.Sprintf(`
		SELECT id, tool_id, cycle_id, reason, performed_by
		FROM %s
		WHERE tool_id = ?
		ORDER BY id DESC
	`, TableNameToolRegenerations)

	rows, err := s.DB.Query(query, toolID)
	if err != nil {
		return nil, s.GetSelectError(err)
	}
	defer rows.Close()

	regenerations, err := ScanRows(rows, scanToolRegeneration)
	if err != nil {
		return nil, err
	}

	return regenerations, nil
}

func scanToolRegeneration(scannable Scannable) (*models.Regeneration, error) {
	regeneration := &models.Regeneration{}
	var performedBy sql.NullInt64

	err := scannable.Scan(
		&regeneration.ID,
		&regeneration.ToolID,
		&regeneration.CycleID,
		&regeneration.Reason,
		&performedBy,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("scan tool regeneration: %v", err)
	}

	if performedBy.Valid {
		performedBy := models.TelegramID(performedBy.Int64)
		regeneration.PerformedBy = &performedBy
	}

	return regeneration, nil
}
