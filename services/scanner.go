package services

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
)

func ScanRows[T any](rows *sql.Rows, scanFunc func(Scannable) (*T, error)) ([]*T, *errors.DBError) {
	var results []*T

	for rows.Next() {
		item, err := scanFunc(rows)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, errors.NewDBError(err, errors.DBTypeNotFound)
			}
			return nil, errors.NewDBError(err, errors.DBTypeScan)
		}
		results = append(results, item)
	}

	if err := rows.Err(); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewDBError(err, errors.DBTypeNotFound)
		}
		return nil, errors.NewDBError(err, errors.DBTypeScan)
	}

	return results, nil
}

func ScanRow[T any](row *sql.Row, scanFunc func(Scannable) (*T, error)) (*T, *errors.DBError) {
	t, err := scanFunc(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewDBError(err, errors.DBTypeNotFound)
		}
		return t, errors.NewDBError(err, errors.DBTypeScan)
	}
	return t, nil
}

// Scanner functions to pass into `ScanRows` or ScanRow`

func ScanAttachment(scanner Scannable) (*models.Attachment, error) {
	var id int64
	attachment := &models.Attachment{}

	err := scanner.Scan(&id, &attachment.MimeType, &attachment.Data)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("scan attachment: %v", err)
	}

	attachment.ID = fmt.Sprintf("%d", id)
	return attachment, nil
}

func ScanCookie(scanner Scannable) (*models.Cookie, error) {
	cookie := &models.Cookie{}

	err := scanner.Scan(&cookie.UserAgent, &cookie.Value, &cookie.ApiKey, &cookie.LastLogin)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("scan cookie: %v", err)
	}

	return cookie, nil
}

func ScanFeed(scanner Scannable) (*models.Feed, error) {
	feed := &models.Feed{}

	err := scanner.Scan(&feed.ID, &feed.Title, &feed.Content, &feed.UserID, &feed.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("scan feed: %v", err)
	}

	return feed, nil
}

func ScanMetalSheet(scanner Scannable) (*models.MetalSheet, error) {
	sheet := &models.MetalSheet{}
	var identifierStr string

	err := scanner.Scan(&sheet.ID, &sheet.TileHeight, &sheet.Value, &sheet.MarkeHeight,
		&sheet.STF, &sheet.STFMax, &identifierStr, &sheet.ToolID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("scan metal-sheet: %v", err)
	}

	sheet.Identifier = models.MachineType(identifierStr)
	return sheet, nil
}

func ScanModification(scanner Scannable) (*models.Modification[any], error) {
	mod := &models.Modification[any]{}
	err := scanner.Scan(&mod.ID, &mod.UserID, &mod.Data, &mod.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("scan modification: %v", err)
	}
	return mod, nil
}

func ScanNote(scanner Scannable) (*models.Note, error) {
	note := &models.Note{}
	err := scanner.Scan(&note.ID, &note.Level, &note.Content, &note.CreatedAt, &note.Linked)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("scan note: %v", err)
	}
	return note, nil
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
	if err != nil {
		return nil, fmt.Errorf("scan press-cycle: %w", err)
	}

	return cycle, nil
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
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("scan press regeneration: %v", err)
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
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("scan tool-regeneration: %v", err)
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
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("scan tool: %v", err)
	}

	if err := json.Unmarshal(format, &tool.Format); err != nil {
		return nil, fmt.Errorf("unmarshal tool format: %v", err)
	}

	return tool, nil
}

func ScanTroubleReport(scannable Scannable) (*models.TroubleReport, error) {
	report := &models.TroubleReport{}
	var linkedAttachments string

	err := scannable.Scan(&report.ID, &report.Title, &report.Content, &linkedAttachments, &report.UseMarkdown)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("scan trouble-report: %v", err)
	}

	if err := json.Unmarshal([]byte(linkedAttachments), &report.LinkedAttachments); err != nil {
		return nil, fmt.Errorf("unmarshal linked attachments: %v", err)
	}

	return report, nil
}
