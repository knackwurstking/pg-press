package helper

import (
	"github.com/knackwurstking/pg-press/internal/common"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"
)

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

func ListAvailableCassettesForBinding(db *common.DB, toolID shared.EntityID) ([]*shared.Cassette, *errors.MasterError) {
	tool, merr := db.Tool.Tool.GetByID(toolID)
	if merr != nil {
		return nil, merr
	}

	cassettes, merr := db.Tool.Cassette.List()
	if merr != nil {
		return nil, merr
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

// TODO: BindCassetteToTool
