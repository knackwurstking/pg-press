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
	sqlCreatePressesTable string = `
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

	sqlGetPressByToolID string = `
		SELECT
			id,
		FROM presses
		WHERE slot_up = :tool_id OR slot_down = :tool_id
		LIMIT 1;
	`
)

func GetPressNumberForTool(toolID shared.EntityID) (shared.PressNumber, *errors.MasterError) {
	var pressNumber shared.PressNumber = -1

	err := dbPress.QueryRow(sqlGetPressByToolID, sql.Named("tool_id", toolID)).Scan(&pressNumber)
	if err != nil {
		return pressNumber, errors.NewMasterError(err)
	}

	return pressNumber, nil
}

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
		return nil, errors.NewMasterError(err)
	}
	return press, nil
}
