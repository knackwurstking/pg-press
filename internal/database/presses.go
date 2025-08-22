package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

const (
	createPressCyclesTableQuery = `
		CREATE TABLE IF NOT EXISTS press_cycles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			press_number INTEGER NOT NULL CHECK(press_number >= 0 AND press_number <= 5),
			tool_id INTEGER NOT NULL,
			from_date DATETIME NOT NULL,
			to_date DATETIME,
			total_cycles INTEGER NOT NULL DEFAULT 0,
			partial_cycles INTEGER NOT NULL DEFAULT 0,
			mods BLOB,
			FOREIGN KEY (tool_id) REFERENCES tools(id)
		);
		CREATE INDEX IF NOT EXISTS idx_press_cycles_tool_id ON press_cycles(tool_id);
		CREATE INDEX IF NOT EXISTS idx_press_cycles_press_number ON press_cycles(press_number);
		CREATE INDEX IF NOT EXISTS idx_press_cycles_dates ON press_cycles(from_date, to_date);
	`
)

type PressNumber int8

func (pn *PressNumber) IsValid() bool {
	return pn == nil
}

type PressCycle struct {
	ID            int64                      `json:"id"`
	PressNumber   PressNumber                `json:"press_number"`
	ToolID        int64                      `json:"tool_id"`
	FromDate      time.Time                  `json:"from_date"`
	ToDate        *time.Time                 `json:"to_date"`
	TotalCycles   int64                      `json:"total_cycles"`
	PartialCycles int64                      `json:"partial_cycles"`
	Mods          []*Modified[PressCycleMod] `json:"mods"`
}

type PressCycleMod struct {
	TotalCycles   int64 `json:"total_cycles"`
	PartialCycles int64 `json:"partial_cycles"`
}

type Presses struct {
	db    *sql.DB
	feeds *Feeds
}

func NewPresses(db *sql.DB, feeds *Feeds) *Presses {
	p := &Presses{
		db:    db,
		feeds: feeds,
	}
	p.init()
	return p
}

func (p *Presses) init() {
	// Create press_cycles table
	if _, err := p.db.Exec(createPressCyclesTableQuery); err != nil {
		panic(fmt.Errorf("failed to create press_cycles table: %w", err))
	}
}

// StartToolUsage records when a tool starts being used on a press
func (p *Presses) StartToolUsage(toolID int64, pressNumber int) (*PressCycle, error) {
	// Validate press number
	if pressNumber < 0 || pressNumber > 5 {
		return nil, fmt.Errorf("invalid press number %d: must be between 0 and 5", pressNumber)
	}

	// First, end any current usage of this tool on other presses
	if err := p.EndToolUsage(toolID); err != nil {
		return nil, fmt.Errorf("failed to end previous tool usage: %w", err)
	}

	// Create new press cycle entry
	query := `
		INSERT INTO press_cycles (press_number, tool_id, from_date, total_cycles, partial_cycles, mods)
		VALUES (?, ?, ?, 0, 0, ?)
		RETURNING id, press_number, tool_id, from_date, to_date, total_cycles, partial_cycles, mods
	`

	var cycle PressCycle
	var toDate sql.NullTime
	var modsData []byte

	// Initialize empty mods array
	modsJSON, _ := json.Marshal([]*Modified[PressCycleMod]{})

	err := p.db.QueryRow(query, pressNumber, toolID, time.Now(), modsJSON).Scan(
		&cycle.ID,
		&cycle.PressNumber,
		&cycle.ToolID,
		&cycle.FromDate,
		&toDate,
		&cycle.TotalCycles,
		&cycle.PartialCycles,
		&modsData,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to start tool usage: %w", err)
	}

	if toDate.Valid {
		cycle.ToDate = &toDate.Time
	}

	// Unmarshal mods
	if err := json.Unmarshal(modsData, &cycle.Mods); err != nil {
		cycle.Mods = []*Modified[PressCycleMod]{}
	}

	// Create feed entry
	if p.feeds != nil {
		p.feeds.Add(NewFeed(
			FeedTypeToolUpdate,
			&FeedToolUpdate{
				ID:         toolID,
				Tool:       fmt.Sprintf("Werkzeug #%d wurde an Presse %d angebracht", toolID, pressNumber),
				ModifiedBy: nil, // System update
			},
		))
	}

	return &cycle, nil
}

// EndToolUsage ends the current usage of a tool on any press
func (p *Presses) EndToolUsage(toolID int64) error {
	query := `
		UPDATE press_cycles
		SET to_date = ?
		WHERE tool_id = ? AND to_date IS NULL
	`

	_, err := p.db.Exec(query, time.Now(), toolID)
	if err != nil {
		return fmt.Errorf("failed to end tool usage: %w", err)
	}

	return nil
}

// UpdateCycles updates the cycle counts for a currently active tool on a press
func (p *Presses) UpdateCycles(toolID int64, totalCycles, partialCycles int64) error {
	// First get the current cycle to preserve and update mods
	current, err := p.GetCurrentToolUsage(toolID)
	if err != nil {
		return fmt.Errorf("failed to get current usage: %w", err)
	}
	if current == nil {
		return fmt.Errorf("no active press cycle found for tool %d", toolID)
	}

	// Add modification record if values changed
	if current.TotalCycles != totalCycles || current.PartialCycles != partialCycles {
		mod := NewModified[PressCycleMod](nil, PressCycleMod{
			TotalCycles:   current.TotalCycles,
			PartialCycles: current.PartialCycles,
		})
		current.Mods = append(current.Mods, mod)
	}

	// Marshal mods
	modsJSON, err := json.Marshal(current.Mods)
	if err != nil {
		return fmt.Errorf("failed to marshal mods: %w", err)
	}

	query := `
		UPDATE press_cycles
		SET total_cycles = ?, partial_cycles = ?, mods = ?
		WHERE tool_id = ? AND to_date IS NULL
	`

	result, err := p.db.Exec(query, totalCycles, partialCycles, modsJSON, toolID)
	if err != nil {
		return fmt.Errorf("failed to update cycles: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("no active press cycle found for tool %d", toolID)
	}

	return nil
}

// GetCurrentToolUsage gets the current active press cycle for a tool
func (p *Presses) GetCurrentToolUsage(toolID int64) (*PressCycle, error) {
	query := `
		SELECT id, press_number, tool_id, from_date, to_date, total_cycles, partial_cycles, mods
		FROM press_cycles
		WHERE tool_id = ? AND to_date IS NULL
		LIMIT 1
	`

	var cycle PressCycle
	var toDate sql.NullTime
	var modsData []byte

	err := p.db.QueryRow(query, toolID).Scan(
		&cycle.ID,
		&cycle.PressNumber,
		&cycle.ToolID,
		&cycle.FromDate,
		&toDate,
		&cycle.TotalCycles,
		&cycle.PartialCycles,
		&modsData,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get current tool usage: %w", err)
	}

	if toDate.Valid {
		cycle.ToDate = &toDate.Time
	}

	// Unmarshal mods
	if err := json.Unmarshal(modsData, &cycle.Mods); err != nil {
		cycle.Mods = []*Modified[PressCycleMod]{}
	}

	return &cycle, nil
}

// GetToolHistory gets the press usage history for a tool
func (p *Presses) GetToolHistory(toolID int64) ([]*PressCycle, error) {
	query := `
		SELECT id, press_number, tool_id, from_date, to_date, total_cycles, partial_cycles, mods
		FROM press_cycles
		WHERE tool_id = ?
		ORDER BY from_date DESC
	`

	rows, err := p.db.Query(query, toolID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tool history: %w", err)
	}
	defer rows.Close()

	var cycles []*PressCycle
	for rows.Next() {
		var cycle PressCycle
		var toDate sql.NullTime
		var modsData []byte

		err := rows.Scan(
			&cycle.ID,
			&cycle.PressNumber,
			&cycle.ToolID,
			&cycle.FromDate,
			&toDate,
			&cycle.TotalCycles,
			&cycle.PartialCycles,
			&modsData,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan press cycle: %w", err)
		}

		if toDate.Valid {
			cycle.ToDate = &toDate.Time
		}

		// Unmarshal mods
		if err := json.Unmarshal(modsData, &cycle.Mods); err != nil {
			cycle.Mods = []*Modified[PressCycleMod]{}
		}

		cycles = append(cycles, &cycle)
	}

	return cycles, nil
}

// GetToolHistorySinceRegeneration gets press cycles since the last tool regeneration
func (p *Presses) GetToolHistorySinceRegeneration(toolID int64, lastRegenerationDate *time.Time) ([]*PressCycle, error) {
	var query string
	var args []interface{}

	if lastRegenerationDate != nil {
		query = `
			SELECT id, press_number, tool_id, from_date, to_date, total_cycles, partial_cycles, mods
			FROM press_cycles
			WHERE tool_id = ? AND from_date >= ?
			ORDER BY from_date DESC
		`
		args = []interface{}{toolID, *lastRegenerationDate}
	} else {
		// If no regeneration date, get all history
		query = `
			SELECT id, press_number, tool_id, from_date, to_date, total_cycles, partial_cycles, mods
			FROM press_cycles
			WHERE tool_id = ?
			ORDER BY from_date DESC
		`
		args = []interface{}{toolID}
	}

	rows, err := p.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get tool history since regeneration: %w", err)
	}
	defer rows.Close()

	var cycles []*PressCycle
	for rows.Next() {
		var cycle PressCycle
		var toDate sql.NullTime
		var modsData []byte

		err := rows.Scan(
			&cycle.ID,
			&cycle.PressNumber,
			&cycle.ToolID,
			&cycle.FromDate,
			&toDate,
			&cycle.TotalCycles,
			&cycle.PartialCycles,
			&modsData,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan press cycle: %w", err)
		}

		if toDate.Valid {
			cycle.ToDate = &toDate.Time
		}

		// Unmarshal mods
		if err := json.Unmarshal(modsData, &cycle.Mods); err != nil {
			cycle.Mods = []*Modified[PressCycleMod]{}
		}

		cycles = append(cycles, &cycle)
	}

	return cycles, nil
}

// GetTotalCyclesSinceRegeneration calculates total cycles since last regeneration
func (p *Presses) GetTotalCyclesSinceRegeneration(toolID int64, lastRegenerationDate *time.Time) (int64, error) {
	var query string
	var args []interface{}

	if lastRegenerationDate != nil {
		query = `
			SELECT COALESCE(SUM(total_cycles), 0)
			FROM press_cycles
			WHERE tool_id = ? AND from_date >= ?
		`
		args = []interface{}{toolID, *lastRegenerationDate}
	} else {
		query = `
			SELECT COALESCE(SUM(total_cycles), 0)
			FROM press_cycles
			WHERE tool_id = ?
		`
		args = []interface{}{toolID}
	}

	var totalCycles int64
	err := p.db.QueryRow(query, args...).Scan(&totalCycles)
	if err != nil {
		return 0, fmt.Errorf("failed to get total cycles: %w", err)
	}

	return totalCycles, nil
}

// GetCurrentToolsOnPress gets all tools currently active on a specific press
func (p *Presses) GetCurrentToolsOnPress(pressNumber int) ([]int64, error) {
	// Validate press number
	if pressNumber < 0 || pressNumber > 5 {
		return nil, fmt.Errorf("invalid press number %d: must be between 0 and 5", pressNumber)
	}

	query := `
		SELECT tool_id
		FROM press_cycles
		WHERE press_number = ? AND to_date IS NULL
	`

	rows, err := p.db.Query(query, pressNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get current tools on press: %w", err)
	}
	defer rows.Close()

	var toolIDs []int64
	for rows.Next() {
		var toolID int64
		if err := rows.Scan(&toolID); err != nil {
			return nil, fmt.Errorf("failed to scan tool ID: %w", err)
		}
		toolIDs = append(toolIDs, toolID)
	}

	return toolIDs, nil
}

// GetPressUtilization gets current tool count for each press (0-5)
func (p *Presses) GetPressUtilization() (map[int][]int64, error) {
	utilization := make(map[int][]int64)

	// Initialize all presses (0-5) with empty slices
	for i := 0; i <= 5; i++ {
		utilization[i] = []int64{}
	}

	// Get all currently active tool assignments
	query := `
		SELECT press_number, tool_id
		FROM press_cycles
		WHERE to_date IS NULL
		ORDER BY press_number, tool_id
	`

	rows, err := p.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get press utilization: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var pressNumber int
		var toolID int64
		if err := rows.Scan(&pressNumber, &toolID); err != nil {
			return nil, fmt.Errorf("failed to scan utilization data: %w", err)
		}
		utilization[pressNumber] = append(utilization[pressNumber], toolID)
	}

	return utilization, nil
}

// MarkToolRegeneration marks when a tool has been regenerated (resets cycles)
func (p *Presses) MarkToolRegeneration(toolID int64) error {
	// End any current usage
	if err := p.EndToolUsage(toolID); err != nil {
		return fmt.Errorf("failed to end tool usage for regeneration: %w", err)
	}

	// Create feed entry
	if p.feeds != nil {
		p.feeds.Add(NewFeed(
			FeedTypeToolUpdate,
			&FeedToolUpdate{
				ID:         toolID,
				Tool:       fmt.Sprintf("Werkzeug #%d wurde regeneriert", toolID),
				ModifiedBy: nil, // System update
			},
		))
	}

	return nil
}

// GetPressCycleStats gets statistics for all presses
func (p *Presses) GetPressCycleStats() (map[int]struct {
	TotalCycles    int64
	ActiveTools    int
	TotalToolsUsed int
}, error) {
	stats := make(map[int]struct {
		TotalCycles    int64
		ActiveTools    int
		TotalToolsUsed int
	})

	// Initialize stats for all presses (0-5)
	for i := 0; i <= 5; i++ {
		stats[i] = struct {
			TotalCycles    int64
			ActiveTools    int
			TotalToolsUsed int
		}{}
	}

	// Get statistics per press
	query := `
		SELECT
			press_number,
			SUM(total_cycles) as total_cycles,
			COUNT(DISTINCT tool_id) as total_tools_used,
			SUM(CASE WHEN to_date IS NULL THEN 1 ELSE 0 END) as active_tools
		FROM press_cycles
		GROUP BY press_number
	`

	rows, err := p.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get press cycle stats: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var pressNumber int
		var totalCycles sql.NullInt64
		var totalToolsUsed, activeTools int

		err := rows.Scan(&pressNumber, &totalCycles, &totalToolsUsed, &activeTools)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stats: %w", err)
		}

		stat := stats[pressNumber]
		if totalCycles.Valid {
			stat.TotalCycles = totalCycles.Int64
		}
		stat.TotalToolsUsed = totalToolsUsed
		stat.ActiveTools = activeTools
		stats[pressNumber] = stat
	}

	return stats, nil
}
