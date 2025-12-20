package db

import (
	"database/sql"
	"sync"
	"time"

	"github.com/knackwurstking/pg-press/internal/common"
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

	SQLCreateToolRegenerationsTable string = `
		CREATE TABLE IF NOT EXISTS tool_regenerations (
			id 		INTEGER NOT NULL,
			tool_id INTEGER NOT NULL,
			start 	INTEGER NOT NULL,
			stop 	INTEGER NOT NULL,
			cycles 	INTEGER NOT NULL,

			PRIMARY KEY("id" AUTOINCREMENT)
		);
	`

	SQLCreateToolsTable string = `
		CREATE TABLE IF NOT EXISTS tools (
			id					INTEGER NOT NULL,
			position 			INTEGER NOT NULL,
			width 				INTEGER NOT NULL,
			height 				INTEGER NOT NULL,
			type 				TEXT NOT NULL,
			code 				TEXT NOT NULL,
			cycles_offset 		INTEGER NOT NULL DEFAULT 0,
			cycles 				INTEGER NOT NULL DEFAULT 0,
			is_dead 			INTEGER NOT NULL DEFAULT 0,
			cassette			INTEGER NOT NULL DEFAULT 0,
			min_thickness		REAL NOT NULL DEFAULT 0,
			max_thickness		REAL NOT NULL DEFAULT 0,
			model_type			TEXT NOT NULL, -- e.g.: "tool", "cassette",

			PRIMARY KEY("id" AUTOINCREMENT)
		);
	`
)

// -----------------------------------------------------------------------------
// Tool Helpers
// ------------------------------------------------------------------------------

const sqlGetToolPositionByID = `SELECT position FROM tools WHERE id = :id;`

// GetToolByID retrieves a tool by its ID from the "tools" table and if it fails from the cassettes table.
func GetToolByID(db *common.DB, id shared.EntityID) (shared.ModelTool, *errors.MasterError) {
	toolsDB := db.Tool.UpperTools.DB() // Does not matter which database we use here

	var position shared.Slot
	err := toolsDB.QueryRow(sqlGetToolPositionByID, sql.Named("id", id)).Scan(&position)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}

	var dbFn func(shared.EntityID) (shared.ModelTool, *errors.MasterError)
	switch position {
	case shared.SlotUpper:
		dbFn = db.Tool.UpperTools.GetByID
	case shared.SlotLower:
		dbFn = db.Tool.LowerTools.GetByID
	case shared.SlotUpperCassette:
		dbFn = db.Tool.Cassettes.GetByID
	default:
		return nil, errors.NewValidationError(
			"invalid tool position for tool ID %d: %d", id, position,
		).MasterError()
	}
	return dbFn(id)
}

func ListTools(db *common.DB) ([]shared.ModelTool, *errors.MasterError) {
	allTools := make([]shared.ModelTool, 0)

	queryFn := []func() ([]shared.ModelTool, *errors.MasterError){
		db.Tool.UpperTools.List,
		db.Tool.LowerTools.List,
		db.Tool.Cassettes.List,
	}

	for _, fn := range queryFn {
		tools, merr := fn()
		if merr != nil {
			return nil, merr
		}

		allTools = append(allTools, tools...)
	}

	return allTools, nil
}

func ListDeadTools(db *common.DB) ([]shared.ModelTool, *errors.MasterError) {
	deadTools := make([]shared.ModelTool, 0)

	queryFn := []func() ([]shared.ModelTool, *errors.MasterError){
		db.Tool.UpperTools.List,
		db.Tool.LowerTools.List,
		db.Tool.Cassettes.List,
	}

	for _, fn := range queryFn {
		tools, merr := fn()
		if merr != nil {
			return nil, merr
		}

		for _, t := range tools {
			if t.GetBase().IsDead {
				deadTools = append(deadTools, t)
			}
		}
	}

	return deadTools, nil
}

func ListAvailableCassettesForBinding(db *common.DB, toolID shared.EntityID) ([]*shared.Cassette, *errors.MasterError) {
	tool, merr := GetToolByID(db, toolID)
	if merr != nil {
		return nil, merr.Wrap("could not get tool for ID %d", toolID)
	}
	if tool.GetBase().Position != shared.SlotUpper {
		return nil, errors.NewValidationError("tool ID %d is not an upper tool", toolID).MasterError()
	}

	cassettes, merr := db.Tool.Cassettes.List()
	if merr != nil {
		return nil, merr.Wrap("could not list cassettes")
	}

	// Filter cassettes based on the tool width and height
	bindableCassettes := make([]*shared.Cassette, 0)
	width := tool.GetBase().Width
	height := tool.GetBase().Height
	for _, c := range cassettes {
		// Skip dead cassettes or those that do not match the tool dimensions
		if c.GetBase().IsDead || c.GetBase().Width != width || c.GetBase().Height != height {
			continue
		}
		bindableCassettes = append(bindableCassettes, c.(*shared.Cassette))
	}

	return bindableCassettes, nil
}

// -----------------------------------------------------------------------------
// Tool Cassette Binding Helpers
// ------------------------------------------------------------------------------

var bindMutex = &sync.Mutex{}

func BindCassetteToTool(db *common.DB, toolID, cassetteID shared.EntityID) *errors.MasterError {
	// First, check if cassette exists
	_, merr := db.Tool.Cassettes.GetByID(cassetteID)
	if merr != nil {
		return errors.NewValidationError("cassette ID %d does not exist", cassetteID).MasterError()
	}

	bindMutex.Lock()
	defer bindMutex.Unlock()

	// Get the tool and check if it's an upper tool without a cassette bound to it
	mTool, merr := GetToolByID(db, toolID)
	if merr != nil {
		return merr
	}
	tool, ok := mTool.(*shared.Tool)
	if !ok {
		return errors.NewValidationError("tool ID %d is not an upper tool", toolID).MasterError()
	}
	if tool.Cassette > 0 {
		return errors.NewValidationError("tool already has a cassette bound").MasterError()
	}

	tool.Cassette = cassetteID
	merr = db.Tool.UpperTools.Update(tool)
	if merr != nil {
		return merr
	}

	return nil
}

func UnbindCassetteFromTool(db *common.DB, toolID shared.EntityID) *errors.MasterError {
	bindMutex.Lock()
	defer bindMutex.Unlock()

	mTool, merr := db.Tool.UpperTools.GetByID(toolID)
	if merr != nil {
		return merr
	}
	tool, _ := mTool.(*shared.Tool)

	tool.Cassette = 0
	merr = db.Tool.UpperTools.Update(tool)
	if merr != nil {
		return merr
	}

	return nil
}

// -----------------------------------------------------------------------------
// Tool Metal Sheet Helpers
// ------------------------------------------------------------------------------

func ListUpperMetalSheetsForTool(db *common.DB, toolID shared.EntityID) ([]*shared.UpperMetalSheet, *errors.MasterError) {
	metalSheets, merr := db.Tool.UpperMetalSheets.List()
	if merr != nil {
		return metalSheets, merr
	}

	i := 0
	for _, ms := range metalSheets {
		if ms.ToolID != toolID {
			continue
		}

		metalSheets[i] = ms
		i++
	}

	return metalSheets[:i], nil
}

func ListLowerMetalSheetsForTool(db *common.DB, toolID shared.EntityID) ([]*shared.LowerMetalSheet, *errors.MasterError) {
	metalSheets, merr := db.Tool.LowerMetalSheets.List()
	if merr != nil {
		return metalSheets, merr
	}

	i := 0
	for _, ms := range metalSheets {
		if ms.ToolID != toolID {
			continue
		}

		metalSheets[i] = ms
		i++
	}

	return metalSheets[:i], nil
}

// -----------------------------------------------------------------------------
// Tool Regeneration Helpers
// ------------------------------------------------------------------------------

const sqlSelectOngoingRegenerationForTool = `
	SELECT id, tool_id, start, stop, cycles FROM tool_regenerations WHERE tool_id = ? AND stop = 0
`

var regenerationsMutex = &sync.Mutex{}

func GetRegenerationsForTool(db *common.DB, toolID shared.EntityID) (
	regenerations []*shared.ToolRegeneration, merr *errors.MasterError,
) {
	regenerations, merr = db.Tool.Regenerations.List()
	if merr != nil {
		return nil, merr
	}

	n := 0
	for _, r := range regenerations {
		if r.ToolID != toolID {
			continue
		}

		regenerations[n] = r
		n++
	}

	return regenerations[:n], nil
}

func StartToolRegeneration(db *common.DB, toolID shared.EntityID) *errors.MasterError {
	_, merr := db.Tool.Regenerations.Create(&shared.ToolRegeneration{
		ToolID: toolID,
		Start:  shared.NewUnixMilli(time.Now()),
	})
	if merr != nil {
		return merr
	}
	return nil
}

func StopToolRegeneration(db *common.DB, toolID shared.EntityID) *errors.MasterError {
	regenerationsMutex.Lock()
	defer regenerationsMutex.Unlock()

	regeneration := &shared.ToolRegeneration{}
	err := db.Tool.Regenerations.DB().QueryRow(sqlSelectOngoingRegenerationForTool, toolID).Scan(
		&regeneration.ID,
		&regeneration.ToolID,
		&regeneration.Start,
		&regeneration.Stop,
		&regeneration.Cycles,
	)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	regeneration.Stop = shared.NewUnixMilli(time.Now())
	totalCycles, merr := GetTotalCyclesForTool(db, toolID)
	if merr != nil {
		return merr
	}
	regeneration.Cycles = totalCycles
	merr = db.Tool.Regenerations.Update(regeneration)
	if merr != nil {
		return merr
	}
	return nil
}

func AbortToolRegeneration(db *common.DB, toolID shared.EntityID) *errors.MasterError {
	regenerationsMutex.Lock()
	defer regenerationsMutex.Unlock()

	regeneration := &shared.ToolRegeneration{}
	err := db.Tool.Regenerations.DB().QueryRow(sqlSelectOngoingRegenerationForTool, toolID).Scan(
		&regeneration.ID,
		&regeneration.ToolID,
		&regeneration.Start,
		&regeneration.Stop,
		&regeneration.Cycles,
	)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	if regeneration.Stop != 0 {
		return errors.NewValidationError("cannot abort a completed regeneration").MasterError()
	}

	merr := db.Tool.Regenerations.Delete(regeneration.ID)
	if merr != nil {
		return merr
	}
	return nil
}
