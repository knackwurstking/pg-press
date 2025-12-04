package services

import (
	"fmt"

	"github.com/knackwurstking/pg-press/errors"
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
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			press_number INTEGER NOT NULL,
			tool_id INTEGER NOT NULL,
			tool_position TEXT NOT NULL,
			total_cycles INTEGER NOT NULL DEFAULT 0,
			date DATETIME NOT NULL,
			performed_by INTEGER NOT NULL
		);
	`, TableNamePressCycles)

	if _, err := s.DB.Exec(query); err != nil {
		return errors.Wrap(err, "create %s table", TableNamePressCycles)
	}

	return nil
}
