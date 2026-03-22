package db

import (
	"database/sql"
	"fmt"

	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"
)

// -----------------------------------------------------------------------------
// Table Creation Statements
// -----------------------------------------------------------------------------

const (
	sqlCreateCyclesTable string = `
CREATE TABLE IF NOT EXISTS cycles (
	id INTEGER NOT NULL,
	tool_id INTEGER NOT NULL,
	press_id INTEGER NOT NULL,
	cycles INTEGER NOT NULL, -- Cycles at stop time
	stop INTEGER NOT NULL,

	PRIMARY KEY("id" AUTOINCREMENT)
);

CREATE INDEX IF NOT EXISTS idx_cycles_press_stop ON cycles(press_id, stop DESC);`

	sqlAddCycle string = `
INSERT INTO cycles (tool_id, press_id, cycles, stop)
VALUES (:tool_id, :press_id, :cycles, :stop)`

	sqlAddCycleWithID string = `
INSERT INTO cycles (id, tool_id, press_id, cycles, stop)
VALUES (:id, :tool_id, :press_id, :cycles, :stop)`

	sqlUpdateCycle string = `
UPDATE cycles
SET
	tool_id = :tool_id,
	press_id = :press_id,
	cycles = :cycles,
	stop = :stop
WHERE id = :id`

	sqlDeleteCycle string = `
DELETE FROM cycles
WHERE id = :id;`

	sqlGetCycle string = `
SELECT id, tool_id, press_id, cycles, stop
FROM cycles
WHERE id = :id`

	sqlListToolCycles string = `
SELECT id, tool_id, press_id, cycles, stop
FROM cycles
WHERE tool_id = :tool_id
ORDER BY stop DESC;`

	sqlListCyclesByPressID string = `
SELECT id, tool_id, press_id, cycles, stop
FROM cycles
WHERE press_id = :press_id
ORDER BY stop DESC;`

	sqlListPrevCycles string = `
SELECT tool_id, cycles, stop
FROM cycles
WHERE press_id = ? AND stop < ?
ORDER BY stop DESC;`
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
		sql.Named("press_id", cycle.PressID),
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
		sql.Named("press_id", cycle.PressID),
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
	_, err := dbPress.Exec(sqlDeleteCycle, sql.Named("id", id))
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

func ListCyclesByPressID(pressID shared.EntityID) ([]*shared.Cycle, *errors.HTTPError) {
	rows, err := dbPress.Query(sqlListCyclesByPressID, sql.Named("press_id", pressID))
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
	// Helper to fetch the position of a tool, needed to determine if the cycle
	// is relevant for the current position
	fetchPosition := func(toolID shared.EntityID) (shared.Slot, *errors.HTTPError) {
		var position shared.Slot
		err := dbTool.QueryRow(
			`SELECT position FROM tools WHERE id = ?;`,
			toolID,
		).Scan(&position)
		if err != nil {
			return 0, errors.NewHTTPError(fmt.Errorf(
				"failed to fetch position for tool ID %d: %w",
				toolID, err,
			))
		}
		return position, nil
	}

	// Helper to fetch the last cycles for all tools in the same press before
	// the current stop
	type prevCycle struct {
		ToolID     shared.EntityID
		LastCycles int64
		LastStop   int64
	}

	fetchLastCycles := func(pressID shared.EntityID, stop int64) (data []prevCycle, herr *errors.HTTPError) {
		r, err := dbPress.Query(sqlListPrevCycles, pressID, stop)
		if err != nil {
			herr = errors.NewHTTPError(err)
		}
		defer r.Close()
		// Scan rows
		for r.Next() {
			pc := prevCycle{}
			if err = r.Scan(&pc.ToolID, &pc.LastCycles, &pc.LastStop); err != nil {
				herr = errors.NewHTTPError(err)
				return
			}
			data = append(data, pc)
		}
		return
	}

	// Helper to determine if a slot matches the current position, only the
	// upper cassette can match the upper tool
	isPosition := func(slot, currentPosition shared.Slot) bool {
		return slot == currentPosition ||
			(currentPosition == shared.SlotUpperCassette &&
				(slot == shared.SlotUpper || slot == shared.SlotUpperCassette))
	}

	// Get the cycles offset for the press, needed to calculate the partial cycles
	var cycleOffset int64 = 0
	press, herr := GetPress(cycle.PressID)
	if herr != nil && !herr.IsNotFoundError() {
		return herr.Wrap("failed to get press ID %d for cycle injection", cycle.PressID)
	}
	if press != nil { // Press could be not found
		cycleOffset = press.CyclesOffset
	}

	// Get the current position of the tool, needed to determine if the cycle
	// is relevant for the current position
	currentPosition, herr := fetchPosition(cycle.ToolID)
	if herr != nil {
		return herr
	}

	// Get the last cycles for all tools in the same press before the current stop
	prevCycles, herr := fetchLastCycles(cycle.PressID, int64(cycle.Stop))
	if herr != nil && herr.Err() != sql.ErrNoRows {
		return errors.NewHTTPError(herr)
	}

	if len(prevCycles) == 0 {
		cycle.PartialCycles = cycle.PressCycles - cycleOffset
		cycle.Start = cycle.Stop
		return nil
	}

	// Iterate over the previous cycles to find the most recent one that matches
	// the current position and tool, then calculate the partial cycles.
	for i, pc := range prevCycles {
		if i == len(prevCycles)-1 {
			// If we are at the last previous cycle, we can calculate the partial cycles
			// using the cycle offset, as there are no more previous cycles to compare to
			cycle.PartialCycles = cycle.PressCycles - cycleOffset
			cycle.Start = cycle.Stop
			break
		}

		// Check the tool_id for if the position is matching, only the
		// upper cassette can match the upper tool too
		slot, herr := fetchPosition(pc.ToolID)
		if herr != nil {
			return herr
		}

		if isPosition(slot, currentPosition) && (cycle.ToolID != pc.ToolID || int64(cycle.Stop) != pc.LastStop) {
			cycle.PartialCycles = cycle.PressCycles - pc.LastCycles
			cycle.Start = shared.UnixMilli(pc.LastStop)
			break
		}
	}

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
		&cycle.PressID,
		&cycle.PressCycles,
		&cycle.Stop,
	)
	if err != nil {
		return nil, errors.NewHTTPError(err)
	}
	return cycle, nil
}
