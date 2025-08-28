package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/knackwurstking/pgpress/internal/logger"
)

const (
	createPressCyclesTableQuery = `
		DROP TABLE IF EXISTS press_cycles;
		CREATE TABLE IF NOT EXISTS press_cycles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			press_number INTEGER NOT NULL CHECK(press_number >= 0 AND press_number <= 5),
			tool_id INTEGER NOT NULL,
			date DATETIME NOT NULL,
			total_cycles INTEGER NOT NULL DEFAULT 0,
			performed_by INTEGER NOT NULL,
			FOREIGN KEY (tool_id) REFERENCES tools(id),
			FOREIGN KEY (performed_by) REFERENCES users(id) ON DELETE SET NULL
		);
		CREATE INDEX IF NOT EXISTS idx_press_cycles_tool_id ON press_cycles(tool_id);
		CREATE INDEX IF NOT EXISTS idx_press_cycles_press_number ON press_cycles(press_number);
		CREATE INDEX IF NOT EXISTS idx_press_cycles_date ON press_cycles(date);
		INSERT INTO press_cycles (press_number, tool_id, date, total_cycles, performed_by)
		VALUES
			(0, 1, '2023-01-01',     0, -1),
			(0, 2, '2023-01-01',     0, -1),
			(0, 1, '2023-02-01',  1000, -1),
			(0, 2, '2023-02-01',  1000, -1),
			(0, 1, '2023-03-01',  2000, -1),
			(0, 2, '2023-03-01',  2000, -1),
			(0, 1, '2023-04-01',  3000, -1),
			(0, 2, '2023-04-01',  3000, -1),
			(0, 1, '2023-05-01',  4000, -1),
			(0, 2, '2023-05-01',  4000, -1),
			(0, 1, '2023-06-01',  5000, -1),
			(0, 2, '2023-06-01',  5000, -1),
			(0, 1, '2023-07-01',  6000, -1),
			(0, 2, '2023-07-01',  6000, -1),
			(0, 1, '2023-08-01',  7000, -1),
			(0, 2, '2023-08-01',  7000, -1),
			(0, 1, '2023-09-01',  8000, -1),
			(0, 2, '2023-09-01',  8000, -1),
			(0, 1, '2023-10-01',  9000, -1),
			(0, 2, '2023-10-01',  9000, -1),
			(0, 1, '2023-11-01', 10000, -1),
			(0, 2, '2023-11-01', 10000, -1);
	`

	insertPressCycleQuery = `
		INSERT INTO press_cycles (press_number, tool_id, date, total_cycles, performed_by)
		VALUES (?, ?, ?, ?, ?)
		RETURNING id, press_number, tool_id, date, total_cycles, performed_by
	`

	// This query is no longer needed since we don't track to_date
	// endToolUsageQuery is deprecated

	updatePressCyclesQuery = `
		UPDATE press_cycles
		SET total_cycles = ?, performed_by = ?
		WHERE id = (
			SELECT id
			FROM press_cycles
			WHERE tool_id = ?
			ORDER BY date DESC
			LIMIT 1
		)
	`

	selectCurrentToolUsageQuery = `
		SELECT id, press_number, tool_id, date, total_cycles, performed_by
		FROM press_cycles
		WHERE tool_id = ?
		ORDER BY date DESC
		LIMIT 1
	`

	selectToolHistoryQuery = `
		SELECT id, press_number, tool_id, date, total_cycles, performed_by
		FROM press_cycles
		WHERE tool_id = ?
		ORDER BY date DESC
		LIMIT ? OFFSET ?
	`

	selectToolHistorySinceRegenerationQuery = `
		SELECT pc.id, pc.press_number, pc.tool_id, pc.date, pc.total_cycles,
		       (pc.total_cycles - COALESCE(
		           (SELECT cycles_at_regeneration
		            FROM tool_regenerations
		            WHERE tool_id = pc.tool_id AND regenerated_at <= pc.date
		            ORDER BY regenerated_at DESC
		            LIMIT 1), 0)) as partial_cycles,
		       pc.performed_by
		FROM press_cycles pc
		WHERE pc.tool_id = ? AND pc.date >= ?
		ORDER BY pc.date DESC
	`

	selectAllToolHistoryQuery = `
		SELECT pc.id, pc.press_number, pc.tool_id, pc.date, pc.total_cycles,
		       (pc.total_cycles - COALESCE(
		           (SELECT cycles_at_regeneration
		            FROM tool_regenerations
		            WHERE tool_id = pc.tool_id AND regenerated_at <= pc.date
		            ORDER BY regenerated_at DESC
		            LIMIT 1), 0)) as partial_cycles,
		       pc.performed_by
		FROM press_cycles pc
		WHERE pc.tool_id = ?
		ORDER BY pc.date DESC
	`

	selectTotalCyclesSinceRegenerationQuery = `
		SELECT COALESCE(
		    (SELECT MAX(total_cycles) FROM press_cycles WHERE tool_id = ?) -
		    COALESCE((SELECT cycles_at_regeneration
		              FROM tool_regenerations
		              WHERE tool_id = ? AND regenerated_at = ?
		              ORDER BY regenerated_at DESC
		              LIMIT 1), 0), 0)
	`

	selectTotalCyclesAllTimeQuery = `
		SELECT COALESCE(MAX(total_cycles), 0)
		FROM press_cycles
		WHERE tool_id = ?
	`

	selectCurrentToolsOnPressQuery = `
		SELECT tool_id
		FROM (
			SELECT tool_id, press_number,
			       ROW_NUMBER() OVER (PARTITION BY tool_id ORDER BY date DESC) as rn
			FROM press_cycles
		)
		WHERE press_number = ? AND rn = 1
	`

	selectPressCyclesForPressQuery = `
		SELECT id, press_number, tool_id, date, total_cycles, performed_by
		FROM press_cycles
		WHERE press_number = ?
		ORDER BY date DESC
		LIMIT ? OFFSET ?
	`

	selectActivePressCyclesForPressQuery = `
		SELECT DISTINCT pc1.id, pc1.press_number, pc1.tool_id, pc1.date, pc1.total_cycles, pc1.performed_by
		FROM press_cycles pc1
		WHERE pc1.press_number = ?
		  AND pc1.date = (
		    SELECT MAX(pc2.date)
		    FROM press_cycles pc2
		    WHERE pc2.tool_id = pc1.tool_id
		  )
		ORDER BY pc1.date DESC
		LIMIT ? OFFSET ?
	`

	selectPressUtilizationQuery = `
		SELECT press_number, tool_id
		FROM (
			SELECT press_number, tool_id,
			       ROW_NUMBER() OVER (PARTITION BY tool_id ORDER BY date DESC) as rn
			FROM press_cycles
		)
		WHERE rn = 1
		ORDER BY press_number, tool_id
	`

	selectPressCycleStatsQuery = `
		SELECT
			press_number,
			SUM(total_cycles) as total_cycles,
			COUNT(DISTINCT tool_id) as total_tools_used,
			(SELECT COUNT(DISTINCT tool_id)
			 FROM (
			   SELECT tool_id,
			          ROW_NUMBER() OVER (PARTITION BY tool_id ORDER BY date DESC) as rn
			   FROM press_cycles pc2
			   WHERE pc2.press_number = press_cycles.press_number
			 ) WHERE rn = 1) as active_tools
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
func (p *PressCycles) StartToolUsage(toolID int64, pressNumber PressNumber, user *User) (*PressCycle, error) {
	logger.DBPressCycles().Info("Starting tool usage: tool_id=%d, press_number=%d", toolID, pressNumber)

	// Validate press number
	if pressNumber < MinPressNumber || pressNumber > MaxPressNumber {
		return nil, fmt.Errorf("invalid press number %d: must be between 0 and 5", pressNumber)
	}

	// No need to end previous usage since we're not tracking to_date anymore

	// Create new press cycle entry
	var performedBy *int64
	if user != nil {
		performedBy = &user.TelegramID
	}

	row := p.db.QueryRow(insertPressCycleQuery, pressNumber, toolID, time.Now(), 0, performedBy)
	cycle, err := p.scanPressCyclesRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("failed to start tool usage: %w", err)
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

	return cycle, nil
}

// EndToolUsage is deprecated - we no longer track end dates
// Kept for backward compatibility but does nothing
func (p *PressCycles) EndToolUsage(toolID int64) error {
	logger.DBPressCycles().Info("EndToolUsage called (deprecated): tool_id=%d", toolID)
	// No-op - we don't track to_date anymore
	return nil
}

// AddCycle adds a new press cycle entry for a tool
func (p *PressCycles) AddCycle(toolID int64, pressNumber PressNumber, totalCycles int64, user *User) (*PressCycle, error) {
	logger.DBPressCycles().Info("Adding new cycle: tool_id=%d, press_number=%d, total_cycles=%d", toolID, pressNumber, totalCycles)

	var performedBy *int64
	if user != nil {
		performedBy = &user.TelegramID
	}

	row := p.db.QueryRow(insertPressCycleQuery, pressNumber, toolID, time.Now(), totalCycles, performedBy)
	cycle, err := p.scanPressCyclesRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to add cycle: %w", err)
	}

	return cycle, nil
}

// UpdateCycles updates the cycle counts for a currently active tool on a press
// UpdateCycles updates the total and partial cycles for an active tool
func (p *PressCycles) UpdateCycles(toolID int64, totalCycles int64, user *User) error {
	logger.DBPressCycles().Debug("Updating cycles: tool_id=%d, total=%d", toolID, totalCycles)

	var performedBy *int64
	if user != nil {
		performedBy = &user.TelegramID
	}

	result, err := p.db.Exec(updatePressCyclesQuery, totalCycles, performedBy, toolID)
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
	logger.DBPressCycles().Debug("Getting current tool usage: tool_id=%d", toolID)

	row := p.db.QueryRow(selectCurrentToolUsageQuery, toolID)
	cycle, err := p.scanPressCyclesRow(row)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get current tool usage: %w", err)
	}

	return cycle, nil
}

// GetToolHistory gets the press usage history for a tool
// GetToolHistory retrieves all press cycles for a specific tool
func (p *PressCycles) GetToolHistory(toolID int64, limit, offset int) ([]*PressCycle, error) {
	logger.DBPressCycles().Debug("Getting tool history: tool_id=%d", toolID)

	rows, err := p.db.Query(selectToolHistoryQuery, toolID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get tool history: %w", err)
	}
	defer rows.Close()

	cycles, err := p.scanPressCyclesRows(rows)
	if err != nil {
		return nil, fmt.Errorf("failed to scan press cycles: %w", err)
	}

	return cycles, nil
}

// GetPressCyclesForTool gets all press cycles for a specific tool
func (p *PressCycles) GetPressCyclesForTool(toolID int64) ([]*PressCycle, error) {
	logger.DBPressCycles().Debug("Getting press cycles for tool: tool_id=%d", toolID)

	rows, err := p.db.Query(selectAllToolHistoryQuery, toolID)
	if err != nil {
		return nil, fmt.Errorf("failed to get press cycles for tool %d: %w", toolID, err)
	}
	defer rows.Close()

	cycles, err := p.scanPressCyclesRowsWithPartial(rows)
	if err != nil {
		return nil, fmt.Errorf("failed to scan press cycles: %w", err)
	}

	return cycles, nil
}

// GetToolHistorySinceRegeneration gets press cycles since the last tool regeneration
func (p *PressCycles) GetToolHistorySinceRegeneration(toolID int64, lastRegenerationDate *time.Time) ([]*PressCycle, error) {
	logger.DBPressCycles().Debug("Getting tool history since regeneration: tool_id=%d", toolID)

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
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tool history since regeneration: %w", err)
	}
	defer rows.Close()

	cycles, err := p.scanPressCyclesRowsWithPartial(rows)
	if err != nil {
		return nil, fmt.Errorf("failed to scan press cycles: %w", err)
	}

	return cycles, nil
}

// GetTotalCyclesSinceRegeneration calculates total cycles since last regeneration
func (p *PressCycles) GetTotalCyclesSinceRegeneration(toolID int64, lastRegenerationDate *time.Time) (int64, error) {
	logger.DBPressCycles().Debug("Getting total cycles since regeneration: tool_id=%d", toolID)

	var query string
	var args []any

	if lastRegenerationDate != nil {
		query = selectTotalCyclesSinceRegenerationQuery
		args = []any{toolID, toolID, *lastRegenerationDate}
	} else {
		// If no regeneration date, get total cycles all time
		query = selectTotalCyclesAllTimeQuery
		args = []any{toolID}
	}

	var totalCycles int64
	err := p.db.QueryRow(query, args...).Scan(&totalCycles)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get total cycles: %w", err)
	}

	return totalCycles, nil
}

// GetCurrentTotalCycles gets the current absolute total cycles for a tool
func (p *PressCycles) GetCurrentTotalCycles(toolID int64) (int64, error) {
	logger.DBPressCycles().Debug("Getting current total cycles: tool_id=%d", toolID)

	var totalCycles int64
	err := p.db.QueryRow(selectTotalCyclesAllTimeQuery, toolID).Scan(&totalCycles)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get current total cycles: %w", err)
	}

	return totalCycles, nil
}

// GetPressCycles gets all press cycles (current and historical) for a specific press
func (p *PressCycles) GetPressCycles(pressNumber PressNumber, limit, offset int) ([]*PressCycle, error) {
	logger.DBPressCycles().Debug("Getting press cycles: press_number=%d", pressNumber)

	// Validate press number
	if !pressNumber.IsValid() {
		return nil, fmt.Errorf("invalid press number %d: must be between 0 and 5", pressNumber)
	}

	rows, err := p.db.Query(selectPressCyclesForPressQuery, pressNumber, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get press cycles for press %d: %w", pressNumber, err)
	}
	defer rows.Close()

	cycles, err := p.scanPressCyclesRows(rows)
	if err != nil {
		return nil, fmt.Errorf("failed to scan press cycles: %w", err)
	}

	return cycles, nil
}

// GetActivePressCycles gets only the currently active press cycles for a specific press
func (p *PressCycles) GetActivePressCycles(pressNumber PressNumber, limit, offset int) ([]*PressCycle, error) {
	logger.DBPressCycles().Debug("Getting active press cycles: press_number=%d", pressNumber)

	// Validate press number
	if !pressNumber.IsValid() {
		return nil, fmt.Errorf("invalid press number %d: must be between 0 and 5", pressNumber)
	}

	rows, err := p.db.Query(selectActivePressCyclesForPressQuery, pressNumber, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get active press cycles for press %d: %w", pressNumber, err)
	}
	defer rows.Close()

	cycles, err := p.scanPressCyclesRows(rows)
	if err != nil {
		return nil, fmt.Errorf("failed to scan press cycles: %w", err)
	}

	return cycles, nil
}

// GetCurrentToolsOnPress gets all tools currently active on a specific press
func (p *PressCycles) GetCurrentToolsOnPress(pressNumber PressNumber) ([]int64, error) {
	logger.DBPressCycles().Debug("Getting current tools on press: press_number=%d", pressNumber)

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
	logger.DBPressCycles().Debug("Getting press utilization for all presses")

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
	logger.DBPressCycles().Info("Marking tool regeneration: tool_id=%d", toolID)

	// No need to end usage since we don't track to_date anymore

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
	logger.DBPressCycles().Debug("Getting press cycle statistics for all presses")

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

// scanPressCyclesRows scans multiple press cycles from sql.Rows (without partial_cycles)
func (p *PressCycles) scanPressCyclesRows(rows *sql.Rows) ([]*PressCycle, error) {
	cycles := make([]*PressCycle, 0)
	for rows.Next() {
		cycle := &PressCycle{}
		var performedBy sql.NullInt64

		err := rows.Scan(
			&cycle.ID,
			&cycle.PressNumber,
			&cycle.ToolID,
			&cycle.Date,
			&cycle.TotalCycles,
			&performedBy,
		)
		if err != nil {
			return nil, err
		}

		if performedBy.Valid {
			cycle.PerformedBy = performedBy.Int64
		}

		// Partial cycles will be calculated when needed
		cycle.PartialCycles = 0

		cycles = append(cycles, cycle)
	}
	return cycles, nil
}

// scanPressCyclesRowsWithPartial scans multiple press cycles from sql.Rows (with calculated partial_cycles)
func (p *PressCycles) scanPressCyclesRowsWithPartial(rows *sql.Rows) ([]*PressCycle, error) {
	cycles := make([]*PressCycle, 0)
	for rows.Next() {
		cycle := &PressCycle{}
		var performedBy sql.NullInt64

		err := rows.Scan(
			&cycle.ID,
			&cycle.PressNumber,
			&cycle.ToolID,
			&cycle.Date,
			&cycle.TotalCycles,
			&cycle.PartialCycles,
			&performedBy,
		)
		if err != nil {
			return nil, err
		}

		if performedBy.Valid {
			cycle.PerformedBy = performedBy.Int64
		}

		cycles = append(cycles, cycle)
	}
	return cycles, nil
}

// scanPressCyclesRow scans a single press cycle from a sql.Row (without partial_cycles)
func (p *PressCycles) scanPressCyclesRow(scanner *sql.Row) (*PressCycle, error) {
	cycle := &PressCycle{}
	var performedBy sql.NullInt64

	err := scanner.Scan(
		&cycle.ID,
		&cycle.PressNumber,
		&cycle.ToolID,
		&cycle.Date,
		&cycle.TotalCycles,
		&performedBy,
	)
	if err != nil {
		return nil, err
	}

	if performedBy.Valid {
		cycle.PerformedBy = performedBy.Int64
	}

	// Partial cycles will be calculated when needed
	cycle.PartialCycles = 0

	return cycle, nil
}
