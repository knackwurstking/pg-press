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
	id INTEGER NOT NULL,
	tool_id INTEGER NOT NULL,
	tile_height REAL NOT NULL,
	value REAL NOT NULL,
	type TEXT NOT NULL, -- 'upper' | 'lower'
	marke_height INTEGER,
	stf REAL,
	stf_max REAL,
	identifier TEXT, -- e.g. 'SACMI' | 'SITI'

	PRIMARY KEY("id" AUTOINCREMENT),
	FOREIGN KEY(tool_id) REFERENCES tools(id) ON DELETE CASCADE
);`

	// -----------------------------------------------------------------------------
	// Upper Metal Sheets Queries
	// -----------------------------------------------------------------------------

	sqlAddUpperMetalSheet string = `
INSERT INTO metal_sheets (tool_id, tile_height, value, type)
VALUES (:tool_id, :tile_height, :value, 'upper');`

	sqlAddUpperMetalSheetWithID string = `
INSERT INTO metal_sheets (id, tool_id, tile_height, value, type)
VALUES (:id, :tool_id, :tile_height, :value, 'upper');`

	sqlUpdateUpperMetalSheet string = `
UPDATE metal_sheets
SET
	tool_id = :tool_id,
	tile_height = :tile_height,
	value = :value
WHERE id = :id AND type = 'upper';`

	sqlGetUpperMetalSheet string = `
SELECT id, tool_id, tile_height, value
FROM metal_sheets
WHERE id = :id AND type = 'upper'
ORDER BY tile_height ASC, value ASC;`

	sqlListUpperMetalSheetsByTool string = `
SELECT id, tool_id, tile_height, value
FROM metal_sheets
WHERE tool_id = :tool_id AND type = 'upper'
ORDER BY tile_height ASC, value ASC;`

	// -----------------------------------------------------------------------------
	// Lower Metal Sheets Queries
	// -----------------------------------------------------------------------------

	sqlAddLowerMetalSheet string = `
INSERT INTO metal_sheets (tool_id, tile_height, value, type, marke_height, stf, stf_max, identifier)
VALUES (:tool_id, :tile_height, :value, 'lower', :marke_height, :stf, :stf_max, :identifier);`
	sqlAddLowerMetalSheetWithID string = `
INSERT INTO metal_sheets (id, tool_id, tile_height, value, type, marke_height, stf, stf_max, identifier)
VALUES (:id, :tool_id, :tile_height, :value, 'lower', :marke_height, :stf, :stf_max, :identifier);`

	sqlUpdateLowerMetalSheet string = `
UPDATE metal_sheets
SET
	tool_id = :tool_id,
	tile_height = :tile_height,
	value = :value,
	marke_height = :marke_height,
	stf = :stf,
	stf_max = :stf_max,
	identifier = :identifier
WHERE id = :id AND type = 'lower';`

	sqlGetLowerMetalSheet string = `
SELECT id, tool_id, tile_height, value, marke_height, stf, stf_max, identifier
FROM metal_sheets
WHERE id = :id AND type = 'lower'
ORDER BY tile_height ASC, value ASC;`

	sqlListLowerMetalSheetsByTool string = `
SELECT id, tool_id, tile_height, value, marke_height, stf, stf_max, identifier
FROM metal_sheets
WHERE tool_id = :tool_id AND type = 'lower'
ORDER BY tile_height ASC, value ASC;`
)

// -----------------------------------------------------------------------------
// Upper Metal Sheets
// -----------------------------------------------------------------------------

// AddUpperMetalSheet adds a new upper metal sheet to the database
func AddUpperMetalSheet(ums *shared.UpperMetalSheet) *errors.HTTPError {
	if verr := ums.Validate(); verr != nil {
		return verr.HTTPError().Wrap("invalid upper metal sheet")
	}

	var query string
	if ums.ID > 0 {
		query = sqlAddUpperMetalSheetWithID
	} else {
		query = sqlAddUpperMetalSheet
	}

	var queryArgs []any
	if ums.ID > 0 {
		queryArgs = append(queryArgs, sql.Named("id", ums.ID))
	}
	queryArgs = append(queryArgs,
		sql.Named("tool_id", ums.ToolID),
		sql.Named("tile_height", ums.TileHeight),
		sql.Named("value", ums.Value),
	)

	if _, err := dbTool.Exec(query, queryArgs...); err != nil {
		return errors.NewHTTPError(err)
	}
	return nil
}

// UpdateUpperMetalSheet updates an existing upper metal sheet in the database
func UpdateUpperMetalSheet(ums *shared.UpperMetalSheet) *errors.HTTPError {
	if verr := ums.Validate(); verr != nil {
		return verr.HTTPError().Wrap("invalid upper metal sheet")
	}

	_, err := dbTool.Exec(sqlUpdateUpperMetalSheet,
		sql.Named("id", ums.ID),
		sql.Named("tool_id", ums.ToolID),
		sql.Named("tile_height", ums.TileHeight),
		sql.Named("value", ums.Value),
	)
	if err != nil {
		return errors.NewHTTPError(err)
	}
	return nil
}

// GetUpperMetalSheet retrieves an upper metal sheet by its ID
func GetUpperMetalSheet(metalSheetID shared.EntityID) (*shared.UpperMetalSheet, *errors.HTTPError) {
	r := dbTool.QueryRow(sqlGetUpperMetalSheet, sql.Named("id", metalSheetID))
	ums, merr := ScanUpperMetalSheet(r)
	if merr != nil {
		return nil, merr
	}
	return ums, nil
}

// ListUpperMetalSheetsByTool retrieves all upper metal sheets for a given tool
func ListUpperMetalSheetsByTool(toolID shared.EntityID) ([]*shared.UpperMetalSheet, *errors.HTTPError) {
	rows, err := dbTool.Query(sqlListUpperMetalSheetsByTool, sql.Named("tool_id", toolID))
	if err != nil {
		return nil, errors.NewHTTPError(err)
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

// DeleteUpperMetalSheet removes an upper metal sheet from the database
func DeleteUpperMetalSheet(id shared.EntityID) *errors.HTTPError {
	_, err := dbTool.Exec("DELETE FROM metal_sheets WHERE id = ? AND type = 'upper'", id)
	if err != nil {
		return errors.NewHTTPError(err)
	}
	return nil
}

// -----------------------------------------------------------------------------
// Lower Metal Sheets
// -----------------------------------------------------------------------------

// AddLowerMetalSheet adds a new lower metal sheet to the database
func AddLowerMetalSheet(lms *shared.LowerMetalSheet) *errors.HTTPError {
	if verr := lms.Validate(); verr != nil {
		return verr.HTTPError().Wrap("invalid lower metal sheet")
	}

	var query string
	{
		if lms.ID > 0 {
			query = sqlAddLowerMetalSheetWithID
		} else {
			query = sqlAddLowerMetalSheet
		}
	}

	var queryArgs []any
	{
		if lms.ID > 0 {
			queryArgs = append(queryArgs, sql.Named("id", lms.ID))
		}
		queryArgs = append(queryArgs,
			sql.Named("tool_id", lms.ToolID),
			sql.Named("tile_height", lms.TileHeight),
			sql.Named("value", lms.Value),
			sql.Named("marke_height", lms.MarkeHeight),
			sql.Named("stf", lms.STF),
			sql.Named("stf_max", lms.STFMax),
			sql.Named("identifier", lms.Identifier),
		)
	}

	if _, err := dbTool.Exec(query, queryArgs...); err != nil {
		return errors.NewHTTPError(err)
	}

	return nil
}

// UpdateLowerMetalSheet updates an existing lower metal sheet in the database
func UpdateLowerMetalSheet(lms *shared.LowerMetalSheet) *errors.HTTPError {
	if verr := lms.Validate(); verr != nil {
		return verr.HTTPError().Wrap("invalid lower metal sheet")
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
		return errors.NewHTTPError(err)
	}
	return nil
}

// GetLowerMetalSheet retrieves a lower metal sheet by its ID
func GetLowerMetalSheet(metalSheetID shared.EntityID) (*shared.LowerMetalSheet, *errors.HTTPError) {
	r := dbTool.QueryRow(sqlGetLowerMetalSheet, sql.Named("id", metalSheetID))
	lms, merr := ScanLowerMetalSheet(r)
	if merr != nil {
		return nil, merr
	}
	return lms, nil
}

// ListLowerMetalSheetsByTool retrieves all lower metal sheets for a given tool
func ListLowerMetalSheetsByTool(toolID shared.EntityID) ([]*shared.LowerMetalSheet, *errors.HTTPError) {
	rows, err := dbTool.Query(sqlListLowerMetalSheetsByTool, sql.Named("tool_id", toolID))
	if err != nil {
		return nil, errors.NewHTTPError(err)
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

// DeleteLowerMetalSheet removes a lower metal sheet from the database
func DeleteLowerMetalSheet(id shared.EntityID) *errors.HTTPError {
	_, err := dbTool.Exec("DELETE FROM metal_sheets WHERE id = ? AND type = 'lower'", id)
	if err != nil {
		return errors.NewHTTPError(err)
	}
	return nil
}

// -----------------------------------------------------------------------------
// Scan Helpers
// -----------------------------------------------------------------------------

// ScanUpperMetalSheet scans a database row into an UpperMetalSheet struct
func ScanUpperMetalSheet(row Scannable) (*shared.UpperMetalSheet, *errors.HTTPError) {
	var ums shared.UpperMetalSheet
	err := row.Scan(
		&ums.ID,
		&ums.ToolID,
		&ums.TileHeight,
		&ums.Value,
	)
	if err != nil {
		return nil, errors.NewHTTPError(err)
	}
	return &ums, nil
}

// ScanLowerMetalSheet scans a database row into a LowerMetalSheet struct
func ScanLowerMetalSheet(row Scannable) (*shared.LowerMetalSheet, *errors.HTTPError) {
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
		return nil, errors.NewHTTPError(err)
	}
	return &lms, nil
}
