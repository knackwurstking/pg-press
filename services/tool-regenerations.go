package services

import (
	"database/sql"
	"fmt"

	"github.com/knackwurstking/pgpress/errors"
	"github.com/knackwurstking/pgpress/logger"
	"github.com/knackwurstking/pgpress/models"
)

const TableNameToolRegenerations = "tool_regenerations"

type ToolRegenerations struct {
	*Base
}

func NewToolRegenerations(r *Registry) *ToolRegenerations {
	base := NewBase(r, logger.NewComponentLogger("Service: ToolRegenerations"))

	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %[1]s (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tool_id INTEGER NOT NULL,
			cycle_id INTEGER NOT NULL,
			reason TEXT,
			performed_by INTEGER NOT NULL,
			FOREIGN KEY (tool_id) REFERENCES %[2]s(id) ON DELETE CASCADE,
			FOREIGN KEY (performed_by) REFERENCES users(telegram_id) ON DELETE SET NULL,
			FOREIGN KEY (cycle_id) REFERENCES press_cycles(id) ON DELETE SET NULL
		);
		CREATE INDEX IF NOT EXISTS idx_%[1]s_tool_id ON %[1]s(tool_id);
		CREATE INDEX IF NOT EXISTS idx_%[1]s_cycle_id ON %[1]s(cycle_id);
	`, TableNameToolRegenerations, TableNameTools)

	if err := base.CreateTable(query, TableNameToolRegenerations); err != nil {
		panic(err)
	}

	return &ToolRegenerations{
		Base: base,
	}
}

func (s *ToolRegenerations) Get(id models.RegenerationID) (*models.Regeneration, error) {
	s.Log.Debug("Getting tool regeneration by ID: %d", id)

	query := fmt.Sprintf(`SELECT * FROM %s WHERE id = ?`, TableNameToolRegenerations)
	row := s.DB.QueryRow(query, id)

	regeneration, err := ScanSingleRow(row, scanToolRegeneration)
	if err != nil {
		return nil, s.GetSelectError(err)
	}

	return regeneration, nil
}

func (s *ToolRegenerations) Add(toolID int64, cycleID models.CycleID, reason string, user *models.User) (models.RegenerationID, error) {
	s.Log.Debug("Adding tool regeneration by %s (%d): tool: %d, cycle: %d, reason: %s",
		user.Name, user.TelegramID, toolID, cycleID, reason)

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
		return 0, fmt.Errorf("failed to get last insert ID: %v", err)
	}

	return models.RegenerationID(id), nil
}

func (s *ToolRegenerations) Update(r *models.Regeneration, user *models.User) error {
	s.Log.Debug("Updating tool regeneration by %s (%d): id: %d",
		user.Name, user.TelegramID, r.ID)

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
	s.Log.Debug("Deleting tool regeneration: %d", id)

	query := fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, TableNameToolRegenerations)
	_, err := s.DB.Exec(query, id)
	if err != nil {
		return s.GetDeleteError(err)
	}

	return nil
}

func (s *ToolRegenerations) StartToolRegeneration(toolID int64, reason string, user *models.User) (models.RegenerationID, error) {
	s.Log.Debug("Starting tool regeneration by %s (%d): tool: %d",
		user.Name, user.TelegramID, toolID)

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

func (s *ToolRegenerations) StopToolRegeneration(toolID int64, user *models.User) error {
	s.Log.Debug("Stopping tool regeneration by %s (%d): tool: %d",
		user.Name, user.TelegramID, toolID)

	if err := user.Validate(); err != nil {
		return err
	}

	if err := s.Registry.Tools.UpdateRegenerating(toolID, false, user); err != nil {
		return fmt.Errorf("failed to update tool regeneration status: %v", err)
	}

	return nil
}

func (s *ToolRegenerations) AbortToolRegeneration(toolID int64, user *models.User) error {
	s.Log.Debug("Aborting tool regeneration by %s (%d): tool: %d",
		user.Name, user.TelegramID, toolID)

	if err := user.Validate(); err != nil {
		return err
	}

	lastRegen, err := s.GetLastRegeneration(toolID)
	if err != nil {
		if !errors.IsNotFoundError(err) {
			return fmt.Errorf("failed to get last regeneration record: %v", err)
		}
		s.Log.Debug("No regeneration record found to abort: tool: %d", toolID)
	} else {
		tool, err := s.Registry.Tools.Get(toolID)
		if err != nil {
			return fmt.Errorf("failed to get tool: %v", err)
		}
		if !tool.Regenerating {
			return fmt.Errorf("tool is not regenerating")
		}

		s.Log.Debug("Deleting regeneration record: id: %d", lastRegen.ID)
		if err := s.Delete(lastRegen.ID); err != nil {
			return fmt.Errorf("failed to delete regeneration record: %v", err)
		}
	}

	s.Log.Debug("Setting tool to non-regenerating status: tool: %d", toolID)
	if err := s.Registry.Tools.UpdateRegenerating(toolID, false, user); err != nil {
		return fmt.Errorf("failed to update tool regeneration status: %v", err)
	}

	return nil
}

func (s *ToolRegenerations) GetLastRegeneration(toolID int64) (*models.Regeneration, error) {
	s.Log.Debug("Getting last regeneration for tool: %d", toolID)

	query := fmt.Sprintf(`
		SELECT id, tool_id, cycle_id, reason, performed_by
		FROM %s
		WHERE tool_id = ?
		ORDER BY id DESC
		LIMIT 1
	`, TableNameToolRegenerations)

	row := s.DB.QueryRow(query, toolID)
	regen, err := ScanSingleRow(row, scanToolRegeneration)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError(
				fmt.Sprintf("tool regeneration for tool_id: %d", toolID),
			)
		}
		return nil, err
	}

	return regen, nil
}

func (s *ToolRegenerations) HasRegenerationsForCycle(cycleID models.CycleID) (bool, error) {
	s.Log.Debug("Checking if cycle has regenerations: %d", cycleID)

	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE cycle_id = ?`, TableNameToolRegenerations)

	var count int
	err := s.DB.QueryRow(query, cycleID).Scan(&count)
	if err != nil {
		return false, s.GetSelectError(err)
	}

	return count > 0, nil
}

func (s *ToolRegenerations) GetRegenerationHistory(toolID int64) ([]*models.Regeneration, error) {
	s.Log.Debug("Getting regeneration history for tool: %d", toolID)

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
	regen := &models.Regeneration{}
	var performedBy sql.NullInt64

	err := scannable.Scan(
		&regen.ID,
		&regen.ToolID,
		&regen.CycleID,
		&regen.Reason,
		&performedBy,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("failed to scan tool regeneration: %v", err)
	}

	if performedBy.Valid {
		regen.PerformedBy = &performedBy.Int64
	}

	return regen, nil
}
