package metalsheet

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/knackwurstking/pgpress/internal/database/dberror"
	"github.com/knackwurstking/pgpress/internal/database/interfaces"
	feedmodels "github.com/knackwurstking/pgpress/internal/database/models/feed"
	metalsheetmodels "github.com/knackwurstking/pgpress/internal/database/models/metalsheet"
	modmodels "github.com/knackwurstking/pgpress/internal/database/models/mod"
	notemodels "github.com/knackwurstking/pgpress/internal/database/models/note"
	usermodels "github.com/knackwurstking/pgpress/internal/database/models/user"
	"github.com/knackwurstking/pgpress/internal/database/services/feed"
	"github.com/knackwurstking/pgpress/internal/database/services/note"
	"github.com/knackwurstking/pgpress/internal/logger"
)

// Service represents a collection of metal sheets in the database.
type Service struct {
	db    *sql.DB
	feeds *feed.Service
	notes *note.Service
}

var _ interfaces.DataOperations[*metalsheetmodels.MetalSheet] = (*Service)(nil)

// NewMetalSheets creates a new MetalSheets instance
func New(db *sql.DB, feeds *feed.Service, notes *note.Service) *Service {
	query := `
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
	`

	if _, err := db.Exec(query); err != nil {
		panic(
			dberror.NewDatabaseError(
				"create_table",
				"metal_sheets",
				"failed to create metal_sheets table",
				err,
			),
		)
	}

	return &Service{
		db:    db,
		feeds: feeds,
		notes: notes,
	}
}

// List returns all metal sheets
func (s *Service) List() ([]*metalsheetmodels.MetalSheet, error) {
	logger.DBMetalSheets().Info("Listing metal sheets")

	query := `
		SELECT id, tile_height, value, marke_height, stf, stf_max, tool_id, notes, mods
		FROM metal_sheets
		ORDER BY id DESC;
	`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, dberror.NewDatabaseError("select", "metal_sheets",
			"failed to query metal sheets", err)
	}
	defer rows.Close()

	var sheets []*metalsheetmodels.MetalSheet

	for rows.Next() {
		sheet, err := s.scanMetalSheet(rows)
		if err != nil {
			return nil, dberror.WrapError(err, "failed to scan metal sheet")
		}
		sheets = append(sheets, sheet)
	}

	if err := rows.Err(); err != nil {
		return nil, dberror.NewDatabaseError("select", "metal_sheets",
			"error iterating over rows", err)
	}

	logger.DBMetalSheets().Debug("Listed %d metal sheets", len(sheets))
	return sheets, nil
}

// Get returns a metal sheet by ID
func (s *Service) Get(id int64) (*metalsheetmodels.MetalSheet, error) {
	logger.DBMetalSheets().Info("Getting metal sheet, id: %d", id)

	query := `
		SELECT id, tile_height, value, marke_height, stf, stf_max, tool_id, notes, mods
		FROM metal_sheets
		WHERE id = $1;
	`
	row := s.db.QueryRow(query, id)

	sheet, err := s.scanMetalSheet(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, dberror.ErrNotFound
		}
		return nil, dberror.NewDatabaseError("select", "metal_sheets",
			fmt.Sprintf("failed to get metal sheet with ID %d", id), err)
	}

	return sheet, nil
}

// GetWithNotes returns a metal sheet with its related notes loaded
func (s *Service) GetWithNotes(id int64) (*metalsheetmodels.MetalSheetWithNotes, error) {
	logger.DBMetalSheets().Info("Getting metal sheet with notes, id: %d", id)

	sheet, err := s.Get(id)
	if err != nil {
		return nil, err
	}

	result := &metalsheetmodels.MetalSheetWithNotes{
		MetalSheet:  sheet,
		LoadedNotes: []*notemodels.Note{},
	}

	// Load notes if there are any linked
	if len(sheet.LinkedNotes) > 0 && s.notes != nil {
		logger.DBMetalSheets().Debug("Loading %d notes for metal sheet %d", len(sheet.LinkedNotes), id)
		notes, err := s.notes.GetByIDs(sheet.LinkedNotes)
		if err != nil {
			logger.DBMetalSheets().Error("Failed to load notes for metal sheet %d: %v", id, err)
			// Don't fail the entire operation if notes can't be loaded
		} else {
			result.LoadedNotes = notes
		}
	}

	return result, nil
}

// GetByToolID returns all metal sheets assigned to a specific tool
func (s *Service) GetByToolID(toolID int64) ([]*metalsheetmodels.MetalSheet, error) {
	logger.DBMetalSheets().Info("Getting metal sheets for tool, id: %d", toolID)

	query := `
		SELECT id, tile_height, value, marke_height, stf, stf_max, tool_id, notes, mods
		FROM metal_sheets
		WHERE tool_id = $1
		ORDER BY id DESC;
	`
	rows, err := s.db.Query(query, toolID)
	if err != nil {
		return nil, dberror.NewDatabaseError("select", "metal_sheets",
			fmt.Sprintf("failed to query metal sheets for tool ID %d", toolID), err)
	}
	defer rows.Close()

	var sheets []*metalsheetmodels.MetalSheet

	for rows.Next() {
		sheet, err := s.scanMetalSheet(rows)
		if err != nil {
			return nil, dberror.WrapError(err, "failed to scan metal sheet")
		}
		sheets = append(sheets, sheet)
	}

	if err := rows.Err(); err != nil {
		return nil, dberror.NewDatabaseError("select", "metal_sheets",
			"error iterating over rows", err)
	}

	logger.DBMetalSheets().Debug("Found %d metal sheets for tool %d", len(sheets), toolID)
	return sheets, nil
}

// GetAvailable returns all available metal sheets
func (s *Service) GetAvailable() ([]*metalsheetmodels.MetalSheet, error) {
	logger.DBMetalSheets().Info("Getting available metal sheets")

	query := `
		SELECT id, tile_height, value, marke_height, stf, stf_max, tool_id, notes, mods
		FROM metal_sheets
		WHERE tool_id IS NULL
		ORDER BY id DESC;
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, dberror.NewDatabaseError("select", "metal_sheets",
			"failed to query available metal sheets", err)
	}
	defer rows.Close()

	var sheets []*metalsheetmodels.MetalSheet

	for rows.Next() {
		sheet, err := s.scanMetalSheet(rows)
		if err != nil {
			return nil, dberror.WrapError(err, "failed to scan metal sheet")
		}
		sheets = append(sheets, sheet)
	}

	if err := rows.Err(); err != nil {
		return nil, dberror.NewDatabaseError("select", "metal_sheets",
			"error iterating over rows", err)
	}

	logger.DBMetalSheets().Debug("Found %d available metal sheets", len(sheets))
	return sheets, nil
}

// Add inserts a new metal sheet
func (s *Service) Add(sheet *metalsheetmodels.MetalSheet, user *usermodels.User) (int64, error) {
	logger.DBMetalSheets().Info("Adding metal sheet: %s", sheet.String())

	// Ensure initial mod entry exists
	if len(sheet.Mods) == 0 {
		initialMod := modmodels.NewMod(user, metalsheetmodels.MetalSheetMod{
			TileHeight:  sheet.TileHeight,
			Value:       sheet.Value,
			MarkeHeight: sheet.MarkeHeight,
			STF:         sheet.STF,
			STFMax:      sheet.STFMax,
			ToolID:      sheet.ToolID,
			LinkedNotes: sheet.LinkedNotes,
		})
		sheet.Mods = []*modmodels.Mod[metalsheetmodels.MetalSheetMod]{initialMod}
	}

	// Marshal JSON fields
	notesBytes, err := json.Marshal(sheet.LinkedNotes)
	if err != nil {
		return 0, dberror.NewDatabaseError("insert", "metal_sheets",
			"failed to marshal notes", err)
	}

	modsBytes, err := json.Marshal(sheet.Mods)
	if err != nil {
		return 0, dberror.NewDatabaseError("insert", "metal_sheets",
			"failed to marshal mods", err)
	}

	query := `
		INSERT INTO metal_sheets (tile_height, value, marke_height, stf, stf_max, tool_id, notes, mods)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8);
	`

	result, err := s.db.Exec(query,
		sheet.TileHeight, sheet.Value, sheet.MarkeHeight, sheet.STF, sheet.STFMax,
		sheet.ToolID, notesBytes, modsBytes)
	if err != nil {
		return 0, dberror.NewDatabaseError("insert", "metal_sheets",
			"failed to insert metal sheet", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, dberror.NewDatabaseError("insert", "metal_sheets", "failed to get last insert ID", err)
	}

	// Set sheet ID for return
	sheet.ID = id

	logger.DBMetalSheets().Debug("Added metal sheet with ID %d", id)

	// Trigger feed update
	if s.feeds != nil {
		s.feeds.Add(feedmodels.New(
			feedmodels.TypeMetalSheetAdd,
			&feedmodels.MetalSheetAdd{
				ID:         id,
				MetalSheet: sheet.String(),
				ModifiedBy: user,
			},
		))
	}

	return id, nil
}

// Update updates an existing metal sheet
func (s *Service) Update(sheet *metalsheetmodels.MetalSheet, user *usermodels.User) error {
	logger.DBMetalSheets().Info("Updating metal sheet: %d", sheet.ID)

	// Get current sheet to compare for changes
	current, err := s.Get(sheet.ID)
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

		mod := modmodels.NewMod(user, metalsheetmodels.MetalSheetMod{
			TileHeight:  current.TileHeight,
			Value:       current.Value,
			MarkeHeight: current.MarkeHeight,
			STF:         current.STF,
			STFMax:      current.STFMax,
			ToolID:      current.ToolID,
			LinkedNotes: current.LinkedNotes,
		})
		// Prepend new mod to keep most recent first
		sheet.Mods = append([]*modmodels.Mod[metalsheetmodels.MetalSheetMod]{mod}, sheet.Mods...)
	}

	// Marshal JSON fields
	notesBytes, err := json.Marshal(sheet.LinkedNotes)
	if err != nil {
		return dberror.NewDatabaseError("update", "metal_sheets",
			"failed to marshal notes", err)
	}

	modsBytes, err := json.Marshal(sheet.Mods)
	if err != nil {
		return dberror.NewDatabaseError("update", "metal_sheets",
			"failed to marshal mods", err)
	}

	query := `
		UPDATE metal_sheets
		SET tile_height = $1, value = $2, marke_height = $3, stf = $4, stf_max = $5,
		    tool_id = $6, notes = $7, mods = $8, updated_at = CURRENT_TIMESTAMP
		WHERE id = $9;
	`

	_, err = s.db.Exec(query,
		sheet.TileHeight, sheet.Value, sheet.MarkeHeight, sheet.STF, sheet.STFMax,
		sheet.ToolID, notesBytes, modsBytes, sheet.ID)
	if err != nil {
		return dberror.NewDatabaseError("update", "metal_sheets",
			fmt.Sprintf("failed to update metal sheet with ID %d", sheet.ID), err)
	}

	logger.DBMetalSheets().Debug("Updated metal sheet with ID %d", sheet.ID)

	// Trigger feed update
	if s.feeds != nil {
		s.feeds.Add(feedmodels.New(
			feedmodels.TypeMetalSheetUpdate,
			&feedmodels.MetalSheetUpdate{
				ID:         sheet.ID,
				MetalSheet: sheet.String(),
				ModifiedBy: user,
			},
		))
	}

	return nil
}

// AssignTool assigns a metal sheet to a tool
func (s *Service) AssignTool(sheetID int64, toolID *int64, user *usermodels.User) error {
	logger.DBMetalSheets().Info("Assigning tool to metal sheet: sheet_id=%d, tool_id=%v", sheetID, toolID)

	// Get current sheet to track changes
	sheet, err := s.Get(sheetID)
	if err != nil {
		return fmt.Errorf("failed to get sheet for tool assignment: %w", err)
	}

	// Add modification record before changing
	mod := modmodels.NewMod(user, metalsheetmodels.MetalSheetMod{
		TileHeight:  sheet.TileHeight,
		Value:       sheet.Value,
		MarkeHeight: sheet.MarkeHeight,
		STF:         sheet.STF,
		STFMax:      sheet.STFMax,
		ToolID:      sheet.ToolID,
		LinkedNotes: sheet.LinkedNotes,
	})
	// Prepend new mod to keep most recent first
	sheet.Mods = append(modmodels.Mods[metalsheetmodels.MetalSheetMod]{mod}, sheet.Mods...)

	// Update the tool assignment
	sheet.ToolID = toolID

	// Marshal mods for database update
	modsBytes, err := json.Marshal(sheet.Mods)
	if err != nil {
		return dberror.NewDatabaseError("update", "metal_sheets",
			"failed to marshal mods", err)
	}

	// Update both tool_id and mods in database
	_, err = s.db.Exec(
		`UPDATE metal_sheets
		SET tool_id = $1, mods = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3`,
		toolID, modsBytes, sheetID,
	)
	if err != nil {
		return dberror.NewDatabaseError("update", "metal_sheets",
			fmt.Sprintf("failed to assign tool to metal sheet ID %d", sheetID), err)
	}

	logger.DBMetalSheets().Debug("Assigned tool %v to metal sheet %d", toolID, sheetID)

	// Trigger feed update
	if s.feeds != nil {
		s.feeds.Add(feedmodels.New(
			feedmodels.TypeMetalSheetToolAssignment,
			&feedmodels.MetalSheetToolAssignment{
				SheetID:    sheetID,
				ToolID:     toolID,
				ModifiedBy: user,
			},
		))
	}

	return nil
}

// Delete deletes a metal sheet
func (s *Service) Delete(id int64, user *usermodels.User) error {
	logger.DBMetalSheets().Info("Deleting metal sheet: %d", id)

	query := `
		DELETE FROM metal_sheets WHERE id = $1;
	`
	_, err := s.db.Exec(query, id)
	if err != nil {
		return dberror.NewDatabaseError("delete", "metal_sheets",
			fmt.Sprintf("failed to delete metal sheet with ID %d", id), err)
	}

	logger.DBMetalSheets().Debug("Deleted metal sheet with ID %d", id)

	// Trigger feed update
	if s.feeds != nil {
		s.feeds.Add(feedmodels.New(
			feedmodels.TypeMetalSheetDelete,
			&feedmodels.MetalSheetDelete{
				ID:         id,
				ModifiedBy: user,
			},
		))
	}

	return nil
}

func (s *Service) scanMetalSheet(scanner interfaces.Scannable) (*metalsheetmodels.MetalSheet, error) {
	logger.DBMetalSheets().Debug("Scanning metal sheet")
	sheet := &metalsheetmodels.MetalSheet{}

	var (
		linkedNotes []byte
		mods        []byte
		toolID      sql.NullInt64
	)

	if err := scanner.Scan(&sheet.ID, &sheet.TileHeight, &sheet.Value, &sheet.MarkeHeight, &sheet.STF, &sheet.STFMax,
		&toolID, &linkedNotes, &mods); err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, dberror.NewDatabaseError("scan", "metal_sheets",
			"failed to scan row", err)
	}

	// Handle nullable tool_id
	if toolID.Valid {
		sheet.ToolID = &toolID.Int64
	}

	if err := json.Unmarshal(linkedNotes, &sheet.LinkedNotes); err != nil {
		return nil, dberror.NewDatabaseError("scan", "metal_sheets",
			"failed to unmarshal notes", err)
	}

	if err := json.Unmarshal(mods, &sheet.Mods); err != nil {
		return nil, dberror.WrapError(err, "failed to unmarshal mods data")
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
