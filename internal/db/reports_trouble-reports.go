package db

import (
	"database/sql"
	"encoding/json"

	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"
)

// -----------------------------------------------------------------------------
// Table Creation Statements
// -----------------------------------------------------------------------------

const (
	sqlCreateTroubleReportsTable string = `
		CREATE TABLE IF NOT EXISTS trouble_reports (
			id                  INTEGER NOT NULL,
			title               TEXT NOT NULL,
			content             TEXT NOT NULL,
			linked_attachments  TEXT,
			use_markdown        INTEGER NOT NULL DEFAULT 0,

			PRIMARY KEY("id" AUTOINCREMENT)
		);
	`

	sqlAddTroubleReport string = `
		INSERT INTO trouble_reports (title, content, linked_attachments, use_markdown)
		VALUES (:title, :content, :linked_attachments, :use_markdown);
	`

	sqlUpdateTroubleReport string = `
		UPDATE trouble_reports
		SET title               = :title,
		    content             = :content,
		    linked_attachments  = :linked_attachments,
		    use_markdown        = :use_markdown
		WHERE id = :id;
	`

	sqlGetTroubleReport string = `
		SELECT id, title, content, linked_attachments, use_markdown
		FROM trouble_reports
		WHERE id = :id;
	`

	sqlListTroubleReports string = `
		SELECT id, title, content, linked_attachments, use_markdown
		FROM trouble_reports
		ORDER BY id DESC;
	`

	sqlDeleteTroubleReport string = `
		DELETE FROM trouble_reports
		WHERE id = :id;
	`
)

// -----------------------------------------------------------------------------
// Trouble Report Functions
// -----------------------------------------------------------------------------

// AddTroubleReport adds a new trouble report to the database
func AddTroubleReport(report *shared.TroubleReport) *errors.HTTPError {
	if verr := report.Validate(); verr != nil {
		return verr.HTTPError().Wrap("invalid trouble report data")
	}

	// Convert linked_attachments slice to JSON string
	linkedAttachmentsJSON, err := json.Marshal(report.LinkedAttachments)
	if err != nil {
		return errors.NewHTTPError(err).Wrap("failed to marshal linked attachments")
	}

	_, err = dbReports.Exec(sqlAddTroubleReport,
		sql.Named("title", report.Title),
		sql.Named("content", report.Content),
		sql.Named("linked_attachments", string(linkedAttachmentsJSON)),
		sql.Named("use_markdown", boolToInt(report.UseMarkdown)),
	)
	if err != nil {
		return errors.NewHTTPError(err)
	}

	return nil
}

// UpdateTroubleReport updates an existing trouble report in the database
func UpdateTroubleReport(report *shared.TroubleReport) *errors.HTTPError {
	if verr := report.Validate(); verr != nil {
		return verr.HTTPError().Wrap("invalid trouble report data")
	}

	// Convert linked_attachments slice to JSON string
	linkedAttachmentsJSON, err := json.Marshal(report.LinkedAttachments)
	if err != nil {
		return errors.NewHTTPError(err).Wrap("failed to marshal linked attachments")
	}

	_, err = dbReports.Exec(sqlUpdateTroubleReport,
		sql.Named("id", report.ID),
		sql.Named("title", report.Title),
		sql.Named("content", report.Content),
		sql.Named("linked_attachments", string(linkedAttachmentsJSON)),
		sql.Named("use_markdown", boolToInt(report.UseMarkdown)),
	)
	if err != nil {
		return errors.NewHTTPError(err)
	}

	return nil
}

// GetTroubleReport retrieves a trouble report by its ID
func GetTroubleReport(id shared.EntityID) (*shared.TroubleReport, *errors.HTTPError) {
	return ScanTroubleReport(dbReports.QueryRow(sqlGetTroubleReport, sql.Named("id", id)))
}

// ListTroubleReports retrieves all trouble reports from the database
func ListTroubleReports() (reports []*shared.TroubleReport, merr *errors.HTTPError) {
	rows, err := dbReports.Query(sqlListTroubleReports)
	if err != nil {
		return nil, errors.NewHTTPError(err)
	}
	defer rows.Close()

	for rows.Next() {
		report, merr := ScanTroubleReport(rows)
		if merr != nil {
			return nil, merr
		}
		reports = append(reports, report)
	}

	return reports, nil
}

// DeleteTroubleReport removes a trouble report from the database
func DeleteTroubleReport(id shared.EntityID) *errors.HTTPError {
	_, err := dbReports.Exec(sqlDeleteTroubleReport, sql.Named("id", id))
	if err != nil {
		return errors.NewHTTPError(err)
	}
	return nil
}

// -----------------------------------------------------------------------------
// Scan Helpers
// -----------------------------------------------------------------------------

// ScanTroubleReport scans a database row into a TroubleReport struct
func ScanTroubleReport(row Scannable) (*shared.TroubleReport, *errors.HTTPError) {
	var (
		report  shared.TroubleReport
		jsonStr string
	)
	err := row.Scan(
		&report.ID,
		&report.Title,
		&report.Content,
		&jsonStr,
		&report.UseMarkdown,
	)
	if err != nil {
		return nil, errors.NewHTTPError(err)
	}

	// Parse linked_attachments JSON string back to slice
	if jsonStr != "" {
		err = json.Unmarshal([]byte(jsonStr), &report.LinkedAttachments)
		if err != nil {
			return nil, errors.NewHTTPError(err).Wrap("failed to unmarshal linked attachments")
		}
	}

	return &report, nil
}

// Helper function to convert bool to int for database storage
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
