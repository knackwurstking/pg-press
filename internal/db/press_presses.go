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
	sqlCreatePressesTable string = `
		CREATE TABLE IF NOT EXISTS presses (
			id 					INTEGER NOT NULL,
			slot_up 			INTEGER NOT NULL,
			slot_down 			INTEGER NOT NULL,
			cycles_offset 		INTEGER NOT NULL,
			type 				TEXT NOT NULL,

			PRIMARY KEY("id")
		);
	`

	sqlAddPress string = `
		INSERT INTO presses (
			id,
			slot_up,
			slot_down,
			cycles_offset,
			type
		) VALUES (
			:id,
			:slot_up,
			:slot_down,
			:cycles_offset,
			:type
		)
	`

	sqlUpdatePress string = `
		UPDATE presses
		SET
			slot_up = :slot_up,
			slot_down = :slot_down,
			cycles_offset = :cycles_offset,
			type = :type
		WHERE id = :id
	`

	sqlGetPress string = `
		SELECT
			id,
			slot_up,
			slot_down,
			cycles_offset,
			type
		FROM presses
		WHERE id = :id
	`

	sqlGetPressNumberForTool string = `
		SELECT id
		FROM presses
		WHERE slot_up = :tool_id OR slot_down = :tool_id
		LIMIT 1;
	`

	sqlGetPressUtilization string = `
		SELECT
			id,
			slot_up,
			slot_down,
			cycles_offset,
			type
		FROM presses
		WHERE id = :press_number;
	`

	sqlDeletePress string = `
		DELETE FROM presses
		WHERE id = :id
	`
)

// -----------------------------------------------------------------------------
// Press Functions
// -----------------------------------------------------------------------------

// AddPress adds a new press to the database
func AddPress(press *shared.Press) *errors.HTTPError {
	if verr := press.Validate(); verr != nil {
		return verr.HTTPError()
	}

	_, err := dbPress.Exec(sqlAddPress,
		sql.Named("id", press.ID),
		sql.Named("slot_up", press.SlotUp),
		sql.Named("slot_down", press.SlotDown),
		sql.Named("cycles_offset", press.CyclesOffset),
		sql.Named("type", press.Type),
	)
	if err != nil {
		return errors.NewHTTPError(err)
	}
	return nil
}

// UpdatePress updates an existing press in the database
func UpdatePress(press *shared.Press) *errors.HTTPError {
	if verr := press.Validate(); verr != nil {
		return verr.HTTPError()
	}

	_, err := dbPress.Exec(sqlUpdatePress,
		sql.Named("id", press.ID),
		sql.Named("slot_up", press.SlotUp),
		sql.Named("slot_down", press.SlotDown),
		sql.Named("cycles_offset", press.CyclesOffset),
		sql.Named("type", press.Type),
	)
	if err != nil {
		return errors.NewHTTPError(err)
	}
	return nil
}

// GetPress retrieves a press by its ID
func GetPress(id shared.PressNumber) (*shared.Press, *errors.HTTPError) {
	return ScanPress(dbPress.QueryRow(sqlGetPress, sql.Named("id", id)))
}

// GetPressNumberForTool retrieves the press number that has the given tool in either slot
func GetPressNumberForTool(toolID shared.EntityID) (shared.PressNumber, *errors.HTTPError) {
	var pressNumber shared.PressNumber = -1

	err := dbPress.QueryRow(sqlGetPressNumberForTool, sql.Named("tool_id", toolID)).Scan(&pressNumber)
	if err != nil && err != sql.ErrNoRows {
		return pressNumber, errors.NewHTTPError(err)
	}

	return pressNumber, nil
}

func GetPressUtilization(pressNumber shared.PressNumber) (*shared.PressUtilization, *errors.HTTPError) {
	var (
		slotUpper shared.EntityID
		slotLower shared.EntityID
	)
	pu := &shared.PressUtilization{}
	err := dbPress.QueryRow(sqlGetPressUtilization, sql.Named("press_number", pressNumber)).Scan(
		&pu.PressNumber,
		&slotUpper,
		&slotLower,
		&pu.CyclesOffset,
		&pu.Type,
	)
	if err != nil {
		return pu, errors.NewHTTPError(fmt.Errorf("error query press utilization for press %d failed: %v", pressNumber, err))
	}

	if slotUpper > 0 {
		tool, merr := GetTool(slotUpper)
		if merr != nil {
			return nil, merr
		}
		pu.SlotUpper = tool

		if tool.Cassette > 0 {
			cassette, merr := GetTool(tool.Cassette)
			if merr != nil {
				return nil, merr.Wrap(
					"error getting upper cassette tool (%d) for tool ID %d",
					tool.Cassette, tool.ID,
				)
			}
			pu.SlotUpperCassette = cassette
		}
	}
	if slotLower > 0 {
		tool, merr := GetTool(slotLower)
		if merr != nil {
			return nil, merr
		}
		pu.SlotLower = tool
	}

	return pu, nil
}

// GetPressUtilizations retrieves all press utilizations with tool information
func GetPressUtilizations(pressNumbers ...shared.PressNumber) (
	pu map[shared.PressNumber]*shared.PressUtilization, merr *errors.HTTPError,
) {
	pu = make(map[shared.PressNumber]*shared.PressUtilization)
	for _, pn := range pressNumbers {
		u, merr := GetPressUtilization(pn)
		if merr != nil {
			return pu, merr.Wrap("%d", pn)
		}
		pu[pn] = u
	}

	return pu, nil
}

// DeletePress removes a press from the database
func DeletePress(id shared.PressNumber) *errors.HTTPError {
	_, err := dbPress.Exec(sqlDeletePress, sql.Named("id", id))
	if err != nil {
		return errors.NewHTTPError(err)
	}
	return nil
}

// -----------------------------------------------------------------------------
// Scan Helpers
// -----------------------------------------------------------------------------

// ScanPress scans a database row into a Press struct
func ScanPress(row Scannable) (press *shared.Press, merr *errors.HTTPError) {
	press = &shared.Press{}
	err := row.Scan(
		&press.ID,
		&press.SlotUp,
		&press.SlotDown,
		&press.CyclesOffset,
		&press.Type,
	)
	if err != nil {
		return nil, errors.NewHTTPError(err)
	}
	return press, nil
}
