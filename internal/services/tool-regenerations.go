package services

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/knackwurstking/pgpress/internal/interfaces"
	"github.com/knackwurstking/pgpress/pkg/logger"
	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"
)

type ToolRegenerations struct {
	db    *sql.DB
	tools *Tools
	log   *logger.Logger
}

func NewToolRegenerations(db *sql.DB, tools *Tools) *ToolRegenerations {
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

	if _, err := db.Exec(query); err != nil {
		panic(fmt.Errorf("failed to create tool_regenerations table: %v", err))
	}

	return &ToolRegenerations{
		db:    db,
		tools: tools,
		log:   logger.GetComponentLogger("Service: Tool Regenerations"),
	}
}

// Create records a new tool regeneration event
func (r *ToolRegenerations) Add(regeneration *models.Regeneration, user *models.User) (*models.Regeneration, error) {
	r.log.Info("Creating tool regeneration: tool_id=%d, cycle_id=%d, reason=%s", regeneration.ToolID, regeneration.CycleID, regeneration.Reason)

	if user == nil {
		return nil, utils.NewValidationError("user: user is required")
	}

	if regeneration.ToolID <= 0 {
		return nil, utils.NewValidationError("tool_id: tool_id is required")
	}

	if regeneration.CycleID <= 0 {
		return nil, utils.NewValidationError("cycle_id: cycle_id is required")
	}

	query := `
		INSERT INTO tool_regenerations (tool_id, cycle_id, reason, performed_by)
		VALUES (?, ?, ?, ?)
		RETURNING id, tool_id, cycle_id, reason, performed_by
	`

	regeneration, err := r.scanToolRegeneration(
		r.db.QueryRow(query,
			regeneration.ToolID,
			regeneration.CycleID,
			regeneration.Reason,
			user.TelegramID,
		),
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError(fmt.Sprintf("tool regeneration: %d", regeneration.ID))
		}

		return nil, fmt.Errorf("insert error: tool_regenerations: %v", err)
	}

	return regeneration, nil
}

// Update updates an existing regeneration record
func (r *ToolRegenerations) Update(regeneration *models.Regeneration, user *models.User) error {
	r.log.Info("Updating tool regeneration: id=%d", regeneration.ID)

	if user == nil {
		return utils.NewValidationError("user: user is required")
	}

	query := `
		UPDATE tool_regenerations
		SET cycle_id = ?, reason = ?, performed_by = ?
		WHERE id = ?
	`

	_, err := r.db.Exec(query,
		regeneration.CycleID,
		regeneration.Reason,
		user.TelegramID,
		regeneration.ID,
	)
	if err != nil {
		return fmt.Errorf("update error: tool_regenerations: %v", err)
	}

	return nil
}

// Delete removes a regeneration record (should be used carefully)
func (r *ToolRegenerations) Delete(id int64) error {
	r.log.Info("Deleting regeneration record: id=%d", id)

	query := `DELETE FROM tool_regenerations WHERE id = ?`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("delete error: tool_regenerations: %v", err)
	}

	return nil
}

func (r *ToolRegenerations) AddToolRegeneration(cycleID, toolID int64, reason string, user *models.User) (*models.Regeneration, error) {
	r.log.Info("Starting tool regeneration: tool_id=%d", toolID)

	// Update the tool's regeneration status
	r.log.Debug("Updating tool regeneration status to regenerating: tool_id=%d", toolID)
	if err := r.tools.UpdateRegenerating(toolID, true, user); err != nil {
		return nil, fmt.Errorf("failed to update tool regeneration status: %v", err)
	}

	// After this, create a new regeneration record
	r.log.Debug("Creating new regeneration record: tool_id=%d", toolID)
	regeneration, err := r.Add(
		models.NewRegeneration(toolID, cycleID, reason, &user.TelegramID),
		user,
	)
	if err != nil {
		// Undo the tool's regeneration status
		r.log.Error("Failed to create new regeneration record: tool_id=%d", toolID)
		r.log.Debug("Undoing tool regeneration status: tool_id=%d", toolID)
		return nil, r.tools.UpdateRegenerating(toolID, false, user)
	}

	return regeneration, nil
}

// Stop stops the tool regeneration process for the given tool ID
func (r *ToolRegenerations) StopToolRegeneration(toolID int64, user *models.User) error {
	r.log.Info("Stopping tool regeneration: tool_id=%d", toolID)

	if toolID <= 0 {
		return errors.New("invalid tool ID")
	}

	// Just set the tool's regeneration status to false
	r.log.Debug("Undoing tool regeneration status: tool_id=%d", toolID)
	if err := r.tools.UpdateRegenerating(toolID, false, user); err != nil {
		return fmt.Errorf("failed to update tool regeneration status: %v", err)
	}

	r.log.Info("Tool regeneration stopped: tool_id=%d", toolID)
	return nil
}

// AbortToolRegeneration aborts the tool regeneration process and removes the regeneration record
func (r *ToolRegenerations) AbortToolRegeneration(toolID int64, user *models.User) error {
	r.log.Info("Aborting tool regeneration: tool_id=%d", toolID)

	if toolID <= 0 {
		return errors.New("invalid tool ID")
	}

	// First, get the last regeneration record to delete it
	lastRegen, err := r.GetLastRegeneration(toolID)
	if err != nil {
		r.log.Warn("No regeneration record found to abort for tool_id=%d: %v", toolID, err)
		// Continue with status update even if no record found
	} else {
		// Delete the regeneration record
		r.log.Debug("Deleting regeneration record: id=%d", lastRegen.ID)
		if err := r.Delete(lastRegen.ID); err != nil {
			r.log.Error("Failed to delete regeneration record: id=%d, error=%v", lastRegen.ID, err)
			return fmt.Errorf("failed to delete regeneration record: %v", err)
		}
	}

	// Set the tool's regeneration status to false
	r.log.Debug("Undoing tool regeneration status: tool_id=%d", toolID)
	if err := r.tools.UpdateRegenerating(toolID, false, user); err != nil {
		return fmt.Errorf("failed to update tool regeneration status: %v", err)
	}

	r.log.Info("Tool regeneration aborted: tool_id=%d", toolID)
	return nil
}

// GetLastRegeneration gets the most recent regeneration for a tool
func (r *ToolRegenerations) GetLastRegeneration(toolID int64) (*models.Regeneration, error) {
	r.log.Info("Getting last regeneration for tool: tool_id=%d", toolID)

	query := `
		SELECT id, tool_id, cycle_id, reason, performed_by
		FROM tool_regenerations
		WHERE tool_id = ?
		ORDER BY id DESC
		LIMIT 1
	`

	regen, err := r.scanToolRegeneration(r.db.QueryRow(query, toolID))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError(fmt.Sprintf("tool_id=%d", toolID))
		}

		return nil, fmt.Errorf("select error: tool_regenerations: %v", err)
	}

	return regen, nil
}

// GetRegenerationHistory gets all regenerations for a tool
func (r *ToolRegenerations) GetRegenerationHistory(toolID int64) ([]*models.Regeneration, error) {
	r.log.Info("Getting regeneration history for tool: tool_id=%d", toolID)

	query := `
		SELECT id, tool_id, cycle_id, reason, performed_by
		FROM tool_regenerations
		WHERE tool_id = ?
		ORDER BY id DESC
	`
	rows, err := r.db.Query(query, toolID)
	if err != nil {
		return nil, fmt.Errorf("select error: tool_regenerations: %v", err)
	}
	defer rows.Close()

	var regenerations []*models.Regeneration
	for rows.Next() {
		regen, err := r.scanToolRegeneration(rows)
		if err != nil {
			return nil, fmt.Errorf("scan error: tool_regenerations: %v", err)
		}

		regenerations = append(regenerations, regen)
	}

	return regenerations, nil
}

func (r *ToolRegenerations) scanToolRegeneration(scanner interfaces.Scannable) (*models.Regeneration, error) {
	regen := &models.Regeneration{}
	var performedBy sql.NullInt64

	err := scanner.Scan(
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
		return nil, err
	}

	if performedBy.Valid {
		regen.PerformedBy = &performedBy.Int64
	}

	return regen, nil
}
