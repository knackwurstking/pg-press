package tool

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

var _ interfaces.DataOperations[*models.Tool] = (*Service)(nil)

func New(db *sql.DB, feeds *feed.Service) *Service {
	query := `
		DROP TABLE IF EXISTS tools;
		CREATE TABLE IF NOT EXISTS tools (
			id INTEGER NOT NULL,
			position TEXT NOT NULL,
			format BLOB NOT NULL,
			type TEXT NOT NULL,
			code TEXT NOT NULL,
			regenerating BOOLEAN NOT NULL DEFAULT 0,
			press INTEGER,
			notes BLOB NOT NULL,
			mods BLOB NOT NULL,
			PRIMARY KEY("id" AUTOINCREMENT)
		);
	`
	if _, err := db.Exec(query); err != nil {
		panic(
			dberror.NewDatabaseError(
				"create_table",
				"tools",
				"failed to create tools table",
				err,
			),
		)
	}

	return &Service{
		db:    db,
		feeds: feeds,
	}
}

func (t *Service) List() ([]*models.Tool, error) {
	logger.DBTools().Info("Listing tools")

	query := `
		SELECT id, position, format, type, code, regenerating, press, notes, mods FROM tools;
	`
	rows, err := t.db.Query(query)
	if err != nil {
		return nil, dberror.NewDatabaseError("select", "tools",
			"failed to query tools", err)
	}
	defer rows.Close()

	var tools []*models.Tool

	for rows.Next() {
		tool, err := t.scanTool(rows)
		if err != nil {
			return nil, dberror.WrapError(err, "failed to scan tool")
		}

		tools = append(tools, tool)
	}

	if err := rows.Err(); err != nil {
		return nil, dberror.NewDatabaseError("select", "tools",
			"error iterating over rows", err)
	}

	return tools, nil
}

func (t *Service) Get(id int64) (*models.Tool, error) {
	logger.DBTools().Info("Getting tool, id: %d", id)

	query := `
		SELECT id, position, format, type, code, regenerating, press, notes, mods FROM tools WHERE id = $1;
	`
	row := t.db.QueryRow(query, id)

	tool, err := t.scanTool(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, dberror.ErrNotFound
		}
		return nil, dberror.NewDatabaseError("select", "tools",
			fmt.Sprintf("failed to get tool with ID %d", id), err)
	}

	return tool, nil
}

func (t *Service) Add(tool *models.Tool, user *models.User) (int64, error) {
	logger.DBTools().Info("Adding tool: %s", tool.String())

	// Marshal format for the existence check and insert
	formatBytes, err := json.Marshal(tool.Format)
	if err != nil {
		return 0, dberror.NewDatabaseError("insert", "tools",
			"failed to marshal format", err)
	}

	if err := t.exists(tool, formatBytes); err != nil {
		return 0, err
	}

	// Marshal other JSON fields
	notesBytes, err := json.Marshal(tool.LinkedNotes)
	if err != nil {
		return 0, dberror.NewDatabaseError("insert", "tools",
			"failed to marshal notes", err)
	}

	modsBytes, err := json.Marshal(tool.Mods)
	if err != nil {
		return 0, dberror.NewDatabaseError("insert", "tools",
			"failed to marshal mods", err)
	}

	t.updateMods(user, tool)

	query := `
		INSERT INTO tools (position, format, type, code, regenerating, press, notes, mods) VALUES ($1, $2, $3, $4, $5, $6, $7, $8);
	`
	result, err := t.db.Exec(query,
		tool.Position, formatBytes, tool.Type, tool.Code, tool.Regenerating, tool.Press, notesBytes, modsBytes)
	if err != nil {
		return 0, dberror.NewDatabaseError("insert", "tools",
			"failed to insert tool", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, dberror.NewDatabaseError("insert", "tools",
			"failed to get last insert ID", err)
	}

	// Set tool ID for return
	tool.ID = id

	// Trigger feed update
	if t.feeds != nil {
		t.feeds.Add(models.NewFeed(
			models.FeedTypeToolAdd,
			&models.FeedToolAdd{
				ID:         id,
				Tool:       tool.String(),
				ModifiedBy: user,
			},
		))
	}

	return id, nil
}

func (t *Service) Update(tool *models.Tool, user *models.User) error {
	logger.DBTools().Info("Updating tool: %d", tool.ID)

	// Marshal format for the existence check and update
	formatBytes, err := json.Marshal(tool.Format)
	if err != nil {
		return dberror.NewDatabaseError("update", "tools",
			"failed to marshal format for existence check", err)
	}

	if err := t.exists(tool, formatBytes); err != nil {
		return err
	}

	// Marshal other JSON fields
	notesBytes, err := json.Marshal(tool.LinkedNotes)
	if err != nil {
		return dberror.NewDatabaseError("update", "tools",
			"failed to marshal notes", err)
	}

	modsBytes, err := json.Marshal(tool.Mods)
	if err != nil {
		return dberror.NewDatabaseError("update", "tools",
			"failed to marshal mods", err)
	}

	t.updateMods(user, tool)

	query := `
		UPDATE tools SET position = $1, format = $2, type = $3, code = $4, regenerating = $5, press = $6, notes = $7, mods = $8 WHERE id = $9;
	`
	_, err = t.db.Exec(query,
		tool.Position, formatBytes, tool.Type, tool.Code, tool.Regenerating, tool.Press, notesBytes, modsBytes, tool.ID)
	if err != nil {
		return dberror.NewDatabaseError("update", "tools",
			fmt.Sprintf("failed to update tool with ID %d", tool.ID), err)
	}

	// Trigger feed update
	if t.feeds != nil {
		t.feeds.Add(models.NewFeed(
			models.FeedTypeToolUpdate,
			&models.FeedToolUpdate{
				ID:         tool.ID,
				Tool:       tool.String(),
				ModifiedBy: user,
			},
		))
	}

	return nil
}

func (t *Service) Delete(id int64, user *models.User) error {
	logger.DBTools().Info("Deleting tool: %d", id)

	// Get tool info before deletion for feed
	tool, err := t.Get(id)
	if err != nil {
		return dberror.WrapError(err, "failed to get tool before deletion")
	}

	query := `
		DELETE FROM tools WHERE id = $1;
	`
	_, err = t.db.Exec(query, id)
	if err != nil {
		return dberror.NewDatabaseError("delete", "tools",
			fmt.Sprintf("failed to delete tool with ID %d", id), err)
	}

	// Trigger feed update
	if t.feeds != nil {
		t.feeds.Add(models.NewFeed(
			models.FeedTypeToolDelete,
			&models.FeedToolDelete{
				ID:         id,
				Tool:       tool.String(),
				ModifiedBy: user,
			},
		))
	}

	return nil
}

func (t *Service) exists(tool *models.Tool, formatBytes []byte) error {
	var count int

	query := `
		SELECT COUNT(*) FROM tools WHERE position = $1 AND format = $2 AND type = $3 AND code = $4
	`
	err := t.db.QueryRow(query,
		tool.Position, formatBytes, tool.Type, tool.Code).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check for existing tool: %#v", err)
	}

	if count > 0 {
		return dberror.ErrAlreadyExists
	}

	return nil
}

func (t *Service) scanTool(scanner interfaces.Scannable) (*models.Tool, error) {
	tool := &models.Tool{}

	var (
		format      []byte
		linkedNotes []byte
		mods        []byte
	)

	if err := scanner.Scan(&tool.ID, &tool.Position, &format, &tool.Type,
		&tool.Code, &tool.Regenerating, &tool.Press, &linkedNotes, &mods); err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, dberror.NewDatabaseError("scan", "tools",
			"failed to scan row", err)
	}

	if err := json.Unmarshal(format, &tool.Format); err != nil {
		return nil, dberror.NewDatabaseError("scan", "tools",
			"failed to unmarshal format", err)
	}

	if err := json.Unmarshal(linkedNotes, &tool.LinkedNotes); err != nil {
		return nil, dberror.NewDatabaseError("scan", "tools",
			"failed to unmarshal notes", err)
	}

	if err := json.Unmarshal(mods, &tool.Mods); err != nil {
		return nil, dberror.WrapError(err, "failed to unmarshal mods data")
	}

	return tool, nil
}

func (t *Service) updateMods(user *models.User, tool *models.Tool) {
	if user == nil {
		return
	}

	tool.Mods.Add(user, models.ToolMod{
		Position:     tool.Position,
		Format:       tool.Format,
		Type:         tool.Type,
		Code:         tool.Code,
		Regenerating: tool.Regenerating,
		Press:        tool.Press,
		LinkedNotes:  tool.LinkedNotes,
	})
}
