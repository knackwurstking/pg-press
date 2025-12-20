package db

import (
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"
)

// -----------------------------------------------------------------------------
// Table Creation Statements
// -----------------------------------------------------------------------------

const (
	SQLCreatePressesTable string = `
		CREATE TABLE IF NOT EXISTS presses (
			id 					INTEGER NOT NULL,
			slot_up 			INTEGER NOT NULL,
			slot_down 			INTEGER NOT NULL,
			last_regeneration 	INTEGER NOT NULL,
			start_cycles 		INTEGER NOT NULL,
			cycles 				INTEGER NOT NULL,
			type 				TEXT NOT NULL,

			PRIMARY KEY("id")
		);
	`
)

// -----------------------------------------------------------------------------
// Table Helpers: "presses"
// -----------------------------------------------------------------------------

// -----------------------------------------------------------------------------
// Scan Helpers
// -----------------------------------------------------------------------------

func ScanPress(row Scannable) (press *shared.Press, merr *errors.MasterError) {
	press = &shared.Press{}
	err := row.Scan(
		&press.ID,
		&press.SlotUp,
		&press.SlotDown,
		&press.LastRegeneration,
		&press.StartCycles,
		&press.Cycles,
		&press.Type,
	)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	return press, nil
}
