package services

import (
	"database/sql"
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
