package db

import (
	"database/sql"

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

func ListCyclesByPressNumber(pressNumber shared.PressNumber) ([]*shared.Cycle, *errors.MasterError) {
	rows, err := DBUser.Query(SQLListCyclesByPressNumber, sql.Named("press_number", int64(pressNumber)))
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	defer rows.Close()

	var cycles []*shared.Cycle
	for rows.Next() {
		cycle, merr := ScanCycle(rows)
		if merr != nil {
			return nil, merr
		}
		cycles = append(cycles, cycle)
	}

	var merr *errors.MasterError
	for _, c := range cycles {
		if merr = InjectPartialCycles(c); merr != nil {
			return nil, merr.Wrap("failed to inject partial cycles for ID %d", c.ID)
		}
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

const SQLTotalToolCycles string = `
`

func TotalToolCycles(id shared.EntityID) (int64, *errors.MasterError) // TODO: ...

const SQLGetPrevCycle string = `
	SELECT cycles
	FROM cycles
	WHERE press_number = ? AND stop <= ?
	ORDER BY stop DESC
	LIMIT 1;
`

// TODO: Take into account the last press regeneration when calculating partial cycles
func InjectPartialCycles(cycle *shared.Cycle) *errors.MasterError {
	var lastKnownCycles int64 = 0
	err := DBPress.QueryRow(SQLGetPrevCycle, cycle.PressNumber, cycle.Start).Scan(&lastKnownCycles)
	if err != nil {
		if err == sql.ErrNoRows {
			// No previous cycles found, return full cycles
			cycle.PartialCycles = cycle.PressCycles
		} else {
			cycle.PartialCycles = 0
		}
		return errors.NewMasterError(err, 0)
	}

	cycle.PartialCycles = cycle.PressCycles - lastKnownCycles
	return nil
}

// -----------------------------------------------------------------------------
// Table Helpers: "press_regenerations"
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

func ScanPressRegeneration(row Scannable) (pr *shared.PressRegeneration, merr *errors.MasterError) {
	pr = &shared.PressRegeneration{}
	err := row.Scan(
		&pr.ID,
		&pr.PressNumber,
		&pr.Start,
		&pr.Stop,
		&pr.Cycles,
	)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	return pr, nil
}
