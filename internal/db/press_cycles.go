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

	sqlAddCycleWithID string = `
		INSERT INTO cycles (id, tool_id, press_number, cycles, stop)
		VALUES (:id, :tool_id, :press_number, :cycles, :stop)
	`

	sqlUpdateCycle string = `
		UPDATE cycles
		SET
			tool_id = :tool_id,
			press_number = :press_number,
			cycles = :cycles,
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
// Cycle Functions
// -----------------------------------------------------------------------------

// AddCycle adds a new cycle entry to the database
func AddCycle(cycle *shared.Cycle) *errors.HTTPError {
	if err := cycle.Validate(); err != nil {
		return err.HTTPError()
	}

	var query string
	if cycle.ID > 0 {
		query = sqlAddCycleWithID
	} else {
		query = sqlAddCycle
	}

	var queryArgs []any
	if cycle.ID > 0 {
		queryArgs = append(queryArgs, sql.Named("id", cycle.ID))
	}
	queryArgs = append(queryArgs,
		sql.Named("tool_id", cycle.ToolID),
		sql.Named("press_number", cycle.PressNumber),
		sql.Named("cycles", cycle.PressCycles),
		sql.Named("stop", cycle.Stop),
	)

	if _, err := dbPress.Exec(query, queryArgs...); err != nil {
		return errors.NewHTTPError(err)
	}

	return nil
}

// UpdateCycle updates an existing cycle entry in the database
func UpdateCycle(cycle *shared.Cycle) *errors.HTTPError {
	if err := cycle.Validate(); err != nil {
		return err.HTTPError()
	}

	_, err := dbPress.Exec(sqlUpdateCycle,
		sql.Named("id", cycle.ID),
		sql.Named("tool_id", cycle.ToolID),
		sql.Named("press_number", cycle.PressNumber),
		sql.Named("cycles", cycle.PressCycles),
		sql.Named("stop", cycle.Stop),
	)
	if err != nil {
		return errors.NewHTTPError(err)
	}
	return nil
}

// DeleteCycle removes a cycle entry from the database
func DeleteCycle(id shared.EntityID) *errors.HTTPError {
	_, err := dbPress.Exec(sqlDeleteCycle, sql.Named("id", int64(id)))
	if err != nil {
		return errors.NewHTTPError(err)
	}
	return nil
}

// GetCycle retrieves a cycle entry by its ID
func GetCycle(id shared.EntityID) (*shared.Cycle, *errors.HTTPError) {
	return ScanCycle(dbPress.QueryRow(sqlGetCycle, sql.Named("id", id)))
}

// TotalToolCycles since last tool regeneration
func GetTotalToolCycles(toolID shared.EntityID) (int64, *errors.HTTPError) {
	cycles, herr := ListToolCycles(toolID)
	if herr != nil {
		return 0, herr.Wrap("failed to list tool cycles for tool ID %d", toolID)
	}

	// Filter out cycles before last tool regeneration, if any
	regenerations, herr := ListToolRegenerationsByTool(toolID)
	if herr != nil && !herr.IsNotFoundError() {
		return 0, herr.Wrap("failed to list tool regenerations for tool ID %d", toolID)
	}
	if len(regenerations) > 0 {
		lastRegeneration := regenerations[0]
		var filteredCycles []*shared.Cycle
		for _, c := range cycles {
			if c.Stop > lastRegeneration.Stop {
				filteredCycles = append(filteredCycles, c)
			}
		}
		cycles = filteredCycles
	}

	// Inject partial cycles for each cycle
	var totalCycles int64 = 0
	for _, c := range cycles {
		totalCycles += c.PartialCycles
	}

	return totalCycles, nil
}

// ListToolCycles retrieves all cycle entries for a specific tool
func ListToolCycles(toolID shared.EntityID) ([]*shared.Cycle, *errors.HTTPError) {
	rows, err := dbPress.Query(sqlListToolCycles, sql.Named("tool_id", int64(toolID)))
	if err != nil {
		return nil, errors.NewHTTPError(err)
	}

	var cycles []*shared.Cycle
	for rows.Next() {
		cycle, herr := ScanCycle(rows)
		if herr != nil {
			rows.Close()
			return nil, herr
		}
		cycles = append(cycles, cycle)
	}
	rows.Close()

	var herr *errors.HTTPError
	for _, c := range cycles {
		if herr = CycleInject(c); herr != nil {
			return nil, herr.Wrap("failed to inject into cycle with ID %d", c.ID)
		}
	}

	return cycles, nil
}

// ListCyclesByPressNumber retrieves all cycle entries for a specific press number
func ListCyclesByPressNumber(pressNumber shared.PressNumber) ([]*shared.Cycle, *errors.HTTPError) {
	rows, err := dbPress.Query(sqlListCyclesByPressNumber, sql.Named("press_number", int64(pressNumber)))
	if err != nil {
		return nil, errors.NewHTTPError(err)
	}
	defer rows.Close()

	var cycles []*shared.Cycle
	for rows.Next() {
		cycle, herr := ScanCycle(rows)
		if herr != nil {
			return nil, herr
		}
		cycles = append(cycles, cycle)
	}

	var herr *errors.HTTPError
	for _, c := range cycles {
		if herr = CycleInject(c); herr != nil {
			return nil, herr.Wrap("failed to inject into cycle with ID %d", c.ID)
		}
	}

	return cycles, nil
}

// CycleInject injects "start" and `PartialCycles` into cycle
func CycleInject(cycle *shared.Cycle) *errors.HTTPError {
	var cycleOffset int64 = 0
	press, herr := GetPress(cycle.PressNumber)
	if herr != nil && !herr.IsNotFoundError() {
		return herr.Wrap("failed to get press %d for cycle injection", cycle.PressNumber)
	}
	if press != nil { // Press could be not found
		cycleOffset = press.CyclesOffset
	}

	// Inject partial cycles (press cycle offset)
	var lastCycles int64 = 0
	var lastStop int64 = 0
	err := dbPress.QueryRow(sqlGetPrevCycle, cycle.PressNumber, cycle.Stop).Scan(&lastCycles, &lastStop)
	if err != nil && err == sql.ErrNoRows {
		cycle.PartialCycles = cycle.PressCycles - cycleOffset
	} else if err != nil {
		return errors.NewHTTPError(err)
	} else {
		cycle.PartialCycles = cycle.PressCycles - lastCycles
	}

	// Inject start time
	cycle.Start = shared.UnixMilli(lastStop)

	return nil
}

// -----------------------------------------------------------------------------
// Scan Helpers
// -----------------------------------------------------------------------------

// ScanCycle scans a database row into a Cycle struct
func ScanCycle(row Scannable) (cycle *shared.Cycle, herr *errors.HTTPError) {
	cycle = &shared.Cycle{}
	err := row.Scan(
		&cycle.ID,
		&cycle.ToolID,
		&cycle.PressNumber,
		&cycle.PressCycles,
		&cycle.Stop,
	)
	if err != nil {
		return nil, errors.NewHTTPError(err)
	}
	return cycle, nil
}
