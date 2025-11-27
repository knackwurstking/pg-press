package services

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
)

const TableNamePressRegenerations = "press_regenerations"

type PressRegenerations struct {
	*Base
}

func NewPressRegenerations(r *Registry) *PressRegenerations {
	base := NewBase(r)

	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %[1]s (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			press_number INTEGER NOT NULL,
			started_at DATETIME NOT NULL,
			completed_at DATETIME,
			reason TEXT
		);
	`, TableNamePressRegenerations)

	if err := base.CreateTable(query, TableNamePressRegenerations); err != nil {
		panic(err)
	}

	return &PressRegenerations{
		Base: base,
	}
}

func (s *PressRegenerations) Get(id models.PressRegenerationID) (*models.PressRegeneration, error) {
	slog.Debug("Getting press regeneration by ID", "id", id)

	query := fmt.Sprintf(`SELECT * FROM %s WHERE id = ?`, TableNamePressRegenerations)
	row := s.DB.QueryRow(query, id)

	regeneration, err := ScanSingleRow(row, scanPressRegeneration)
	if err != nil {
		return nil, s.GetSelectError(err)
	}

	return regeneration, nil
}

func (s *PressRegenerations) Add(r *models.PressRegeneration) (models.PressRegenerationID, error) {
	slog.Debug(
		"Adding press regeneration",
		"press", r.PressNumber,
		"started_at", r.StartedAt, "completed_at", r.CompletedAt,
		"reason", r.Reason,
	)

	if err := r.Validate(); err != nil {
		return 0, err
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (press_number, started_at, completed_at, reason)
		VALUES (?, ?, ?, ?)
	`, TableNamePressRegenerations)

	result, err := s.DB.Exec(query, r.PressNumber, r.StartedAt, r.CompletedAt, r.Reason)
	if err != nil {
		return 0, s.GetInsertError(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get last insert ID: %v", err)
	}

	return models.PressRegenerationID(id), nil
}

func (s *PressRegenerations) Update(r *models.PressRegeneration) error {
	slog.Debug(
		"Updating press regeneration",
		"id", r.ID, "press", r.PressNumber,
		"started_at", r.StartedAt, "completed_at", r.CompletedAt,
		"reason", r.Reason,
	)

	var (
		err   error
		query string
	)

	if err = r.Validate(); err != nil {
		return err
	}

	query = fmt.Sprintf(`
		UPDATE %s
		SET started_at = ?, completed_at = ?, reason = ?
		WHERE id = ?
	`, TableNamePressRegenerations)

	if _, err = s.DB.Exec(query, r.StartedAt, r.CompletedAt, r.Reason, r.ID); err != nil {
		return s.GetUpdateError(err)
	}

	return nil
}

func (s *PressRegenerations) Delete(id models.PressRegenerationID) error {
	slog.Debug("Deleting press regeneration", "id", id)

	query := fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, TableNamePressRegenerations)
	_, err := s.DB.Exec(query, id)
	if err != nil {
		return s.GetDeleteError(err)
	}

	return nil
}

func (s *PressRegenerations) StartPressRegeneration(pn models.PressNumber, reason string) (models.PressRegenerationID, error) {
	slog.Debug("Starting press regeneration", "press_number", pn, "reason", reason)

	regenerationID, err := s.Add(models.NewPressRegeneration(pn, time.Now(), reason))
	if err != nil {
		return 0, err
	}

	return regenerationID, nil
}

func (s *PressRegenerations) StopPressRegeneration(id models.PressRegenerationID) error {
	slog.Debug("Stopping press regeneration", "id", id)

	regeneration, err := s.Get(id)
	if err != nil {
		return err
	}

	regeneration.Stop()

	if err := regeneration.Validate(); err != nil {
		return err
	}

	query := fmt.Sprintf(`
		UPDATE %s
		SET completed_at = ?
		WHERE id = ?
	`, TableNamePressRegenerations)

	_, err = s.DB.Exec(query, regeneration.CompletedAt, id)
	if err != nil {
		return s.GetUpdateError(err)
	}

	return nil
}

func (s *PressRegenerations) GetLastRegeneration(pressNumber models.PressNumber) (*models.PressRegeneration, error) {
	query := fmt.Sprintf(`
		SELECT id, press_number, started_at, completed_at, reason
		FROM %s
		WHERE press_number = ?
		ORDER BY id DESC
		LIMIT 1
	`, TableNamePressRegenerations)

	row := s.DB.QueryRow(query, pressNumber)
	regeneration, err := ScanSingleRow(row, scanPressRegeneration)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError(
				fmt.Sprintf("press regeneration for press_number: %d", pressNumber),
			)
		}
		return nil, err
	}

	slog.Debug("Got last regeneration for press", "press", pressNumber, "regeneration", regeneration)
	return regeneration, nil
}

func (s *PressRegenerations) GetRegenerationHistory(pressNumber models.PressNumber) ([]*models.PressRegeneration, error) {
	slog.Debug("Getting regeneration history for press", "press", pressNumber)

	query := fmt.Sprintf(`
		SELECT id, press_number, started_at, completed_at, reason
		FROM %s
		WHERE press_number = ?
		ORDER BY id DESC
	`, TableNamePressRegenerations)

	rows, err := s.DB.Query(query, pressNumber)
	if err != nil {
		return nil, s.GetSelectError(err)
	}
	defer rows.Close()

	regenerations, err := ScanRows(rows, scanPressRegeneration)
	if err != nil {
		return nil, err
	}

	return regenerations, nil
}

func scanPressRegeneration(scannable Scannable) (*models.PressRegeneration, error) {
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
