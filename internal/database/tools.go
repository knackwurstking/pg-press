package database

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/knackwurstking/pgpress/internal/logger"
)

const (
	createToolsTableQuery = `
		DROP TABLE IF EXISTS tools;
		CREATE TABLE IF NOT EXISTS tools (
			id INTEGER NOT NULL,
			position TEXT NOT NULL,
			format BLOB NOT NULL,
			type TEXT NOT NULL,
			code TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'active',
			press INTEGER,
			notes BLOB NOT NULL,
			mods BLOB NOT NULL,
			PRIMARY KEY("id" AUTOINCREMENT)
		);
		INSERT INTO tools (position, format, type, code, status, press, notes, mods)
		VALUES
			('top', '{"width": 100, "height": 100}', 'MASS', 'G01', "active", 0, '[]', '[]'),
			('bottom', '{"width": 100, "height": 100}', 'MASS', 'G01', "active", 0, '[]', '[]');
	`

	selectAllToolsQuery = `
		SELECT id, position, format, type, code, status, press, notes, mods FROM tools;
	`

	selectToolByIDQuery = `
		SELECT id, position, format, type, code, status, press, notes, mods FROM tools WHERE id = $1;
	`

	selectToolsByPressQuery = `
		SELECT id, position, format, type, code, status, press, notes, mods FROM tools WHERE press = $1 AND status = 'active';
	`

	checkToolsExistenceQuery = `
		SELECT COUNT(*) FROM tools WHERE position = $1 AND format = $2 AND type = $3 AND code = $4
	`

	insertToolQuery = `
		INSERT INTO tools (position, format, type, code, status, press, notes, mods) VALUES ($1, $2, $3, $4, $5, $6, $7, $8);
	`

	updateToolQuery = `
		UPDATE tools SET position = $1, format = $2, type = $3, code = $4, status = $5, press = $6, notes = $7, mods = $8 WHERE id = $9;
	`

	deleteToolQuery = `
		DELETE FROM tools WHERE id = $1;
	`
)

// Tools represents a collection of tools in the database.
type Tools struct {
	db    *sql.DB
	feeds *Feeds
}

var _ DataOperations[*Tool] = (*Tools)(nil)

func NewTools(db *sql.DB, feeds *Feeds) *Tools {
	if _, err := db.Exec(createToolsTableQuery); err != nil {
		panic(
			NewDatabaseError(
				"create_table",
				"tools",
				"failed to create tools table",
				err,
			),
		)
	}

	return &Tools{
		db:    db,
		feeds: feeds,
	}
}

func (t *Tools) List() ([]*Tool, error) {
	logger.DBTools().Info("Listing tools")

	rows, err := t.db.Query(selectAllToolsQuery)
	if err != nil {
		return nil, NewDatabaseError("select", "tools",
			"failed to query tools", err)
	}
	defer rows.Close()

	var tools []*Tool

	for rows.Next() {
		tool, err := t.scanToolFromRows(rows)
		if err != nil {
			return nil, WrapError(err, "failed to scan tool")
		}
		tools = append(tools, tool)
	}

	if err := rows.Err(); err != nil {
		return nil, NewDatabaseError("select", "tools",
			"error iterating over rows", err)
	}

	return tools, nil
}

func (t *Tools) Get(id int64) (*Tool, error) {
	logger.DBTools().Info("Getting tool, id: %d", id)

	row := t.db.QueryRow(selectToolByIDQuery, id)

	tool, err := t.scanToolFromRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, NewDatabaseError("select", "tools",
			fmt.Sprintf("failed to get tool with ID %d", id), err)
	}

	return tool, nil
}

// GetByPress returns all active tools for a specific press (0-5)
func (t *Tools) GetByPress(pressNumber *PressNumber) ([]*Tool, error) {
	if pressNumber != nil && !(*pressNumber).IsValid() {
		return nil, fmt.Errorf("invalid press number: %d (must be 0-5)", *pressNumber)
	}

	if pressNumber == nil {
		logger.DBTools().Info("Getting inactive tools")
	} else {
		logger.DBTools().Info("Getting active tools for press: %d", *pressNumber)
	}

	rows, err := t.db.Query(selectToolsByPressQuery, pressNumber)
	if err != nil {
		return nil, NewDatabaseError("select", "tools",
			fmt.Sprintf("failed to query tools for press %d", pressNumber), err)
	}
	defer rows.Close()

	var tools []*Tool

	for rows.Next() {
		tool, err := t.scanToolFromRows(rows)
		if err != nil {
			return nil, WrapError(err, "failed to scan tool")
		}
		tools = append(tools, tool)
	}

	if err := rows.Err(); err != nil {
		return nil, NewDatabaseError("select", "tools",
			"error iterating over rows", err)
	}

	return tools, nil
}

func (t *Tools) Add(tool *Tool, user *User) (int64, error) {
	logger.DBTools().Info("Adding tool: %s", tool.String())

	// Marshal format for the existence check and insert
	formatBytes, err := json.Marshal(tool.Format)
	if err != nil {
		return 0, NewDatabaseError("insert", "tools",
			"failed to marshal format", err)
	}

	if err := t.exists(tool, formatBytes); err != nil {
		return 0, err
	}

	// Marshal other JSON fields
	notesBytes, err := json.Marshal(tool.LinkedNotes)
	if err != nil {
		return 0, NewDatabaseError("insert", "tools",
			"failed to marshal notes", err)
	}

	modsBytes, err := json.Marshal(tool.Mods)
	if err != nil {
		return 0, NewDatabaseError("insert", "tools",
			"failed to marshal mods", err)
	}

	t.updateMods(user, tool)

	result, err := t.db.Exec(insertToolQuery,
		tool.Position, formatBytes, tool.Type, tool.Code, string(tool.Status), tool.Press, notesBytes, modsBytes)
	if err != nil {
		return 0, NewDatabaseError("insert", "tools",
			"failed to insert tool", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, NewDatabaseError("insert", "tools",
			"failed to get last insert ID", err)
	}

	// Set tool ID for return
	tool.ID = id

	// Trigger feed update
	if t.feeds != nil {
		t.feeds.Add(NewFeed(
			FeedTypeToolAdd,
			&FeedToolAdd{
				ID:         id,
				Tool:       tool.String(),
				ModifiedBy: user,
			},
		))
	}

	return id, nil
}

func (t *Tools) Update(tool *Tool, user *User) error {
	logger.DBTools().Info("Updating tool: %d", tool.ID)

	// Marshal format for the existence check and update
	formatBytes, err := json.Marshal(tool.Format)
	if err != nil {
		return NewDatabaseError("update", "tools",
			"failed to marshal format for existence check", err)
	}

	if err := t.exists(tool, formatBytes); err != nil {
		return err
	}

	// Marshal other JSON fields
	notesBytes, err := json.Marshal(tool.LinkedNotes)
	if err != nil {
		return NewDatabaseError("update", "tools",
			"failed to marshal notes", err)
	}

	modsBytes, err := json.Marshal(tool.Mods)
	if err != nil {
		return NewDatabaseError("update", "tools",
			"failed to marshal mods", err)
	}

	t.updateMods(user, tool)

	_, err = t.db.Exec(updateToolQuery,
		tool.Position, formatBytes, tool.Type, tool.Code, string(tool.Status), tool.Press, notesBytes, modsBytes, tool.ID)
	if err != nil {
		return NewDatabaseError("update", "tools",
			fmt.Sprintf("failed to update tool with ID %d", tool.ID), err)
	}

	// Trigger feed update
	if t.feeds != nil {
		t.feeds.Add(NewFeed(
			FeedTypeToolUpdate,
			&FeedToolUpdate{
				ID:         tool.ID,
				Tool:       tool.String(),
				ModifiedBy: user,
			},
		))
	}

	return nil
}

func (t *Tools) Delete(id int64, user *User) error {
	logger.DBTools().Info("Deleting tool: %d", id)

	// Get tool info before deletion for feed
	tool, err := t.Get(id)
	if err != nil {
		return WrapError(err, "failed to get tool before deletion")
	}

	_, err = t.db.Exec(deleteToolQuery, id)
	if err != nil {
		return NewDatabaseError("delete", "tools",
			fmt.Sprintf("failed to delete tool with ID %d", id), err)
	}

	// Trigger feed update
	if t.feeds != nil {
		t.feeds.Add(NewFeed(
			FeedTypeToolDelete,
			&FeedToolDelete{
				ID:         id,
				Tool:       tool.String(),
				ModifiedBy: user,
			},
		))
	}

	return nil
}

// UpdateStatus updates only the status field of a tool
func (t *Tools) UpdateStatus(toolID int64, status ToolStatus, user *User) error {
	logger.DBTools().Info("Updating tool status: %d to %s", toolID, status)

	// Get current tool to track changes
	tool, err := t.Get(toolID)
	if err != nil {
		return fmt.Errorf("failed to get tool for status update: %w", err)
	}

	if tool.Status == status {
		return nil
	}

	// Update tool
	tool.Status = status

	// Update mods
	t.updateMods(user, tool)

	// Marshal mods for database update
	modsBytes, err := json.Marshal(tool.Mods)
	if err != nil {
		return NewDatabaseError("update", "tools",
			"failed to marshal mods", err)
	}

	query := `UPDATE tools SET status = ?, mods = ? WHERE id = ?`
	_, err = t.db.Exec(query, tool.Status, modsBytes, tool.ID)
	if err != nil {
		return NewDatabaseError("update", "tools",
			fmt.Sprintf("failed to update status for tool %d", tool.ID), err)
	}

	// Trigger feed update
	if t.feeds != nil {
		tool.Status = status // Update status for correct display
		t.feeds.Add(NewFeed(
			FeedTypeToolUpdate,
			&FeedToolUpdate{
				ID:         tool.ID,
				Tool:       tool.String(),
				ModifiedBy: user,
			},
		))
	}

	return nil
}

// UpdatePress updates only the press field of a tool
func (t *Tools) UpdatePress(toolID int64, press *PressNumber, user *User) error {
	logger.DBTools().Info("Updating tool press: %d", toolID)

	// Get current tool to track changes
	tool, err := t.Get(toolID)
	if err != nil {
		return fmt.Errorf("failed to get tool for press update: %w", err)
	}

	if tool.Press == press {
		return nil
	}

	// Update tool
	if err := tool.SetPress(press); err != nil {
		return fmt.Errorf("failed to set press for tool %d: %w", toolID, err)
	}

	// Update mods
	t.updateMods(user, tool)

	// Marshal mods for database update
	modsBytes, err := json.Marshal(tool.Mods)
	if err != nil {
		return NewDatabaseError("update", "tools",
			"failed to marshal mods", err)
	}

	query := `UPDATE tools SET press = ?, mods = ? WHERE id = ?`
	_, err = t.db.Exec(query, press, modsBytes, toolID)
	if err != nil {
		return NewDatabaseError("update", "tools",
			fmt.Sprintf("failed to update press for tool %d", toolID), err)
	}

	// Trigger feed update
	if t.feeds != nil {
		tool.Press = press // Update press for correct display
		t.feeds.Add(NewFeed(
			FeedTypeToolUpdate,
			&FeedToolUpdate{
				ID:         toolID,
				Tool:       tool.String(),
				ModifiedBy: user,
			},
		))
	}

	return nil
}

func (t *Tools) exists(tool *Tool, formatBytes []byte) error {
	var count int

	err := t.db.QueryRow(checkToolsExistenceQuery,
		tool.Position, formatBytes, tool.Type, tool.Code).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check for existing tool: %#v", err)
	}

	if count > 0 {
		return ErrAlreadyExists
	}

	return nil
}

func (t *Tools) scanToolFromRows(rows *sql.Rows) (*Tool, error) {
	tool := &Tool{}

	var (
		format      []byte
		linkedNotes []byte
		mods        []byte
	)

	var status string
	if err := rows.Scan(&tool.ID, &tool.Position, &format, &tool.Type,
		&tool.Code, &status, &tool.Press, &linkedNotes, &mods); err != nil {
		return nil, NewDatabaseError("scan", "tools",
			"failed to scan row", err)
	}

	tool.Status = ToolStatus(status)

	if err := json.Unmarshal(format, &tool.Format); err != nil {
		return nil, NewDatabaseError("scan", "tools",
			"failed to unmarshal format", err)
	}

	if err := json.Unmarshal(linkedNotes, &tool.LinkedNotes); err != nil {
		return nil, NewDatabaseError("scan", "tools",
			"failed to unmarshal notes", err)
	}

	if err := json.Unmarshal(mods, &tool.Mods); err != nil {
		return nil, WrapError(err, "failed to unmarshal mods data")
	}

	return tool, nil
}

func (t *Tools) scanToolFromRow(row *sql.Row) (*Tool, error) {
	tool := &Tool{}

	var (
		format      []byte
		linkedNotes []byte
		mods        []byte
	)

	var status string
	if err := row.Scan(&tool.ID, &tool.Position, &format, &tool.Type,
		&tool.Code, &status, &tool.Press, &linkedNotes, &mods); err != nil {
		return nil, err
	}

	tool.Status = ToolStatus(status)

	if err := json.Unmarshal(format, &tool.Format); err != nil {
		return nil, NewDatabaseError("scan", "tools",
			"failed to unmarshal format", err)
	}

	if err := json.Unmarshal(linkedNotes, &tool.LinkedNotes); err != nil {
		return nil, NewDatabaseError("scan", "tools",
			"failed to unmarshal notes", err)
	}

	if err := json.Unmarshal(mods, &tool.Mods); err != nil {
		return nil, WrapError(err, "failed to unmarshal mods data")
	}

	return tool, nil
}

func (t *Tools) updateMods(user *User, tool *Tool) {
	if user == nil {
		return
	}

	tool.Mods.Add(user, ToolMod{
		Position:    tool.Position,
		Format:      tool.Format,
		Type:        tool.Type,
		Code:        tool.Code,
		Status:      tool.Status,
		Press:       tool.Press,
		LinkedNotes: tool.LinkedNotes,
	})
}

// equalPressNumbers compares two press number pointers for equality
func equalPressNumbers(a, b *PressNumber) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
