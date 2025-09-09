package regeneration

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/knackwurstking/pgpress/internal/database/dberror"
	"github.com/knackwurstking/pgpress/internal/database/interfaces"
	feedmodels "github.com/knackwurstking/pgpress/internal/database/models/feed"
	toolmodels "github.com/knackwurstking/pgpress/internal/database/models/tool"
	usermodels "github.com/knackwurstking/pgpress/internal/database/models/user"
	"github.com/knackwurstking/pgpress/internal/database/services/feed"
	"github.com/knackwurstking/pgpress/internal/database/services/tool"
	"github.com/knackwurstking/pgpress/internal/logger"
)

type Service struct {
	db    *sql.DB
	tools *tool.Service
	feeds *feed.Service

	log *logger.Logger
}

func New(db *sql.DB, tools *tool.Service, feeds *feed.Service) *Service {
	//dropQuery := `DROP TABLE IF EXISTS tool_regenerations;`
	//if _, err := db.Exec(dropQuery); err != nil {
	//	panic(fmt.Errorf("failed to drop existing press_cycles table: %w", err))
	//}

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
		panic(fmt.Errorf("failed to create tool_regenerations table: %w", err))
	}

	return &Service{
		db:    db,
		tools: tools,
		feeds: feeds,
		log:   logger.DBToolRegenerations(),
	}
}

// Create records a new tool regeneration event
func (s *Service) Add(regeneration *toolmodels.Regeneration, user *usermodels.User) (*toolmodels.Regeneration, error) {
	s.log.Info("Creating tool regeneration: tool_id=%d, cycle_id=%d, reason=%s", regeneration.ToolID, regeneration.CycleID, regeneration.Reason)

	if user == nil {
		return nil, fmt.Errorf("user is required")
	}

	if regeneration.ToolID <= 0 {
		return nil, fmt.Errorf("tool_id is required")
	}

	if regeneration.CycleID <= 0 {
		return nil, fmt.Errorf("cycle_id is required")
	}

	query := `
		INSERT INTO tool_regenerations (tool_id, cycle_id, reason, performed_by)
		VALUES (?, ?, ?, ?)
		RETURNING id, tool_id, cycle_id, reason, performed_by
	`

	r, err := s.scanToolRegeneration(s.db.QueryRow(query,
		regeneration.ToolID,
		regeneration.CycleID,
		regeneration.Reason,
		user.TelegramID,
	))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, dberror.ErrNotFound
		}

		return nil, fmt.Errorf("failed to create regeneration record: %w", err)
	}

	// Create feed entry
	if s.feeds != nil {
		feed := feedmodels.New(
			"Regenerierung hinzugefügt",
			"Eine neue Regenerierung wurde hinzugefügt.",
			0, // No specific user for regeneration entries
		)
		s.feeds.Add(feed)
	}

	return r, nil
}

// Update updates an existing regeneration record
func (s *Service) Update(regeneration *toolmodels.Regeneration, user *usermodels.User) error {
	s.log.Info("Updating tool regeneration: id=%d", regeneration.ID)

	if user == nil {
		return fmt.Errorf("user is required")
	}

	query := `
		UPDATE tool_regenerations
		SET cycle_id = ?, reason = ?, performed_by = ?
		WHERE id = ?
	`

	_, err := s.db.Exec(query,
		regeneration.CycleID,
		regeneration.Reason,
		user.TelegramID,
		regeneration.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update regeneration: %w", err)
	}

	return nil
}

// Delete removes a regeneration record (should be used carefully)
func (s *Service) Delete(id int64) error {
	s.log.Info("Deleting regeneration record: id=%d", id)

	query := `DELETE FROM tool_regenerations WHERE id = ?`
	_, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete regeneration record: %w", err)
	}

	return nil
}

func (s *Service) Start(cycleID, toolID int64, reason string, user *usermodels.User) (*toolmodels.Regeneration, error) {
	s.log.Info("Starting tool regeneration: tool_id=%d", toolID)

	// Get the tool from the database
	s.log.Debug("Getting tool from database: tool_id=%d", toolID)
	tool, err := s.tools.Get(toolID)
	if err != nil {
		return nil, err
	}

	// Check tools status and verify that this tool is not currently active
	if tool.Status() == toolmodels.StatusRegenerating {
		return nil, fmt.Errorf("tool is already regenerating")
	}

	// Update the tool's regeneration status
	s.log.Debug("Updating tool regeneration status to regenerating: tool_id=%d", toolID)
	if err = s.tools.UpdateRegenerating(toolID, true, user); err != nil {
		return nil, err
	}

	// After this, create a new regeneration record
	s.log.Debug("Creating new regeneration record: tool_id=%d", toolID)
	r, err := s.Add(toolmodels.NewRegeneration(toolID, cycleID, reason, &user.TelegramID), user)
	if err != nil {
		// Undo the tool's regeneration status
		s.log.Error("Failed to create new regeneration record: tool_id=%d", toolID)
		s.log.Debug("Undoing tool regeneration status: tool_id=%d", toolID)
		return nil, s.tools.UpdateRegenerating(toolID, false, user)
	}

	return r, nil
}

// TODO: ...
func (s *Service) Stop(toolID int64) error {
	s.log.Info("Stopping tool regeneration: tool_id=%d", toolID)

	if toolID <= 0 {
		return errors.New("invalid tool ID")
	}

	return errors.New("under construction")
}

// GetLastRegeneration gets the most recent regeneration for a tool
func (s *Service) GetLastRegeneration(toolID int64) (*toolmodels.Regeneration, error) {
	s.log.Info("Getting last regeneration for tool: tool_id=%d", toolID)

	query := `
		SELECT id, tool_id, cycle_id, reason, performed_by
		FROM tool_regenerations
		WHERE tool_id = ?
		ORDER BY id DESC
		LIMIT 1
	`

	regen, err := s.scanToolRegeneration(s.db.QueryRow(query, toolID))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, dberror.ErrNotFound
		}

		return nil, fmt.Errorf("failed to get last regeneration: %w", err)
	}

	return regen, nil
}

// GetRegenerationHistory gets all regenerations for a tool
func (s *Service) GetRegenerationHistory(toolID int64) ([]*toolmodels.Regeneration, error) {
	s.log.Info("Getting regeneration history for tool: tool_id=%d", toolID)

	query := `
		SELECT id, tool_id, cycle_id, reason, performed_by
		FROM tool_regenerations
		WHERE tool_id = ?
		ORDER BY id DESC
	`
	rows, err := s.db.Query(query, toolID)
	if err != nil {
		return nil, fmt.Errorf("failed to get regeneration history: %w", err)
	}
	defer rows.Close()

	var regenerations []*toolmodels.Regeneration
	for rows.Next() {
		regen, err := s.scanToolRegeneration(rows)
		if err != nil {
			return nil, dberror.NewDatabaseError("scan", "tool_regenerations",
				"failed to get regeneration history", err)
		}

		regenerations = append(regenerations, regen)
	}

	return regenerations, nil
}

// GetRegenerationCount gets the total number of regenerations for a tool
func (s *Service) GetRegenerationCount(toolID int64) (int, error) {
	s.log.Info("Getting regeneration count for tool: tool_id=%d", toolID)

	var count int
	query := `
		SELECT COUNT(*) FROM tool_regenerations WHERE tool_id = ?
	`
	err := s.db.QueryRow(query, toolID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get regeneration count: %w", err)
	}

	return count, nil
}

// GetAllRegenerations gets all regenerations across all tools
func (s *Service) GetAllRegenerations(limit, offset int) ([]*toolmodels.Regeneration, error) {
	s.log.Info("Getting all regenerations: limit=%d offset=%d", limit, offset)

	query := `
		SELECT id, tool_id, cycle_id, reason, performed_by
		FROM tool_regenerations
		ORDER BY id DESC
		LIMIT ? OFFSET ?
	`

	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get all regenerations: %w", err)
	}
	defer rows.Close()

	var regenerations []*toolmodels.Regeneration
	for rows.Next() {
		regen, err := s.scanToolRegeneration(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tool regeneration: %w", err)
		}

		regenerations = append(regenerations, regen)
	}

	return regenerations, nil
}

// GetToolsWithMostRegenerations gets tools sorted by regeneration count
func (s *Service) GetToolsWithMostRegenerations(limit int) ([]struct {
	ToolID      int64
	RegCount    int
	LastRegenAt *time.Time
}, error) {
	s.log.Info("Getting tools with most regenerations: limit=%d", limit)

	query := `
		SELECT
			t.tool_id,
			COUNT(t.id) as regen_count,
			MAX(p.date) as last_regen
		FROM tool_regenerations t
		JOIN press_cycles p ON t.cycle_id = p.id
		GROUP BY t.tool_id
		ORDER BY regen_count DESC
		LIMIT ?
	`

	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get tools with most regenerations: %w", err)
	}
	defer rows.Close()

	var results []struct {
		ToolID      int64
		RegCount    int
		LastRegenAt *time.Time
	}

	for rows.Next() {
		var result struct {
			ToolID      int64
			RegCount    int
			LastRegenAt *time.Time
		}
		var lastRegen sql.NullTime

		err := rows.Scan(&result.ToolID, &result.RegCount, &lastRegen)
		if err != nil {
			return nil, fmt.Errorf("failed to scan result: %w", err)
		}

		if lastRegen.Valid {
			result.LastRegenAt = &lastRegen.Time
		}

		results = append(results, result)
	}

	return results, nil
}

func (t *Service) scanToolRegeneration(scanner interfaces.Scannable) (*toolmodels.Regeneration, error) {
	regen := &toolmodels.Regeneration{}
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
