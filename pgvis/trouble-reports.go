// ai: Organize
package pgvis

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

const (
	// SQL queries for trouble reports table
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

// TroubleReports handles database operations for trouble reports
type TroubleReports struct {
	db    *sql.DB
	feeds *Feeds
}

// NewTroubleReports creates a new TroubleReports instance and initializes the database table
func NewTroubleReports(db *sql.DB, feeds *Feeds) *TroubleReports {
	if _, err := db.Exec(createTroubleReportsTableQuery); err != nil {
		panic(fmt.Errorf("failed to create trouble_reports table: %w", err))
	}

	return &TroubleReports{
		db:    db,
		feeds: feeds,
	}
}

// List retrieves all trouble reports ordered by ID in descending order
func (tr *TroubleReports) List() ([]*TroubleReport, error) {
	rows, err := tr.db.Query(selectAllTroubleReportsQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query trouble reports: %w", err)
	}
	defer rows.Close()

	var reports []*TroubleReport

	for rows.Next() {
		report, err := tr.scanTroubleReport(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan trouble report: %w", err)
		}
		reports = append(reports, report)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return reports, nil
}

// Get retrieves a specific trouble report by ID
func (tr *TroubleReports) Get(id int64) (*TroubleReport, error) {
	row := tr.db.QueryRow(selectTroubleReportByIDQuery, id)

	report, err := tr.scanTroubleReportRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get trouble report with ID %d: %w", id, err)
	}

	return report, nil
}

// Add creates a new trouble report and adds a corresponding feed entry
func (tr *TroubleReports) Add(report *TroubleReport) error {
	linkedAttachments, err := json.Marshal(report.LinkedAttachments)
	if err != nil {
		return fmt.Errorf("failed to marshal linked attachments: %w", err)
	}

	modified, err := json.Marshal(report.Modified)
	if err != nil {
		return fmt.Errorf("failed to marshal modified data: %w", err)
	}

	_, err = tr.db.Exec(
		insertTroubleReportQuery,
		report.Title, report.Content, linkedAttachments, modified,
	)
	if err != nil {
		return fmt.Errorf("failed to insert trouble report: %w", err)
	}

	// Create feed entry for the new trouble report
	feed := NewTroubleReportAddFeed(report)
	if err := tr.feeds.Add(feed); err != nil {
		return fmt.Errorf("failed to add feed entry: %w", err)
	}

	return nil
}

// Update modifies an existing trouble report
func (tr *TroubleReports) Update(id int64, report *TroubleReport) error {
	linkedAttachments, err := json.Marshal(report.LinkedAttachments)
	if err != nil {
		return fmt.Errorf("failed to marshal linked attachments: %w", err)
	}

	modified, err := json.Marshal(report.Modified)
	if err != nil {
		return fmt.Errorf("failed to marshal modified data: %w", err)
	}

	_, err = tr.db.Exec(
		updateTroubleReportQuery,
		report.Title, report.Content, linkedAttachments, modified, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update trouble report with ID %d: %w", id, err)
	}

	// Create feed entry for the updated trouble report
	feed := NewTroubleReportUpdateFeed(report)
	if err := tr.feeds.Add(feed); err != nil {
		return fmt.Errorf("failed to add feed entry: %w", err)
	}

	return nil
}

// Remove deletes a trouble report by ID
func (tr *TroubleReports) Remove(id int64) error {
	// Get the trouble report before deleting it.
	// Ignore any error, just check if it exists before creating a feed entry
	report, _ := tr.Get(id)

	result, err := tr.db.Exec(deleteTroubleReportQuery, id)
	if err != nil {
		return fmt.Errorf("failed to delete trouble report with ID %d: %w", id, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	// Create feed entry for the removed trouble report
	if report != nil {
		feed := NewTroubleReportRemoveFeed(report)
		if err := tr.feeds.Add(feed); err != nil {
			return fmt.Errorf("failed to add feed entry: %w", err)
		}
	}

	return nil
}

func (tr *TroubleReports) scanTroubleReport(rows *sql.Rows) (*TroubleReport, error) {
	report := &TroubleReport{}
	var linkedAttachments, modified []byte

	if err := rows.Scan(&report.ID, &report.Title, &report.Content, &linkedAttachments, &modified); err != nil {
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	if err := json.Unmarshal(linkedAttachments, &report.LinkedAttachments); err != nil {
		return nil, fmt.Errorf("failed to unmarshal linked attachments: %w", err)
	}

	if err := json.Unmarshal(modified, &report.Modified); err != nil {
		return nil, fmt.Errorf("failed to unmarshal modified data: %w", err)
	}

	return report, nil
}

// scanTroubleReportRow scans a trouble report from a single row
func (tr *TroubleReports) scanTroubleReportRow(row *sql.Row) (*TroubleReport, error) {
	report := &TroubleReport{}
	var linkedAttachments, modified []byte

	if err := row.Scan(&report.ID, &report.Title, &report.Content, &linkedAttachments, &modified); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(linkedAttachments, &report.LinkedAttachments); err != nil {
		return nil, fmt.Errorf("failed to unmarshal linked attachments: %w", err)
	}

	if err := json.Unmarshal(modified, &report.Modified); err != nil {
		return nil, fmt.Errorf("failed to unmarshal modified data: %w", err)
	}

	return report, nil
}
