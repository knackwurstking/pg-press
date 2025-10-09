package services

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/knackwurstking/pgpress/internal/interfaces"
	"github.com/knackwurstking/pgpress/pkg/models"
)

// ScanNote scans a database row into a Note model
func ScanNote(scanner interfaces.Scannable) (*models.Note, error) {
	note := &models.Note{}
	err := scanner.Scan(&note.ID, &note.Level, &note.Content, &note.CreatedAt, &note.Linked)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("failed to scan note: %v", err)
	}
	return note, nil
}

// ScanNoteWithNullable scans a database row into a Note model with nullable linked field
func ScanNoteWithNullable(scanner interfaces.Scannable) (*models.Note, error) {
	note := &models.Note{}
	var nullableLinked sql.NullString

	err := scanner.Scan(&note.ID, &note.Level, &note.Content, &note.CreatedAt, &nullableLinked)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("failed to scan note: %v", err)
	}

	// Handle nullable linked field
	if nullableLinked.Valid {
		note.Linked = nullableLinked.String
	} else {
		note.Linked = ""
	}

	return note, nil
}

// ScanNotesFromRows scans multiple note rows
func ScanNotesFromRows(rows *sql.Rows) ([]*models.Note, error) {
	return ScanRows(rows, ScanNote)
}

// ScanNotesFromRowsWithNullable scans multiple note rows with nullable fields
func ScanNotesFromRowsWithNullable(rows *sql.Rows) ([]*models.Note, error) {
	return ScanRows(rows, ScanNoteWithNullable)
}

// ScanNotesIntoMap scans notes into a map by ID
func ScanNotesIntoMap(rows *sql.Rows) (map[int64]*models.Note, error) {
	return ScanIntoMap(rows, ScanNote, func(note *models.Note) int64 {
		return note.ID
	})
}

// ScanUsersIntoMap scans users into a map by Telegram ID
func ScanUsersIntoMap(rows *sql.Rows) (map[int64]*models.User, error) {
	return ScanIntoMap(rows, ScanUser, func(user *models.User) int64 {
		return user.TelegramID
	})
}

// ScanFeedsIntoMap scans feeds into a map by ID
func ScanFeedsIntoMap(rows *sql.Rows) (map[int64]*models.Feed, error) {
	return ScanIntoMap(rows, ScanFeed, func(feed *models.Feed) int64 {
		return feed.ID
	})
}

// ScanTool scans a database row into a Tool model
func ScanTool(scanner interfaces.Scannable) (*models.Tool, error) {
	tool := &models.Tool{}
	var format []byte

	err := scanner.Scan(&tool.ID, &tool.Position, &format, &tool.Type,
		&tool.Code, &tool.Regenerating, &tool.Press)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("failed to scan tool: %v", err)
	}

	// Unmarshal the format data
	if err := json.Unmarshal(format, &tool.Format); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tool format data: %v", err)
	}

	return tool, nil
}

// ScanToolsFromRows scans multiple tool rows
func ScanToolsFromRows(rows *sql.Rows) ([]*models.Tool, error) {
	return ScanRows(rows, ScanTool)
}

// ScanToolsIntoMap scans tools into a map by ID
func ScanToolsIntoMap(rows *sql.Rows) (map[int64]*models.Tool, error) {
	return ScanIntoMap(rows, ScanTool, func(tool *models.Tool) int64 {
		return tool.ID
	})
}

// ScanMetalSheet scans a database row into a MetalSheet model
func ScanMetalSheet(scanner interfaces.Scannable) (*models.MetalSheet, error) {
	sheet := &models.MetalSheet{}
	var identifierStr string
	var toolID int64

	err := scanner.Scan(&sheet.ID, &sheet.TileHeight, &sheet.Value, &sheet.MarkeHeight,
		&sheet.STF, &sheet.STFMax, &identifierStr, &toolID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("failed to scan metal sheet: %v", err)
	}

	// Convert string identifier to MachineType
	sheet.Identifier = models.MachineType(identifierStr)
	sheet.ToolID = toolID

	return sheet, nil
}

// ScanMetalSheetsFromRows scans multiple metal sheet rows
func ScanMetalSheetsFromRows(rows *sql.Rows) ([]*models.MetalSheet, error) {
	return ScanRows(rows, ScanMetalSheet)
}

// ScanMetalSheetsIntoMap scans metal sheets into a map by ID
func ScanMetalSheetsIntoMap(rows *sql.Rows) (map[int64]*models.MetalSheet, error) {
	return ScanIntoMap(rows, ScanMetalSheet, func(sheet *models.MetalSheet) int64 {
		return sheet.ID
	})
}

// ScanModification scans a database row into a Modification model
func ScanModification(scanner interfaces.Scannable) (*models.Modification[any], error) {
	mod := &models.Modification[any]{}
	err := scanner.Scan(&mod.ID, &mod.UserID, &mod.Data, &mod.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("failed to scan modification: %v", err)
	}
	return mod, nil
}

// ScanModificationsFromRows scans multiple modification rows
func ScanModificationsFromRows(rows *sql.Rows) ([]*models.Modification[any], error) {
	return ScanRows(rows, ScanModification)
}

// ScanModificationsIntoMap scans modifications into a map by ID
func ScanModificationsIntoMap(rows *sql.Rows) (map[int64]*models.Modification[any], error) {
	return ScanIntoMap(rows, ScanModification, func(mod *models.Modification[any]) int64 {
		return mod.ID
	})
}

// ScanPressCycle scans a database row into a Cycle model
func ScanPressCycle(scanner interfaces.Scannable) (*models.Cycle, error) {
	cycle := &models.Cycle{}
	var performedBy sql.NullInt64

	err := scanner.Scan(
		&cycle.ID,
		&cycle.PressNumber,
		&cycle.ToolID,
		&cycle.ToolPosition,
		&cycle.TotalCycles,
		&cycle.Date,
		&performedBy,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("failed to scan press cycle: %v", err)
	}

	if performedBy.Valid {
		cycle.PerformedBy = performedBy.Int64
	}

	return cycle, nil
}

// ScanPressCyclesFromRows scans multiple press cycle rows
func ScanPressCyclesFromRows(rows *sql.Rows) ([]*models.Cycle, error) {
	return ScanRows(rows, ScanPressCycle)
}

// ScanPressCyclesIntoMap scans press cycles into a map by ID
func ScanPressCyclesIntoMap(rows *sql.Rows) (map[int64]*models.Cycle, error) {
	return ScanIntoMap(rows, ScanPressCycle, func(cycle *models.Cycle) int64 {
		return cycle.ID
	})
}

// ScanToolRegeneration scans a database row into a Regeneration model
func ScanToolRegeneration(scanner interfaces.Scannable) (*models.Regeneration, error) {
	regen := &models.Regeneration{}
	var performedBy sql.NullInt64

	err := scanner.Scan(
		&regen.ID,
		&regen.ToolID,
		&regen.CycleID,
		&regen.Reason,
		&performedBy,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("failed to scan tool regeneration: %v", err)
	}

	if performedBy.Valid {
		regen.PerformedBy = &performedBy.Int64
	}

	return regen, nil
}

// ScanToolRegenerationsFromRows scans multiple tool regeneration rows
func ScanToolRegenerationsFromRows(rows *sql.Rows) ([]*models.Regeneration, error) {
	return ScanRows(rows, ScanToolRegeneration)
}

// ScanToolRegenerationsIntoMap scans tool regenerations into a map by ID
func ScanToolRegenerationsIntoMap(rows *sql.Rows) (map[int64]*models.Regeneration, error) {
	return ScanIntoMap(rows, ScanToolRegeneration, func(regen *models.Regeneration) int64 {
		return regen.ID
	})
}

// ScanTroubleReport scans a database row into a TroubleReport model
func ScanTroubleReport(scanner interfaces.Scannable) (*models.TroubleReport, error) {
	report := &models.TroubleReport{}
	var linkedAttachments string

	err := scanner.Scan(&report.ID, &report.Title, &report.Content, &linkedAttachments, &report.UseMarkdown)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("failed to scan trouble report: %v", err)
	}

	// Unmarshal the linked attachments JSON
	if err := json.Unmarshal([]byte(linkedAttachments), &report.LinkedAttachments); err != nil {
		return nil, fmt.Errorf("failed to unmarshal linked attachments: %v", err)
	}

	return report, nil
}

// ScanTroubleReportsFromRows scans multiple trouble report rows
func ScanTroubleReportsFromRows(rows *sql.Rows) ([]*models.TroubleReport, error) {
	return ScanRows(rows, ScanTroubleReport)
}

// ScanTroubleReportsIntoMap scans trouble reports into a map by ID
func ScanTroubleReportsIntoMap(rows *sql.Rows) (map[int64]*models.TroubleReport, error) {
	return ScanIntoMap(rows, ScanTroubleReport, func(report *models.TroubleReport) int64 {
		return report.ID
	})
}
