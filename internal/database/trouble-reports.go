// Package database provides trouble reports management.
package database

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/knackwurstking/pgpress/internal/logger"
)

const (
	createTroubleReportsTableQuery = `
		CREATE TABLE IF NOT EXISTS trouble_reports (
			id INTEGER NOT NULL,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			linked_attachments TEXT NOT NULL,
			mods BLOB NOT NULL,
			PRIMARY KEY("id" AUTOINCREMENT)
		);
	`

	selectAllTroubleReportsQuery = `SELECT * FROM trouble_reports ORDER BY id DESC`
	selectTroubleReportByIDQuery = `SELECT * FROM trouble_reports WHERE id = ?`
	insertTroubleReportQuery     = `INSERT INTO trouble_reports
		(title, content, linked_attachments, mods) VALUES (?, ?, ?, ?)`
	updateTroubleReportQuery = `UPDATE trouble_reports
		SET title = ?, content = ?, linked_attachments = ?, mods = ? WHERE id = ?`
	deleteTroubleReportQuery = `DELETE FROM trouble_reports WHERE id = ?`
)

// TroubleReports provides database operations for managing trouble reports.
type TroubleReports struct {
	db    *sql.DB
	feeds *Feeds
}

// NewTroubleReports creates a new TroubleReports instance and initializes the database table.
func NewTroubleReports(db *sql.DB, feeds *Feeds) *TroubleReports {
	if _, err := db.Exec(createTroubleReportsTableQuery); err != nil {
		panic(NewDatabaseError(
			"create_table",
			"trouble_reports",
			"failed to create trouble_reports table",
			err,
		))
	}

	return &TroubleReports{
		db:    db,
		feeds: feeds,
	}
}

// List retrieves all trouble reports ordered by ID descending.
func (tr *TroubleReports) List() ([]*TroubleReport, error) {
	logger.TroubleReport().Info("Listing trouble reports")

	rows, err := tr.db.Query(selectAllTroubleReportsQuery)
	if err != nil {
		return nil, NewDatabaseError("select", "trouble_reports",
			"failed to query trouble reports", err)
	}
	defer rows.Close()

	var reports []*TroubleReport

	for rows.Next() {
		report, err := tr.scanTroubleReport(rows)
		if err != nil {
			return nil, WrapError(err, "failed to scan trouble report")
		}
		reports = append(reports, report)
	}

	if err := rows.Err(); err != nil {
		return nil, NewDatabaseError("select", "trouble_reports",
			"error iterating over rows", err)
	}

	return reports, nil
}

// Get retrieves a specific trouble report by ID.
func (tr *TroubleReports) Get(id int64) (*TroubleReport, error) {
	logger.TroubleReport().Debug("Getting trouble report, id: %d", id)

	row := tr.db.QueryRow(selectTroubleReportByIDQuery, id)

	report, err := tr.scanTroubleReportRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, NewDatabaseError("select", "trouble_reports",
			fmt.Sprintf("failed to get trouble report with ID %d", id), err)
	}

	return report, nil
}

// Add creates a new trouble report and generates a corresponding activity feed entry.
func (tr *TroubleReports) Add(troubleReport *TroubleReport) error {
	logger.TroubleReport().Info("Adding trouble report: %+v", troubleReport)

	if troubleReport == nil {
		return NewValidationError("report", "trouble report cannot be nil", nil)
	}

	if err := troubleReport.Validate(); err != nil {
		return err
	}

	linkedAttachments, err := json.Marshal(troubleReport.LinkedAttachments)
	if err != nil {
		return WrapError(err, "failed to marshal linked attachments")
	}

	mods, err := json.Marshal(troubleReport.Mods)
	if err != nil {
		return WrapError(err, "failed to marshal mods data")
	}

	// After exec this insert query, i need to get the id
	result, err := tr.db.Exec(
		insertTroubleReportQuery,
		troubleReport.Title, troubleReport.Content, string(linkedAttachments), mods,
	)
	if err != nil {
		return NewDatabaseError("insert", "trouble_reports",
			"failed to insert trouble report", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return NewDatabaseError("insert", "trouble_reports",
			"failed to get last insert ID", err)
	}

	feed := NewFeed(
		FeedTypeTroubleReportAdd,
		&FeedTroubleReportAdd{
			ID:         id,
			Title:      troubleReport.Title,
			ModifiedBy: troubleReport.Mods.Current().User,
		},
	)
	if err := tr.feeds.Add(feed); err != nil {
		return WrapError(err, "failed to add feed entry")
	}

	return nil
}

// Update modifies an existing trouble report and generates an activity feed entry.
func (tr *TroubleReports) Update(id int64, troubleReport *TroubleReport) error {
	logger.TroubleReport().Info("Updating trouble report, id: %d, data: %+v", id, troubleReport)

	if troubleReport == nil {
		return NewValidationError("report", "trouble report cannot be nil", nil)
	}

	if err := troubleReport.Validate(); err != nil {
		return err
	}

	linkedAttachments, err := json.Marshal(troubleReport.LinkedAttachments)
	if err != nil {
		return WrapError(err, "failed to marshal linked attachments")
	}

	mods, err := json.Marshal(troubleReport.Mods)
	if err != nil {
		return WrapError(err, "failed to marshal mods data")
	}

	_, err = tr.db.Exec(
		updateTroubleReportQuery,
		troubleReport.Title, troubleReport.Content, string(linkedAttachments), mods, id,
	)
	if err != nil {
		return NewDatabaseError("update", "trouble_reports",
			fmt.Sprintf("failed to update trouble report with ID %d", id), err)
	}

	feed := NewFeed(
		FeedTypeTroubleReportUpdate,
		&FeedTroubleReportUpdate{
			ID:         id,
			Title:      troubleReport.Title,
			ModifiedBy: troubleReport.Mods.Current().User,
		},
	)
	if err := tr.feeds.Add(feed); err != nil {
		return WrapError(err, "failed to add feed entry")
	}

	return nil
}

// Remove deletes a trouble report by ID and generates an activity feed entry.
func (tr *TroubleReports) Remove(id int64) error {
	logger.TroubleReport().Info("Removing trouble report, id: %d", id)

	result, err := tr.db.Exec(deleteTroubleReportQuery, id)
	if err != nil {
		return NewDatabaseError("delete", "trouble_reports",
			fmt.Sprintf("failed to delete trouble report with ID %d", id), err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return NewDatabaseError("delete", "trouble_reports",
			"failed to get rows affected", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (tr *TroubleReports) scanTroubleReport(rows *sql.Rows) (*TroubleReport, error) {
	report := &TroubleReport{}
	var linkedAttachments string
	var mods []byte

	if err := rows.Scan(&report.ID, &report.Title, &report.Content,
		&linkedAttachments, &mods); err != nil {
		return nil, NewDatabaseError("scan", "trouble_reports",
			"failed to scan row", err)
	}

	// Try to unmarshal as new format (array of int64 IDs) first
	if err := json.Unmarshal([]byte(linkedAttachments), &report.LinkedAttachments); err != nil {
		// If that fails, try to handle as old format or empty/invalid data
		if linkedAttachments == "" || linkedAttachments == "[]" {
			report.LinkedAttachments = make([]int64, 0)
		} else {
			// Try to unmarshal as old format (full attachment objects) and extract IDs
			var oldAttachments []*Attachment
			if err2 := json.Unmarshal([]byte(linkedAttachments), &oldAttachments); err2 == nil {
				// Convert old format to new format (this shouldn't happen after migration)
				logger.TroubleReport().Warn(
					"Found old format attachments for report %d, converting to IDs", report.ID)
				report.LinkedAttachments = make([]int64, 0)
				for _, att := range oldAttachments {
					if att != nil && att.GetID() > 0 {
						report.LinkedAttachments = append(report.LinkedAttachments, att.GetID())
					}
				}
			} else {
				// Neither format worked, log and use empty array
				logger.TroubleReport().Error(
					"Failed to unmarshal linked attachments for report %d: %v, data: %s",
					report.ID, err, linkedAttachments)
				report.LinkedAttachments = make([]int64, 0)
			}
		}
	}

	if err := json.Unmarshal(mods, &report.Mods); err != nil {
		return nil, WrapError(err, "failed to unmarshal mods data")
	}

	return report, nil
}

func (tr *TroubleReports) scanTroubleReportRow(row *sql.Row) (*TroubleReport, error) {
	report := &TroubleReport{}
	var linkedAttachments string
	var mods []byte

	if err := row.Scan(&report.ID, &report.Title, &report.Content,
		&linkedAttachments, &mods); err != nil {
		return nil, err
	}

	// Try to unmarshal as new format (array of int64 IDs) first
	if err := json.Unmarshal([]byte(linkedAttachments), &report.LinkedAttachments); err != nil {
		// If that fails, try to handle as old format or empty/invalid data
		if linkedAttachments == "" || linkedAttachments == "[]" {
			report.LinkedAttachments = make([]int64, 0)
		} else {
			// Try to unmarshal as old format (full attachment objects) and extract IDs
			var oldAttachments []*Attachment
			if err2 := json.Unmarshal([]byte(linkedAttachments), &oldAttachments); err2 == nil {
				// Convert old format to new format (this shouldn't happen after migration)
				logger.TroubleReport().Warn(
					"Found old format attachments for report %d, converting to IDs", report.ID)
				report.LinkedAttachments = make([]int64, 0)
				for _, att := range oldAttachments {
					if att != nil && att.GetID() > 0 {
						report.LinkedAttachments = append(report.LinkedAttachments, att.GetID())
					}
				}
			} else {
				// Neither format worked, log and use empty array
				logger.TroubleReport().Error(
					"Failed to unmarshal linked attachments for report %d: %v, data: %s",
					report.ID, err, linkedAttachments)
				report.LinkedAttachments = make([]int64, 0)
			}
		}
	}

	if err := json.Unmarshal(mods, &report.Mods); err != nil {
		return nil, WrapError(err, "failed to unmarshal mods data")
	}

	return report, nil
}
