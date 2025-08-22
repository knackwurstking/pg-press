package database

import (
	"database/sql"
	"fmt"
	"time"
)

const (
	// TODO: Mods missing
	createToolRegenerationsTableQuery = `
		CREATE TABLE IF NOT EXISTS tool_regenerations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tool_id INTEGER NOT NULL,
			regenerated_at DATETIME NOT NULL,
			cycles_at_regeneration INTEGER NOT NULL DEFAULT 0,
			reason TEXT,
			performed_by TEXT,
			notes TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (tool_id) REFERENCES tools(id) ON DELETE CASCADE
		);
		CREATE INDEX IF NOT EXISTS idx_tool_regenerations_tool_id ON tool_regenerations(tool_id);
		CREATE INDEX IF NOT EXISTS idx_tool_regenerations_date ON tool_regenerations(regenerated_at);
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
	t := &ToolRegenerations{
		db:          db,
		feeds:       feeds,
		pressCycles: pressCycles,
	}
	t.init()
	return t
}

func (t *ToolRegenerations) init() {
	if _, err := t.db.Exec(createToolRegenerationsTableQuery); err != nil {
		panic(fmt.Errorf("failed to create tool_regenerations table: %w", err))
	}
}

// Create records a new tool regeneration event
func (t *ToolRegenerations) Create(toolID int64, reason, performedBy, notes string) (*ToolRegeneration, error) {
	// Get current total cycles for the tool before regeneration
	totalCycles, err := t.pressCycles.GetTotalCyclesSinceRegeneration(toolID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get total cycles: %w", err)
	}

	query := `
		INSERT INTO tool_regenerations (tool_id, regenerated_at, cycles_at_regeneration, reason, performed_by, notes)
		VALUES (?, ?, ?, ?, ?, ?)
		RETURNING id, tool_id, regenerated_at, cycles_at_regeneration, reason, performed_by, notes, created_at
	`

	var regen ToolRegeneration
	err = t.db.QueryRow(query,
		toolID,
		time.Now(),
		totalCycles,
		reason,
		performedBy,
		notes,
	).Scan(
		&regen.ID,
		&regen.ToolID,
		&regen.RegeneratedAt,
		&regen.CyclesAtRegeneration,
		&regen.Reason,
		&regen.PerformedBy,
		&regen.Notes,
		&regen.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create regeneration record: %w", err)
	}

	// Create feed entry
	if t.feeds != nil {
		t.feeds.Add(NewFeed(
			FeedTypeToolUpdate,
			&FeedToolUpdate{
				ID:         toolID,
				Tool:       fmt.Sprintf("Werkzeug #%d wurde regeneriert (Grund: %s, Durchgeführt von: %s)", toolID, reason, performedBy),
				ModifiedBy: nil, // System update
			},
		))
	}

	return &regen, nil
}

// GetLastRegeneration gets the most recent regeneration for a tool
func (t *ToolRegenerations) GetLastRegeneration(toolID int64) (*ToolRegeneration, error) {
	query := `
		SELECT id, tool_id, regenerated_at, cycles_at_regeneration, reason, performed_by, notes, created_at
		FROM tool_regenerations
		WHERE tool_id = ?
		ORDER BY regenerated_at DESC
		LIMIT 1
	`

	var regen ToolRegeneration
	err := t.db.QueryRow(query, toolID).Scan(
		&regen.ID,
		&regen.ToolID,
		&regen.RegeneratedAt,
		&regen.CyclesAtRegeneration,
		&regen.Reason,
		&regen.PerformedBy,
		&regen.Notes,
		&regen.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get last regeneration: %w", err)
	}

	return &regen, nil
}

// GetRegenerationHistory gets all regenerations for a tool
func (t *ToolRegenerations) GetRegenerationHistory(toolID int64) ([]*ToolRegeneration, error) {
	query := `
		SELECT id, tool_id, regenerated_at, cycles_at_regeneration, reason, performed_by, notes, created_at
		FROM tool_regenerations
		WHERE tool_id = ?
		ORDER BY regenerated_at DESC
	`

	rows, err := t.db.Query(query, toolID)
	if err != nil {
		return nil, fmt.Errorf("failed to get regeneration history: %w", err)
	}
	defer rows.Close()

	var regenerations []*ToolRegeneration
	for rows.Next() {
		var regen ToolRegeneration
		err := rows.Scan(
			&regen.ID,
			&regen.ToolID,
			&regen.RegeneratedAt,
			&regen.CyclesAtRegeneration,
			&regen.Reason,
			&regen.PerformedBy,
			&regen.Notes,
			&regen.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan regeneration: %w", err)
		}
		regenerations = append(regenerations, &regen)
	}

	return regenerations, nil
}

// GetRegenerationCount gets the total number of regenerations for a tool
func (t *ToolRegenerations) GetRegenerationCount(toolID int64) (int, error) {
	query := `SELECT COUNT(*) FROM tool_regenerations WHERE tool_id = ?`

	var count int
	err := t.db.QueryRow(query, toolID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get regeneration count: %w", err)
	}

	return count, nil
}

// GetRegenerationsBetween gets regenerations for a tool within a time period
func (t *ToolRegenerations) GetRegenerationsBetween(toolID int64, from, to time.Time) ([]*ToolRegeneration, error) {
	query := `
		SELECT id, tool_id, regenerated_at, cycles_at_regeneration, reason, performed_by, notes, created_at
		FROM tool_regenerations
		WHERE tool_id = ? AND regenerated_at BETWEEN ? AND ?
		ORDER BY regenerated_at DESC
	`

	rows, err := t.db.Query(query, toolID, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to get regenerations between dates: %w", err)
	}
	defer rows.Close()

	var regenerations []*ToolRegeneration
	for rows.Next() {
		var regen ToolRegeneration
		err := rows.Scan(
			&regen.ID,
			&regen.ToolID,
			&regen.RegeneratedAt,
			&regen.CyclesAtRegeneration,
			&regen.Reason,
			&regen.PerformedBy,
			&regen.Notes,
			&regen.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan regeneration: %w", err)
		}
		regenerations = append(regenerations, &regen)
	}

	return regenerations, nil
}

// GetAllRegenerations gets all regenerations across all tools
func (t *ToolRegenerations) GetAllRegenerations(limit, offset int) ([]*ToolRegeneration, error) {
	query := `
		SELECT id, tool_id, regenerated_at, cycles_at_regeneration, reason, performed_by, notes, created_at
		FROM tool_regenerations
		ORDER BY regenerated_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := t.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get all regenerations: %w", err)
	}
	defer rows.Close()

	var regenerations []*ToolRegeneration
	for rows.Next() {
		var regen ToolRegeneration
		err := rows.Scan(
			&regen.ID,
			&regen.ToolID,
			&regen.RegeneratedAt,
			&regen.CyclesAtRegeneration,
			&regen.Reason,
			&regen.PerformedBy,
			&regen.Notes,
			&regen.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan regeneration: %w", err)
		}
		regenerations = append(regenerations, &regen)
	}

	return regenerations, nil
}

// Delete removes a regeneration record (should be used carefully)
func (t *ToolRegenerations) Delete(id int64) error {
	query := `DELETE FROM tool_regenerations WHERE id = ?`

	_, err := t.db.Exec(query, id)
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
	query := `
		SELECT
			tool_id,
			COUNT(*) as regen_count,
			MAX(regenerated_at) as last_regen
		FROM tool_regenerations
		GROUP BY tool_id
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
