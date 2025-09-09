package tool

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/knackwurstking/pgpress/internal/database/dberror"
	"github.com/knackwurstking/pgpress/internal/database/interfaces"
	feedmodels "github.com/knackwurstking/pgpress/internal/database/models/feed"
	notemodels "github.com/knackwurstking/pgpress/internal/database/models/note"
	toolmodels "github.com/knackwurstking/pgpress/internal/database/models/tool"
	usermodels "github.com/knackwurstking/pgpress/internal/database/models/user"
	"github.com/knackwurstking/pgpress/internal/database/services/feed"
	"github.com/knackwurstking/pgpress/internal/database/services/note"
	"github.com/knackwurstking/pgpress/internal/logger"
)

type Service struct {
	db    *sql.DB
	notes *note.Service
	feeds *feed.Service
}

var _ interfaces.DataOperations[*toolmodels.Tool] = (*Service)(nil)

func New(db *sql.DB, notes *note.Service, feeds *feed.Service) *Service {
	query := `
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
		notes: notes,
		feeds: feeds,
	}
}

func (t *Service) List() ([]*toolmodels.Tool, error) {
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

	var tools []*toolmodels.Tool

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

func (t *Service) Get(id int64) (*toolmodels.Tool, error) {
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

func (t *Service) Add(tool *toolmodels.Tool, user *usermodels.User) (int64, error) {
	logger.DBTools().Info("Adding tool: %s", tool.String())

	// Marshal format for the existence check and insert
	formatBytes, err := json.Marshal(tool.Format)
	if err != nil {
		return 0, dberror.NewDatabaseError("insert", "tools",
			"failed to marshal format", err)
	}

	// Vaidate tool
	var count int
	query := `
		SELECT COUNT(*) FROM tools WHERE position = $1 AND format = $2 AND type = $3 AND code = $4
	`
	err = t.db.QueryRow(query, tool.Position, formatBytes, tool.Type, tool.Code).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to check for existing tool: %#v", err)
	}

	if count > 0 {
		return 0, dberror.ErrAlreadyExists
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

	query = `
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
		feed := feedmodels.New(
			"Neues Werkzeug hinzugefügt",
			fmt.Sprintf("Benutzer %s hat ein neues Werkzeug %s zur Werkzeugliste hinzugefügt.", user.Name, tool.String()),
			user.TelegramID,
		)
		t.feeds.Add(feed)
	}

	return id, nil
}

func (t *Service) Update(tool *toolmodels.Tool, user *usermodels.User) error {
	logger.DBTools().Info("Updating tool: %d", tool.ID)

	// Marshal format for the existence check and update
	formatBytes, err := json.Marshal(tool.Format)
	if err != nil {
		return dberror.NewDatabaseError("update", "tools",
			"failed to marshal format for existence check", err)
	}

	// Validate tool
	var count int
	query := `
		SELECT COUNT(*) FROM tools WHERE id != $1 AND position = $2 AND format = $3 AND type = $4 AND code = $5
	`
	err = t.db.QueryRow(query, tool.ID, tool.Position, formatBytes, tool.Type, tool.Code).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check for existing tool: %#v", err)
	}

	if count > 0 {
		return dberror.ErrAlreadyExists
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

	query = `
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
		feed := feedmodels.New(
			"Werkzeug aktualisiert",
			fmt.Sprintf("Benutzer %s hat das Werkzeug %s aktualisiert.", user.Name, tool.String()),
			user.TelegramID,
		)
		t.feeds.Add(feed)
	}

	return nil
}

func (t *Service) Delete(id int64, user *usermodels.User) error {
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
		feed := feedmodels.New(
			"Werkzeug entfernt",
			fmt.Sprintf("Benutzer %s hat das Werkzeug %s entfernt.", user.Name, tool.String()),
			user.TelegramID,
		)
		t.feeds.Add(feed)
	}

	return nil
}

func (s *Service) GetWithNotes(id int64) (*toolmodels.ToolWithNotes, error) {
	logger.DBTools().Debug(
		"Getting tools with notes, id: %d", id)

	// Get the tool
	tool, err := s.Get(id)
	if err != nil {
		return nil, err
	}

	// Load notes
	notes, err := s.notes.GetByIDs(tool.LinkedNotes)
	if err != nil {
		return nil, dberror.WrapError(err, "failed to load attachments for trouble report")
	}

	return &toolmodels.ToolWithNotes{
		Tool:        tool,
		LoadedNotes: notes,
	}, nil
}

// ListWithNotes retrieves all tools and loads their associated notes.
func (s *Service) ListWithNotes() ([]*toolmodels.ToolWithNotes, error) {
	logger.DBTools().Debug("Listing tools with notes")

	// Get all tools
	tools, err := s.List()
	if err != nil {
		return nil, err
	}

	var result []*toolmodels.ToolWithNotes

	for _, tool := range tools {
		// Load notes for each tool
		notes, err := s.notes.GetByIDs(tool.LinkedNotes)
		if err != nil {
			return nil, dberror.WrapError(err,
				fmt.Sprintf("failed to load notes for tool %d", tool.ID))
		}

		result = append(result, &toolmodels.ToolWithNotes{
			Tool:        tool,
			LoadedNotes: notes,
		})
	}

	return result, nil
}

// AddWithNotes creates a new tool and its associated notes in a single transaction.
func (s *Service) AddWithNotes(tool *toolmodels.Tool, user *usermodels.User, notes ...*notemodels.Note) (*toolmodels.ToolWithNotes, error) {
	logger.DBTools().Debug("Adding tool with notes")

	// First, add all notes and collect their IDs
	var noteIDs []int64
	for _, note := range notes {
		noteID, err := s.notes.Add(note, user)
		if err != nil {
			return nil, dberror.WrapError(err, "failed to add note")
		}
		noteIDs = append(noteIDs, noteID)
	}

	// Set the note IDs in the tool
	tool.LinkedNotes = noteIDs

	// Add the tool
	toolID, err := s.Add(tool, user)
	if err != nil {
		if err == dberror.ErrAlreadyExists {
			return nil, err
		}
		return nil, dberror.WrapError(err, "failed to add tool")
	}

	// Set the tool ID
	tool.ID = toolID

	// Return the tool with loaded notes
	return &toolmodels.ToolWithNotes{
		Tool:        tool,
		LoadedNotes: notes,
	}, nil
}

// GetByPress returns all active tools for a specific press (0-5).
func (s *Service) GetByPress(pressNumber *toolmodels.PressNumber) ([]*toolmodels.Tool, error) {
	if !toolmodels.IsValidPressNumber(pressNumber) {
		return nil, fmt.Errorf("invalid press number: %d (must be 0-5)", *pressNumber)
	}

	if pressNumber == nil {
		logger.DBTools().Info("Getting inactive tools")
	} else {
		logger.DBTools().Info("Getting active tools for press: %d", *pressNumber)
	}

	const query = `
		SELECT id, position, format, type, code, regenerating, press, notes, mods FROM tools WHERE press = $1 AND regenerating = 0;
	`
	rows, err := s.db.Query(query, pressNumber)
	if err != nil {
		return nil, dberror.NewDatabaseError("select", "tools",
			fmt.Sprintf("failed to query tools for press %v", pressNumber), err)
	}
	defer rows.Close()

	var tools []*toolmodels.Tool

	for rows.Next() {
		tool, err := s.scanTool(rows)
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

// UpdateRegenerating updates only the regenerating field of a tool.
func (s *Service) UpdateRegenerating(toolID int64, regenerating bool, user *usermodels.User) error {
	logger.DBTools().Info("Updating tool regenerating status: %d to %v", toolID, regenerating)

	// Get current tool to track changes
	tool, err := s.Get(toolID)
	if err != nil {
		return fmt.Errorf("failed to get tool for regenerating status update: %w", err)
	}

	if tool.Regenerating == regenerating {
		return nil
	}

	// Update tool
	tool.Regenerating = regenerating

	// Update mods
	s.updateMods(user, tool)

	// Marshal mods for database update
	modsBytes, err := json.Marshal(tool.Mods)
	if err != nil {
		return dberror.NewDatabaseError("update", "tools",
			"failed to marshal mods", err)
	}

	const query = `UPDATE tools SET regenerating = ?, mods = ? WHERE id = ?`
	_, err = s.db.Exec(query, tool.Regenerating, modsBytes, tool.ID)
	if err != nil {
		return dberror.NewDatabaseError("update", "tools",
			fmt.Sprintf("failed to update regenerating for tool %d", tool.ID), err)
	}

	// Trigger feed update
	if s.feeds != nil {
		feed := feedmodels.New(
			"Werkzeug Regenerierung aktualisiert",
			fmt.Sprintf("Benutzer %s hat die Regenerierung für Werkzeug %s aktualisiert.", user.Name, tool.String()),
			user.TelegramID,
		)
		s.feeds.Add(feed)
	}

	return nil
}

// UpdatePress updates only the press field of a tool.
func (s *Service) UpdatePress(toolID int64, press *toolmodels.PressNumber, user *usermodels.User) error {
	logger.DBTools().Info("Updating tool press: %d", toolID)

	// Get current tool to track changes
	tool, err := s.Get(toolID)
	if err != nil {
		return fmt.Errorf("failed to get tool for press update: %w", err)
	}

	if equalPressNumbers(tool.Press, press) {
		return nil
	}

	// Update tool
	if err := tool.SetPress(press); err != nil {
		return fmt.Errorf("failed to set press for tool %d: %w", toolID, err)
	}

	// Update mods
	s.updateMods(user, tool)

	// Marshal mods for database update
	modsBytes, err := json.Marshal(tool.Mods)
	if err != nil {
		return dberror.NewDatabaseError("update", "tools",
			"failed to marshal mods", err)
	}

	const query = `UPDATE tools SET press = ?, mods = ? WHERE id = ?`
	_, err = s.db.Exec(query, press, modsBytes, toolID)
	if err != nil {
		return dberror.NewDatabaseError("update", "tools",
			fmt.Sprintf("failed to update press for tool %d", toolID), err)
	}

	// Trigger feed update
	if s.feeds != nil {
		tool.Press = press // Update press for correct display
		feed := feedmodels.New(
			"Werkzeug Pressendaten aktualisiert",
			fmt.Sprintf("Benutzer %s hat die Pressendaten für Werkzeug %s aktualisiert.", user.Name, tool.String()),
			user.TelegramID,
		)
		s.feeds.Add(feed)
	}

	return nil
}

func (t *Service) scanTool(scanner interfaces.Scannable) (*toolmodels.Tool, error) {
	tool := &toolmodels.Tool{}

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

func (t *Service) updateMods(user *usermodels.User, tool *toolmodels.Tool) {
	if user == nil {
		return
	}

	tool.Mods.Add(user, toolmodels.ToolMod{
		Position:     tool.Position,
		Format:       tool.Format,
		Type:         tool.Type,
		Code:         tool.Code,
		Regenerating: tool.Regenerating,
		Press:        tool.Press,
		LinkedNotes:  tool.LinkedNotes,
	})
}

// equalPressNumbers compares two press number pointers for equality
func equalPressNumbers(a, b *toolmodels.PressNumber) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
