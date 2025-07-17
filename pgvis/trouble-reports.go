// Package pgvis trouble reports management.
//
// This file provides database operations for managing trouble reports,
// including CRUD operations and integration with the activity feed system.
// Trouble reports are used to track issues, problems, and their resolutions
// within the system.
package pgvis

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

// SQL queries for trouble reports operations
const (
	// createTroubleReportsTableQuery creates the trouble_reports table if it doesn't exist
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

	// selectAllTroubleReportsQuery retrieves all trouble reports ordered by ID descending
	selectAllTroubleReportsQuery = `SELECT * FROM trouble_reports ORDER BY id DESC`

	// selectTroubleReportByIDQuery retrieves a specific trouble report by ID
	selectTroubleReportByIDQuery = `SELECT * FROM trouble_reports WHERE id = ?`

	// insertTroubleReportQuery creates a new trouble report
	insertTroubleReportQuery = `INSERT INTO trouble_reports (title, content, linked_attachments, modified) VALUES (?, ?, ?, ?)`

	// updateTroubleReportQuery updates an existing trouble report
	updateTroubleReportQuery = `UPDATE trouble_reports SET title = ?, content = ?, linked_attachments = ?, modified = ? WHERE id = ?`

	// deleteTroubleReportQuery removes a trouble report by ID
	deleteTroubleReportQuery = `DELETE FROM trouble_reports WHERE id = ?`
)

// TroubleReports provides database operations for managing trouble reports.
// It handles CRUD operations and maintains integration with the activity feed system
// to track trouble report lifecycle events (create, update, delete).
type TroubleReports struct {
	db    *sql.DB // Database connection
	feeds *Feeds  // Feed system for activity tracking
}

// NewTroubleReports creates a new TroubleReports instance and initializes the database table.
// It creates the trouble_reports table if it doesn't exist and sets up the necessary
// database schema for trouble report management.
//
// Parameters:
//   - db: Active database connection
//   - feeds: Feed system for activity tracking
//
// Returns:
//   - *TroubleReports: Initialized trouble reports handler
//
// Panics if the database table cannot be created.
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

// List retrieves all trouble reports ordered by ID in descending order (newest first).
//
// Returns:
//   - []*TroubleReport: Slice of all trouble reports
//   - error: Database or scanning error
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
//
// Parameters:
//   - id: The unique identifier of the trouble report
//
// Returns:
//   - *TroubleReport: The requested trouble report
//   - error: ErrNotFound if not found, or database error
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
//
// Parameters:
//   - report: The trouble report to create
//
// Returns:
//   - error: Validation, database, or feed creation error
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

	// Create feed entry for the new trouble report
	feed := NewTroubleReportAddFeed(report)
	if err := tr.feeds.Add(feed); err != nil {
		return WrapError(err, "failed to add feed entry")
	}

	return nil
}

// Update modifies an existing trouble report and generates an activity feed entry.
//
// Parameters:
//   - id: The ID of the trouble report to update
//   - report: The updated trouble report data
//
// Returns:
//   - error: Validation, database, or feed creation error
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

	// Create feed entry for the updated trouble report
	feed := NewTroubleReportUpdateFeed(report)
	if err := tr.feeds.Add(feed); err != nil {
		return WrapError(err, "failed to add feed entry")
	}

	return nil
}

// Remove deletes a trouble report by ID and generates an activity feed entry.
//
// Parameters:
//   - id: The ID of the trouble report to delete
//
// Returns:
//   - error: ErrNotFound if not found, database error, or feed creation error
func (tr *TroubleReports) Remove(id int64) error {
	// Get the trouble report before deleting it for the feed entry
	// Ignore any error, just check if it exists before creating a feed entry
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

	// Create feed entry for the removed trouble report
	if report != nil {
		feed := NewTroubleReportRemoveFeed(report)
		if err := tr.feeds.Add(feed); err != nil {
			return WrapError(err, "failed to add feed entry")
		}
	}

	return nil
}

// scanTroubleReport scans a trouble report from database rows.
// This is a helper method for handling multiple rows from SELECT queries.
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

// scanTroubleReportRow scans a trouble report from a single database row.
// This is a helper method for handling single row results from QueryRow.
func (tr *TroubleReports) scanTroubleReportRow(row *sql.Row) (*TroubleReport, error) {
	report := &TroubleReport{}
	var linkedAttachments, modified []byte

	if err := row.Scan(&report.ID, &report.Title, &report.Content, &linkedAttachments, &modified); err != nil {
		return nil, err // Return raw error for ErrNoRows detection
	}

	if err := json.Unmarshal(linkedAttachments, &report.LinkedAttachments); err != nil {
		return nil, WrapError(err, "failed to unmarshal linked attachments")
	}

	if err := json.Unmarshal(modified, &report.Modified); err != nil {
		return nil, WrapError(err, "failed to unmarshal modified data")
	}

	return report, nil
}
