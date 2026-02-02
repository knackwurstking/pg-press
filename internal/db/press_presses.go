package db

import (
	"database/sql"

	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"
)

// -----------------------------------------------------------------------------
// Table Creation Statements
// -----------------------------------------------------------------------------

// SQL statements for managing press records in the database.
const (
	// sqlCreatePressesTable creates the presses table with all required columns.
	sqlCreatePressesTable string = `
		CREATE TABLE IF NOT EXISTS presses (
			id 					INTEGER NOT NULL,
			type 				TEXT NOT NULL,
			code 				TEXT NOT NULL,
			slot_up 			INTEGER NOT NULL,
			slot_down 			INTEGER NOT NULL,
			cycles_offset 		INTEGER NOT NULL,

			PRIMARY KEY (id, type, code),
			UNIQUE (id, type, code)
		);
	`

	// sqlAddPress inserts a new press record into the database.
	sqlAddPress string = `
		INSERT INTO presses (
			id,
			type,
			code,
			slot_up,
			slot_down,
			cycles_offset
		) VALUES (
			:id,
			:type,
			:code,
			:slot_up,
			:slot_down,
			:cycles_offset
		)
	`

	// sqlUpdatePress updates an existing press record in the database.
	sqlUpdatePress string = `
		UPDATE presses
		SET
			type = :type
			code = :code,
			slot_up = :slot_up,
			slot_down = :slot_down,
			cycles_offset = :cycles_offset
		WHERE id = :id
	`

	// sqlGetPress retrieves a single press record by ID.
	sqlGetPress string = `
		SELECT
			id,
			type
			code,
			slot_up,
			slot_down,
			cycles_offset
		FROM presses
		WHERE id = :id
	`

	// sqlGetPressNumberForTool finds the press number that contains a specific tool in either slot.
	sqlGetPressNumberForTool string = `
		SELECT id, type, code
		FROM presses
		WHERE slot_up = :tool_id OR slot_down = :tool_id
		LIMIT 1;
	`

	// sqlGetPressUtilization retrieves press details for utilization reporting.
	sqlGetPressUtilization string = `
		SELECT
			id,
			type,
			code,
			slot_up,
			slot_down,
			cycles_offset
		FROM presses
		WHERE id = :press_number, type = :type, code = :code;
	`

	// sqlListPress retrieves all press records from the database.
	sqlListPress string = `
		SELECT
			id,
			type,
			code,
			slot_up,
			slot_down,
			cycles_offset
		FROM presses
		ORDER BY id ASC
	`

	// sqlListPressNumbers retrieves all press numbers from the database.
	sqlListPressNumbers = `
		SELECT id, type, code FROM presses ORDER BY id ASC
	`

	// sqlDeletePress removes a press record from the database.
	sqlDeletePress string = `
		DELETE FROM presses
		WHERE id = :id, type = :type, code = :code
	`
)

// -----------------------------------------------------------------------------
// Press Functions
// -----------------------------------------------------------------------------

// AddPress adds a new press to the database.
//
// It validates the press data before insertion and returns appropriate HTTP errors
// if validation fails or database operations encounter issues.
//
// Parameters:
//   - press: Pointer to the Press struct to be added
//
// Returns:
//   - *errors.HTTPError: Error if operation fails, nil on success
func AddPress(press *shared.Press) *errors.HTTPError {
	if verr := press.Validate(); verr != nil {
		return verr.HTTPError()
	}

	_, err := dbPress.Exec(sqlAddPress,
		sql.Named("id", press.ID),
		sql.Named("type", press.Type),
		sql.Named("code", press.Code),
		sql.Named("slot_up", press.SlotUp),
		sql.Named("slot_down", press.SlotDown),
		sql.Named("cycles_offset", press.CyclesOffset),
	)
	if err != nil {
		return errors.NewHTTPError(err)
	}
	return nil
}

// UpdatePress updates an existing press in the database.
//
// It validates the press data before updating and returns appropriate HTTP errors
// if validation fails or database operations encounter issues.
//
// Parameters:
//   - press: Pointer to the Press struct to be updated
//
// Returns:
//   - *errors.HTTPError: Error if operation fails, nil on success
func UpdatePress(press *shared.Press) *errors.HTTPError {
	if verr := press.Validate(); verr != nil {
		return verr.HTTPError()
	}

	_, err := dbPress.Exec(sqlUpdatePress,
		sql.Named("id", press.ID),
		sql.Named("type", press.Type),
		sql.Named("code", press.Code),
		sql.Named("slot_up", press.SlotUp),
		sql.Named("slot_down", press.SlotDown),
		sql.Named("cycles_offset", press.CyclesOffset),
	)
	if err != nil {
		return errors.NewHTTPError(err)
	}
	return nil
}

// GetPress retrieves a press by its ID.
//
// It queries the database for a specific press record and returns it as a Press struct.
// Returns an error if no record is found or database query fails.
//
// Parameters:
//   - id: The press number to retrieve
//
// Returns:
//   - *shared.Press: Pointer to the retrieved Press struct, or nil if not found
//   - *errors.HTTPError: Error if operation fails, nil on success
func GetPress(id shared.PressNumber) (*shared.Press, *errors.HTTPError) {
	return ScanPress(dbPress.QueryRow(sqlGetPress, sql.Named("id", id)))
}

// GetPressNumberForTool retrieves the press number that has the given tool in either slot.
//
// It searches for a press record where the specified tool ID appears in either slot_up or slot_down.
// Returns -1 if no press is found for the tool.
//
// Parameters:
//   - toolID: The entity ID of the tool to search for
//
// Returns:
//   - shared.PressNumber: The press number containing the tool, or -1 if not found
//   - *errors.HTTPError: Error if database query fails, nil on success
func GetPressNumberForTool(toolID shared.EntityID) (shared.PressNumber, *errors.HTTPError) {
	var pressNumber shared.PressNumber = -1

	err := dbPress.QueryRow(sqlGetPressNumberForTool, sql.Named("tool_id", toolID)).Scan(&pressNumber)
	if err != nil && err != sql.ErrNoRows {
		return pressNumber, errors.NewHTTPError(err)
	}

	return pressNumber, nil
}

// GetPressUtilization retrieves detailed utilization information for a specific press.
//
// It fetches the press details and resolves tool entities in both slots to provide
// complete information about tools currently installed on the press.
//
// Parameters:
//   - pressNumber: The press number to retrieve utilization for
//
// Returns:
//   - *shared.PressUtilization: Pointer to the populated PressUtilization struct
//   - *errors.HTTPError: Error if operation fails, nil on success
func GetPressUtilization(pressNumber shared.PressNumber) (*shared.PressUtilization, *errors.HTTPError) {
	var (
		slotUpper shared.EntityID
		slotLower shared.EntityID
	)
	pu := &shared.PressUtilization{}
	err := dbPress.QueryRow(sqlGetPressUtilization,
		sql.Named("press_number", pressNumber),
		sql.Named("type", pressNumber),
		sql.Named("code", pressNumber),
	).Scan(
		&pu.PressNumber,
		&pu.Type,
		&pu.Code,
		&slotUpper,
		&slotLower,
		&pu.CyclesOffset,
	)
	if err != nil {
		return pu, errors.NewHTTPError(err).Wrap("error query press utilization for press %d failed", pressNumber)
	}

	if slotUpper > 0 {
		tool, herr := GetTool(slotUpper)
		if herr != nil {
			return nil, herr
		}
		pu.SlotUpper = tool

		if tool.Cassette > 0 {
			cassette, herr := GetTool(tool.Cassette)
			if herr != nil {
				return nil, herr.Wrap(
					"error getting upper cassette tool (%d) for tool ID %d",
					tool.Cassette, tool.ID,
				)
			}
			pu.SlotUpperCassette = cassette
		}
	}
	if slotLower > 0 {
		tool, herr := GetTool(slotLower)
		if herr != nil {
			return nil, herr
		}
		pu.SlotLower = tool
	}

	return pu, nil
}

// GetPressUtilizations retrieves all press utilizations with tool information.
//
// It fetches utilization details for multiple presses and resolves tool entities
// to provide comprehensive information about all press configurations.
//
// Parameters:
//   - pressNumbers: Slice of press numbers to retrieve utilization for
//
// Returns:
//   - map[shared.PressNumber]*shared.PressUtilization: Map of press numbers to utilization info
//   - *errors.HTTPError: Error if operation fails, nil on success
func GetPressUtilizations(pressNumbers ...shared.PressNumber) (
	pu map[shared.PressNumber]*shared.PressUtilization, herr *errors.HTTPError,
) {
	pu = make(map[shared.PressNumber]*shared.PressUtilization)
	for _, pn := range pressNumbers {
		u, herr := GetPressUtilization(pn)
		if herr != nil {
			return pu, herr.Wrap("%d", pn)
		}
		pu[pn] = u
	}

	return pu, nil
}

// ListPress retrieves all presses from the database.
//
// It queries the database for all press records and returns them as a slice of Press structs.
// Returns an error if the query fails.
//
// Returns:
//   - []*shared.Press: Slice of pointers to Press structs
//   - *errors.HTTPError: Error if operation fails, nil on success
func ListPress() (presses []*shared.Press, herr *errors.HTTPError) {
	r, err := dbPress.Query(sqlListPress)
	if err != nil {
		return nil, errors.NewHTTPError(err)
	}
	defer r.Close()

	for r.Next() {
		press, herr := ScanPress(r)
		if herr != nil {
			return nil, herr.Wrap("scanning press row failed")
		}
		presses = append(presses, press)
	}

	if err := r.Err(); err != nil {
		return nil, errors.NewHTTPError(err)
	}

	return presses, nil
}

// ListPressNumbers retrieves all press numbers from the database.
//
// It queries the database for all press IDs and returns them as a slice.
// Returns an error if the query fails.
//
// Returns:
//   - []shared.PressNumber: Slice of press numbers
//   - *errors.HTTPError: Error if operation fails, nil on success
func ListPressNumbers() (pressNumbers []shared.PressNumber, herr *errors.HTTPError) {
	r, err := dbPress.Query(sqlListPressNumbers)
	if err != nil {
		return nil, errors.NewHTTPError(err)
	}
	defer r.Close()

	for r.Next() {
		var pn shared.PressNumber
		err := r.Scan(&pn)
		if err != nil {
			return nil, errors.NewHTTPError(err).Wrap("scanning press number row failed")
		}
		pressNumbers = append(pressNumbers, pn)
	}

	if err := r.Err(); err != nil {
		return nil, errors.NewHTTPError(err)
	}

	return pressNumbers, nil
}

// DeletePress removes a press from the database.
//
// It deletes the specified press record and returns an error if the operation fails.
//
// Parameters:
//   - id: The press number to delete
//
// Returns:
//   - *errors.HTTPError: Error if operation fails, nil on success
func DeletePress(id shared.PressNumber) *errors.HTTPError {
	_, err := dbPress.Exec(sqlDeletePress, sql.Named("id", id), sql.Named("type", id), sql.Named("code", id))
	if err != nil {
		return errors.NewHTTPError(err)
	}
	return nil
}

// -----------------------------------------------------------------------------
// Scan Helpers
// -----------------------------------------------------------------------------

// ScanPress scans a database row into a Press struct.
//
// It takes a scannable database result and maps the columns to a Press struct,
// returning any errors encountered during the scan operation.
//
// Parameters:
//   - row: Scannable database row to read from
//
// Returns:
//   - *shared.Press: Pointer to the populated Press struct, or nil if not found
//   - *errors.HTTPError: Error if operation fails, nil on success
func ScanPress(row Scannable) (press *shared.Press, herr *errors.HTTPError) {
	press = &shared.Press{}
	err := row.Scan(
		&press.ID,
		&press.Type,
		&press.Code,
		&press.SlotUp,
		&press.SlotDown,
		&press.CyclesOffset,
	)
	if err != nil {
		return nil, errors.NewHTTPError(err)
	}
	return press, nil
}
