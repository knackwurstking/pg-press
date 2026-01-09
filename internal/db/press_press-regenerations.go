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
			id 				INTEGER NOT NULL,
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
)

// -----------------------------------------------------------------------------
// Press Regeneration Functions
// -----------------------------------------------------------------------------

// ListPressRegenerationsByPress retrieves all press regenerations for a specific press number
func ListPressRegenerationsByPress(pressNumber shared.PressNumber) ([]*shared.PressRegeneration, *errors.MasterError) {
	if !pressNumber.IsValid() {
		return nil, errors.NewValidationError("invalid press_number").MasterError()
	}

	r, err := dbPress.Query(sqlListPressRegenerationsByPress,
		sql.Named("press_number", pressNumber))
	if err != nil {
		return nil, errors.NewMasterError(err)
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
		return nil, errors.NewMasterError(err)
	}

	return regenerations, nil
}

// -----------------------------------------------------------------------------
// Scan Helpers
// -----------------------------------------------------------------------------

// ScanPressRegeneration scans a database row into a PressRegeneration struct
func ScanPressRegeneration(row Scannable) (pr *shared.PressRegeneration, merr *errors.MasterError) {
	pr = &shared.PressRegeneration{}
	err := row.Scan(
		&pr.ID,
		&pr.PressNumber,
		&pr.Start,
		&pr.Stop,
	)
	if err != nil {
		return nil, errors.NewMasterError(err)
	}
	return pr, nil
}
