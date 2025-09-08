package presscycle

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/knackwurstking/pgpress/internal/database/dberror"
	"github.com/knackwurstking/pgpress/internal/database/interfaces"
	feedmodels "github.com/knackwurstking/pgpress/internal/database/models/feed"
	pressmodels "github.com/knackwurstking/pgpress/internal/database/models/press"
	usermodels "github.com/knackwurstking/pgpress/internal/database/models/user"
	"github.com/knackwurstking/pgpress/internal/database/services/feed"
	"github.com/knackwurstking/pgpress/internal/logger"
)

type Service struct {
	db    *sql.DB
	feeds *feed.Service
}

// Just to make sure it fits
var _ interfaces.DataOperations[*pressmodels.Cycle] = (*Service)(nil)

func New(db *sql.DB, feeds *feed.Service) *Service {
	query := `
		CREATE TABLE IF NOT EXISTS press_cycles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			press_number INTEGER NOT NULL CHECK(press_number >= 0 AND press_number <= 5),
			slot_top INTEGER NOT NULL,
			slot_top_cassette INTEGER NOT NULL,
			slot_bottom INTEGER NOT NULL,
			total_cycles INTEGER NOT NULL DEFAULT 0,
			date DATETIME NOT NULL,
			performed_by INTEGER NOT NULL,
			FOREIGN KEY (slot_top) REFERENCES tools(id),
			FOREIGN KEY (slot_top_cassette) REFERENCES tools(id),
			FOREIGN KEY (slot_bottom) REFERENCES tools(id),
			FOREIGN KEY (performed_by) REFERENCES users(id) ON DELETE SET NULL
		);
		CREATE INDEX IF NOT EXISTS idx_press_cycles_slot_top ON press_cycles(slot_top);
		CREATE INDEX IF NOT EXISTS idx_press_cycles_slot_top_cassette ON press_cycles(slot_top_cassette);
		CREATE INDEX IF NOT EXISTS idx_press_cycles_slot_bottom ON press_cycles(slot_bottom);
		CREATE INDEX IF NOT EXISTS idx_press_cycles_press_number ON press_cycles(press_number);
	`

	if _, err := db.Exec(query); err != nil {
		panic(fmt.Errorf("failed to create press_cycles table: %w", err))
	}

	return &Service{
		db:    db,
		feeds: feeds,
	}
}

func (s *Service) GetPartialCycles(cycle *pressmodels.Cycle) int64 {
	logger.DBPressCycles().Debug("Getting partial cycles: %#v", cycle)

	var query string
	if cycle.SlotTop > 0 {
		query += "AND slot_top > 0"
	}
	if cycle.SlotTopCassette > 0 {
		query += "AND slot_top_cassette > 0"
	}
	if cycle.SlotBottom > 0 {
		query += "AND slot_bottom > 0"
	}

	// Get the total_cycles from the previous entry on the same press (regardless of tool_id)
	previousQuery := fmt.Sprintf(`
		SELECT
			total_cycles
		FROM
	 		press_cycles
		WHERE
			press_number = ?
			%s
			AND id < ?
		ORDER BY
			id DESC
		LIMIT 1
	`, query)

	var previousTotalCycles int64
	err := s.db.QueryRow(previousQuery, cycle.PressNumber, cycle.ID).Scan(&previousTotalCycles)
	if err != nil {
		if err == sql.ErrNoRows {
			// No previous entry found, so partial cycles equals total cycles
			logger.DBPressCycles().Debug("No previous entry found for press %d, partial cycles = total cycles (%d)", cycle.PressNumber, cycle.TotalCycles)
		} else {
			logger.DBPressCycles().Error("Failed to get previous total cycles for press %d: %v", cycle.PressNumber, err)
		}
		return cycle.TotalCycles
	}

	partialCycles := cycle.TotalCycles - previousTotalCycles
	logger.DBPressCycles().Debug("Partial cycles calculated: %d (current: %d - previous: %d)", partialCycles, cycle.TotalCycles, previousTotalCycles)

	return partialCycles
}

// Get retrieves a specific press cycle by its ID.
func (p *Service) Get(id int64) (*pressmodels.Cycle, error) {
	logger.DBPressCycles().Debug("Getting press cycle by id: %d", id)

	query := `
		SELECT id, press_number, slot_top, slot_top_cassette, slot_bottom, total_cycles, date, performed_by
		FROM press_cycles
		WHERE id = ?
	`

	row := p.db.QueryRow(query, id)
	cycle, err := p.scanPressCycle(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, dberror.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get press cycle %d: %w", id, err)
	}
	cycle.PartialCycles = p.GetPartialCycles(cycle)

	return cycle, nil
}

// List retrieves all press cycles from the database, ordered by ID descending.
func (p *Service) List() ([]*pressmodels.Cycle, error) {
	logger.DBPressCycles().Debug("Listing all press cycles")

	query := `
		SELECT id, press_number, slot_top, slot_top_cassette, slot_bottom, total_cycles, date, performed_by
		FROM press_cycles
		ORDER BY id DESC
	`

	rows, err := p.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list press cycles: %w", err)
	}
	defer rows.Close()

	return p.scanPressCyclesRows(rows)
}

// Add creates a new press cycle entry in the database.
func (p *Service) Add(cycle *pressmodels.Cycle, user *usermodels.User) (int64, error) {
	logger.DBPressCycles().Info(
		"Adding new cycle: slot_top=%d, slot_top_cassette=%d, slot_bottom=%d, press_number=%d, total_cycles=%d",
		cycle.SlotTop, cycle.SlotTopCassette, cycle.SlotBottom, cycle.PressNumber, cycle.TotalCycles,
	)

	if cycle.Date.IsZero() {
		cycle.Date = time.Now()
	}

	query := `
		INSERT INTO press_cycles (press_number, slot_top, slot_top_cassette, slot_bottom, total_cycles, date, performed_by)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	result, err := p.db.Exec(query,
		cycle.PressNumber,
		cycle.SlotTop, cycle.SlotTopCassette, cycle.SlotBottom,
		cycle.TotalCycles,
		cycle.Date,
		user.TelegramID,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to add cycle: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id for cycle: %w", err)
	}
	cycle.ID = id

	// Create feed entry
	if p.feeds != nil {
		p.feeds.Add(feedmodels.New(
			feedmodels.TypePressCycleAdd,
			&feedmodels.PressCycleAdd{
				SlotTop:         cycle.SlotTop,
				SlotTopCassette: cycle.SlotTopCassette,
				SlotBottom:      cycle.SlotBottom,
				TotalCycles:     cycle.TotalCycles,
				ModifiedBy:      user,
			},
		))
	}

	return id, nil
}

// Update modifies an existing press cycle entry.
func (p *Service) Update(cycle *pressmodels.Cycle, user *usermodels.User) error {
	logger.DBPressCycles().Info("Updating press cycle: id=%d", cycle.ID)

	if cycle.Date.IsZero() {
		cycle.Date = time.Now()
	}

	query := `
		UPDATE press_cycles
		SET total_cycles = ?, slot_top = ?, slot_top_cassette = ?, slot_bottom = ?, performed_by = ?, press_number = ?, date = ?
		WHERE id = ?
	`

	result, err := p.db.Exec(query,
		cycle.TotalCycles,
		cycle.SlotTop, cycle.SlotTopCassette, cycle.SlotBottom,
		user.TelegramID,
		cycle.PressNumber,
		cycle.Date,
		cycle.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update press cycle with id %d: %w", cycle.ID, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("no press cycle found with id %d", cycle.ID)
	}

	// Create feed entry
	if p.feeds != nil {
		p.feeds.Add(feedmodels.New(
			feedmodels.TypePressCycleUpdate,
			&feedmodels.PressCycleUpdate{
				SlotTop:         cycle.SlotTop,
				SlotTopCassette: cycle.SlotTopCassette,
				SlotBottom:      cycle.SlotBottom,
				TotalCycles:     cycle.TotalCycles,
				ModifiedBy:      user,
			},
		))
	}

	return nil
}

// Delete removes a press cycle from the database.
func (p *Service) Delete(id int64, user *usermodels.User) error {
	logger.DBPressCycles().Info("Deleting press cycle: id=%d", id)

	// Get cycle for feed before deleting
	cycle, err := p.Get(id)
	if err != nil {
		return fmt.Errorf("failed to get press cycle for deletion: %w", err)
	}

	query := `
		DELETE FROM press_cycles WHERE id = ?
	`

	result, err := p.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete press cycle with id %d: %w", id, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected for delete: %w", err)
	}
	if rows == 0 {
		return dberror.ErrNotFound
	}

	// Create feed entry
	if p.feeds != nil {
		p.feeds.Add(feedmodels.New(
			feedmodels.TypePressCycleDelete,
			&feedmodels.PressCycleUpdate{ // TODO: Create Delete type
				SlotTop:         cycle.SlotTop,
				SlotTopCassette: cycle.SlotTopCassette,
				SlotBottom:      cycle.SlotBottom,
				TotalCycles:     cycle.TotalCycles,
				ModifiedBy:      user,
			},
		))
	}

	return nil
}

// GetPressCyclesForTool gets all press cycles for a specific tool
func (s *Service) GetPressCyclesForTool(toolID int64) ([]*pressmodels.Cycle, error) {
	logger.DBPressCycles().Debug("Getting press cycles for tool: tool_id=%d", toolID)

	query := `
		SELECT id, press_number, slot_top, slot_top_cassette, slot_bottom, total_cycles, date, performed_by
		FROM press_cycles
		WHERE slot_top = ? OR slot_top_cassette = ? OR slot_bottom = ?
		ORDER BY id DESC
	`

	rows, err := s.db.Query(query, toolID, toolID, toolID)
	if err != nil {
		return nil, fmt.Errorf("failed to get press cycles for tool %d: %w", toolID, err)
	}
	defer rows.Close()

	cycles, err := s.scanPressCyclesRows(rows)
	if err != nil {
		return nil, fmt.Errorf("failed to scan press cycles: %w", err)
	}

	return cycles, nil
}

// GetPressCycles gets all press cycles (current and historical) for a specific presForPresss
func (s *Service) GetPressCycles(pressNumber pressmodels.PressNumber, limit, offset int) ([]*pressmodels.Cycle, error) {
	logger.DBPressCycles().Debug("Getting press cycles: press_number=%d, limit=%d, offset=%d", pressNumber, limit, offset)

	if !pressmodels.IsValidPressNumber(&pressNumber) {
		return nil, fmt.Errorf("invalid press number %d: must be between 0 and 5", pressNumber)
	}

	query := `
		SELECT id, press_number, slot_top, slot_top_cassette, slot_bottom, total_cycles, date, performed_by
		FROM press_cycles
		WHERE press_number = ?
		ORDER BY id DESC
		LIMIT ? OFFSET ?
	`

	rows, err := s.db.Query(query, pressNumber, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get press cycles for press %d: %w", pressNumber, err)
	}
	defer rows.Close()

	cycles, err := s.scanPressCyclesRows(rows)
	if err != nil {
		return nil, fmt.Errorf("failed to scan press cycles: %w", err)
	}

	return cycles, nil
}

// scanPressCyclesRows scans multiple press cycles from sql.Rows (without partial_cycles)
func (p *Service) scanPressCyclesRows(rows *sql.Rows) ([]*pressmodels.Cycle, error) {
	cycles := make([]*pressmodels.Cycle, 0)
	for rows.Next() {
		cycle, err := p.scanPressCycle(rows)
		if err != nil {
			return nil, err
		}
		cycle.PartialCycles = p.GetPartialCycles(cycle)
		cycles = append(cycles, cycle)
	}
	return cycles, nil
}

func (p *Service) scanPressCycle(scanner interfaces.Scannable) (*pressmodels.Cycle, error) {
	cycle := &pressmodels.Cycle{}
	var performedBy sql.NullInt64

	err := scanner.Scan(
		&cycle.ID,
		&cycle.PressNumber,
		&cycle.SlotTop,
		&cycle.SlotTopCassette,
		&cycle.SlotBottom,
		&cycle.TotalCycles,
		&cycle.Date,
		&performedBy,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}

	if performedBy.Valid {
		cycle.PerformedBy = performedBy.Int64
	}

	return cycle, nil
}
