package tool

import (
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

	return bindableCassettes, nil
}
