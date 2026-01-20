package db

import (
	"database/sql"

	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"
)

// -----------------------------------------------------------------------------
// Table Creation Statements
// -----------------------------------------------------------------------------

const (
	sqlCreatePressRegenerationsTable string = `
		CREATE TABLE IF NOT EXISTS press_regenerations (
			id				INTEGER NOT NULL,
			press_number 	INTEGER NOT NULL,
			start 			INTEGER NOT NULL,
			stop 			INTEGER NOT NULL,

			PRIMARY KEY("id" AUTOINCREMENT)
		);
	`

	sqlListPressRegenerationsByPress string = `
		SELECT
			id,
			press_number,
			start,
			stop
		FROM
			press_regenerations
		WHERE
			press_number = :press_number
		ORDER BY
			start DESC;
	`

	sqlAddPressRegeneration string = `
		INSERT INTO press_regenerations (
			press_number,
			start,
			stop
		) VALUES (
			:press_number,
			:start,
			:stop
		)
	`

	sqlGetPressRegeneration string = `
		SELECT
			id,
			press_number,
			start,
			stop
		FROM
			press_regenerations
		WHERE
			id = :id
	`

	sqlUpdatePressRegeneration string = `
		UPDATE press_regenerations
		SET
			press_number = :press_number,
			start = :start,
			stop = :stop
		WHERE
			id = :id
	`

	sqlDeletePressRegeneration string = `
		DELETE FROM press_regenerations
		WHERE
			id = :id
	`
)

// -----------------------------------------------------------------------------
// Press Regeneration Functions
// -----------------------------------------------------------------------------

// AddPressRegeneration adds a new press regeneration to the database
func AddPressRegeneration(pr *shared.PressRegeneration) *errors.HTTPError {
	if verr := pr.Validate(); verr != nil {
		return verr.HTTPError()
	}

	_, err := dbPress.Exec(sqlAddPressRegeneration,
		sql.Named("press_number", pr.PressNumber),
		sql.Named("start", pr.Start),
		sql.Named("stop", pr.Stop),
	)
	if err != nil {
		return errors.NewHTTPError(err)
	}
	return nil
}

// GetPressRegeneration retrieves a press regeneration by its ID
func GetPressRegeneration(id shared.EntityID) (*shared.PressRegeneration, *errors.HTTPError) {
	return ScanPressRegeneration(dbPress.QueryRow(sqlGetPressRegeneration, sql.Named("id", id)))
}

// UpdatePressRegeneration updates an existing press regeneration in the database
func UpdatePressRegeneration(pr *shared.PressRegeneration) *errors.HTTPError {
	if verr := pr.Validate(); verr != nil {
		return verr.HTTPError()
	}

	_, err := dbPress.Exec(sqlUpdatePressRegeneration,
		sql.Named("id", pr.ID),
		sql.Named("press_number", pr.PressNumber),
		sql.Named("start", pr.Start),
		sql.Named("stop", pr.Stop),
	)
	if err != nil {
		return errors.NewHTTPError(err)
	}
	return nil
}

// DeletePressRegeneration removes a press regeneration from the database
func DeletePressRegeneration(id shared.EntityID) *errors.HTTPError {
	_, err := dbPress.Exec(sqlDeletePressRegeneration, sql.Named("id", id))
	if err != nil {
		return errors.NewHTTPError(err)
	}
	return nil
}

// ListPressRegenerationsByPress retrieves all press regenerations for a specific press number
func ListPressRegenerationsByPress(pressNumber shared.PressNumber) ([]*shared.PressRegeneration, *errors.HTTPError) {
	if !pressNumber.IsValid() {
		return nil, errors.NewValidationError("invalid press_number").HTTPError()
	}

	r, err := dbPress.Query(sqlListPressRegenerationsByPress,
		sql.Named("press_number", pressNumber))
	if err != nil {
		return nil, errors.NewHTTPError(err)
	}
	defer r.Close()

	var regenerations []*shared.PressRegeneration
	for r.Next() {
		pr, merr := ScanPressRegeneration(r)
		if merr != nil {
			return nil, merr
		}
		regenerations = append(regenerations, pr)
	}
	if err = r.Err(); err != nil {
		return nil, errors.NewHTTPError(err)
	}

	return regenerations, nil
}

// -----------------------------------------------------------------------------
// Scan Helpers
// -----------------------------------------------------------------------------

// ScanPressRegeneration scans a database row into a PressRegeneration struct
func ScanPressRegeneration(row Scannable) (pr *shared.PressRegeneration, merr *errors.HTTPError) {
	pr = &shared.PressRegeneration{}
	err := row.Scan(
		&pr.ID,
		&pr.PressNumber,
		&pr.Start,
		&pr.Stop,
	)
	if err != nil {
		return nil, errors.NewHTTPError(err)
	}
	return pr, nil
}
