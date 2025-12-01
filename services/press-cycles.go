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
	return s.Base.CreateTable(fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			press_number INTEGER NOT NULL,
			tool_id INTEGER NOT NULL,
			tool_position TEXT NOT NULL,
			total_cycles INTEGER NOT NULL DEFAULT 0,
			date DATETIME NOT NULL,
			performed_by INTEGER NOT NULL
		);
	`, TableNamePressCycles), TableNamePressCycles)
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
