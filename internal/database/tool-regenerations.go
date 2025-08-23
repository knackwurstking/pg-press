package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

const (
	createToolRegenerationsTableQuery = `
		CREATE TABLE IF NOT EXISTS tool_regenerations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tool_id INTEGER NOT NULL,
			regenerated_at DATETIME NOT NULL,
			cycles_at_regeneration INTEGER NOT NULL DEFAULT 0,
			reason TEXT,
			performed_by TEXT,
			notes TEXT,
			mods BLOB,
			FOREIGN KEY (tool_id) REFERENCES tools(id) ON DELETE CASCADE
		);
		CREATE INDEX IF NOT EXISTS idx_tool_regenerations_tool_id ON tool_regenerations(tool_id);
		CREATE INDEX IF NOT EXISTS idx_tool_regenerations_date ON tool_regenerations(regenerated_at);
	`

	insertToolRegenerationQuery = `
		INSERT INTO tool_regenerations (tool_id, regenerated_at, cycles_at_regeneration, reason, performed_by, notes, mods)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		RETURNING id, tool_id, regenerated_at, cycles_at_regeneration, reason, performed_by, notes, mods
	`

	selectLastRegenerationQuery = `
		SELECT id, tool_id, regenerated_at, cycles_at_regeneration, reason, performed_by, notes, mods
		FROM tool_regenerations
		WHERE tool_id = ?
		ORDER BY regenerated_at DESC
		LIMIT 1
	`

	selectRegenerationHistoryQuery = `
		SELECT id, tool_id, regenerated_at, cycles_at_regeneration, reason, performed_by, notes, mods
		FROM tool_regenerations
		WHERE tool_id = ?
		ORDER BY regenerated_at DESC
	`

	selectRegenerationCountQuery = `
		SELECT COUNT(*) FROM tool_regenerations WHERE tool_id = ?
	`

	selectRegenerationsBetweenQuery = `
		SELECT id, tool_id, regenerated_at, cycles_at_regeneration, reason, performed_by, notes, mods
		FROM tool_regenerations
		WHERE tool_id = ? AND regenerated_at BETWEEN ? AND ?
		ORDER BY regenerated_at DESC
	`

	selectAllRegenerationsQuery = `
		SELECT id, tool_id, regenerated_at, cycles_at_regeneration, reason, performed_by, notes, mods
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
		SET cycles_at_regeneration = ?, reason = ?, performed_by = ?, notes = ?, mods = ?
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

	// Initialize empty mods array
	modsJSON, _ := json.Marshal([]*Modified[ToolRegenerationMod]{})

	var regen ToolRegeneration
	var modsData []byte

	err = t.db.QueryRow(insertToolRegenerationQuery,
		toolID,
		time.Now(),
		totalCycles,
		reason,
		performedBy,
		notes,
		modsJSON,
	).Scan(
		&regen.ID,
		&regen.ToolID,
		&regen.RegeneratedAt,
		&regen.CyclesAtRegeneration,
		&regen.Reason,
		&regen.PerformedBy,
		&regen.Notes,
		&modsData,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create regeneration record: %w", err)
	}

	// Unmarshal mods
	if err := json.Unmarshal(modsData, &regen.Mods); err != nil {
		regen.Mods = []*Modified[ToolRegenerationMod]{}
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

// Update updates an existing regeneration record (mainly for mods)
func (t *ToolRegenerations) Update(regen *ToolRegeneration) error {
	// Add modification record if values changed
	existingRegen, err := t.getByID(regen.ID)
	if err != nil {
		return fmt.Errorf("failed to get existing regeneration: %w", err)
	}

	if existingRegen.CyclesAtRegeneration != regen.CyclesAtRegeneration ||
		existingRegen.Reason != regen.Reason ||
		existingRegen.PerformedBy != regen.PerformedBy ||
		existingRegen.Notes != regen.Notes {
		mod := NewModified(nil, ToolRegenerationMod{
			ToolID:               existingRegen.ToolID,
			RegeneratedAt:        existingRegen.RegeneratedAt,
			CyclesAtRegeneration: existingRegen.CyclesAtRegeneration,
			Reason:               existingRegen.Reason,
			PerformedBy:          existingRegen.PerformedBy,
			Notes:                existingRegen.Notes,
		})
		// Prepend new mod to keep most recent first
		regen.Mods = append([]*Modified[ToolRegenerationMod]{mod}, regen.Mods...)
	}

	// Marshal mods
	modsJSON, err := json.Marshal(regen.Mods)
	if err != nil {
		return fmt.Errorf("failed to marshal mods: %w", err)
	}

	_, err = t.db.Exec(updateToolRegenerationQuery,
		regen.CyclesAtRegeneration,
		regen.Reason,
		regen.PerformedBy,
		regen.Notes,
		modsJSON,
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
		SELECT id, tool_id, regenerated_at, cycles_at_regeneration, reason, performed_by, notes, mods
		FROM tool_regenerations
		WHERE id = ?
	`

	var regen ToolRegeneration
	var modsData []byte

	err := t.db.QueryRow(query, id).Scan(
		&regen.ID,
		&regen.ToolID,
		&regen.RegeneratedAt,
		&regen.CyclesAtRegeneration,
		&regen.Reason,
		&regen.PerformedBy,
		&regen.Notes,
		&modsData,
	)

	if err != nil {
		return nil, err
	}

	// Unmarshal mods
	if err := json.Unmarshal(modsData, &regen.Mods); err != nil {
		regen.Mods = []*Modified[ToolRegenerationMod]{}
	}

	return &regen, nil
}

// GetLastRegeneration gets the most recent regeneration for a tool
func (t *ToolRegenerations) GetLastRegeneration(toolID int64) (*ToolRegeneration, error) {
	var regen ToolRegeneration
	var modsData []byte

	err := t.db.QueryRow(selectLastRegenerationQuery, toolID).Scan(
		&regen.ID,
		&regen.ToolID,
		&regen.RegeneratedAt,
		&regen.CyclesAtRegeneration,
		&regen.Reason,
		&regen.PerformedBy,
		&regen.Notes,
		&modsData,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get last regeneration: %w", err)
	}

	// Unmarshal mods
	if err := json.Unmarshal(modsData, &regen.Mods); err != nil {
		regen.Mods = []*Modified[ToolRegenerationMod]{}
	}

	return &regen, nil
}

// GetRegenerationHistory gets all regenerations for a tool
func (t *ToolRegenerations) GetRegenerationHistory(toolID int64) ([]*ToolRegeneration, error) {
	rows, err := t.db.Query(selectRegenerationHistoryQuery, toolID)
	if err != nil {
		return nil, fmt.Errorf("failed to get regeneration history: %w", err)
	}
	defer rows.Close()

	var regenerations []*ToolRegeneration
	for rows.Next() {
		var regen ToolRegeneration
		var modsData []byte

		err := rows.Scan(
			&regen.ID,
			&regen.ToolID,
			&regen.RegeneratedAt,
			&regen.CyclesAtRegeneration,
			&regen.Reason,
			&regen.PerformedBy,
			&regen.Notes,
			&modsData,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan regeneration: %w", err)
		}

		// Unmarshal mods
		if err := json.Unmarshal(modsData, &regen.Mods); err != nil {
			regen.Mods = []*Modified[ToolRegenerationMod]{}
		}

		regenerations = append(regenerations, &regen)
	}

	return regenerations, nil
}

// GetByToolID gets all regenerations for a specific tool (alias for GetRegenerationHistory)
func (t *ToolRegenerations) GetByToolID(toolID int64) ([]*ToolRegeneration, error) {
	return t.GetRegenerationHistory(toolID)
}

// GetRegenerationCount gets the total number of regenerations for a tool
func (t *ToolRegenerations) GetRegenerationCount(toolID int64) (int, error) {
	var count int
	err := t.db.QueryRow(selectRegenerationCountQuery, toolID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get regeneration count: %w", err)
	}

	return count, nil
}

// GetRegenerationsBetween gets regenerations for a tool within a time period
func (t *ToolRegenerations) GetRegenerationsBetween(toolID int64, from, to time.Time) ([]*ToolRegeneration, error) {
	rows, err := t.db.Query(selectRegenerationsBetweenQuery, toolID, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to get regenerations between dates: %w", err)
	}
	defer rows.Close()

	var regenerations []*ToolRegeneration
	for rows.Next() {
		var regen ToolRegeneration
		var modsData []byte

		err := rows.Scan(
			&regen.ID,
			&regen.ToolID,
			&regen.RegeneratedAt,
			&regen.CyclesAtRegeneration,
			&regen.Reason,
			&regen.PerformedBy,
			&regen.Notes,
			&modsData,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan regeneration: %w", err)
		}

		// Unmarshal mods
		if err := json.Unmarshal(modsData, &regen.Mods); err != nil {
			regen.Mods = []*Modified[ToolRegenerationMod]{}
		}

		regenerations = append(regenerations, &regen)
	}

	return regenerations, nil
}

// GetAllRegenerations gets all regenerations across all tools
func (t *ToolRegenerations) GetAllRegenerations(limit, offset int) ([]*ToolRegeneration, error) {
	rows, err := t.db.Query(selectAllRegenerationsQuery, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get all regenerations: %w", err)
	}
	defer rows.Close()

	var regenerations []*ToolRegeneration
	for rows.Next() {
		var regen ToolRegeneration
		var modsData []byte

		err := rows.Scan(
			&regen.ID,
			&regen.ToolID,
			&regen.RegeneratedAt,
			&regen.CyclesAtRegeneration,
			&regen.Reason,
			&regen.PerformedBy,
			&regen.Notes,
			&modsData,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan regeneration: %w", err)
		}

		// Unmarshal mods
		if err := json.Unmarshal(modsData, &regen.Mods); err != nil {
			regen.Mods = []*Modified[ToolRegenerationMod]{}
		}

		regenerations = append(regenerations, &regen)
	}

	return regenerations, nil
}

// Delete removes a regeneration record (should be used carefully)
func (t *ToolRegenerations) Delete(id int64) error {
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
