package db

import (
	"database/sql"

	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"
)

// -----------------------------------------------------------------------------
// Table Creation Statements
// ------------------------------------------------------------------------------

const (
	SQLCreateMetalSheetsTable string = `
		CREATE TABLE IF NOT EXISTS metal_sheets (
			id 				INTEGER NOT NULL,
			tool_id 		INTEGER NOT NULL,
			tile_height 	REAL NOT NULL,
			value 			REAL NOT NULL,
			type 			TEXT NOT NULL,
			marke_height 	INTEGER,
			stf 			REAL,
			stf_max 		REAL,
			identifier 		TEXT,

			PRIMARY KEY("id" AUTOINCREMENT),
			FOREIGN KEY(tool_id) REFERENCES tools(id) ON DELETE CASCADE
		);
	`
)

// -----------------------------------------------------------------------------
// Scan Helpers
// -----------------------------------------------------------------------------

func ScanUpperMetalSheet(row Scannable) (*shared.UpperMetalSheet, *errors.MasterError) {
	var ums shared.UpperMetalSheet
	err := row.Scan(
		&ums.ID,
		&ums.ToolID,
		&ums.TileHeight,
		&ums.Value,
	)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	return &ums, nil
}

func ScanLowerMetalSheet(row Scannable) (*shared.LowerMetalSheet, *errors.MasterError) {
	var lms shared.LowerMetalSheet
	err := row.Scan(
		&lms.ID,
		&lms.ToolID,
		&lms.TileHeight,
		&lms.Value,
		&lms.MarkeHeight,
		&lms.STF,
		&lms.STFMax,
		&lms.Identifier,
	)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	return &lms, nil
}
