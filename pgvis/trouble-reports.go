// Package pgvis provides trouble reports management.
package pgvis

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

const (
	createTroubleReportsTableQuery = `
		CREATE TABLE IF NOT EXISTS trouble_reports (
			id INTEGER NOT NULL,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			linked_attachments BLOB NOT NULL,
			modified BLOB NOT NULL,
			PRIMARY KEY("id" AUTOINCREMENT)
		);
	`

	selectAllTroubleReportsQuery = `SELECT * FROM trouble_reports ORDER BY id DESC`
	selectTroubleReportByIDQuery = `SELECT * FROM trouble_reports WHERE id = ?`
	insertTroubleReportQuery     = `INSERT INTO trouble_reports (title, content, linked_attachments, modified) VALUES (?, ?, ?, ?)`
	updateTroubleReportQuery     = `UPDATE trouble_reports SET title = ?, content = ?, linked_attachments = ?, modified = ? WHERE id = ?`
	deleteTroubleReportQuery     = `DELETE FROM trouble_reports WHERE id = ?`
)

// TroubleReports provides database operations for managing trouble reports.
type TroubleReports struct {
	db    *sql.DB
	feeds *Feeds
}

// NewTroubleReports creates a new TroubleReports instance and initializes the database table.
func NewTroubleReports(db *sql.DB, feeds *Feeds) *TroubleReports {
	if _, err := db.Exec(createTroubleReportsTableQuery); err != nil {
		panic(NewDatabaseError("create_table", "trouble_reports",
			"failed to create trouble_reports table", err))
	}

	return &TroubleReports{
		db:    db,
		feeds: feeds,
	}
}

// List retrieves all trouble reports ordered by ID descending.
func (tr *TroubleReports) List() ([]*TroubleReport, error) {
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
func (tr *TroubleReports) Add(report *TroubleReport) error {
	if report == nil {
		return NewValidationError("report", "trouble report cannot be nil", nil)
	}

	linkedAttachments, err := json.Marshal(report.LinkedAttachments)
	if err != nil {
		return WrapError(err, "failed to marshal linked attachments")
	}

	modified, err := json.Marshal(report.Modified)
	if err != nil {
		return WrapError(err, "failed to marshal modified data")
	}

	_, err = tr.db.Exec(
		insertTroubleReportQuery,
		report.Title, report.Content, linkedAttachments, modified,
	)
	if err != nil {
		return NewDatabaseError("insert", "trouble_reports",
			"failed to insert trouble report", err)
	}

	feed := NewFeed(
		FeedTypeTroubleReportAdd,
		&FeedTroubleReportAdd{
			ID:         report.ID,
			Title:      report.Title,
			ModifiedBy: report.Modified.User,
		},
	)
	if err := tr.feeds.Add(feed); err != nil {
		return WrapError(err, "failed to add feed entry")
	}

	return nil
}

// Update modifies an existing trouble report and generates an activity feed entry.
func (tr *TroubleReports) Update(id int64, report *TroubleReport) error {
	if report == nil {
		return NewValidationError("report", "trouble report cannot be nil", nil)
	}

	linkedAttachments, err := json.Marshal(report.LinkedAttachments)
	if err != nil {
		return WrapError(err, "failed to marshal linked attachments")
	}

	modified, err := json.Marshal(report.Modified)
	if err != nil {
		return WrapError(err, "failed to marshal modified data")
	}

	_, err = tr.db.Exec(
		updateTroubleReportQuery,
		report.Title, report.Content, linkedAttachments, modified, id,
	)
	if err != nil {
		return NewDatabaseError("update", "trouble_reports",
			fmt.Sprintf("failed to update trouble report with ID %d", id), err)
	}

	feed := NewFeed(
		FeedTypeTroubleReportUpdate,
		&FeedTroubleReportUpdate{
			ID:         report.ID,
			Title:      report.Title,
			ModifiedBy: report.Modified.User,
		},
	)
	if err := tr.feeds.Add(feed); err != nil {
		return WrapError(err, "failed to add feed entry")
	}

	return nil
}

// Remove deletes a trouble report by ID and generates an activity feed entry.
func (tr *TroubleReports) Remove(id int64) error {
	report, _ := tr.Get(id)

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

	if report != nil {
		feed := NewFeed(
			FeedTypeTroubleReportRemove,
			&FeedTroubleReportRemove{
				ID:         report.ID,
				Title:      report.Title,
				ModifiedBy: report.Modified.User,
			},
		)
		if err := tr.feeds.Add(feed); err != nil {
			return WrapError(err, "failed to add feed entry")
		}
	}

	return nil
}

func (tr *TroubleReports) scanTroubleReport(rows *sql.Rows) (*TroubleReport, error) {
	report := &TroubleReport{}
	var linkedAttachments, modified []byte

	if err := rows.Scan(&report.ID, &report.Title, &report.Content, &linkedAttachments, &modified); err != nil {
		return nil, NewDatabaseError("scan", "trouble_reports",
			"failed to scan row", err)
	}

	if err := json.Unmarshal(linkedAttachments, &report.LinkedAttachments); err != nil {
		return nil, WrapError(err, "failed to unmarshal linked attachments")
	}

	if err := json.Unmarshal(modified, &report.Modified); err != nil {
		return nil, WrapError(err, "failed to unmarshal modified data")
	}

	return report, nil
}

func (tr *TroubleReports) scanTroubleReportRow(row *sql.Row) (*TroubleReport, error) {
	report := &TroubleReport{}
	var linkedAttachments, modified []byte

	if err := row.Scan(&report.ID, &report.Title, &report.Content, &linkedAttachments, &modified); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(linkedAttachments, &report.LinkedAttachments); err != nil {
		return nil, WrapError(err, "failed to unmarshal linked attachments")
	}

	if err := json.Unmarshal(modified, &report.Modified); err != nil {
		return nil, WrapError(err, "failed to unmarshal modified data")
	}

	return report, nil
}
