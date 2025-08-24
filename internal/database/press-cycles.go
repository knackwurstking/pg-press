package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

const (
	// TODO: Add more example data for testing
	createPressCyclesTableQuery = `
		DROP TABLE IF EXISTS press_cycles;
		CREATE TABLE IF NOT EXISTS press_cycles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			press_number INTEGER NOT NULL CHECK(press_number >= 0 AND press_number <= 5),
			tool_id INTEGER NOT NULL,
			from_date DATETIME NOT NULL,
			to_date DATETIME,
			total_cycles INTEGER NOT NULL DEFAULT 0,
			partial_cycles INTEGER NOT NULL DEFAULT 0,
			mods BLOB NOT NULL,
			FOREIGN KEY (tool_id) REFERENCES tools(id)
		);
		CREATE INDEX IF NOT EXISTS idx_press_cycles_tool_id ON press_cycles(tool_id);
		CREATE INDEX IF NOT EXISTS idx_press_cycles_press_number ON press_cycles(press_number);
		CREATE INDEX IF NOT EXISTS idx_press_cycles_dates ON press_cycles(from_date, to_date);
		INSERT INTO press_cycles (press_number, tool_id, from_date, to_date, total_cycles, partial_cycles, mods)
		VALUES
			(0, 1, '2023-01-01', NULL,     0, 	 0, '[]'),
			(0, 2, '2023-01-01', NULL,     0, 	 0, '[]'),
			(0, 1, '2023-02-01', NULL,  1000, 1000, '[]'),
			(0, 2, '2023-02-01', NULL,  1000, 1000, '[]'),
			(0, 1, '2023-03-01', NULL,  2000, 1000, '[]'),
			(0, 2, '2023-03-01', NULL,  2000, 1000, '[]'),
			(0, 1, '2023-04-01', NULL,  3000, 1000, '[]'),
			(0, 2, '2023-04-01', NULL,  3000, 1000, '[]'),
			(0, 1, '2023-05-01', NULL,  4000, 1000, '[]'),
			(0, 2, '2023-05-01', NULL,  4000, 1000, '[]'),
			(0, 1, '2023-06-01', NULL,  5000, 1000, '[]'),
			(0, 2, '2023-06-01', NULL,  5000, 1000, '[]'),
			(0, 1, '2023-07-01', NULL,  6000, 1000, '[]'),
			(0, 2, '2023-07-01', NULL,  6000, 1000, '[]'),
			(0, 1, '2023-08-01', NULL,  7000, 1000, '[]'),
			(0, 2, '2023-08-01', NULL,  7000, 1000, '[]'),
			(0, 1, '2023-09-01', NULL,  8000, 1000, '[]'),
			(0, 2, '2023-09-01', NULL,  8000, 1000, '[]'),
			(0, 1, '2023-10-01', NULL,  9000, 1000, '[]'),
			(0, 2, '2023-10-01', NULL,  9000, 1000, '[]'),
			(0, 1, '2023-11-01', NULL, 10000, 1000, '[]'),
			(0, 2, '2023-11-01', NULL, 10000, 1000, '[]');
	`

	insertPressCycleQuery = `
		INSERT INTO press_cycles (press_number, tool_id, from_date, total_cycles, partial_cycles, mods)
		VALUES (?, ?, ?, 0, 0, ?)
		RETURNING id, press_number, tool_id, from_date, to_date, total_cycles, partial_cycles, mods
	`

	endToolUsageQuery = `
		UPDATE press_cycles
		SET to_date = ?
		WHERE tool_id = ? AND to_date IS NULL
	`

	updatePressCyclesQuery = `
		UPDATE press_cycles
		SET total_cycles = ?, partial_cycles = ?, mods = ?
		WHERE tool_id = ? AND to_date IS NULL
	`

	selectCurrentToolUsageQuery = `
		SELECT id, press_number, tool_id, from_date, to_date, total_cycles, partial_cycles, mods
		FROM press_cycles
		WHERE tool_id = ? AND to_date IS NULL
		LIMIT 1
	`

	selectToolHistoryQuery = `
		SELECT id, press_number, tool_id, from_date, to_date, total_cycles, partial_cycles, mods
		FROM press_cycles
		WHERE tool_id = ?
		ORDER BY from_date DESC
	`

	selectToolHistorySinceRegenerationQuery = `
		SELECT id, press_number, tool_id, from_date, to_date, total_cycles, partial_cycles, mods
		FROM press_cycles
		WHERE tool_id = ? AND from_date >= ?
		ORDER BY from_date DESC
	`

	selectAllToolHistoryQuery = `
		SELECT id, press_number, tool_id, from_date, to_date, total_cycles, partial_cycles, mods
		FROM press_cycles
		WHERE tool_id = ?
		ORDER BY from_date DESC
	`

	selectTotalCyclesSinceRegenerationQuery = `
		SELECT COALESCE(SUM(partial_cycles), 0)
		FROM press_cycles
		WHERE tool_id = ? AND from_date >= ?
	`

	selectTotalCyclesAllTimeQuery = `
		SELECT COALESCE(SUM(partial_cycles), 0)
		FROM press_cycles
		WHERE tool_id = ?
	`

	selectCurrentToolsOnPressQuery = `
		SELECT tool_id
		FROM press_cycles
		WHERE press_number = ? AND to_date IS NULL
	`

	selectPressCyclesForPressQuery = `
		SELECT id, press_number, tool_id, from_date, to_date, total_cycles, partial_cycles, mods
		FROM press_cycles
		WHERE press_number = ?
		ORDER BY from_date DESC
	`

	selectActivePressCyclesForPressQuery = `
		SELECT id, press_number, tool_id, from_date, to_date, total_cycles, partial_cycles, mods
		FROM press_cycles
		WHERE press_number = ? AND to_date IS NULL
		ORDER BY from_date DESC
	`

	selectPressUtilizationQuery = `
		SELECT press_number, tool_id
		FROM press_cycles
		WHERE to_date IS NULL
		ORDER BY press_number, tool_id
	`

	selectPressCycleStatsQuery = `
		SELECT
			press_number,
			SUM(total_cycles) as total_cycles,
			COUNT(DISTINCT tool_id) as total_tools_used,
			SUM(CASE WHEN to_date IS NULL THEN 1 ELSE 0 END) as active_tools
		FROM press_cycles
		GROUP BY press_number
	`
)

// PressCycles manages press cycle data and operations
type PressCycles struct {
	db    *sql.DB
	feeds *Feeds
}

func NewPressCycles(db *sql.DB, feeds *Feeds) *PressCycles {
	p := &PressCycles{
		db:    db,
		feeds: feeds,
	}
	p.init()
	return p
}

func (p *PressCycles) init() {
	// Create press_cycles table
	if _, err := p.db.Exec(createPressCyclesTableQuery); err != nil {
		panic(fmt.Errorf("failed to create press_cycles table: %w", err))
	}
}

// StartToolUsage records when a tool starts being used on a press
func (p *PressCycles) StartToolUsage(toolID int64, pressNumber PressNumber) (*PressCycle, error) {
	// Validate press number
	if pressNumber < MinPressNumber || pressNumber > MaxPressNumber {
		return nil, fmt.Errorf("invalid press number %d: must be between 0 and 5", pressNumber)
	}

	// First, end any current usage of this tool on other presses
	if err := p.EndToolUsage(toolID); err != nil {
		return nil, fmt.Errorf("failed to end previous tool usage: %w", err)
	}

	// Create new press cycle entry

	var cycle PressCycle
	var toDate sql.NullTime
	var modsData []byte

	// Initialize empty mods array
	modsJSON, _ := json.Marshal([]*Modified[PressCycleMod]{})

	err := p.db.QueryRow(insertPressCycleQuery, pressNumber, toolID, time.Now(), modsJSON).Scan(
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
func (p *PressCycles) EndToolUsage(toolID int64) error {
	_, err := p.db.Exec(endToolUsageQuery, time.Now(), toolID)
	if err != nil {
		return fmt.Errorf("failed to end tool usage: %w", err)
	}

	return nil
}

// UpdateCycles updates the cycle counts for a currently active tool on a press
func (p *PressCycles) UpdateCycles(toolID int64, totalCycles, partialCycles int64) error {
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
		mod := NewModified(nil, PressCycleMod{
			PressNumber:   current.PressNumber,
			ToolID:        current.ToolID,
			FromDate:      current.FromDate,
			ToDate:        current.ToDate,
			TotalCycles:   current.TotalCycles,
			PartialCycles: current.PartialCycles,
		})
		current.Mods = append([]*Modified[PressCycleMod]{mod}, current.Mods...)
	}

	// Marshal mods
	modsJSON, err := json.Marshal(current.Mods)
	if err != nil {
		return fmt.Errorf("failed to marshal mods: %w", err)
	}

	result, err := p.db.Exec(updatePressCyclesQuery, totalCycles, partialCycles, modsJSON, toolID)
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
func (p *PressCycles) GetCurrentToolUsage(toolID int64) (*PressCycle, error) {
	var cycle PressCycle
	var toDate sql.NullTime
	var modsData []byte

	err := p.db.QueryRow(selectCurrentToolUsageQuery, toolID).Scan(
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
func (p *PressCycles) GetToolHistory(toolID int64) ([]*PressCycle, error) {
	rows, err := p.db.Query(selectToolHistoryQuery, toolID)
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

// GetPressCyclesForTool gets all press cycles for a specific tool
func (p *PressCycles) GetPressCyclesForTool(toolID int64) ([]*PressCycle, error) {
	rows, err := p.db.Query(selectToolHistoryQuery, toolID)
	if err != nil {
		return nil, fmt.Errorf("failed to get press cycles for tool: %w", err)
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
func (p *PressCycles) GetToolHistorySinceRegeneration(toolID int64, lastRegenerationDate *time.Time) ([]*PressCycle, error) {
	var query string
	var args []any

	if lastRegenerationDate != nil {
		query = selectToolHistorySinceRegenerationQuery
		args = []any{toolID, *lastRegenerationDate}
	} else {
		// If no regeneration date, get all history
		query = selectAllToolHistoryQuery
		args = []any{toolID}
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
func (p *PressCycles) GetTotalCyclesSinceRegeneration(toolID int64, lastRegenerationDate *time.Time) (int64, error) {
	var query string
	var args []any

	if lastRegenerationDate != nil {
		query = selectTotalCyclesSinceRegenerationQuery
		args = []any{toolID, *lastRegenerationDate}
	} else {
		query = selectTotalCyclesAllTimeQuery
		args = []any{toolID}
	}

	var totalCycles int64
	err := p.db.QueryRow(query, args...).Scan(&totalCycles)
	if err != nil {
		return 0, fmt.Errorf("failed to get total cycles: %w", err)
	}

	return totalCycles, nil
}

// GetPressCycles gets all press cycles (current and historical) for a specific press
func (p *PressCycles) GetPressCycles(pressNumber PressNumber) ([]*PressCycle, error) {
	// Validate press number
	if !pressNumber.IsValid() {
		return nil, fmt.Errorf("invalid press number %d: must be between 0 and 5", pressNumber)
	}

	rows, err := p.db.Query(selectPressCyclesForPressQuery, pressNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get press cycles for press %d: %w", pressNumber, err)
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

// GetActivePressCycles gets only the currently active press cycles for a specific press
func (p *PressCycles) GetActivePressCycles(pressNumber PressNumber) ([]*PressCycle, error) {
	// Validate press number
	if !pressNumber.IsValid() {
		return nil, fmt.Errorf("invalid press number %d: must be between 0 and 5", pressNumber)
	}

	rows, err := p.db.Query(selectActivePressCyclesForPressQuery, pressNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get active press cycles for press %d: %w", pressNumber, err)
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

// GetCurrentToolsOnPress gets all tools currently active on a specific press
func (p *PressCycles) GetCurrentToolsOnPress(pressNumber PressNumber) ([]int64, error) {
	// Validate press number
	if pressNumber < MinPressNumber || pressNumber > MaxPressNumber {
		return nil, fmt.Errorf("invalid press number %d: must be between 0 and 5", pressNumber)
	}

	rows, err := p.db.Query(selectCurrentToolsOnPressQuery, pressNumber)
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
func (p *PressCycles) GetPressUtilization() (map[PressNumber][]int64, error) {
	utilization := make(map[PressNumber][]int64)

	// Initialize all presses (0-5) with empty slices
	for i := PressNumber(0); i <= 5; i++ {
		utilization[i] = []int64{}
	}

	// Get all currently active tool assignments
	rows, err := p.db.Query(selectPressUtilizationQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get press utilization: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var pressNumber PressNumber
		var toolID int64
		if err := rows.Scan(&pressNumber, &toolID); err != nil {
			return nil, fmt.Errorf("failed to scan utilization data: %w", err)
		}
		utilization[pressNumber] = append(utilization[pressNumber], toolID)
	}

	return utilization, nil
}

// MarkToolRegeneration marks when a tool has been regenerated (resets cycles)
func (p *PressCycles) MarkToolRegeneration(toolID int64) error {
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
func (p *PressCycles) GetPressCycleStats() (map[PressNumber]struct {
	TotalCycles    int64
	ActiveTools    int
	TotalToolsUsed int
}, error) {
	stats := make(map[PressNumber]struct {
		TotalCycles    int64
		ActiveTools    int
		TotalToolsUsed int
	})

	// Initialize stats for all presses (0-5)
	for i := PressNumber(0); i <= 5; i++ {
		stats[i] = struct {
			TotalCycles    int64
			ActiveTools    int
			TotalToolsUsed int
		}{}
	}

	// Get statistics per press
	rows, err := p.db.Query(selectPressCycleStatsQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get press cycle stats: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var pressNumber PressNumber
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
