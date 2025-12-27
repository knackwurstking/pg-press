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
	sqlCreateMetalSheetsTable string = `
		CREATE TABLE IF NOT EXISTS metal_sheets (
			id				INTEGER NOT NULL,
			tool_id			INTEGER NOT NULL,
			tile_height 	REAL NOT NULL,
			value			REAL NOT NULL,
			type			TEXT NOT NULL,			-- 'upper' | 'lower'
			marke_height 	INTEGER,
			stf				REAL,
			stf_max 		REAL,
			identifier 		TEXT, 					-- e.g. 'SACMI' | 'SITI'

			PRIMARY KEY("id" AUTOINCREMENT),
			FOREIGN KEY(tool_id) REFERENCES tools(id) ON DELETE CASCADE
		);
	`

	// TODO: GetUpperMetalSheet

	sqlListUpperMetalSheetsByTool string = `
		SELECT id, tool_id, tile_height, value
		FROM metal_sheets
		WHERE tool_id = :tool_id AND type = 'upper';
	`

	// TODO: GetLowerMetalSheet

	sqlListLowerMetalSheetsByTool string = `
		SELECT id, tool_id, tile_height, value, marke_height, stf, stf_max, identifier
		FROM metal_sheets
		WHERE tool_id = :tool_id AND type = 'lower';
	`
)

// -----------------------------------------------------------------------------
// Upper Metal Sheets
// -----------------------------------------------------------------------------

// TODO: GetUpperMetalSheet

func ListUpperMetalSheetsByTool(toolID shared.EntityID) ([]*shared.UpperMetalSheet, *errors.MasterError) {
	rows, err := dbTool.Query(sqlListUpperMetalSheetsByTool, sql.Named("tool_id", int64(toolID)))
	if err != nil {
		return nil, errors.NewMasterError(err)
	}
	defer rows.Close()

	var metalSheets []*shared.UpperMetalSheet
	for rows.Next() {
		ums, merr := ScanUpperMetalSheet(rows)
		if merr != nil {
			return nil, merr.Wrap("could not scan upper metal sheet")
		}
		metalSheets = append(metalSheets, ums)
	}
	return metalSheets, nil
}

// -----------------------------------------------------------------------------
// Lower Metal Sheets
// -----------------------------------------------------------------------------

// TODO: GetLowerMetalSheet

func ListLowerMetalSheetsByTool(toolID shared.EntityID) ([]*shared.LowerMetalSheet, *errors.MasterError) {
	rows, err := dbTool.Query(sqlListLowerMetalSheetsByTool, sql.Named("tool_id", int64(toolID)))
	if err != nil {
		return nil, errors.NewMasterError(err)
	}
	defer rows.Close()

	var metalSheets []*shared.LowerMetalSheet
	for rows.Next() {
		lms, merr := ScanLowerMetalSheet(rows)
		if merr != nil {
			return nil, merr.Wrap("could not scan lower metal sheet")
		}
		metalSheets = append(metalSheets, lms)
	}
	return metalSheets, nil
}

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
		return nil, errors.NewMasterError(err)
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
		return nil, errors.NewMasterError(err)
	}
	return &lms, nil
}
