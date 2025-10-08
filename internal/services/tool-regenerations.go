package services

import (
	"database/sql"
	"fmt"

	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"
)

type ToolRegenerations struct {
	*BaseService
	tools *Tools
}

func NewToolRegenerations(db *sql.DB, tools *Tools) *ToolRegenerations {
	base := NewBaseService(db, "Tool Regenerations")

	query := `
		CREATE TABLE IF NOT EXISTS tool_regenerations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tool_id INTEGER NOT NULL,
			cycle_id INTEGER NOT NULL,
			reason TEXT,
			performed_by INTEGER NOT NULL,
			FOREIGN KEY (tool_id) REFERENCES tools(id) ON DELETE CASCADE,
			FOREIGN KEY (performed_by) REFERENCES users(telegram_id) ON DELETE SET NULL,
			FOREIGN KEY (cycle_id) REFERENCES press_cycles(id) ON DELETE SET NULL
		);
		CREATE INDEX IF NOT EXISTS idx_tool_regenerations_tool_id ON tool_regenerations(tool_id);
		CREATE INDEX IF NOT EXISTS idx_tool_regenerations_cycle_id ON tool_regenerations(cycle_id);
	`

	if err := base.CreateTable(query, "tool_regenerations"); err != nil {
		panic(err)
	}

	return &ToolRegenerations{
		BaseService: base,
		tools:       tools,
	}
}

// Add records a new tool regeneration event
func (r *ToolRegenerations) Add(regeneration *models.Regeneration, user *models.User) (*models.Regeneration, error) {
	if err := ValidateToolRegeneration(regeneration); err != nil {
		return nil, err
	}

	if err := ValidateNotNil(user, "user"); err != nil {
		return nil, err
	}

	r.LogOperationWithUser("Adding tool regeneration", createUserInfo(user),
		fmt.Sprintf("tool: %d, cycle: %d, reason: %s", regeneration.ToolID, regeneration.CycleID, regeneration.Reason))

	query := `
		INSERT INTO tool_regenerations (tool_id, cycle_id, reason, performed_by)
		VALUES (?, ?, ?, ?)
		RETURNING id, tool_id, cycle_id, reason, performed_by
	`

	row := r.db.QueryRow(query,
		regeneration.ToolID,
		regeneration.CycleID,
		regeneration.Reason,
		user.TelegramID,
	)

	result, err := ScanSingleRow(row, ScanToolRegeneration, "tool_regenerations")
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError("failed to create tool regeneration")
		}
		return nil, err
	}

	r.LogOperation("Added tool regeneration", fmt.Sprintf("id: %d", result.ID))
	return result, nil
}

// Update updates an existing regeneration record
func (r *ToolRegenerations) Update(regeneration *models.Regeneration, user *models.User) error {
	if err := ValidateToolRegeneration(regeneration); err != nil {
		return err
	}

	if err := ValidateID(regeneration.ID, "regeneration"); err != nil {
		return err
	}

	if err := ValidateNotNil(user, "user"); err != nil {
		return err
	}

	r.LogOperationWithUser("Updating tool regeneration", createUserInfo(user), fmt.Sprintf("id: %d", regeneration.ID))

	query := `
		UPDATE tool_regenerations
		SET cycle_id = ?, reason = ?, performed_by = ?
		WHERE id = ?
	`

	result, err := r.db.Exec(query,
		regeneration.CycleID,
		regeneration.Reason,
		user.TelegramID,
		regeneration.ID,
	)
	if err != nil {
		return r.HandleUpdateError(err, "tool_regenerations")
	}

	if err := r.CheckRowsAffected(result, "tool_regeneration", regeneration.ID); err != nil {
		return err
	}

	r.LogOperation("Updated tool regeneration", fmt.Sprintf("id: %d", regeneration.ID))
	return nil
}

// Delete removes a regeneration record (should be used carefully)
func (r *ToolRegenerations) Delete(id int64) error {
	if err := ValidateID(id, "regeneration"); err != nil {
		return err
	}

	r.LogOperation("Deleting tool regeneration", id)

	query := `DELETE FROM tool_regenerations WHERE id = ?`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return r.HandleDeleteError(err, "tool_regenerations")
	}

	if err := r.CheckRowsAffected(result, "tool_regeneration", id); err != nil {
		return err
	}

	r.LogOperation("Deleted tool regeneration", id)
	return nil
}

// AddToolRegeneration starts the tool regeneration process
func (r *ToolRegenerations) AddToolRegeneration(cycleID, toolID int64, reason string, user *models.User) (*models.Regeneration, error) {
	if err := ValidateID(cycleID, "cycle"); err != nil {
		return nil, err
	}

	if err := ValidateID(toolID, "tool"); err != nil {
		return nil, err
	}

	if err := ValidateNotNil(user, "user"); err != nil {
		return nil, err
	}

	r.LogOperationWithUser("Starting tool regeneration", createUserInfo(user), fmt.Sprintf("tool: %d", toolID))

	// Update the tool's regeneration status
	r.LogOperation("Setting tool to regenerating status", fmt.Sprintf("tool: %d", toolID))
	if err := r.tools.UpdateRegenerating(toolID, true, user); err != nil {
		return nil, fmt.Errorf("failed to update tool regeneration status: %v", err)
	}

	// Create a new regeneration record
	r.LogOperation("Creating regeneration record", fmt.Sprintf("tool: %d", toolID))
	regeneration, err := r.Add(
		models.NewRegeneration(toolID, cycleID, reason, &user.TelegramID),
		user,
	)
	if err != nil {
		// Undo the tool's regeneration status on failure
		r.log.Error("Failed to create regeneration record, undoing status change for tool: %d", toolID)
		if undoErr := r.tools.UpdateRegenerating(toolID, false, user); undoErr != nil {
			r.log.Error("Failed to undo tool regeneration status: %v", undoErr)
		}
		return nil, err
	}

	r.LogOperation("Started tool regeneration successfully", fmt.Sprintf("tool: %d, regen_id: %d", toolID, regeneration.ID))
	return regeneration, nil
}

// StopToolRegeneration stops the tool regeneration process for the given tool ID
func (r *ToolRegenerations) StopToolRegeneration(toolID int64, user *models.User) error {
	if err := ValidateID(toolID, "tool"); err != nil {
		return err
	}

	if err := ValidateNotNil(user, "user"); err != nil {
		return err
	}

	r.LogOperationWithUser("Stopping tool regeneration", createUserInfo(user), fmt.Sprintf("tool: %d", toolID))

	// Set the tool's regeneration status to false
	if err := r.tools.UpdateRegenerating(toolID, false, user); err != nil {
		return fmt.Errorf("failed to update tool regeneration status: %v", err)
	}

	r.LogOperation("Stopped tool regeneration", fmt.Sprintf("tool: %d", toolID))
	return nil
}

// AbortToolRegeneration aborts the tool regeneration process and removes the regeneration record
func (r *ToolRegenerations) AbortToolRegeneration(toolID int64, user *models.User) error {
	if err := ValidateID(toolID, "tool"); err != nil {
		return err
	}

	if err := ValidateNotNil(user, "user"); err != nil {
		return err
	}

	r.LogOperationWithUser("Aborting tool regeneration", createUserInfo(user), fmt.Sprintf("tool: %d", toolID))

	// First, get the last regeneration record to delete it
	lastRegen, err := r.GetLastRegeneration(toolID)
	if err != nil {
		if !utils.IsNotFoundError(err) {
			return fmt.Errorf("failed to get last regeneration record: %v", err)
		}
		r.LogOperation("No regeneration record found to abort", fmt.Sprintf("tool: %d", toolID))
	} else {
		// Delete the regeneration record
		r.LogOperation("Deleting regeneration record", fmt.Sprintf("id: %d", lastRegen.ID))
		if err := r.Delete(lastRegen.ID); err != nil {
			return fmt.Errorf("failed to delete regeneration record: %v", err)
		}
	}

	// Set the tool's regeneration status to false
	r.LogOperation("Setting tool to non-regenerating status", fmt.Sprintf("tool: %d", toolID))
	if err := r.tools.UpdateRegenerating(toolID, false, user); err != nil {
		return fmt.Errorf("failed to update tool regeneration status: %v", err)
	}

	r.LogOperation("Aborted tool regeneration", fmt.Sprintf("tool: %d", toolID))
	return nil
}

// GetLastRegeneration gets the most recent regeneration for a tool
func (r *ToolRegenerations) GetLastRegeneration(toolID int64) (*models.Regeneration, error) {
	if err := ValidateID(toolID, "tool"); err != nil {
		return nil, err
	}

	r.LogOperation("Getting last regeneration for tool", toolID)

	query := `
		SELECT id, tool_id, cycle_id, reason, performed_by
		FROM tool_regenerations
		WHERE tool_id = ?
		ORDER BY id DESC
		LIMIT 1
	`

	row := r.db.QueryRow(query, toolID)
	regen, err := ScanSingleRow(row, ScanToolRegeneration, "tool_regenerations")
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError(fmt.Sprintf("tool regeneration for tool_id: %d", toolID))
		}
		return nil, err
	}

	r.LogOperation("Found last regeneration", fmt.Sprintf("tool: %d, regen_id: %d", toolID, regen.ID))
	return regen, nil
}

// GetRegenerationHistory gets all regenerations for a tool
func (r *ToolRegenerations) GetRegenerationHistory(toolID int64) ([]*models.Regeneration, error) {
	if err := ValidateID(toolID, "tool"); err != nil {
		return nil, err
	}

	r.LogOperation("Getting regeneration history for tool", toolID)

	query := `
		SELECT id, tool_id, cycle_id, reason, performed_by
		FROM tool_regenerations
		WHERE tool_id = ?
		ORDER BY id DESC
	`

	rows, err := r.db.Query(query, toolID)
	if err != nil {
		return nil, r.HandleSelectError(err, "tool_regenerations")
	}
	defer rows.Close()

	regenerations, err := ScanToolRegenerationsFromRows(rows)
	if err != nil {
		return nil, err
	}

	r.LogOperation("Found regeneration history", fmt.Sprintf("tool: %d, count: %d", toolID, len(regenerations)))
	return regenerations, nil
}
