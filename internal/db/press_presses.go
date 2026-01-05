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

	sqlGetPressUtilizations string = `
		SELECT
			id,
			slot_up,
			slot_down,
			type
		FROM presses;
	`

	// TODO: delete SQL query
)

// -----------------------------------------------------------------------------
// Press Functions
// -----------------------------------------------------------------------------

func AddPress(press *shared.Press) *errors.MasterError {
	if verr := press.Validate(); verr != nil {
		return verr.MasterError()
	}

	_, err := dbPress.Exec(sqlAddPress,
		sql.Named("id", press.ID),
		sql.Named("slot_up", press.SlotUp),
		sql.Named("slot_down", press.SlotDown),
		sql.Named("cycles_offset", press.CyclesOffset),
		sql.Named("type", press.Type),
	)
	if err != nil {
		return errors.NewMasterError(err)
	}
	return nil
}

func UpdatePress(press *shared.Press) *errors.MasterError {
	if verr := press.Validate(); verr != nil {
		return verr.MasterError()
	}

	_, err := dbPress.Exec(sqlUpdatePress,
		sql.Named("id", press.ID),
		sql.Named("slot_up", press.SlotUp),
		sql.Named("slot_down", press.SlotDown),
		sql.Named("cycles_offset", press.CyclesOffset),
		sql.Named("type", press.Type),
	)
	if err != nil {
		return errors.NewMasterError(err)
	}
	return nil
}

func GetPress(id shared.EntityID) (*shared.Press, *errors.MasterError) {
	return ScanPress(dbPress.QueryRow(sqlGetPress, sql.Named("id", id)))
}

func GetPressNumberForTool(toolID shared.EntityID) (shared.PressNumber, *errors.MasterError) {
	var pressNumber shared.PressNumber = -1

	err := dbPress.QueryRow(sqlGetPressNumberForTool, sql.Named("tool_id", toolID)).Scan(&pressNumber)
	if err != nil && err != sql.ErrNoRows {
		return pressNumber, errors.NewMasterError(err)
	}

	return pressNumber, nil
}

func GetPressUtilizations(pressNumbers ...shared.PressNumber) (
	pu map[shared.PressNumber]*shared.PressUtilization, merr *errors.MasterError,
) {
	pu = make(map[shared.PressNumber]*shared.PressUtilization)

	r, err := dbPress.Query(sqlGetPressUtilizations)
	if err != nil {
		return nil, errors.NewMasterError(fmt.Errorf("error querying press utilizations: %v", err))
	}
	defer r.Close()

	for r.Next() {
		var (
			pressNumber shared.PressNumber
			slotUp      shared.EntityID
			slotDown    shared.EntityID
			pressType   shared.MachineType
		)
		err := r.Scan(
			&pressNumber,
			&slotUp,
			&slotDown,
			&pressType,
		)
		if err != nil {
			return pu, errors.NewMasterError(err)
		}

		pu[pressNumber] = &shared.PressUtilization{
			PressNumber: pressNumber,
			Type:        pressType,
		}
		if slotUp > 0 {
			pu[pressNumber].SlotUpper = &shared.Tool{ID: slotUp}
		}
		if slotDown > 0 {
			pu[pressNumber].SlotLower = &shared.Tool{ID: slotDown}
		}
	}

	for _, v := range pu {
		if v.SlotUpper != nil {
			tool, merr := GetTool(v.SlotUpper.ID)
			if merr != nil {
				return nil, merr
			}
			v.SlotUpper = tool

			if tool.Cassette > 0 {
				cassette, merr := GetTool(tool.Cassette)
				if merr != nil {
					return nil, merr.Wrap(
						"error getting upper cassette tool (%d) for tool ID %d",
						tool.Cassette, tool.ID,
					)
				}
				v.SlotUpperCassette = cassette
			}
		}

		if v.SlotLower != nil {
			tool, merr := GetTool(v.SlotLower.ID)
			if merr != nil {
				return nil, merr
			}
			v.SlotLower = tool

			// NOTE: Only to upper tool can contain a cassette for now
		}
	}

	if err = r.Err(); err != nil {
		return nil, errors.NewMasterError(fmt.Errorf("error iterating over press utilizations: %v", err))
	}

	return pu, nil
}

// TODO: Delete SQL function

// -----------------------------------------------------------------------------
// Scan Helpers
// -----------------------------------------------------------------------------

func ScanPress(row Scannable) (press *shared.Press, merr *errors.MasterError) {
	press = &shared.Press{}
	err := row.Scan(
		&press.ID,
		&press.SlotUp,
		&press.SlotDown,
		&press.CyclesOffset,
		&press.Type,
	)
	if err != nil {
		return nil, errors.NewMasterError(err)
	}
	return press, nil
}
