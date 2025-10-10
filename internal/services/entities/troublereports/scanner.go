package troublereports

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/knackwurstking/pgpress/internal/interfaces"
	"github.com/knackwurstking/pgpress/internal/services/shared/scanner"
	"github.com/knackwurstking/pgpress/pkg/models"
)

// ScanTroubleReport scans a database row into a TroubleReport model
func ScanTroubleReport(scannable interfaces.Scannable) (*models.TroubleReport, error) {
	report := &models.TroubleReport{}
	var linkedAttachments string

	err := scannable.Scan(&report.ID, &report.Title, &report.Content, &linkedAttachments, &report.UseMarkdown)
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
	return scanner.ScanRows(rows, ScanTroubleReport)
}

// ScanTroubleReportsIntoMap scans trouble reports into a map by ID
func ScanTroubleReportsIntoMap(rows *sql.Rows) (map[int64]*models.TroubleReport, error) {
	return scanner.ScanIntoMap(rows, ScanTroubleReport, func(report *models.TroubleReport) int64 {
		return report.ID
	})
}
