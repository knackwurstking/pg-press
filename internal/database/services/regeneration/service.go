package regeneration

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/knackwurstking/pgpress/internal/database/dberror"
	"github.com/knackwurstking/pgpress/internal/database/interfaces"
	feedmodels "github.com/knackwurstking/pgpress/internal/database/models/feed"
	regenerationmodels "github.com/knackwurstking/pgpress/internal/database/models/regeneration"
	usermodels "github.com/knackwurstking/pgpress/internal/database/models/user"
	"github.com/knackwurstking/pgpress/internal/database/services/feed"
	"github.com/knackwurstking/pgpress/internal/database/services/presscycle"
	"github.com/knackwurstking/pgpress/internal/logger"
)

type Service struct {
	db                *sql.DB
	feeds             *feed.Service
	pressCyclesHelper *presscycle.Service
}

// Just to make sure it fits TODO: Make it fit
//var _ interfaces.DataOperations[*models.ToolRegeneration] = (*Service)(nil)

func New(db *sql.DB, feeds *feed.Service, pressCyclesHelper *presscycle.Service) *Service {
	// DROP TABLE IF EXISTS tool_regenerations;
	query := `
		CREATE TABLE IF NOT EXISTS tool_regenerations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tool_id INTEGER NOT NULL,
			cycle_id INTEGER NOT NULL,
			reason TEXT,
			performed_by INTEGER,
			FOREIGN KEY (tool_id) REFERENCES tools(id) ON DELETE CASCADE,
			FOREIGN KEY (performed_by) REFERENCES users(id) ON DELETE SET NULL,
			FOREIGN KEY (cycle_id) REFERENCES press_cycles(id) ON DELETE CASCADE
		);
		CREATE INDEX IF NOT EXISTS idx_tool_regenerations_tool_id ON tool_regenerations(tool_id);
		CREATE INDEX IF NOT EXISTS idx_tool_regenerations_cycle_id ON tool_regenerations(cycle_id);
	`

	if _, err := db.Exec(query); err != nil {
		panic(
			dberror.NewDatabaseError(
				"create_table",
				"tool_regenerations",
				"failed to create tool_regenerations table",
				err,
			),
		)
	}

	return &Service{
		db:                db,
		feeds:             feeds,
		pressCyclesHelper: pressCyclesHelper,
	}
}

// Create records a new tool regeneration event
func (t *Service) Create(toolID int64, cycleID int64, reason string, user *usermodels.User) (*regenerationmodels.ToolRegeneration, error) {
	logger.DBToolRegenerations().Info(
		"Creating tool regeneration: tool_id=%d, cycle_id=%d, reason=%s",
		toolID, cycleID, reason,
	)

	var performedBy *int64
	if user != nil {
		performedBy = &user.TelegramID
	}

	regen := regenerationmodels.NewToolRegeneration(toolID, cycleID, reason, performedBy)

	query := `
		INSERT INTO tool_regenerations (tool_id, cycle_id, reason, performed_by)
		VALUES (?, ?, ?, ?)
		RETURNING id, tool_id, cycle_id, reason, performed_by
	`
	regen, err := t.scanToolRegeneration(t.db.QueryRow(query,
		regen.ToolID,
		regen.CycleID,
		regen.Reason,
		performedBy,
	))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, dberror.ErrNotFound
		}
		return nil, dberror.NewDatabaseError("insert", "tool_regenerations",
			"failed to create regeneration record", err)
	}

	// Create feed entry
	if t.feeds != nil {
		t.feeds.Add(feedmodels.NewFeed(
			feedmodels.FeedTypeToolUpdate,
			&feedmodels.FeedToolUpdate{
				ID:         toolID,
				Tool:       fmt.Sprintf("Werkzeug #%d wurde regeneriert (Grund: %s)", toolID, reason),
				ModifiedBy: user,
			},
		))
	}

	return regen, nil
}

// Update updates an existing regeneration record
func (t *Service) Update(regen *regenerationmodels.ToolRegeneration, user *usermodels.User) error {
	logger.DBToolRegenerations().Info("Updating tool regeneration: id=%d", regen.ID)

	var performedBy *int64
	if user != nil {
		performedBy = &user.TelegramID
	}

	query := `
		UPDATE tool_regenerations
		SET cycle_id = ?, reason = ?, performed_by = ?
		WHERE id = ?
	`
	_, err := t.db.Exec(query,
		regen.CycleID,
		regen.Reason,
		performedBy,
		regen.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update regeneration: %w", err)
	}

	return nil
}

// GetLastRegeneration gets the most recent regeneration for a tool
func (t *Service) GetLastRegeneration(toolID int64) (*regenerationmodels.ToolRegeneration, error) {
	logger.DBToolRegenerations().Debug("Getting last regeneration for tool: tool_id=%d", toolID)

	query := `
		SELECT id, tool_id, cycle_id, reason, performed_by
		FROM tool_regenerations
		WHERE tool_id = ?
		ORDER BY id DESC
		LIMIT 1
	`
	regen, err := t.scanToolRegeneration(t.db.QueryRow(query, toolID))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, dberror.ErrNotFound
		}
		return nil, dberror.NewDatabaseError("scan", "tool_regenerations",
			"failed to get last regeneration", err)
	}

	return regen, nil
}

// GetRegenerationHistory gets all regenerations for a tool
func (t *Service) GetRegenerationHistory(toolID int64) ([]*regenerationmodels.ToolRegeneration, error) {
	logger.DBToolRegenerations().Debug("Getting regeneration history for tool: tool_id=%d", toolID)

	query := `
		SELECT id, tool_id, cycle_id, reason, performed_by
		FROM tool_regenerations
		WHERE tool_id = ?
		ORDER BY id DESC
	`
	rows, err := t.db.Query(query, toolID)
	if err != nil {
		return nil, fmt.Errorf("failed to get regeneration history: %w", err)
	}
	defer rows.Close()

	var regenerations []*regenerationmodels.ToolRegeneration
	for rows.Next() {
		regen, err := t.scanToolRegeneration(rows)
		if err != nil {
			return nil, dberror.NewDatabaseError("scan", "tool_regenerations",
				"failed to get regeneration history", err)
		}

		regenerations = append(regenerations, regen)
	}

	return regenerations, nil
}

// GetRegenerationCount gets the total number of regenerations for a tool
func (t *Service) GetRegenerationCount(toolID int64) (int, error) {
	logger.DBToolRegenerations().Debug("Getting regeneration count for tool: tool_id=%d", toolID)

	var count int
	query := `
		SELECT COUNT(*) FROM tool_regenerations WHERE tool_id = ?
	`
	err := t.db.QueryRow(query, toolID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get regeneration count: %w", err)
	}

	return count, nil
}

// GetAllRegenerations gets all regenerations across all tools
func (t *Service) GetAllRegenerations(limit, offset int) ([]*regenerationmodels.ToolRegeneration, error) {
	logger.DBToolRegenerations().Debug("Getting all regenerations: limit=%d, offset=%d", limit, offset)

	query := `
		SELECT id, tool_id, cycle_id, reason, performed_by
		FROM tool_regenerations
		ORDER BY id DESC
		LIMIT ? OFFSET ?
	`
	rows, err := t.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get all regenerations: %w", err)
	}
	defer rows.Close()

	var regenerations []*regenerationmodels.ToolRegeneration
	for rows.Next() {
		regen, err := t.scanToolRegeneration(rows)
		if err != nil {
			return nil, dberror.NewDatabaseError("scan", "tool_regenerations",
				"failed to get all regenerations", err)
		}

		regenerations = append(regenerations, regen)
	}

	return regenerations, nil
}

// Delete removes a regeneration record (should be used carefully)
func (t *Service) Delete(id int64) error {
	logger.DBToolRegenerations().Info("Deleting regeneration record: id=%d", id)

	query := `
		DELETE FROM tool_regenerations WHERE id = ?
	`
	_, err := t.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete regeneration record: %w", err)
	}

	return nil
}

// GetToolsWithMostRegenerations gets tools sorted by regeneration count
func (t *Service) GetToolsWithMostRegenerations(limit int) ([]struct {
	ToolID      int64
	RegCount    int
	LastRegenAt *time.Time
}, error) {
	logger.DBToolRegenerations().Debug("Getting tools with most regenerations: limit=%d", limit)

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
	rows, err := t.db.Query(query, limit)
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

func (t *Service) scanToolRegeneration(scanner interfaces.Scannable) (*regenerationmodels.ToolRegeneration, error) {
	regen := &regenerationmodels.ToolRegeneration{}
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
