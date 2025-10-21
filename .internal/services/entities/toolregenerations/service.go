package toolregenerations

import (
	"database/sql"
	"fmt"

	"github.com/knackwurstking/pgpress/internal/services/base"
	"github.com/knackwurstking/pgpress/internal/services/shared/scanner"
	"github.com/knackwurstking/pgpress/internal/services/shared/validation"
	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"
)

type Service struct {
	*base.BaseService
	tools       ToolsService
	pressCycles PressCyclesService
}

// ToolsService defines the interface for tools service methods used by ToolRegenerations
type ToolsService interface {
	Get(id int64) (*models.Tool, error)
	Update(tool *models.Tool, user *models.User) error
	UpdateRegenerating(toolID int64, regenerating bool, user *models.User) error
}

type PressCyclesService interface {
	GetLastToolCycle(toolID int64) (*models.Cycle, error)
}

func NewService(db *sql.DB, tools ToolsService, pressCycles PressCyclesService) *Service {
	baseService := base.NewBaseService(db, "Tool Regenerations")

	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %[1]s (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tool_id INTEGER NOT NULL,
			cycle_id INTEGER NOT NULL,
			reason TEXT,
			performed_by INTEGER NOT NULL,
			FOREIGN KEY (tool_id) REFERENCES tools(id) ON DELETE CASCADE,
			FOREIGN KEY (performed_by) REFERENCES users(telegram_id) ON DELETE SET NULL,
			FOREIGN KEY (cycle_id) REFERENCES press_cycles(id) ON DELETE SET NULL
		);
		CREATE INDEX IF NOT EXISTS idx_%[1]s_tool_id ON %[1]s(tool_id);
		CREATE INDEX IF NOT EXISTS idx_%[1]s_cycle_id ON %[1]s(cycle_id);
	`, TableName)

	if err := baseService.CreateTable(query, TableName); err != nil {
		panic(err)
	}

	return &Service{
		BaseService: baseService,
		tools:       tools,
		pressCycles: pressCycles,
	}
}

func (s *Service) Get(id int64) (*models.Regeneration, error) {
	if err := validation.ValidateID(id, EntityName); err != nil {
		return nil, err
	}

	r := s.DB.QueryRow(
		fmt.Sprintf(`
			SELECT * FROM %s WHERE id = ?
		`, TableName),
		id,
	)

	regeneration, err := scanner.ScanSingleRow(
		r, ScanToolRegeneration,
		EntityName,
	)
	if err != nil {
		return nil, s.HandleScanError(err, EntityName)
	}

	return regeneration, nil
}

// Add records a new tool regeneration event
func (s *Service) Add(toolID, cycleID int64, reason string, user *models.User) (int64, error) {
	if err := validation.ValidateNotNil(user, "user"); err != nil {
		return 0, err
	}

	r := models.NewRegeneration(toolID, cycleID, reason, &user.TelegramID)

	if err := ValidateToolRegeneration(r); err != nil {
		return 0, err
	}

	s.Log.Debug("Adding tool regeneration by %s (%d): tool: %d, cycle: %d, reason: %s",
		user.Name, user.TelegramID, r.ToolID, r.CycleID, r.Reason)

	query := fmt.Sprintf(`
		INSERT INTO %s (tool_id, cycle_id, reason, performed_by)
		VALUES (?, ?, ?, ?)
	`, TableName)

	row, err := s.DB.Exec(query, r.ToolID, r.CycleID, r.Reason, user.TelegramID)
	if err != nil {
		return 0, s.HandleInsertError(err, EntityName)
	}

	id, err := row.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %v", err)
	}

	return id, nil
}

// Update updates an existing regeneration record
func (s *Service) Update(r *models.Regeneration, user *models.User) error {
	if err := validation.ValidateNotNil(user, "user"); err != nil {
		return err
	}

	if err := ValidateToolRegeneration(r); err != nil {
		return err
	}

	s.Log.Debug("Updating tool regeneration by %s (%d): id: %d",
		user.Name, user.TelegramID, r.ID)

	query := fmt.Sprintf(`
		UPDATE %s
		SET cycle_id = ?, reason = ?, performed_by = ?
		WHERE id = ?
	`, TableName)

	result, err := s.DB.Exec(query, r.CycleID, r.Reason, user.TelegramID, r.ID)
	if err != nil {
		return s.HandleUpdateError(err, EntityName)
	}

	if err := s.CheckRowsAffected(result, "tool_regeneration", r.ID); err != nil {
		return err
	}

	return nil
}

// Delete removes a regeneration record (should be used carefully)
func (r *Service) Delete(id int64) error {
	if err := validation.ValidateID(id, "regeneration"); err != nil {
		return err
	}

	r.Log.Debug("Deleting tool regeneration: %d", id)

	query := fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, TableName)
	result, err := r.DB.Exec(query, id)
	if err != nil {
		return r.HandleDeleteError(err, EntityName)
	}

	if err := r.CheckRowsAffected(result, "tool_regeneration", id); err != nil {
		return err
	}

	return nil
}

// AddToolRegeneration starts the tool regeneration process
func (r *Service) StartToolRegeneration(toolID int64, reason string, user *models.User) (int64, error) {
	if err := validation.ValidateNotNil(user, "user"); err != nil {
		return 0, err
	}

	r.Log.Debug("Starting tool regeneration by %s (%d): tool: %d",
		user.Name, user.TelegramID, toolID)

	cycle, err := r.pressCycles.GetLastToolCycle(toolID)
	if err != nil {
		return 0, err
	}

	// Update the tool's regeneration status
	if err := r.tools.UpdateRegenerating(toolID, true, user); err != nil {
		return 0, fmt.Errorf("failed to update tool regeneration status: %v", err)
	}

	// Create a new regeneration record
	r.Log.Debug("Creating regeneration record: tool: %d", toolID)
	regenerationID, err := r.Add(toolID, cycle.ID, reason, user)
	if err != nil {
		// Undo the tool's regeneration status on failure
		r.Log.Error("Failed to create regeneration record, undoing status change for tool: %d", toolID)
		if undoErr := r.tools.UpdateRegenerating(toolID, false, user); undoErr != nil {
			r.Log.Error("Failed to undo tool regeneration status: %v", undoErr)
		}
		return 0, err
	}

	r.Log.Debug(
		"Started tool regeneration successfully: tool: %d, regen_id: %d",
		toolID, regenerationID,
	)
	return regenerationID, nil
}

// StopToolRegeneration stops the tool regeneration process for the given tool ID
func (r *Service) StopToolRegeneration(toolID int64, user *models.User) error {
	if err := validation.ValidateID(toolID, "tool"); err != nil {
		return err
	}

	if err := validation.ValidateNotNil(user, "user"); err != nil {
		return err
	}

	r.Log.Debug("Stopping tool regeneration by %s (%d): tool: %d",
		user.Name, user.TelegramID, toolID)

	// Set the tool's regeneration status to false
	if err := r.tools.UpdateRegenerating(toolID, false, user); err != nil {
		return fmt.Errorf("failed to update tool regeneration status: %v", err)
	}

	r.Log.Debug("Stopped tool regeneration: tool: %d", toolID)
	return nil
}

// AbortToolRegeneration aborts the tool regeneration process and removes the regeneration record
func (r *Service) AbortToolRegeneration(toolID int64, user *models.User) error {
	if err := validation.ValidateID(toolID, "tool"); err != nil {
		return err
	}

	if err := validation.ValidateNotNil(user, "user"); err != nil {
		return err
	}

	r.Log.Debug("Aborting tool regeneration by %s (%d): tool: %d", user.Name, user.TelegramID, toolID)

	// First, get the last regeneration record to delete it
	lastRegen, err := r.GetLastRegeneration(toolID)
	if err != nil {
		if !utils.IsNotFoundError(err) {
			return fmt.Errorf("failed to get last regeneration record: %v", err)
		}
		r.Log.Debug("No regeneration record found to abort: tool: %d", toolID)
	} else {
		// Check this regeneration if its finished
		tool, err := r.tools.Get(toolID)
		if err != nil {
			return fmt.Errorf("failed to get tool: %v", err)
		}
		if !tool.Regenerating {
			// Nope, already finished
			return fmt.Errorf("tool is not regenerating")
		}

		// Delete the regeneration record
		r.Log.Debug("Deleting regeneration record: id: %d", lastRegen.ID)
		if err := r.Delete(lastRegen.ID); err != nil {
			return fmt.Errorf("failed to delete regeneration record: %v", err)
		}
	}

	// Set the tool's regeneration status to false
	r.Log.Debug("Setting tool to non-regenerating status: tool: %d", toolID)
	if err := r.tools.UpdateRegenerating(toolID, false, user); err != nil {
		return fmt.Errorf("failed to update tool regeneration status: %v", err)
	}

	return nil
}

// GetLastRegeneration gets the most recent regeneration for a tool
func (r *Service) GetLastRegeneration(toolID int64) (*models.Regeneration, error) {
	if err := validation.ValidateID(toolID, "tool"); err != nil {
		return nil, err
	}

	r.Log.Debug("Getting last regeneration for tool: %d", toolID)

	query := fmt.Sprintf(`
		SELECT id, tool_id, cycle_id, reason, performed_by
		FROM %s
		WHERE tool_id = ?
		ORDER BY id DESC
		LIMIT 1
	`, TableName)

	row := r.DB.QueryRow(query, toolID)
	regen, err := scanner.ScanSingleRow(row, ScanToolRegeneration, EntityName)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError(fmt.Sprintf("tool regeneration for tool_id: %d", toolID))
		}
		return nil, err
	}

	return regen, nil
}

// GetRegenerationHistory gets all regenerations for a tool
// HasRegenerationsForCycle checks if a cycle has any regenerations associated with it
func (r *Service) HasRegenerationsForCycle(cycleID int64) (bool, error) {
	if err := validation.ValidateID(cycleID, "cycle"); err != nil {
		return false, err
	}

	r.Log.Debug("Checking if cycle has regenerations: %d", cycleID)

	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE cycle_id = ?`, TableName)
	var count int
	err := r.DB.QueryRow(query, cycleID).Scan(&count)
	if err != nil {
		return false, r.HandleSelectError(err, EntityName)
	}

	return count > 0, nil
}

func (r *Service) GetRegenerationHistory(toolID int64) ([]*models.Regeneration, error) {
	if err := validation.ValidateID(toolID, "tool"); err != nil {
		return nil, err
	}

	r.Log.Debug("Getting regeneration history for tool: %d", toolID)

	query := fmt.Sprintf(`
		SELECT id, tool_id, cycle_id, reason, performed_by
		FROM %s
		WHERE tool_id = ?
		ORDER BY id DESC
	`, TableName)

	rows, err := r.DB.Query(query, toolID)
	if err != nil {
		return nil, r.HandleSelectError(err, EntityName)
	}
	defer rows.Close()

	regenerations, err := ScanToolRegenerationsFromRows(rows)
	if err != nil {
		return nil, err
	}

	return regenerations, nil
}
