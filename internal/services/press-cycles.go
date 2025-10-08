package services

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"
)

type PressCycles struct {
	*BaseService
}

func NewPressCycles(db *sql.DB) *PressCycles {
	base := NewBaseService(db, "Press Cycles")

	query := `
		CREATE TABLE IF NOT EXISTS press_cycles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			press_number INTEGER NOT NULL CHECK(press_number >= 0 AND press_number <= 5),
			tool_id INTEGER NOT NULL,
			tool_position TEXT NOT NULL,
			total_cycles INTEGER NOT NULL DEFAULT 0,
			date DATETIME NOT NULL,
			performed_by INTEGER NOT NULL,
			FOREIGN KEY (tool_id) REFERENCES tools(id),
			FOREIGN KEY (performed_by) REFERENCES users(telegram_id) ON DELETE SET NULL
		);
		CREATE INDEX IF NOT EXISTS idx_press_cycles_tool_id ON press_cycles(tool_id);
		CREATE INDEX IF NOT EXISTS idx_press_cycles_tool_position ON press_cycles(tool_position);
		CREATE INDEX IF NOT EXISTS idx_press_cycles_press_number ON press_cycles(press_number);
	`

	if err := base.CreateTable(query, "press_cycles"); err != nil {
		panic(err)
	}

	return &PressCycles{
		BaseService: base,
	}
}

// GetPartialCycles calculates the partial cycles for a given cycle
func (s *PressCycles) GetPartialCycles(cycle *models.Cycle) int64 {
	if err := ValidatePressCycle(cycle); err != nil {
		s.log.Error("Invalid cycle for partial calculation: %v", err)
		return cycle.TotalCycles
	}

	s.LogOperation("Calculating partial cycles",
		fmt.Sprintf("press: %d, tool: %d, position: %s, total: %d",
			cycle.PressNumber, cycle.ToolID, cycle.ToolPosition, cycle.TotalCycles))

	query := `
		SELECT total_cycles
		FROM press_cycles
		WHERE press_number = ? AND tool_id > 0 AND tool_position = ? AND total_cycles < ?
		ORDER BY total_cycles DESC
		LIMIT 1
	`

	var previousTotalCycles int64
	err := s.db.QueryRow(query, cycle.PressNumber, cycle.ToolPosition, cycle.TotalCycles).Scan(&previousTotalCycles)
	if err != nil {
		if err != sql.ErrNoRows {
			s.log.Error("Failed to get previous total cycles: %v", err)
		}
		s.LogOperation("No previous cycles found, using total cycles", fmt.Sprintf("total: %d", cycle.TotalCycles))
		return cycle.TotalCycles
	}

	partialCycles := cycle.TotalCycles - previousTotalCycles
	s.LogOperation("Calculated partial cycles", fmt.Sprintf("partial: %d (total: %d - previous: %d)",
		partialCycles, cycle.TotalCycles, previousTotalCycles))

	return partialCycles
}

// Get retrieves a specific press cycle by its ID.
func (p *PressCycles) Get(id int64) (*models.Cycle, error) {
	if err := ValidateID(id, "press_cycle"); err != nil {
		return nil, err
	}

	p.LogOperation("Getting press cycle", id)

	query := `
		SELECT id, press_number, tool_id, tool_position, total_cycles, date, performed_by
		FROM press_cycles
		WHERE id = ?
	`

	row := p.db.QueryRow(query, id)
	cycle, err := ScanSingleRow(row, ScanPressCycle, "press_cycles")
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError(fmt.Sprintf("Press cycle with ID %d not found", id))
		}
		return nil, err
	}

	cycle.PartialCycles = p.GetPartialCycles(cycle)
	return cycle, nil
}

// List retrieves all press cycles from the database, ordered by total cycles descending.
func (p *PressCycles) List() ([]*models.Cycle, error) {
	p.LogOperation("Listing press cycles")

	query := `
		SELECT id, press_number, tool_id, tool_position, total_cycles, date, performed_by
		FROM press_cycles
		ORDER BY total_cycles DESC
	`

	rows, err := p.db.Query(query)
	if err != nil {
		return nil, p.HandleSelectError(err, "press_cycles")
	}
	defer rows.Close()

	cycles, err := p.scanPressCyclesRows(rows)
	if err != nil {
		return nil, err
	}

	p.LogOperation("Listed press cycles", fmt.Sprintf("count: %d", len(cycles)))
	return cycles, nil
}

// Add creates a new press cycle entry in the database.
func (p *PressCycles) Add(cycle *models.Cycle, user *models.User) (int64, error) {
	if err := ValidatePressCycle(cycle); err != nil {
		return 0, err
	}

	if err := ValidateNotNil(user, "user"); err != nil {
		return 0, err
	}

	if cycle.Date.IsZero() {
		cycle.Date = time.Now()
	}

	p.LogOperationWithUser("Adding press cycle", createUserInfo(user),
		fmt.Sprintf("tool: %d, position: %s, press: %d, cycles: %d",
			cycle.ToolID, cycle.ToolPosition, cycle.PressNumber, cycle.TotalCycles))

	query := `
		INSERT INTO press_cycles (press_number, tool_id, tool_position, total_cycles, date, performed_by)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	result, err := p.db.Exec(query,
		cycle.PressNumber,
		cycle.ToolID,
		cycle.ToolPosition,
		cycle.TotalCycles,
		cycle.Date,
		user.TelegramID,
	)
	if err != nil {
		return 0, p.HandleInsertError(err, "press_cycles")
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, p.HandleInsertError(err, "press_cycles")
	}

	cycle.ID = id
	p.LogOperation("Added press cycle", fmt.Sprintf("id: %d", id))
	return id, nil
}

// Update modifies an existing press cycle entry.
func (p *PressCycles) Update(cycle *models.Cycle, user *models.User) error {
	if err := ValidatePressCycle(cycle); err != nil {
		return err
	}

	if err := ValidateID(cycle.ID, "press_cycle"); err != nil {
		return err
	}

	if err := ValidateNotNil(user, "user"); err != nil {
		return err
	}

	if cycle.Date.IsZero() {
		cycle.Date = time.Now()
	}

	p.LogOperationWithUser("Updating press cycle", createUserInfo(user), fmt.Sprintf("id: %d", cycle.ID))

	query := `
		UPDATE press_cycles
		SET total_cycles = ?, tool_id = ?, tool_position = ?, performed_by = ?, press_number = ?, date = ?
		WHERE id = ?
	`

	result, err := p.db.Exec(query,
		cycle.TotalCycles,
		cycle.ToolID,
		cycle.ToolPosition,
		user.TelegramID,
		cycle.PressNumber,
		cycle.Date,
		cycle.ID,
	)
	if err != nil {
		return p.HandleUpdateError(err, "press_cycles")
	}

	if err := p.CheckRowsAffected(result, "press_cycle", cycle.ID); err != nil {
		return err
	}

	p.LogOperation("Updated press cycle", fmt.Sprintf("id: %d", cycle.ID))
	return nil
}

// Delete removes a press cycle from the database.
func (p *PressCycles) Delete(id int64) error {
	if err := ValidateID(id, "press_cycle"); err != nil {
		return err
	}

	p.LogOperation("Deleting press cycle", id)

	query := `DELETE FROM press_cycles WHERE id = ?`
	result, err := p.db.Exec(query, id)
	if err != nil {
		return p.HandleDeleteError(err, "press_cycles")
	}

	if err := p.CheckRowsAffected(result, "press_cycle", id); err != nil {
		return err
	}

	p.LogOperation("Deleted press cycle", id)
	return nil
}

// GetPressCyclesForTool gets all press cycles for a specific tool
func (s *PressCycles) GetPressCyclesForTool(toolID int64) ([]*models.Cycle, error) {
	if err := ValidateID(toolID, "tool"); err != nil {
		return nil, err
	}

	s.LogOperation("Getting press cycles for tool", toolID)

	query := `
		SELECT id, press_number, tool_id, tool_position, total_cycles, date, performed_by
		FROM press_cycles
		WHERE tool_id = ?
		ORDER BY date DESC
	`

	rows, err := s.db.Query(query, toolID)
	if err != nil {
		return nil, s.HandleSelectError(err, "press_cycles")
	}
	defer rows.Close()

	cycles, err := s.scanPressCyclesRows(rows)
	if err != nil {
		return nil, err
	}

	s.LogOperation("Found press cycles for tool", fmt.Sprintf("tool: %d, count: %d", toolID, len(cycles)))
	return cycles, nil
}

// GetPressCycles gets all press cycles for a specific press with optional pagination
func (s *PressCycles) GetPressCycles(pressNumber models.PressNumber, limit *int, offset *int) ([]*models.Cycle, error) {
	if err := ValidatePressNumber(pressNumber); err != nil {
		return nil, err
	}

	// Validate pagination if provided
	if limit != nil || offset != nil {
		limitVal := 0
		offsetVal := 0
		if limit != nil {
			limitVal = *limit
		}
		if offset != nil {
			offsetVal = *offset
		}
		if err := ValidatePagination(limitVal, offsetVal); err != nil {
			return nil, err
		}
	}

	s.LogOperation("Getting press cycles for press",
		fmt.Sprintf("press: %d, limit: %v, offset: %v", pressNumber, limit, offset))

	query := `
		SELECT id, press_number, tool_id, tool_position, total_cycles, date, performed_by
		FROM press_cycles
		WHERE press_number = ?
		ORDER BY total_cycles DESC
	`

	var args []interface{}
	args = append(args, pressNumber)

	if limit != nil {
		query += " LIMIT ?"
		args = append(args, *limit)
	}
	if offset != nil {
		if limit == nil {
			query += " LIMIT -1"
		}
		query += " OFFSET ?"
		args = append(args, *offset)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, s.HandleSelectError(err, "press_cycles")
	}
	defer rows.Close()

	cycles, err := s.scanPressCyclesRows(rows)
	if err != nil {
		return nil, err
	}

	s.LogOperation("Found press cycles for press",
		fmt.Sprintf("press: %d, count: %d", pressNumber, len(cycles)))
	return cycles, nil
}

// scanPressCyclesRows scans multiple press cycles from sql.Rows and calculates partial cycles
func (p *PressCycles) scanPressCyclesRows(rows *sql.Rows) ([]*models.Cycle, error) {
	cycles, err := ScanPressCyclesFromRows(rows)
	if err != nil {
		return nil, err
	}

	// Calculate partial cycles for each cycle
	for _, cycle := range cycles {
		cycle.PartialCycles = p.GetPartialCycles(cycle)
	}

	p.LogOperation("Scanned and calculated partial cycles", fmt.Sprintf("count: %d", len(cycles)))
	return cycles, nil
}
