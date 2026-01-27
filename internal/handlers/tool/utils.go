package tool

import (
	"slices"

	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/labstack/echo/v4"
)

func listBindableCassettes(tool *shared.Tool) ([]*shared.Tool, *echo.HTTPError) {
	var bindableCassettes []*shared.Tool
	if !tool.IsCassette() {
		var herr *errors.HTTPError
		bindableCassettes, herr = db.ListBindableCassettes(tool.ID)
		if herr != nil {
			return bindableCassettes, herr.Echo()
		}
	}

	if tool.Cassette > 0 {
		cassette, herr := db.GetTool(tool.Cassette)
		if herr != nil {
			return bindableCassettes, herr.Echo()
		}
		bindableCassettes = append(bindableCassettes, cassette)
	}

	slices.SortFunc(bindableCassettes, func(a, b *shared.Tool) int {
		if a.German() < b.German() {
			return -1
		} else if a.German() > b.German() {
			return 1
		}
		return 0
	})

	return bindableCassettes, nil
}
