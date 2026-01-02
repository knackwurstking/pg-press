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
			stop         INTEGER NOT NULL,

			PRIMARY KEY("id")
		);
	`

	sqlAddCycle string = `
		INSERT INTO cycles (tool_id, press_number, cycles, stop)
		VALUES (:tool_id, :press_number, :cycles, :stop)
	`

	sqlUpdateCycle string = `
		UPDATE cycles
		SET
			tool_id = :tool_id,
			press_number = :press_number
			cycles = :cycles
			stop = :stop
		WHERE id = :id
	`

	sqlDeleteCycle string = `
		DELETE FROM cycles
		WHERE id = :id;
	`

	sqlGetCycle string = `
		SELECT id, tool_id, press_number, cycles, stop
		FROM cycles
		WHERE id = :id
	`

	sqlListToolCycles string = `
		SELECT id, tool_id, press_number, cycles, stop
		FROM cycles
		WHERE tool_id = :tool_id
		ORDER BY stop DESC;
	`

	sqlListCyclesByPressNumber string = `
		SELECT id, tool_id, press_number, cycles, stop
		FROM cycles
		WHERE press_number = :press_number
		ORDER BY stop DESC;
	`

	sqlGetPrevCycle string = `
		SELECT cycles, stop
		FROM cycles
		WHERE press_number = ? AND stop < ?
		ORDER BY stop DESC
		LIMIT 1;
	`
)

// -----------------------------------------------------------------------------
// Table Helpers: "cycles"
// -----------------------------------------------------------------------------

func AddCycle(cycle *shared.Cycle) *errors.MasterError {
	if verr := cycle.Validate(); verr != nil {
		return verr.MasterError()
	}

	_, err := dbPress.Exec(sqlAddCycle,
		sql.Named("tool_id", cycle.ToolID),
		sql.Named("press_number", cycle.PressNumber),
		sql.Named("cycles", cycle.PressCycles),
		sql.Named("stop", cycle.Stop),
	)
	if err != nil {
		return errors.NewMasterError(err)
	}

	return nil
}

func UpdateCycle(cycle *shared.Cycle) *errors.MasterError {
	if verr := cycle.Validate(); verr != nil {
		return verr.MasterError()
	}

	_, err := dbPress.Exec(sqlUpdateCycle,
		sql.Named("id", cycle.ID),
		sql.Named("tool_id", cycle.ToolID),
		sql.Named("press_number", cycle.PressNumber),
		sql.Named("cycles", cycle.PressCycles),
		sql.Named("stop", cycle.Stop),
	)
	if err != nil {
		return errors.NewMasterError(err)
	}
	return nil
}

func DeleteCycle(id shared.EntityID) *errors.MasterError {
	_, err := dbPress.Exec(sqlDeleteCycle, sql.Named("id", int64(id)))
	if err != nil {
		return errors.NewMasterError(err)
	}
	return nil
}

func GetCycle(id shared.EntityID) (*shared.Cycle, *errors.MasterError) {
	return ScanCycle(dbPress.QueryRow(sqlGetCycle, sql.Named("id", id)))
}

// TotalToolCycles since last tool regeneration
func GetTotalToolCycles(id shared.EntityID) (int64, *errors.MasterError) {
	cycles, merr := ListToolCycles(id)
	if merr != nil {
		return 0, merr.Wrap("failed to list tool cycles for tool ID %d", id)
	}

	// TODO: Get last tool regeneration

	// Inject partial cycles for each cycle
	var totalCycles int64 = 0
	for _, c := range cycles {
		totalCycles += c.PartialCycles
	}

	return totalCycles, nil
}

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
		if merr = CycleInject(c); merr != nil {
			return nil, merr.Wrap("failed to inject into cycle with ID %d", c.ID)
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
		if merr = CycleInject(c); merr != nil {
			return nil, merr.Wrap("failed to inject into cycle with ID %d", c.ID)
		}
	}

	return cycles, nil
}

// CycleInject injects "start" and `PartialCycles` into cycle
func CycleInject(cycle *shared.Cycle) *errors.MasterError {
	var lastCycles int64 = 0
	var lastStop int64 = 0
	err := dbPress.QueryRow(sqlGetPrevCycle, cycle.PressNumber, cycle.Stop).Scan(&lastCycles, &lastStop)
	if err != nil {
		if err == sql.ErrNoRows {
			// No previous cycles found, return full cycles
			cycle.PartialCycles = cycle.PressCycles
		} else {
			return errors.NewMasterError(err)
		}
	} else {
		cycle.PartialCycles = cycle.PressCycles - lastCycles
	}

	// TODO: Check press regenerations before calculating the partial cycles here

	cycle.Start = shared.UnixMilli(lastStop)

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
		&cycle.Stop,
	)
	if err != nil {
		return nil, errors.NewMasterError(err)
	}
	return cycle, nil
}
