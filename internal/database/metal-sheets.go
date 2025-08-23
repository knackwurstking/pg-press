package database

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/knackwurstking/pgpress/internal/logger"
)

const (
	createMetalSheetsTableQuery = `
		DROP TABLE IF EXISTS metal_sheets;
		CREATE TABLE IF NOT EXISTS metal_sheets (
			id INTEGER NOT NULL,
			tile_height REAL NOT NULL,
			value REAL NOT NULL,
			marke_height INTEGER NOT NULL,
			stf REAL NOT NULL,
			stf_max REAL NOT NULL,
			tool_id INTEGER,
			notes BLOB NOT NULL,
			mods BLOB NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(tool_id) REFERENCES tools(id) ON DELETE SET NULL,
			PRIMARY KEY("id" AUTOINCREMENT)
		);
		INSERT INTO metal_sheets (tile_height, value, marke_height, stf, stf_max, tool_id, notes, mods)
		VALUES
			(6, 4, -1, -1, -1, 1, '[]', '[]'),
			(6, 25, 45, 12.8, 19.8, 2, '[]', '[]');
	`

	selectAllMetalSheetsQuery = `
		SELECT id, tile_height, value, marke_height, stf, stf_max, tool_id, notes, mods
		FROM metal_sheets
		ORDER BY id DESC;
	`

	selectMetalSheetByIDQuery = `
		SELECT id, tile_height, value, marke_height, stf, stf_max, tool_id, notes, mods
		FROM metal_sheets
		WHERE id = $1;
	`

	selectMetalSheetsByToolIDQuery = `
		SELECT id, tile_height, value, marke_height, stf, stf_max, tool_id, notes, mods
		FROM metal_sheets
		WHERE tool_id = $1
		ORDER BY id DESC;
	`

	selectAvailableMetalSheetsQuery = `
		SELECT id, tile_height, value, marke_height, stf, stf_max, tool_id, notes, mods
		FROM metal_sheets
		WHERE tool_id IS NULL
		ORDER BY id DESC;
	`

	insertMetalSheetQuery = `
		INSERT INTO metal_sheets (tile_height, value, marke_height, stf, stf_max, tool_id, notes, mods)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8);
	`

	updateMetalSheetQuery = `
		UPDATE metal_sheets
		SET tile_height = $1, value = $2, marke_height = $3, stf = $4, stf_max = $5,
		    tool_id = $6, notes = $7, mods = $8, updated_at = CURRENT_TIMESTAMP
		WHERE id = $9;
	`

	// Status query removed - no longer have status column

	updateMetalSheetToolQuery = `
		UPDATE metal_sheets
		SET tool_id = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2;
	`

	deleteMetalSheetQuery = `
		DELETE FROM metal_sheets WHERE id = $1;
	`
)

// MetalSheets represents a collection of metal sheets in the database.
type MetalSheets struct {
	db    *sql.DB
	feeds *Feeds
	notes *Notes
}

// NewMetalSheets creates a new MetalSheets instance
func NewMetalSheets(db *sql.DB, feeds *Feeds, notes *Notes) *MetalSheets {
	if _, err := db.Exec(createMetalSheetsTableQuery); err != nil {
		panic(
			NewDatabaseError(
				"create_table",
				"metal_sheets",
				"failed to create metal_sheets table",
				err,
			),
		)
	}

	return &MetalSheets{
		db:    db,
		feeds: feeds,
		notes: notes,
	}
}

// List returns all metal sheets
func (ms *MetalSheets) List() ([]*MetalSheet, error) {
	logger.MetalSheets().Info("Listing metal sheets")

	rows, err := ms.db.Query(selectAllMetalSheetsQuery)
	if err != nil {
		return nil, NewDatabaseError("select", "metal_sheets",
			"failed to query metal sheets", err)
	}
	defer rows.Close()

	var sheets []*MetalSheet

	for rows.Next() {
		sheet, err := ms.scanMetalSheetFromRows(rows)
		if err != nil {
			return nil, WrapError(err, "failed to scan metal sheet")
		}
		sheets = append(sheets, sheet)
	}

	if err := rows.Err(); err != nil {
		return nil, NewDatabaseError("select", "metal_sheets",
			"error iterating over rows", err)
	}

	return sheets, nil
}

// Get returns a metal sheet by ID
func (ms *MetalSheets) Get(id int64) (*MetalSheet, error) {
	logger.MetalSheets().Info("Getting metal sheet, id: %d", id)

	row := ms.db.QueryRow(selectMetalSheetByIDQuery, id)

	sheet, err := ms.scanMetalSheetFromRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, NewDatabaseError("select", "metal_sheets",
			fmt.Sprintf("failed to get metal sheet with ID %d", id), err)
	}

	return sheet, nil
}

// GetWithNotes returns a metal sheet with its related notes loaded
func (ms *MetalSheets) GetWithNotes(id int64) (*MetalSheetWithNotes, error) {
	logger.MetalSheets().Info("Getting metal sheet with notes, id: %d", id)

	sheet, err := ms.Get(id)
	if err != nil {
		return nil, err
	}

	result := &MetalSheetWithNotes{
		MetalSheet:  sheet,
		LoadedNotes: []*Note{},
	}

	// Load notes if there are any linked
	if len(sheet.LinkedNotes) > 0 && ms.notes != nil {
		notes, err := ms.notes.GetByIDs(sheet.LinkedNotes)
		if err != nil {
			logger.MetalSheets().Error("Failed to load notes for metal sheet %d: %v", id, err)
			// Don't fail the entire operation if notes can't be loaded
		} else {
			result.LoadedNotes = notes
		}
	}

	return result, nil
}

// GetByToolID returns all metal sheets assigned to a specific tool
func (ms *MetalSheets) GetByToolID(toolID int64) ([]*MetalSheet, error) {
	logger.MetalSheets().Info("Getting metal sheets for tool, id: %d", toolID)

	rows, err := ms.db.Query(selectMetalSheetsByToolIDQuery, toolID)
	if err != nil {
		return nil, NewDatabaseError("select", "metal_sheets",
			fmt.Sprintf("failed to query metal sheets for tool ID %d", toolID), err)
	}
	defer rows.Close()

	var sheets []*MetalSheet

	for rows.Next() {
		sheet, err := ms.scanMetalSheetFromRows(rows)
		if err != nil {
			return nil, WrapError(err, "failed to scan metal sheet")
		}
		sheets = append(sheets, sheet)
	}

	if err := rows.Err(); err != nil {
		return nil, NewDatabaseError("select", "metal_sheets",
			"error iterating over rows", err)
	}

	return sheets, nil
}

// GetAvailable returns all available metal sheets
func (ms *MetalSheets) GetAvailable() ([]*MetalSheet, error) {
	logger.MetalSheets().Info("Getting available metal sheets")

	rows, err := ms.db.Query(selectAvailableMetalSheetsQuery)
	if err != nil {
		return nil, NewDatabaseError("select", "metal_sheets",
			"failed to query available metal sheets", err)
	}
	defer rows.Close()

	var sheets []*MetalSheet

	for rows.Next() {
		sheet, err := ms.scanMetalSheetFromRows(rows)
		if err != nil {
			return nil, WrapError(err, "failed to scan metal sheet")
		}
		sheets = append(sheets, sheet)
	}

	if err := rows.Err(); err != nil {
		return nil, NewDatabaseError("select", "metal_sheets",
			"error iterating over rows", err)
	}

	return sheets, nil
}

// Add inserts a new metal sheet
func (ms *MetalSheets) Add(sheet *MetalSheet, user *User) (int64, error) {
	logger.MetalSheets().Info("Adding metal sheet: %s", sheet.String())

	// Ensure initial mod entry exists
	if len(sheet.Mods) == 0 {
		initialMod := NewModified(user, MetalSheetMod{
			TileHeight:  sheet.TileHeight,
			Value:       sheet.Value,
			MarkeHeight: sheet.MarkeHeight,
			STF:         sheet.STF,
			STFMax:      sheet.STFMax,
			ToolID:      sheet.ToolID,
			LinkedNotes: sheet.LinkedNotes,
		})
		sheet.Mods = []*Modified[MetalSheetMod]{initialMod}
	}

	// Marshal JSON fields
	notesBytes, err := json.Marshal(sheet.LinkedNotes)
	if err != nil {
		return 0, NewDatabaseError("insert", "metal_sheets",
			"failed to marshal notes", err)
	}

	modsBytes, err := json.Marshal(sheet.Mods)
	if err != nil {
		return 0, NewDatabaseError("insert", "metal_sheets",
			"failed to marshal mods", err)
	}

	result, err := ms.db.Exec(insertMetalSheetQuery,
		sheet.TileHeight, sheet.Value, sheet.MarkeHeight, sheet.STF, sheet.STFMax,
		sheet.ToolID, notesBytes, modsBytes)
	if err != nil {
		return 0, NewDatabaseError("insert", "metal_sheets",
			"failed to insert metal sheet", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, NewDatabaseError("insert", "metal_sheets",
			"failed to get last insert ID", err)
	}

	// Set sheet ID for return
	sheet.ID = id

	// Trigger feed update
	if ms.feeds != nil {
		ms.feeds.Add(NewFeed(
			FeedTypeMetalSheetAdd,
			&FeedMetalSheetAdd{
				ID:         id,
				MetalSheet: sheet.String(),
				ModifiedBy: user,
			},
		))
	}

	return id, nil
}

// Update updates an existing metal sheet
func (ms *MetalSheets) Update(sheet *MetalSheet, user *User) error {
	logger.MetalSheets().Info("Updating metal sheet: %d", sheet.ID)

	// Get current sheet to compare for changes
	current, err := ms.Get(sheet.ID)
	if err != nil {
		return fmt.Errorf("failed to get current sheet: %w", err)
	}

	// Add modification record if values changed
	if current.TileHeight != sheet.TileHeight ||
		current.Value != sheet.Value ||
		current.MarkeHeight != sheet.MarkeHeight ||
		current.STF != sheet.STF ||
		current.STFMax != sheet.STFMax ||
		!equalToolIDs(current.ToolID, sheet.ToolID) ||
		len(current.LinkedNotes) != len(sheet.LinkedNotes) {

		mod := NewModified(user, MetalSheetMod{
			TileHeight:  current.TileHeight,
			Value:       current.Value,
			MarkeHeight: current.MarkeHeight,
			STF:         current.STF,
			STFMax:      current.STFMax,
			ToolID:      current.ToolID,
			LinkedNotes: current.LinkedNotes,
		})
		// Prepend new mod to keep most recent first
		sheet.Mods = append([]*Modified[MetalSheetMod]{mod}, sheet.Mods...)
	}

	// Marshal JSON fields
	notesBytes, err := json.Marshal(sheet.LinkedNotes)
	if err != nil {
		return NewDatabaseError("update", "metal_sheets",
			"failed to marshal notes", err)
	}

	modsBytes, err := json.Marshal(sheet.Mods)
	if err != nil {
		return NewDatabaseError("update", "metal_sheets",
			"failed to marshal mods", err)
	}

	_, err = ms.db.Exec(updateMetalSheetQuery,
		sheet.TileHeight, sheet.Value, sheet.MarkeHeight, sheet.STF, sheet.STFMax,
		sheet.ToolID, notesBytes, modsBytes, sheet.ID)
	if err != nil {
		return NewDatabaseError("update", "metal_sheets",
			fmt.Sprintf("failed to update metal sheet with ID %d", sheet.ID), err)
	}

	// Trigger feed update
	if ms.feeds != nil {
		ms.feeds.Add(NewFeed(
			FeedTypeMetalSheetUpdate,
			&FeedMetalSheetUpdate{
				ID:         sheet.ID,
				MetalSheet: sheet.String(),
				ModifiedBy: user,
			},
		))
	}

	return nil
}

// AssignTool assigns a metal sheet to a tool
func (ms *MetalSheets) AssignTool(sheetID int64, toolID *int64, user *User) error {
	logger.MetalSheets().Info("Assigning tool to metal sheet: sheet_id=%d, tool_id=%v", sheetID, toolID)

	// Get current sheet to track changes
	sheet, err := ms.Get(sheetID)
	if err != nil {
		return fmt.Errorf("failed to get sheet for tool assignment: %w", err)
	}

	// Add modification record before changing
	mod := NewModified(user, MetalSheetMod{
		TileHeight:  sheet.TileHeight,
		Value:       sheet.Value,
		MarkeHeight: sheet.MarkeHeight,
		STF:         sheet.STF,
		STFMax:      sheet.STFMax,
		ToolID:      sheet.ToolID,
		LinkedNotes: sheet.LinkedNotes,
	})
	// Prepend new mod to keep most recent first
	sheet.Mods = append([]*Modified[MetalSheetMod]{mod}, sheet.Mods...)

	// Update the tool assignment
	sheet.ToolID = toolID

	// Marshal mods for database update
	modsBytes, err := json.Marshal(sheet.Mods)
	if err != nil {
		return NewDatabaseError("update", "metal_sheets",
			"failed to marshal mods", err)
	}

	// Update both tool_id and mods in database
	_, err = ms.db.Exec(
		`UPDATE metal_sheets
		SET tool_id = $1, mods = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3`,
		toolID, modsBytes, sheetID,
	)
	if err != nil {
		return NewDatabaseError("update", "metal_sheets",
			fmt.Sprintf("failed to assign tool to metal sheet ID %d", sheetID), err)
	}

	// Trigger feed update
	if ms.feeds != nil {
		ms.feeds.Add(NewFeed(
			FeedTypeMetalSheetToolAssignment,
			&FeedMetalSheetToolAssignment{
				SheetID:    sheetID,
				ToolID:     toolID,
				ModifiedBy: user,
			},
		))
	}

	return nil
}

// Delete deletes a metal sheet
func (ms *MetalSheets) Delete(id int64, user *User) error {
	logger.MetalSheets().Info("Deleting metal sheet: %d", id)

	_, err := ms.db.Exec(deleteMetalSheetQuery, id)
	if err != nil {
		return NewDatabaseError("delete", "metal_sheets",
			fmt.Sprintf("failed to delete metal sheet with ID %d", id), err)
	}

	// Trigger feed update
	if ms.feeds != nil {
		ms.feeds.Add(NewFeed(
			FeedTypeMetalSheetDelete,
			&FeedMetalSheetDelete{
				ID:         id,
				ModifiedBy: user,
			},
		))
	}

	return nil
}

// scanMetalSheetFromRows scans a metal sheet from database rows
func (ms *MetalSheets) scanMetalSheetFromRows(rows *sql.Rows) (*MetalSheet, error) {
	sheet := &MetalSheet{}

	var (
		linkedNotes []byte
		mods        []byte
		toolID      sql.NullInt64
	)

	if err := rows.Scan(&sheet.ID, &sheet.TileHeight, &sheet.Value, &sheet.MarkeHeight, &sheet.STF, &sheet.STFMax,
		&toolID, &linkedNotes, &mods); err != nil {
		return nil, NewDatabaseError("scan", "metal_sheets",
			"failed to scan row", err)
	}

	// Handle nullable tool_id
	if toolID.Valid {
		sheet.ToolID = &toolID.Int64
	}

	if err := json.Unmarshal(linkedNotes, &sheet.LinkedNotes); err != nil {
		return nil, NewDatabaseError("scan", "metal_sheets",
			"failed to unmarshal notes", err)
	}

	if err := json.Unmarshal(mods, &sheet.Mods); err != nil {
		return nil, WrapError(err, "failed to unmarshal mods data")
	}

	return sheet, nil
}

// scanMetalSheetFromRow scans a metal sheet from a database row
func (ms *MetalSheets) scanMetalSheetFromRow(row *sql.Row) (*MetalSheet, error) {
	sheet := &MetalSheet{}

	var (
		linkedNotes []byte
		mods        []byte
		toolID      sql.NullInt64
	)

	if err := row.Scan(&sheet.ID, &sheet.TileHeight, &sheet.Value, &sheet.MarkeHeight, &sheet.STF, &sheet.STFMax,
		&toolID, &linkedNotes, &mods); err != nil {
		return nil, NewDatabaseError("scan", "metal_sheets",
			"failed to scan row", err)
	}

	// Handle nullable tool_id
	if toolID.Valid {
		sheet.ToolID = &toolID.Int64
	}

	if err := json.Unmarshal(linkedNotes, &sheet.LinkedNotes); err != nil {
		return nil, NewDatabaseError("scan", "metal_sheets",
			"failed to unmarshal notes", err)
	}

	if err := json.Unmarshal(mods, &sheet.Mods); err != nil {
		return nil, WrapError(err, "failed to unmarshal mods data")
	}

	return sheet, nil
}

// equalToolIDs compares two tool ID pointers for equality
func equalToolIDs(a, b *int64) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
