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

	// -----------------------------------------------------------------------------
	// Upper Metal Sheets Queries
	// -----------------------------------------------------------------------------

	sqlAddUpperMetalSheet string = `
		INSERT INTO metal_sheets (tool_id, tile_height, value, type)
		VALUES (:tool_id, :tile_height, :value, 'upper');
	`

	sqlUpdateUpperMetalSheet string = `
		UPDATE metal_sheets
		SET 
			tool_id 	= :tool_id,
			tile_height = :tile_height,
			value 		= :value
		WHERE id = :id AND type = 'upper';
	`

	sqlGetUpperMetalSheet string = `
		SELECT id, tool_id, tile_height, value
		FROM metal_sheets
		WHERE id = :id AND type = 'upper'
		ORDER BY tile_height ASC, value ASC;
	`

	sqlListUpperMetalSheetsByTool string = `
		SELECT id, tool_id, tile_height, value
		FROM metal_sheets
		WHERE tool_id = :tool_id AND type = 'upper'
		ORDER BY tile_height ASC, value ASC;
	`

	// -----------------------------------------------------------------------------
	// Lower Metal Sheets Queries
	// -----------------------------------------------------------------------------

	sqlAddLowerMetalSheet string = `
		INSERT INTO metal_sheets (tool_id, tile_height, value, type, marke_height, stf, stf_max, identifier)
		VALUES (:tool_id, :tile_height, :value, 'lower', :marke_height, :stf, :stf_max, :identifier);
	`

	sqlUpdateLowerMetalSheet string = `
		UPDATE metal_sheets
		SET 
			tool_id 		= :tool_id,
			tile_height 	= :tile_height,
			value 			= :value,
			marke_height 	= :marke_height,
			stf 			= :stf,
			stf_max 		= :stf_max,
			identifier 		= :identifier
		WHERE id = :id AND type = 'lower';
	`

	sqlGetLowerMetalSheet string = `
		SELECT id, tool_id, tile_height, value, marke_height, stf, stf_max, identifier
		FROM metal_sheets
		WHERE id = :id AND type = 'lower'
		ORDER BY tile_height ASC, value ASC;
	`

	sqlListLowerMetalSheetsByTool string = `
		SELECT id, tool_id, tile_height, value, marke_height, stf, stf_max, identifier
		FROM metal_sheets
		WHERE tool_id = :tool_id AND type = 'lower'
		ORDER BY tile_height ASC, value ASC;
	`
)

// -----------------------------------------------------------------------------
// Upper Metal Sheets
// -----------------------------------------------------------------------------

func AddUpperMetalSheet(ums *shared.UpperMetalSheet) *errors.MasterError {
	if verr := ums.Validate(); verr != nil {
		return verr.MasterError().Wrap("invalid upper metal sheet")
	}

	_, err := dbTool.Exec(sqlAddUpperMetalSheet,
		sql.Named("tool_id", ums.ToolID),
		sql.Named("tile_height", ums.TileHeight),
		sql.Named("value", ums.Value),
	)
	if err != nil {
		return errors.NewMasterError(err)
	}
	return nil
}

func UpdateUpperMetalSheet(ums *shared.UpperMetalSheet) *errors.MasterError {
	if verr := ums.Validate(); verr != nil {
		return verr.MasterError().Wrap("invalid upper metal sheet")
	}

	_, err := dbTool.Exec(sqlUpdateUpperMetalSheet,
		sql.Named("id", ums.ID),
		sql.Named("tool_id", ums.ToolID),
		sql.Named("tile_height", ums.TileHeight),
		sql.Named("value", ums.Value),
	)
	if err != nil {
		return errors.NewMasterError(err)
	}
	return nil
}

func GetUpperMetalSheet(metalSheetID shared.EntityID) (*shared.UpperMetalSheet, *errors.MasterError) {
	r := dbTool.QueryRow(sqlGetUpperMetalSheet, sql.Named("id", metalSheetID))
	ums, merr := ScanUpperMetalSheet(r)
	if merr != nil {
		return nil, merr
	}
	return ums, nil
}

func ListUpperMetalSheetsByTool(toolID shared.EntityID) ([]*shared.UpperMetalSheet, *errors.MasterError) {
	rows, err := dbTool.Query(sqlListUpperMetalSheetsByTool, sql.Named("tool_id", toolID))
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

func AddLowerMetalSheet(lms *shared.LowerMetalSheet) *errors.MasterError {
	if verr := lms.Validate(); verr != nil {
		return verr.MasterError().Wrap("invalid lower metal sheet")
	}

	_, err := dbTool.Exec(sqlAddLowerMetalSheet,
		sql.Named("tool_id", lms.ToolID),
		sql.Named("tile_height", lms.TileHeight),
		sql.Named("value", lms.Value),
		sql.Named("marke_height", lms.MarkeHeight),
		sql.Named("stf", lms.STF),
		sql.Named("stf_max", lms.STFMax),
		sql.Named("identifier", lms.Identifier),
	)
	if err != nil {
		return errors.NewMasterError(err)
	}
	return nil
}

func UpdateLowerMetalSheet(lms *shared.LowerMetalSheet) *errors.MasterError {
	if verr := lms.Validate(); verr != nil {
		return verr.MasterError().Wrap("invalid lower metal sheet")
	}

	_, err := dbTool.Exec(sqlUpdateLowerMetalSheet,
		sql.Named("id", lms.ID),
		sql.Named("tool_id", lms.ToolID),
		sql.Named("tile_height", lms.TileHeight),
		sql.Named("value", lms.Value),
		sql.Named("marke_height", lms.MarkeHeight),
		sql.Named("stf", lms.STF),
		sql.Named("stf_max", lms.STFMax),
		sql.Named("identifier", lms.Identifier),
	)
	if err != nil {
		return errors.NewMasterError(err)
	}
	return nil
}

func GetLowerMetalSheet(metalSheetID shared.EntityID) (*shared.LowerMetalSheet, *errors.MasterError) {
	r := dbTool.QueryRow(sqlGetLowerMetalSheet, sql.Named("id", metalSheetID))
	lms, merr := ScanLowerMetalSheet(r)
	if merr != nil {
		return nil, merr
	}
	return lms, nil
}

func ListLowerMetalSheetsByTool(toolID shared.EntityID) ([]*shared.LowerMetalSheet, *errors.MasterError) {
	rows, err := dbTool.Query(sqlListLowerMetalSheetsByTool, sql.Named("tool_id", toolID))
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
