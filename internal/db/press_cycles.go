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
	sqlCreateCyclesTable string = `
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

	sqlListToolCycles string = `
		SELECT id, tool_id, press_number, cycles, start, stop
		FROM cycles
		WHERE tool_id = :tool_id;
	`

	sqlListCyclesByPressNumber string = `
		SELECT id, tool_id, press_number, cycles, start, stop
		FROM cycles
		WHERE press_number = :press_number;
	`

	sqlDeleteCycle string = `
		DELETE FROM cycles
		WHERE id = :id;
	`

	sqlGetPrevCycle string = `
		SELECT cycles
		FROM cycles
		WHERE press_number = ? AND stop <= ?
		ORDER BY stop DESC
		LIMIT 1;
	`
)

// -----------------------------------------------------------------------------
// Table Helpers: "cycles"
// -----------------------------------------------------------------------------

func ListToolCycles(toolID shared.EntityID) ([]*shared.Cycle, *errors.MasterError) {
	rows, err := dbPress.Query(sqlListToolCycles, sql.Named("tool_id", int64(toolID)))
	if err != nil {
		return nil, errors.NewMasterError(err)
	}

	var cycles []*shared.Cycle
	for rows.Next() {
		cycle, merr := ScanCycle(rows)
		if merr != nil {
			rows.Close()
			return nil, merr
		}
		cycles = append(cycles, cycle)
	}
	rows.Close()

	var merr *errors.MasterError
	for _, c := range cycles {
		if merr = InjectPartialCyclesIntoCycle(c); merr != nil {
			return nil, merr.Wrap("failed to inject partial cycles for ID %d", c.ID)
		}
	}

	return cycles, nil
}

func ListCyclesByPressNumber(pressNumber shared.PressNumber) ([]*shared.Cycle, *errors.MasterError) {
	rows, err := dbPress.Query(sqlListCyclesByPressNumber, sql.Named("press_number", int64(pressNumber)))
	if err != nil {
		return nil, errors.NewMasterError(err)
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
		if merr = InjectPartialCyclesIntoCycle(c); merr != nil {
			return nil, merr.Wrap("failed to inject partial cycles for ID %d", c.ID)
		}
	}

	return cycles, nil
}

func DeleteCycle(id shared.EntityID) *errors.MasterError {
	_, err := dbPress.Exec(sqlDeleteCycle, sql.Named("id", int64(id)))
	if err != nil {
		return errors.NewMasterError(err)
	}
	return nil
}

// TotalToolCycles since last tool regeneration
//
// TODO: Take into account tool regenerations
func GetTotalToolCycles(id shared.EntityID) (int64, *errors.MasterError) {
	cycles, merr := ListToolCycles(id)
	if merr != nil {
		return 0, merr.Wrap("failed to list tool cycles for tool ID %d", id)
	}

	// Inject partial cycles for each cycle
	var totalCycles int64 = 0
	for _, c := range cycles {
		totalCycles += c.PartialCycles
	}

	return totalCycles, nil
}

// InjectPartialCyclesIntoCycle calculates and injects the partial cycles into the given cycle
//
// TODO: Take into account the last press regeneration when calculating partial cycles
func InjectPartialCyclesIntoCycle(cycle *shared.Cycle) *errors.MasterError {
	var lastKnownCycles int64 = 0
	err := dbPress.QueryRow(sqlGetPrevCycle, cycle.PressNumber, cycle.Start).Scan(&lastKnownCycles)
	if err != nil {
		if err == sql.ErrNoRows {
			// No previous cycles found, return full cycles
			cycle.PartialCycles = cycle.PressCycles
		} else {
			cycle.PartialCycles = 0
		}
		return errors.NewMasterError(err)
	}

	cycle.PartialCycles = cycle.PressCycles - lastKnownCycles
	return nil
}

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
		return nil, errors.NewMasterError(err)
	}
	return cycle, nil
}
