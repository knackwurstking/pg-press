package services

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
)

func (t *Tools) Bind(cassetteID, targetID models.ToolID) *errors.MasterError {
	if !t.validateBindingTools(cassetteID, targetID) {
		return errors.NewMasterError(
			fmt.Errorf("invalid tools for binding: cassette=%d, target=%d",
				cassetteID, targetID),
			http.StatusBadRequest,
		)
	}

	// Get press from the target tool
	targetTool, dberr := t.Get(targetID)
	if dberr != nil {
		return dberr
	}

	// Execute binding operations
	queries := []string{
		// Set bindings
		`UPDATE tools SET binding = :target WHERE id = :cassette`,
		`UPDATE tools SET binding = :cassette WHERE id = :target`,
		// Clear press from other cassettes at the same press
		`UPDATE tools SET press = NULL WHERE press = :press AND position = "cassette top"`,
		// Assign press to cassette
		`UPDATE tools SET press = :press WHERE id = :cassette`,
	}

	for _, query := range queries {
		if _, err := t.DB.Exec(query,
			sql.Named("target", targetID),
			sql.Named("cassette", cassetteID),
			sql.Named("press", targetTool.Press),
		); err != nil {
			return errors.NewMasterError(err, 0)
		}
	}

	return nil
}

func (t *Tools) UnBind(toolID models.ToolID) *errors.MasterError {
	tool, dberr := t.Get(toolID)
	if dberr != nil {
		return dberr
	}

	if tool.Binding == nil {
		return nil
	}

	query := `
		UPDATE tools
		SET binding = NULL
		WHERE id = :toolID OR id = :binding
	`

	if _, err := t.DB.Exec(query,
		sql.Named("toolID", toolID),
		sql.Named("binding", *tool.Binding),
	); err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}
