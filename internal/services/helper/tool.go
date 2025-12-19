package helper

import (
	"net/http"
	"sync"
	"time"

	"github.com/knackwurstking/pg-press/internal/common"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"
)

// -----------------------------------------------------------------------------
// Tool Helpers
// ------------------------------------------------------------------------------

// GetToolByID retrieves a tool by its ID from the "tools" table and if it fails from the cassettes table.
func GetToolByID(db *common.DB, id shared.EntityID) (shared.ModelTool, *errors.MasterError) {
	var tool shared.ModelTool
	tool, merr := db.Tool.Tool.GetByID(id)
	if merr != nil {
		if merr.Code == http.StatusNotFound {
			tool, merr = db.Tool.Cassette.GetByID(id)
			if merr != nil {
				return nil, merr
			}
		} else {
			return nil, merr
		}
	}
	return tool, nil
}

func ListDeadTools(db *common.DB) (tools []*shared.Tool, merr *errors.MasterError) {
	tools, merr = db.Tool.Tool.List()
	if merr != nil {
		return nil, merr
	}

	n := 0
	for _, t := range tools {
		if !t.IsDead {
			continue
		}

		tools[n] = t
		n++
	}
	return tools[:n], nil
}

func ListAvailableCassettesForBinding(db *common.DB, toolID shared.EntityID) ([]*shared.Cassette, *errors.MasterError) {
	tool, merr := db.Tool.Tool.GetByID(toolID)
	if merr != nil {
		return nil, merr.Wrap("could not get tool with ID %d", toolID)
	}

	cassettes, merr := db.Tool.Cassette.List()
	if merr != nil {
		return nil, merr.Wrap("could not list cassettes")
	}

	// Filter cassettes based on the tool width and height
	i := 0
	for _, c := range cassettes {
		if c.IsDead || c.Width != tool.Width || c.Height != tool.Height {
			continue
		}

		cassettes[i] = c
		i++
	}

	return cassettes[:i], nil
}

// -----------------------------------------------------------------------------
// Tool Cassette Binding Helpers
// ------------------------------------------------------------------------------

var bindMutex = &sync.Mutex{}

func BindCassetteToTool(db *common.DB, toolID, cassetteID shared.EntityID) *errors.MasterError {
	// First, check if cassette exists
	_, merr := db.Tool.Cassette.GetByID(cassetteID)
	if merr != nil {
		return merr
	}

	bindMutex.Lock()
	defer bindMutex.Unlock()

	tool, merr := db.Tool.Tool.GetByID(toolID)
	if merr != nil {
		return merr
	}
	if tool.Cassette > 0 {
		return errors.NewValidationError("tool already has a cassette bound").MasterError()
	}

	tool.Cassette = cassetteID
	merr = db.Tool.Tool.Update(tool)
	if merr != nil {
		return merr
	}

	return nil
}

func UnbindCassetteFromTool(db *common.DB, toolID shared.EntityID) *errors.MasterError {
	bindMutex.Lock()
	defer bindMutex.Unlock()

	tool, merr := db.Tool.Tool.GetByID(toolID)
	if merr != nil {
		return merr
	}
	if tool.Cassette == 0 {
		return errors.NewValidationError("tool has no cassette bound").MasterError()
	}

	tool.Cassette = 0
	merr = db.Tool.Tool.Update(tool)
	if merr != nil {
		return merr
	}

	return nil
}

// -----------------------------------------------------------------------------
// Tool Metal Sheet Helpers
// ------------------------------------------------------------------------------

func ListUpperMetalSheetsForTool(db *common.DB, toolID shared.EntityID) ([]*shared.UpperMetalSheet, *errors.MasterError) {
	metalSheets, merr := db.Tool.UpperMetalSheet.List()
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
	metalSheets, merr := db.Tool.LowerMetalSheet.List()
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

var regenerationsMutex = &sync.Mutex{}

func GetRegenerationsForTool(db *common.DB, toolID shared.EntityID) (
	regenerations []*shared.ToolRegeneration, merr *errors.MasterError,
) {
	regenerations, merr = db.Tool.Regeneration.List()
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
	merr := db.Tool.Regeneration.Create(&shared.ToolRegeneration{
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

	regeneration, merr := db.Tool.Regeneration.GetByID(toolID)
	if merr != nil {
		return merr
	}

	regeneration.Stop = shared.NewUnixMilli(time.Now())
	totalCycles, merr := GetTotalCyclesForTool(db, toolID)
	if merr != nil {
		return merr
	}
	regeneration.Cycles = totalCycles
	merr = db.Tool.Regeneration.Update(regeneration)
	if merr != nil {
		return merr
	}
	return nil
}

func AbortToolRegeneration(db *common.DB, toolID shared.EntityID) *errors.MasterError {
	regenerationsMutex.Lock()
	defer regenerationsMutex.Unlock()

	regeneration, merr := db.Tool.Regeneration.GetByID(toolID)
	if merr != nil {
		return merr
	}
	if regeneration.Stop != 0 {
		return errors.NewValidationError("cannot abort a completed regeneration").MasterError()
	}

	merr = db.Tool.Regeneration.Delete(regeneration.ID)
	if merr != nil {
		return merr
	}
	return nil
}
