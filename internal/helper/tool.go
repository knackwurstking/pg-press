package helper

import (
	"log/slog"

	"github.com/knackwurstking/pg-press/internal/common"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"
)

const (
	SQLListDeadTools string = `
		SELECT id, width, height, type, code, cycles_offset, cycles, last_regeneration, regenerating, is_dead, cassette
		FROM tools
		WHERE is_dead = 1;
	`
)

func ListDeadTools(db *common.DB) ([]*shared.Tool, error) {
	deadTools := []*shared.Tool{}

	r, err := db.Tool.Tool.DB().Query(SQLListDeadTools)
	if err != nil {
		return deadTools, errors.NewMasterError(err, 0)
	}

	for r.Next() {
		tool := &shared.Tool{}
		err := r.Scan(
			&tool.ID,
			&tool.Width,
			&tool.Height,
			&tool.Type,
			&tool.Code,
			&tool.CyclesOffset,
			&tool.Cycles,
			&tool.LastRegeneration,
			&tool.Regenerating,
			&tool.IsDead,
			&tool.Cassette,
		)
		if err != nil {
			slog.Error("Failed to scan dead tool", "error", errors.NewMasterError(err, 0))
			continue
		}
		deadTools = append(deadTools, tool)
	}

	err = r.Err()
	if err != nil {
		slog.Error("Error occurred during listing dead tools", "error", errors.NewMasterError(err, 0))
	}

	_ = r.Close()

	return deadTools, nil
}
