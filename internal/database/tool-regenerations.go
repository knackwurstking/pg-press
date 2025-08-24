package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/knackwurstking/pgpress/internal/logger"
)

const (
	createToolRegenerationsTableQuery = `
		CREATE TABLE IF NOT EXISTS tool_regenerations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tool_id INTEGER NOT NULL,
			regenerated_at DATETIME NOT NULL,
			cycles_at_regeneration INTEGER NOT NULL DEFAULT 0,
			reason TEXT,
			performed_by INTEGER,
			FOREIGN KEY (tool_id) REFERENCES tools(id) ON DELETE CASCADE,
			FOREIGN KEY (performed_by) REFERENCES users(id) ON DELETE SET NULL
		);
		CREATE INDEX IF NOT EXISTS idx_tool_regenerations_tool_id ON tool_regenerations(tool_id);
		CREATE INDEX IF NOT EXISTS idx_tool_regenerations_date ON tool_regenerations(regenerated_at);
	`

	insertToolRegenerationQuery = `
		INSERT INTO tool_regenerations (tool_id, regenerated_at, cycles_at_regeneration, reason, performed_by)
		VALUES (?, ?, ?, ?, ?)
		RETURNING id, tool_id, regenerated_at, cycles_at_regeneration, reason, performed_by
	`

	selectLastRegenerationQuery = `
		SELECT id, tool_id, regenerated_at, cycles_at_regeneration, reason, performed_by
		FROM tool_regenerations
		WHERE tool_id = ?
		ORDER BY regenerated_at DESC
		LIMIT 1
	`

	selectRegenerationHistoryQuery = `
		SELECT id, tool_id, regenerated_at, cycles_at_regeneration, reason, performed_by
		FROM tool_regenerations
		WHERE tool_id = ?
		ORDER BY regenerated_at DESC
	`

	selectRegenerationCountQuery = `
		SELECT COUNT(*) FROM tool_regenerations WHERE tool_id = ?
	`

	selectRegenerationsBetweenQuery = `
		SELECT id, tool_id, regenerated_at, cycles_at_regeneration, reason, performed_by
		FROM tool_regenerations
		WHERE tool_id = ? AND regenerated_at BETWEEN ? AND ?
		ORDER BY regenerated_at DESC
	`

	selectAllRegenerationsQuery = `
		SELECT id, tool_id, regenerated_at, cycles_at_regeneration, reason, performed_by
		FROM tool_regenerations
		ORDER BY regenerated_at DESC
		LIMIT ? OFFSET ?
	`

	deleteRegenerationQuery = `
		DELETE FROM tool_regenerations WHERE id = ?
	`

	selectToolsWithMostRegenerationsQuery = `
		SELECT
			tool_id,
			COUNT(*) as regen_count,
			MAX(regenerated_at) as last_regen
		FROM tool_regenerations
		GROUP BY tool_id
		ORDER BY regen_count DESC
		LIMIT ?
	`

	updateToolRegenerationQuery = `
		UPDATE tool_regenerations
		SET cycles_at_regeneration = ?, reason = ?, performed_by = ?
		WHERE id = ?
	`
)

// ToolRegenerations handles tool regeneration tracking
type ToolRegenerations struct {
	db          *sql.DB
	feeds       *Feeds
	pressCycles *PressCycles
}

// NewToolRegenerations creates a new ToolRegenerations instance
func NewToolRegenerations(db *sql.DB, feeds *Feeds, pressCycles *PressCycles) *ToolRegenerations {
	if _, err := db.Exec(createToolRegenerationsTableQuery); err != nil {
		panic(
			NewDatabaseError(
				"create_table",
				"tool_regenerations",
				"failed to create tool_regenerations table",
				err,
			),
		)
	}

	return &ToolRegenerations{
		db:          db,
		feeds:       feeds,
		pressCycles: pressCycles,
	}
}

// Create records a new tool regeneration event
func (t *ToolRegenerations) Create(toolID int64, reason string, user *User) (*ToolRegeneration, error) {
	logger.DBToolRegenerations().Info(
		"Creating tool regeneration: tool_id=%d, reason=%s",
		toolID, reason,
	)

	// Get current total cycles for the tool before regeneration
	totalCycles, err := t.pressCycles.GetTotalCyclesSinceRegeneration(toolID, nil)
	if err != nil {
		return nil, NewDatabaseError("insert", "tool_regenerations",
			"failed to get total cycles", err)
	}

	var performedBy *int64
	if user != nil {
		performedBy = &user.ID
	}

	regen := NewToolRegeneration(toolID, time.Now(), totalCycles, reason, performedBy)

	regen, err = t.scanFromRow(t.db.QueryRow(insertToolRegenerationQuery,
		regen.ToolID,
		regen.RegeneratedAt,
		regen.CyclesAtRegeneration,
		regen.Reason,
		performedBy,
	))
	if err != nil {
		return nil, NewDatabaseError("insert", "tool_regenerations",
			"failed to create regeneration record", err)
	}

	// Create feed entry
	if t.feeds != nil {
		t.feeds.Add(NewFeed(
			FeedTypeToolUpdate,
			&FeedToolUpdate{
				ID:         toolID,
				Tool:       fmt.Sprintf("Werkzeug #%d wurde regeneriert (Grund: %s)", toolID, reason),
				ModifiedBy: user,
			},
		))
	}

	return regen, nil
}

// Update updates an existing regeneration record
func (t *ToolRegenerations) Update(regen *ToolRegeneration, user *User) error {
	logger.DBToolRegenerations().Info("Updating tool regeneration: id=%d", regen.ID)

	var performedBy *int64
	if user != nil {
		performedBy = &user.ID
	}

	_, err := t.db.Exec(updateToolRegenerationQuery,
		regen.CyclesAtRegeneration,
		regen.Reason,
		performedBy,
		regen.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update regeneration: %w", err)
	}

	return nil
}

// getByID is an internal helper to get a regeneration by ID
func (t *ToolRegenerations) getByID(id int64) (*ToolRegeneration, error) {
	query := `
		SELECT id, tool_id, regenerated_at, cycles_at_regeneration, reason, performed_by
		FROM tool_regenerations
		WHERE id = ?
	`

	regen, err := t.scanFromRow(t.db.QueryRow(query, id))
	if err != nil {
		return nil, NewDatabaseError("scan", "tool_regenerations",
			"failed to get regeneration by ID", err)
	}

	return regen, nil
}

// GetLastRegeneration gets the most recent regeneration for a tool
func (t *ToolRegenerations) GetLastRegeneration(toolID int64) (*ToolRegeneration, error) {
	logger.DBToolRegenerations().Debug("Getting last regeneration for tool: tool_id=%d", toolID)

	regen, err := t.scanFromRow(t.db.QueryRow(selectLastRegenerationQuery, toolID))
	if err != nil {
		return nil, NewDatabaseError("scan", "tool_regenerations",
			"failed to get last regeneration", err)
	}

	return regen, nil
}

// GetRegenerationHistory gets all regenerations for a tool
func (t *ToolRegenerations) GetRegenerationHistory(toolID int64) ([]*ToolRegeneration, error) {
	logger.DBToolRegenerations().Debug("Getting regeneration history for tool: tool_id=%d", toolID)

	rows, err := t.db.Query(selectRegenerationHistoryQuery, toolID)
	if err != nil {
		return nil, fmt.Errorf("failed to get regeneration history: %w", err)
	}
	defer rows.Close()

	var regenerations []*ToolRegeneration
	for rows.Next() {
		regen, err := t.scanFromRows(rows)
		if err != nil {
			return nil, NewDatabaseError("scan", "tool_regenerations",
				"failed to get regeneration history", err)
		}

		regenerations = append(regenerations, regen)
	}

	return regenerations, nil
}

// GetRegenerationCount gets the total number of regenerations for a tool
func (t *ToolRegenerations) GetRegenerationCount(toolID int64) (int, error) {
	logger.DBToolRegenerations().Debug("Getting regeneration count for tool: tool_id=%d", toolID)

	var count int
	err := t.db.QueryRow(selectRegenerationCountQuery, toolID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get regeneration count: %w", err)
	}

	return count, nil
}

// GetRegenerationsBetween gets regenerations for a tool within a time period
func (t *ToolRegenerations) GetRegenerationsBetween(toolID int64, from, to time.Time) ([]*ToolRegeneration, error) {
	logger.DBToolRegenerations().Debug("Getting regenerations between dates for tool: tool_id=%d, from=%s, to=%s",
		toolID, from.Format("2006-01-02"), to.Format("2006-01-02"))

	rows, err := t.db.Query(selectRegenerationsBetweenQuery, toolID, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to get regenerations between dates: %w", err)
	}
	defer rows.Close()

	var regenerations []*ToolRegeneration
	for rows.Next() {
		regen, err := t.scanFromRows(rows)
		if err != nil {
			return nil, NewDatabaseError("scan", "tool_regenerations",
				"failed to get regenerations between", err)
		}

		regenerations = append(regenerations, regen)
	}

	return regenerations, nil
}

// GetAllRegenerations gets all regenerations across all tools
func (t *ToolRegenerations) GetAllRegenerations(limit, offset int) ([]*ToolRegeneration, error) {
	logger.DBToolRegenerations().Debug("Getting all regenerations: limit=%d, offset=%d", limit, offset)

	rows, err := t.db.Query(selectAllRegenerationsQuery, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get all regenerations: %w", err)
	}
	defer rows.Close()

	var regenerations []*ToolRegeneration
	for rows.Next() {
		regen, err := t.scanFromRows(rows)
		if err != nil {
			return nil, NewDatabaseError("scan", "tool_regenerations",
				"failed to get all regenerations", err)
		}

		regenerations = append(regenerations, regen)
	}

	return regenerations, nil
}

// Delete removes a regeneration record (should be used carefully)
func (t *ToolRegenerations) Delete(id int64) error {
	logger.DBToolRegenerations().Info("Deleting regeneration record: id=%d", id)

	_, err := t.db.Exec(deleteRegenerationQuery, id)
	if err != nil {
		return fmt.Errorf("failed to delete regeneration record: %w", err)
	}

	return nil
}

// GetToolsWithMostRegenerations gets tools sorted by regeneration count
func (t *ToolRegenerations) GetToolsWithMostRegenerations(limit int) ([]struct {
	ToolID      int64
	RegCount    int
	LastRegenAt *time.Time
}, error) {
	logger.DBToolRegenerations().Debug("Getting tools with most regenerations: limit=%d", limit)

	rows, err := t.db.Query(selectToolsWithMostRegenerationsQuery, limit)
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

func (t *ToolRegenerations) scanFromRow(row *sql.Row) (*ToolRegeneration, error) {
	var regen ToolRegeneration
	var performedBy sql.NullInt64

	err := row.Scan(
		&regen.ID,
		&regen.ToolID,
		&regen.RegeneratedAt,
		&regen.CyclesAtRegeneration,
		&regen.Reason,
		&performedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	if performedBy.Valid {
		regen.PerformedBy = &performedBy.Int64
	}

	return &regen, nil
}

func (t *ToolRegenerations) scanFromRows(rows *sql.Rows) (*ToolRegeneration, error) {
	var regen ToolRegeneration
	var performedBy sql.NullInt64

	err := rows.Scan(
		&regen.ID,
		&regen.ToolID,
		&regen.RegeneratedAt,
		&regen.CyclesAtRegeneration,
		&regen.Reason,
		&performedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan from rows: %w", err)
	}

	if performedBy.Valid {
		regen.PerformedBy = &performedBy.Int64
	}

	return &regen, nil
}
