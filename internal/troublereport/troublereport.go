package troublereport

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/knackwurstking/pgpress/internal/dberror"
	"github.com/knackwurstking/pgpress/internal/feed"
	"github.com/knackwurstking/pgpress/internal/interfaces"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/models"
)

type Service struct {
	db    *sql.DB
	feeds *feed.Service
}

var _ interfaces.DataOperations[*models.TroubleReport] = (*Service)(nil)

func New(db *sql.DB, feeds *feed.Service) *Service {
	query := `
		CREATE TABLE IF NOT EXISTS trouble_reports (
			id INTEGER NOT NULL,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			linked_attachments TEXT NOT NULL,
			mods BLOB NOT NULL,
			PRIMARY KEY("id" AUTOINCREMENT)
		);
	`
	if _, err := db.Exec(query); err != nil {
		panic(dberror.NewDatabaseError(
			"create_table",
			"trouble_reports",
			"failed to create trouble_reports table",
			err,
		))
	}

	return &Service{
		db:    db,
		feeds: feeds,
	}
}

// List retrieves all trouble reports ordered by ID descending.
func (tr *Service) List() ([]*models.TroubleReport, error) {
	logger.DBTroubleReports().Info("Listing trouble reports")

	query := `SELECT * FROM trouble_reports ORDER BY id DESC`
	rows, err := tr.db.Query(query)
	if err != nil {
		return nil, dberror.NewDatabaseError("select", "trouble_reports",
			"failed to query trouble reports", err)
	}
	defer rows.Close()

	var reports []*models.TroubleReport

	for rows.Next() {
		report, err := tr.scanTroubleReport(rows)
		if err != nil {
			return nil, dberror.WrapError(err, "failed to scan trouble report")
		}
		reports = append(reports, report)
	}

	if err := rows.Err(); err != nil {
		return nil, dberror.NewDatabaseError("select", "trouble_reports",
			"error iterating over rows", err)
	}

	return reports, nil
}

// Get retrieves a specific trouble report by ID.
func (tr *Service) Get(id int64) (*models.TroubleReport, error) {
	logger.DBTroubleReports().Debug("Getting trouble report, id: %d", id)

	query := `SELECT * FROM trouble_reports WHERE id = ?`
	row := tr.db.QueryRow(query, id)

	report, err := tr.scanTroubleReport(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, dberror.ErrNotFound
		}
		return nil, dberror.NewDatabaseError("select", "trouble_reports",
			fmt.Sprintf("failed to get trouble report with ID %d", id), err)
	}

	return report, nil
}

// Add creates a new trouble report and generates a corresponding activity feed entry.
func (tr *Service) Add(troubleReport *models.TroubleReport, user *models.User) (int64, error) {
	logger.DBTroubleReports().Info("Adding trouble report: %+v", troubleReport)

	if troubleReport == nil {
		return 0, dberror.NewValidationError("report", "trouble report cannot be nil", nil)
	}

	if err := troubleReport.Validate(); err != nil {
		return 0, err
	}

	tr.updateMods(user, troubleReport)

	linkedAttachments, err := json.Marshal(troubleReport.LinkedAttachments)
	if err != nil {
		return 0, dberror.WrapError(err, "failed to marshal linked attachments")
	}

	mods, err := json.Marshal(troubleReport.Mods)
	if err != nil {
		return 0, dberror.WrapError(err, "failed to marshal mods data")
	}

	query := `INSERT INTO trouble_reports
		(title, content, linked_attachments, mods) VALUES (?, ?, ?, ?)`
	// After exec this insert query, i need to get the id
	result, err := tr.db.Exec(
		query,
		troubleReport.Title, troubleReport.Content, string(linkedAttachments), mods,
	)
	if err != nil {
		return 0, dberror.NewDatabaseError("insert", "trouble_reports",
			"failed to insert trouble report", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, dberror.NewDatabaseError("insert", "trouble_reports",
			"failed to get last insert ID", err)
	}
	troubleReport.ID = id

	feed := models.NewFeed(
		models.FeedTypeTroubleReportAdd,
		&models.FeedTroubleReportAdd{
			ID:         id,
			Title:      troubleReport.Title,
			ModifiedBy: troubleReport.Mods.Current().User,
		},
	)
	if err := tr.feeds.Add(feed); err != nil {
		return id, dberror.WrapError(err, "failed to add feed entry")
	}

	return id, nil
}

// Update modifies an existing trouble report and generates an activity feed entry.
func (tr *Service) Update(troubleReport *models.TroubleReport, user *models.User) error {
	id := troubleReport.ID
	logger.DBTroubleReports().Info("Updating trouble report, id: %d, data: %+v", id, troubleReport)

	if troubleReport == nil {
		return dberror.NewValidationError("report", "trouble report cannot be nil", nil)
	}

	if err := troubleReport.Validate(); err != nil {
		return err
	}

	tr.updateMods(user, troubleReport)

	linkedAttachments, err := json.Marshal(troubleReport.LinkedAttachments)
	if err != nil {
		return dberror.WrapError(err, "failed to marshal linked attachments")
	}

	mods, err := json.Marshal(troubleReport.Mods)
	if err != nil {
		return dberror.WrapError(err, "failed to marshal mods data")
	}

	query := `UPDATE trouble_reports
		SET title = ?, content = ?, linked_attachments = ?, mods = ? WHERE id = ?`
	_, err = tr.db.Exec(
		query,
		troubleReport.Title, troubleReport.Content, string(linkedAttachments), mods, id,
	)
	if err != nil {
		return dberror.NewDatabaseError("update", "trouble_reports",
			fmt.Sprintf("failed to update trouble report with ID %d", id), err)
	}

	feed := models.NewFeed(
		models.FeedTypeTroubleReportUpdate,
		&models.FeedTroubleReportUpdate{
			ID:         id,
			Title:      troubleReport.Title,
			ModifiedBy: troubleReport.Mods.Current().User,
		},
	)
	if err := tr.feeds.Add(feed); err != nil {
		return dberror.WrapError(err, "failed to add feed entry")
	}

	return nil
}

// Delete deletes a trouble report by ID and generates an activity feed entry.
func (tr *Service) Delete(id int64, user *models.User) error {
	logger.DBTroubleReports().Info("Removing trouble report, id: %d", id)

	// Get the user before deleting for the feed entry
	report, err := tr.Get(id)
	if err != nil {
		return dberror.WrapError(err, "failed to get trouble report before deletion")
	}

	query := `DELETE FROM trouble_reports WHERE id = ?`
	result, err := tr.db.Exec(query, id)
	if err != nil {
		return dberror.NewDatabaseError("delete", "trouble_reports",
			fmt.Sprintf("failed to delete trouble report with ID %d", id), err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return dberror.NewDatabaseError("delete", "trouble_reports",
			"failed to get rows affected", err)
	}

	if rowsAffected == 0 {
		return dberror.ErrNotFound
	}

	// Create feed entry for the removed user
	if user != nil {
		feed := models.NewFeed(
			models.FeedTypeTroubleReportRemove,
			&models.FeedTroubleReportRemove{
				ID:        report.ID,
				Title:     report.Title,
				RemovedBy: user,
			},
		)
		if err := tr.feeds.Add(feed); err != nil {
			return dberror.WrapError(err, "failed to add feed entry")
		}
	}

	return nil
}

func (tr *Service) scanTroubleReport(scanner interfaces.Scannable) (*models.TroubleReport, error) {
	report := &models.TroubleReport{}
	var linkedAttachments string
	var mods []byte

	if err := scanner.Scan(&report.ID, &report.Title, &report.Content,
		&linkedAttachments, &mods); err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, dberror.NewDatabaseError("scan", "trouble_reports",
			"failed to scan row", err)
	}

	// Try to unmarshal as new format (array of int64 IDs) first
	if err := json.Unmarshal([]byte(linkedAttachments), &report.LinkedAttachments); err != nil {
		return nil, dberror.WrapError(err, "failed to unmarshal linked attachments")
	}

	if err := json.Unmarshal(mods, &report.Mods); err != nil {
		return nil, dberror.WrapError(err, "failed to unmarshal mods data")
	}

	return report, nil
}

func (tr *Service) updateMods(user *models.User, report *models.TroubleReport) {
	if user == nil {
		return
	}

	report.Mods.Add(user, models.TroubleReportMod{
		Title:             report.Title,
		Content:           report.Content,
		LinkedAttachments: report.LinkedAttachments,
	})
}
