package db

import (
	"database/sql"
	"net/http"

	"github.com/knackwurstking/pg-press/internal/common"
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

	SQLCreateCyclesTable string = `
		CREATE TABLE IF NOT EXISTS cycles (
			id           INTEGER NOT NULL,
			tool_id      INTEGER NOT NULL,
			press_number INTEGER NOT NULL,
			cycles       INTEGER NOT NULL, -- Cycles at stop time
			start        INTEGER NOT NULL,
			stop         INTEGER NOT NULL,

			PRIMARY KEY("id")
		);
	`

	SQLCreatePressRegenerationsTable string = `
		CREATE TABLE IF NOT EXISTS press_regenerations (
			id 				INTEGER NOT NULL,
			press_number 	INTEGER NOT NULL,
			start 			INTEGER NOT NULL,
			stop 			INTEGER NOT NULL,
			cycles 			INTEGER NOT NULL,

			PRIMARY KEY("id" AUTOINCREMENT)
		);
	`
)

// -----------------------------------------------------------------------------
// Table Helpers: "presses"
// -----------------------------------------------------------------------------

// -----------------------------------------------------------------------------
// Table Helpers: "cycles"
// -----------------------------------------------------------------------------

const SQLListCyclesByPressNumber string = `
	SELECT id, tool_id, press_number, cycles, start, stop
	FROM cycles
	WHERE press_number = :press_number;
`

func ListCyclesByPressNumber(pressNumber shared.PressNumber) ([]shared.Cycle, *errors.MasterError) {
	rows, err := DBUser.Query(SQLListCyclesByPressNumber, sql.Named("press_number", int64(pressNumber)))
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	defer rows.Close()

	var cycles []shared.Cycle
	for rows.Next() {
		cycle, merr := ScanCycle(rows)
		if merr != nil {
			return nil, merr
		}
		cycles = append(cycles, *cycle)
	}

	return cycles, nil
}

const SQLDeleteCycle string = `
	DELETE FROM cycles
	WHERE id = :id;
`

func DeleteCycle(id shared.EntityID) *errors.MasterError {
	_, err := DBUser.Exec(SQLDeleteCycle, sql.Named("id", int64(id)))
	if err != nil {
		return errors.NewMasterError(err, 0)
	}
	return nil
}

// -----------------------------------------------------------------------------
// Table Helpers: "press_regenerations"
// -----------------------------------------------------------------------------

// -----------------------------------------------------------------------------
// Scan Helpers
// -----------------------------------------------------------------------------

func ScanCycle(row Scannable) (cycle *shared.Cycle, merr *errors.MasterError) {
	cycle = &shared.Cycle{}
	err := row.Scan(
		&cycle.ID,
		&cycle.ToolID,
		&cycle.PressNumber,
		&cycle.PressCycles,
		&cycle.Start,
		&cycle.Stop,
	)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	return cycle, nil
}
