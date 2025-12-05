package services

import (
	"database/sql"
	"fmt"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
)

func (t *Tools) Bind(cassetteID, targetID models.ToolID) *errors.MasterError {
	if !t.validateBindingTools(cassetteID, targetID) {
		return errors.NewMasterError(errors.ErrValidation)
	}

	// Get press from the target tool
	targetTool, dberr := t.Get(targetID)
	if dberr != nil {
		return dberr
	}

	// Execute binding operations
	queries := []string{
		// Set bindings
		fmt.Sprintf(`UPDATE %s SET binding = :target WHERE id = :cassette`, TableNameTools),
		fmt.Sprintf(`UPDATE %s SET binding = :cassette WHERE id = :target`, TableNameTools),
		// Clear press from other cassettes at the same press
		fmt.Sprintf(`UPDATE %s SET press = NULL WHERE press = :press AND position = "cassette top"`, TableNameTools),
		// Assign press to cassette
		fmt.Sprintf(`UPDATE %s SET press = :press WHERE id = :cassette`, TableNameTools),
	}

	for _, query := range queries {
		if _, err := t.DB.Exec(query,
			sql.Named("target", targetID),
			sql.Named("cassette", cassetteID),
			sql.Named("press", targetTool.Press),
		); err != nil {
			return errors.NewMasterError(err)
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

	query := fmt.Sprintf(`
		UPDATE %s
		SET binding = NULL
		WHERE id = :toolID OR id = :binding`,
		TableNameTools)

	if _, err := t.DB.Exec(query,
		sql.Named("toolID", toolID),
		sql.Named("binding", *tool.Binding),
	); err != nil {
		return errors.NewMasterError(err)
	}

	return nil
}
