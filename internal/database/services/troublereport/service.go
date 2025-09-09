package troublereport

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/knackwurstking/pgpress/internal/database/dberror"
	"github.com/knackwurstking/pgpress/internal/database/interfaces"
	attachmentmodels "github.com/knackwurstking/pgpress/internal/database/models/attachment"
	feedmodels "github.com/knackwurstking/pgpress/internal/database/models/feed"
	trmodels "github.com/knackwurstking/pgpress/internal/database/models/troublereport"
	usermodels "github.com/knackwurstking/pgpress/internal/database/models/user"
	"github.com/knackwurstking/pgpress/internal/database/services/attachment"
	"github.com/knackwurstking/pgpress/internal/database/services/feed"
	"github.com/knackwurstking/pgpress/internal/logger"
)

type Service struct {
	db          *sql.DB
	attachments *attachment.Service
	feeds       *feed.Service
}

var _ interfaces.DataOperations[*trmodels.TroubleReport] = (*Service)(nil)

func New(db *sql.DB, attachments *attachment.Service, feeds *feed.Service) *Service {
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
		db:          db,
		attachments: attachments,
		feeds:       feeds,
	}
}

// List retrieves all trouble reports ordered by ID descending.
func (tr *Service) List() ([]*trmodels.TroubleReport, error) {
	logger.DBTroubleReports().Info("Listing trouble reports")

	query := `SELECT * FROM trouble_reports ORDER BY id DESC`
	rows, err := tr.db.Query(query)
	if err != nil {
		return nil, dberror.NewDatabaseError("select", "trouble_reports",
			"failed to query trouble reports", err)
	}
	defer rows.Close()

	var reports []*trmodels.TroubleReport

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
func (tr *Service) Get(id int64) (*trmodels.TroubleReport, error) {
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
func (tr *Service) Add(troubleReport *trmodels.TroubleReport, user *usermodels.User) (int64, error) {
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

	feed := feedmodels.New(
		"Neuer Problembericht",
		fmt.Sprintf("Benutzer %s hat einen neuen Problembericht '%s' hinzugefügt.",
			troubleReport.Mods.Current().User.Name, troubleReport.Title),
		troubleReport.Mods.Current().User.TelegramID,
	)
	if err := tr.feeds.Add(feed); err != nil {
		return id, dberror.WrapError(err, "failed to add feed entry")
	}

	return id, nil
}

// Update modifies an existing trouble report and generates an activity feed entry.
func (tr *Service) Update(troubleReport *trmodels.TroubleReport, user *usermodels.User) error {
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

	feed := feedmodels.New(
		"Problembericht aktualisiert",
		fmt.Sprintf("Benutzer %s hat den Problembericht '%s' aktualisiert.",
			troubleReport.Mods.Current().User.Name, troubleReport.Title),
		troubleReport.Mods.Current().User.TelegramID,
	)
	if err := tr.feeds.Add(feed); err != nil {
		return dberror.WrapError(err, "failed to add feed entry")
	}

	return nil
}

// Delete deletes a trouble report by ID and generates an activity feed entry.
func (tr *Service) Delete(id int64, user *usermodels.User) error {
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
		feed := feedmodels.New(
			"Problembericht entfernt",
			fmt.Sprintf("Benutzer %s hat den Problembericht '%s' entfernt.",
				user.Name, report.Title),
			user.TelegramID,
		)
		if err := tr.feeds.Add(feed); err != nil {
			return dberror.WrapError(err, "failed to add feed entry")
		}
	}

	return nil
}

// GetWithAttachments retrieves a trouble report and loads its attachments.
func (s *Service) GetWithAttachments(
	id int64,
) (*trmodels.TroubleReportWithAttachments, error) {
	logger.DBTroubleReports().Debug(
		"Getting trouble report with attachments, id: %d", id)

	// Get the trouble report
	tr, err := s.Get(id)
	if err != nil {
		return nil, err
	}

	// Load attachments
	attachments, err := s.attachments.GetByIDs(tr.LinkedAttachments)
	if err != nil {
		return nil, dberror.WrapError(err, "failed to load attachments for trouble report")
	}

	return &trmodels.TroubleReportWithAttachments{
		TroubleReport:     tr,
		LoadedAttachments: attachments,
	}, nil
}

// ListWithAttachments retrieves all trouble reports and loads their attachments.
func (s *Service) ListWithAttachments() ([]*trmodels.TroubleReportWithAttachments, error) {
	logger.DBTroubleReports().Debug("Listing trouble reports with attachments")

	// Get all trouble reports
	reports, err := s.List()
	if err != nil {
		return nil, err
	}

	var result []*trmodels.TroubleReportWithAttachments

	for _, tr := range reports {
		// Load attachments for each report
		attachments, err := s.attachments.GetByIDs(tr.LinkedAttachments)
		if err != nil {
			return nil, dberror.WrapError(err,
				fmt.Sprintf("failed to load attachments for trouble report %d", tr.ID))
		}

		result = append(result, &trmodels.TroubleReportWithAttachments{
			TroubleReport:     tr,
			LoadedAttachments: attachments,
		})
	}

	return result, nil
}

// AddWithAttachments creates a new trouble report and its attachments.
func (s *Service) AddWithAttachments(
	user *usermodels.User,
	troubleReport *trmodels.TroubleReport,
	attachments []*attachmentmodels.Attachment,
) error {
	logger.DBTroubleReports().Info("Adding trouble report with %d attachments", len(attachments))

	if troubleReport == nil {
		return dberror.NewValidationError("report", "trouble report cannot be nil", nil)
	}

	// First, add the attachments and collect their IDs
	var attachmentIDs []int64
	for _, attachment := range attachments {
		if attachment == nil {
			continue
		}

		id, err := s.attachments.Add(attachment, user)
		if err != nil {
			// Cleanup already added attachments on failure
			for _, addedID := range attachmentIDs {
				s.attachments.Delete(addedID, user)
			}
			return dberror.WrapError(err, "failed to add attachment")
		}
		attachmentIDs = append(attachmentIDs, id)
	}

	// Set the attachment IDs in the trouble report
	troubleReport.LinkedAttachments = attachmentIDs

	// Add the trouble report
	if _, err := s.Add(troubleReport, user); err != nil {
		// Cleanup attachments on failure
		for _, id := range attachmentIDs {
			s.attachments.Delete(id, user)
		}
		return dberror.WrapError(err, "failed to add trouble report")
	}

	return nil
}

// UpdateWithAttachments updates a trouble report and manages its attachments.
func (s *Service) UpdateWithAttachments(
	user *usermodels.User,
	id int64,
	troubleReport *trmodels.TroubleReport,
	newAttachments []*attachmentmodels.Attachment,
) error {
	logger.DBTroubleReports().Info(
		"Updating trouble report %d with %d new attachments", id, len(newAttachments))

	if troubleReport == nil {
		return dberror.NewValidationError("report", "trouble report cannot be nil", nil)
	}

	// Add new attachments
	var newAttachmentIDs []int64
	for _, attachment := range newAttachments {
		if attachment == nil {
			continue
		}

		attachmentID, err := s.attachments.Add(attachment, user)
		if err != nil {
			// Cleanup already added attachments on failure
			for _, addedID := range newAttachmentIDs {
				s.attachments.Delete(addedID, user)
			}
			return dberror.WrapError(err, "failed to add new attachment")
		}
		newAttachmentIDs = append(newAttachmentIDs, attachmentID)
	}

	// Combine existing and new attachment IDs
	allAttachmentIDs := append(troubleReport.LinkedAttachments, newAttachmentIDs...)
	troubleReport.LinkedAttachments = allAttachmentIDs
	troubleReport.ID = id

	// Update the trouble report
	if err := s.Update(troubleReport, user); err != nil {
		// Cleanup new attachments on failure
		for _, attachmentID := range newAttachmentIDs {
			s.attachments.Delete(attachmentID, user)
		}
		return dberror.WrapError(err, "failed to update trouble report")
	}

	return nil
}

// RemoveWithAttachments removes a trouble report and its attachments.
func (s *Service) RemoveWithAttachments(id int64, user *usermodels.User) (*trmodels.TroubleReport, error) {
	logger.DBTroubleReports().Info("Removing trouble report %d with attachments", id)

	// Get the trouble report to find its attachments
	tr, err := s.Get(id)
	if err != nil {
		return tr, dberror.WrapError(err, "failed to get trouble report for removal")
	}

	// Remove the trouble report first
	if err := s.Delete(id, user); err != nil {
		return tr, dberror.WrapError(err, "failed to remove trouble report")
	}

	// Remove associated attachments
	for _, attachmentID := range tr.LinkedAttachments {
		if err := s.attachments.Delete(attachmentID, user); err != nil {
			logger.DBTroubleReports().Warn("Failed to remove attachment %d: %v", attachmentID, err)
		}
	}

	return tr, nil
}

// LoadAttachments loads attachments for a trouble report.
func (s *Service) LoadAttachments(tr *trmodels.TroubleReport) ([]*attachmentmodels.Attachment, error) {
	logger.DBTroubleReports().Debug("Loading attachments for trouble report")

	if tr == nil {
		return nil, dberror.NewValidationError("report", "trouble report cannot be nil", nil)
	}

	return s.attachments.GetByIDs(tr.LinkedAttachments)
}

// GetAttachment retrieves a specific attachment by ID.
func (s *Service) GetAttachment(id int64) (*attachmentmodels.Attachment, error) {
	logger.DBTroubleReports().Debug("Getting attachment with ID %d", id)
	return s.attachments.Get(id)
}

func (tr *Service) scanTroubleReport(scanner interfaces.Scannable) (*trmodels.TroubleReport, error) {
	report := &trmodels.TroubleReport{}
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

func (tr *Service) updateMods(user *usermodels.User, report *trmodels.TroubleReport) {
	if user == nil {
		return
	}

	report.Mods.Add(user, trmodels.TroubleReportMod{
		Title:             report.Title,
		Content:           report.Content,
		LinkedAttachments: report.LinkedAttachments,
	})
}
