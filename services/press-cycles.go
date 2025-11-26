package services

import (
	"fmt"

	"github.com/knackwurstking/pg-press/models"
)

const TableNamePressCycles = "press_cycles"

type PressCycles struct {
	*Base
}

func NewPressCycles(r *Registry) *PressCycles {
	s := &PressCycles{Base: NewBase(r)}

	if err := s.CreateTable(); err != nil {
		panic(err)
	}

	return s
}

func (s *PressCycles) CreateTable() error {

	if s.HasTable("tool_regenerations") {
		// Drop foreign keys from the tool_regenerations table
		// `FOREIGN KEY (cycle_id) REFERENCES press_cycles(id) ON DELETE SET NULL`
		_, err := s.DB.Exec(`
			PRAGMA foreign_keys=off;

			ALTER TABLE tool_regenerations RENAME TO tool_regenerations_old;

			CREATE TABLE tool_regenerations (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				tool_id INTEGER NOT NULL,
				cycle_id INTEGER NOT NULL,
				reason TEXT,
				performed_by INTEGER NOT NULL
			);

			INSERT INTO tool_regenerations SELECT * FROM tool_regenerations_old;

			DROP TABLE tool_regenerations_old;

			PRAGMA foreign_keys=on;
		`)
		if err != nil {
			return fmt.Errorf("alter tool_regenerations table: %v", err)
		}
	}

	if s.HasTable("press_cycles") {
		s.DB.Exec(`DROP INDEX idx_press_cycles_tool_id`)
		s.DB.Exec(`DROP INDEX idx_press_cycles_tool_position`)
		s.DB.Exec(`DROP INDEX idx_press_cycles_press_number`)

		query := `
		PRAGMA foreign_keys=off;

		ALTER TABLE press_cycles RENAME TO press_cycles_old;

		CREATE TABLE press_cycles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			press_number INTEGER NOT NULL,
			tool_id INTEGER NOT NULL,
			tool_position TEXT NOT NULL,
			total_cycles INTEGER NOT NULL DEFAULT 0,
			date DATETIME NOT NULL,
			performed_by INTEGER NOT NULL
		);

		INSERT INTO press_cycles SELECT * FROM press_cycles_old;

		DROP TABLE press_cycles_old;

		PRAGMA foreign_keys=on;
	`
		if err := s.Base.CreateTable(query, "press_cycles"); err != nil {
			return err
		}
	} else {
		if err := s.Base.CreateTable(`
			CREATE TABLE press_cycles (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				press_number INTEGER NOT NULL,
				tool_id INTEGER NOT NULL,
				tool_position TEXT NOT NULL,
				total_cycles INTEGER NOT NULL DEFAULT 0,
				date DATETIME NOT NULL,
				performed_by INTEGER NOT NULL
			);
		`, "press_cycles"); err != nil {
			return err
		}
	}

	return nil
}

func scanCycle(scannable Scannable) (*models.Cycle, error) {
	cycle := &models.Cycle{}

	err := scannable.Scan(
		&cycle.ID,
		&cycle.PressNumber,
		&cycle.ToolID,
		&cycle.ToolPosition,
		&cycle.TotalCycles,
		&cycle.Date,
		&cycle.PerformedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("scan press cycle: %w", err)
	}

	return cycle, nil
}
