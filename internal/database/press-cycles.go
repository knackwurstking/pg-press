package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/knackwurstking/pgpress/internal/dberror"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/models"
)

// PressCycles manages press cycle data and operations
type PressCycles struct {
	db    *sql.DB
	feeds *Feeds
}

var _ DataOperations[*PressCycle] = (*PressCycles)(nil)

func NewPressCycles(db *sql.DB, feeds *Feeds) *PressCycles {
	p := &PressCycles{
		db:    db,
		feeds: feeds,
	}
	p.init()
	return p
}

func (p *PressCycles) init() {
	query := `
		DROP TABLE IF EXISTS press_cycles;
		CREATE TABLE IF NOT EXISTS press_cycles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			press_number INTEGER NOT NULL CHECK(press_number >= 0 AND press_number <= 5),
			tool_id INTEGER NOT NULL,
			total_cycles INTEGER NOT NULL DEFAULT 0,
			date DATETIME NOT NULL,
			performed_by INTEGER NOT NULL,
			FOREIGN KEY (tool_id) REFERENCES tools(id),
			FOREIGN KEY (performed_by) REFERENCES users(id) ON DELETE SET NULL
		);
		CREATE INDEX IF NOT EXISTS idx_press_cycles_tool_id ON press_cycles(tool_id);
		CREATE INDEX IF NOT EXISTS idx_press_cycles_press_number ON press_cycles(press_number);
	`

	if _, err := p.db.Exec(query); err != nil {
		panic(fmt.Errorf("failed to create press_cycles table: %w", err))
	}
}

// Get retrieves a specific press cycle by its ID.
func (p *PressCycles) Get(id int64) (*PressCycle, error) {
	logger.DBPressCycles().Debug("Getting press cycle by id: %d", id)

	query := `
		SELECT id, press_number, tool_id, total_cycles, date, performed_by FROM press_cycles WHERE id = ?
	`

	row := p.db.QueryRow(query, id)
	cycle, err := p.scanPressCycle(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, dberror.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get press cycle %d: %w", id, err)
	}

	return cycle, nil
}

// List retrieves all press cycles from the database, ordered by ID descending.
func (p *PressCycles) List() ([]*PressCycle, error) {
	logger.DBPressCycles().Debug("Listing all press cycles")

	query := `
		SELECT id, press_number, tool_id, total_cycles, date, performed_by FROM press_cycles ORDER BY id DESC
	`

	rows, err := p.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list press cycles: %w", err)
	}
	defer rows.Close()

	return p.scanPressCyclesRows(rows)
}

// Add creates a new press cycle entry in the database.
func (p *PressCycles) Add(cycle *PressCycle, user *models.User) (int64, error) {
	logger.DBPressCycles().Info("Adding new cycle: tool_id=%d, press_number=%d, total_cycles=%d", cycle.ToolID, cycle.PressNumber, cycle.TotalCycles)

	if cycle.Date.IsZero() {
		cycle.Date = time.Now()
	}

	query := `
		INSERT INTO press_cycles (press_number, tool_id, total_cycles, date, performed_by)
		VALUES (?, ?, ?, ?, ?)
	`

	result, err := p.db.Exec(query, cycle.PressNumber, cycle.ToolID, cycle.TotalCycles, cycle.Date, user.TelegramID)
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
		p.feeds.Add(models.NewFeed(
			models.FeedTypePressCycleAdd,
			&models.FeedPressCycleAdd{
				ToolID:      cycle.ToolID,
				TotalCycles: cycle.TotalCycles,
				ModifiedBy:  user,
			},
		))
	}

	return id, nil
}

// Update modifies an existing press cycle entry.
func (p *PressCycles) Update(cycle *PressCycle, user *models.User) error {
	logger.DBPressCycles().Info("Updating press cycle: id=%d", cycle.ID)

	if cycle.Date.IsZero() {
		cycle.Date = time.Now()
	}

	query := `
		UPDATE press_cycles
		SET total_cycles = ?, performed_by = ?, press_number = ?, date = ?
		WHERE id = ?
	`

	result, err := p.db.Exec(query, cycle.TotalCycles, user.TelegramID, cycle.PressNumber, cycle.Date, cycle.ID)
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
		p.feeds.Add(models.NewFeed(
			models.FeedTypePressCycleUpdate,
			&models.FeedPressCycleUpdate{
				ToolID:      cycle.ToolID,
				TotalCycles: cycle.TotalCycles,
				ModifiedBy:  user,
			},
		))
	}

	return nil
}

// Delete removes a press cycle from the database.
func (p *PressCycles) Delete(id int64, user *models.User) error {
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
		p.feeds.Add(models.NewFeed(
			models.FeedTypePressCycleDelete,
			&models.FeedPressCycleUpdate{ // Using Update here, as Delete doesn't exist
				ToolID:      cycle.ToolID,
				TotalCycles: cycle.TotalCycles,
				ModifiedBy:  user,
			},
		))
	}

	return nil
}

// scanPressCyclesRows scans multiple press cycles from sql.Rows (without partial_cycles)
func (p *PressCycles) scanPressCyclesRows(rows *sql.Rows) ([]*PressCycle, error) {
	cycles := make([]*PressCycle, 0)
	for rows.Next() {
		cycle, err := p.scanPressCycle(rows)
		if err != nil {
			return nil, err
		}
		cycles = append(cycles, cycle)
	}
	return cycles, nil
}

func (p *PressCycles) scanPressCycle(scanner scannable) (*PressCycle, error) {
	cycle := &PressCycle{}
	var performedBy sql.NullInt64

	err := scanner.Scan(
		&cycle.ID,
		&cycle.PressNumber,
		&cycle.ToolID,
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
