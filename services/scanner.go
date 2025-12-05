package services

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
)

func ScanRows[T any](rows *sql.Rows, scanFunc func(Scannable) (*T, error)) ([]*T, *errors.MasterError) {
	var results []*T

	for rows.Next() {
		item, err := scanFunc(rows)
		if err != nil {
			return nil, errors.NewMasterErrorDB(errors.ErrorTypeDB, err)
		}
		results = append(results, item)
	}

	err := rows.Err()
	if err != nil {
		return results, errors.NewMasterErrorDB(errors.ErrorTypeDB, err)
	}

	return results, nil
}

// Scanner functions to pass into `ScanRows` or ScanRow`

func ScanAttachment(scanner Scannable) (*models.Attachment, error) {
	var id int64
	attachment := &models.Attachment{}

	err := scanner.Scan(&id, &attachment.MimeType, &attachment.Data)
	if err != nil {
		return nil, err
	}

	attachment.ID = fmt.Sprintf("%d", id)
	return attachment, nil
}

func ScanCookie(scanner Scannable) (*models.Cookie, error) {
	cookie := &models.Cookie{}

	err := scanner.Scan(&cookie.UserAgent, &cookie.Value, &cookie.ApiKey, &cookie.LastLogin)
	return cookie, err
}

func ScanFeed(scanner Scannable) (*models.Feed, error) {
	feed := &models.Feed{}

	err := scanner.Scan(&feed.ID, &feed.Title, &feed.Content, &feed.UserID, &feed.CreatedAt)
	return feed, err
}

func ScanMetalSheet(scanner Scannable) (*models.MetalSheet, error) {
	sheet := &models.MetalSheet{}
	var identifierStr string

	err := scanner.Scan(&sheet.ID, &sheet.TileHeight, &sheet.Value, &sheet.MarkeHeight,
		&sheet.STF, &sheet.STFMax, &identifierStr, &sheet.ToolID)
	if err != nil {
		return nil, err
	}

	sheet.Identifier = models.MachineType(identifierStr)
	return sheet, nil
}

func ScanModification(scanner Scannable) (*models.Modification[any], error) {
	mod := &models.Modification[any]{}
	err := scanner.Scan(&mod.ID, &mod.UserID, &mod.Data, &mod.CreatedAt)
	return mod, err
}

func ScanNote(scanner Scannable) (*models.Note, error) {
	note := &models.Note{}
	err := scanner.Scan(&note.ID, &note.Level, &note.Content, &note.CreatedAt, &note.Linked)
	return note, err
}

func ScanCycle(scannable Scannable) (*models.Cycle, error) {
	cycle := &models.Cycle{}

	err := scannable.Scan(
		&cycle.ID,
		&cycle.PressNumber,
		&cycle.ToolID,
		&cycle.ToolPosition,
		&cycle.TotalCycles,
		&cycle.Date,
		&cycle.PerformedBy,
	)
	return cycle, err
}

func ScanPressRegeneration(scannable Scannable) (*models.PressRegeneration, error) {
	regeneration := &models.PressRegeneration{}
	var completedAt sql.NullTime

	err := scannable.Scan(
		&regeneration.ID,
		&regeneration.PressNumber,
		&regeneration.StartedAt,
		&completedAt,
		&regeneration.Reason,
	)
	if err != nil {
		return nil, err
	}

	if completedAt.Valid {
		regeneration.CompletedAt = completedAt.Time
	}

	return regeneration, nil
}

func ScanToolRegeneration(scannable Scannable) (*models.ToolRegeneration, error) {
	regeneration := &models.ToolRegeneration{}
	var performedBy sql.NullInt64

	err := scannable.Scan(
		&regeneration.ID,
		&regeneration.ToolID,
		&regeneration.CycleID,
		&regeneration.Reason,
		&performedBy,
	)
	if err != nil {
		return nil, err
	}

	if performedBy.Valid {
		performedBy := models.TelegramID(performedBy.Int64)
		regeneration.PerformedBy = &performedBy
	}

	return regeneration, nil
}

func ScanTool(scannable Scannable) (*models.Tool, error) {
	tool := &models.Tool{}
	var format []byte

	err := scannable.Scan(&tool.ID, &tool.Position, &format, &tool.Type,
		&tool.Code, &tool.Regenerating, &tool.IsDead, &tool.Press, &tool.Binding)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(format, &tool.Format)
	if err != nil {
		return nil, err
	}

	return tool, nil
}

func ScanTroubleReport(scannable Scannable) (*models.TroubleReport, error) {
	report := &models.TroubleReport{}
	var linkedAttachments string

	err := scannable.Scan(&report.ID, &report.Title, &report.Content, &linkedAttachments, &report.UseMarkdown)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(linkedAttachments), &report.LinkedAttachments)
	if err != nil {
		return nil, err
	}

	return report, nil
}

func ScanUser(scanner Scannable) (*models.User, error) {
	user := &models.User{}
	err := scanner.Scan(&user.TelegramID, &user.Name, &user.ApiKey, &user.LastFeed)
	return user, err
}
