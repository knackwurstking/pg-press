package services

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/knackwurstking/pgpress/internal/interfaces"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"
)

type Tool struct {
	db    *sql.DB
	notes *Note
	feeds *Feed
}

func NewTool(db *sql.DB, notes *Note, feeds *Feed) *Tool {
	const createTableQuery = `
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

	if _, err := db.Exec(createTableQuery); err != nil {
		panic(fmt.Errorf("failed to create tools table: %w", err))
	}

	logger.DBTools().Info("Tool service initialized")
	return &Tool{db: db, notes: notes, feeds: feeds}
}

func (t *Tool) Add(tool *models.Tool, user *models.User) (int64, error) {
	logger.DBTools().Info("Adding new tool: %s (user: %s)", tool.String(), user.Name)

	if err := t.validateToolUniqueness(tool, 0); err != nil {
		logger.DBTools().Warn("Tool validation failed: %v", err)
		return 0, err
	}

	formatBytes, notesBytes, modsBytes, err := t.marshalToolData(tool, user)
	if err != nil {
		logger.DBTools().Error("Failed to marshal tool data: %v", err)
		return 0, err
	}

	const insertQuery = `
		INSERT INTO tools (position, format, type, code, regenerating, press, notes, mods)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	result, err := t.db.Exec(insertQuery, tool.Position, formatBytes, tool.Type, tool.Code,
		tool.Regenerating, tool.Press, notesBytes, modsBytes)
	if err != nil {
		logger.DBTools().Error("Failed to insert tool: %v", err)
		return 0, utils.NewDatabaseError("insert", "tools", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		logger.DBTools().Error("Failed to get last insert ID: %v", err)
		return 0, utils.NewDatabaseError("insert", "tools", err)
	}

	tool.ID = id
	t.createFeedUpdate("Neues Werkzeug hinzugefügt",
		fmt.Sprintf("Benutzer %s hat ein neues Werkzeug %s zur Werkzeugliste hinzugefügt.", user.Name, tool.String()), user)

	logger.DBTools().Info("Successfully added tool with ID: %d", id)
	return id, nil
}

func (t *Tool) AddWithNotes(tool *models.Tool, user *models.User, notes ...*models.Note) (*models.ToolWithNotes, error) {
	logger.DBTools().Debug("Adding tool with %d notes (user: %s)", len(notes), user.Name)

	var noteIDs []int64
	for _, note := range notes {
		noteID, err := t.notes.Add(note, user)
		if err != nil {
			return nil, err
		}
		noteIDs = append(noteIDs, noteID)
	}

	tool.LinkedNotes = noteIDs
	toolID, err := t.Add(tool, user)
	if err != nil {
		return nil, err
	}

	tool.ID = toolID
	logger.DBTools().Debug("Successfully added tool with notes, ID: %d", toolID)
	return &models.ToolWithNotes{Tool: tool, LoadedNotes: notes}, nil
}

func (t *Tool) Delete(id int64, user *models.User) error {
	logger.DBTools().Info("Deleting tool ID %d (user: %s)", id, user.Name)

	tool, err := t.Get(id)
	if err != nil {
		return err
	}

	const deleteQuery = `DELETE FROM tools WHERE id = $1`
	_, err = t.db.Exec(deleteQuery, id)
	if err != nil {
		logger.DBTools().Error("Failed to delete tool %d: %v", id, err)
		return utils.NewDatabaseError("delete", "tools", err)
	}

	t.createFeedUpdate(
		"Werkzeug entfernt",
		fmt.Sprintf(
			"Benutzer %s hat das Werkzeug %s entfernt.",
			user.Name, tool.String(),
		),
		user,
	)

	logger.DBTools().Info("Successfully deleted tool ID: %d", id)
	return nil
}

func (t *Tool) Get(id int64) (*models.Tool, error) {
	logger.DBTools().Debug("Getting tool with ID: %d", id)

	const query = `SELECT id, position, format, type, code, regenerating, press, notes, mods FROM tools WHERE id = $1`
	row := t.db.QueryRow(query, id)

	tool, err := t.scanTool(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError(fmt.Sprintf("tool with ID %d not found", id))
		}
		return nil, utils.NewDatabaseError("select", "tools", err)
	}

	logger.DBTools().Debug("Successfully retrieved tool: %s", tool.String())
	return tool, nil
}

func (t *Tool) GetActiveToolsForPress(pressNumber models.PressNumber) []*models.Tool {
	const query = `
		SELECT id, position, format, type, code, regenerating, press, notes, mods
		FROM tools WHERE regenerating = 0 AND press = ?
	`
	rows, err := t.db.Query(query, pressNumber)
	if err != nil {
		logger.DBTools().Error("Failed to query active tools: %v", err)
		return nil
	}
	defer rows.Close()

	var tools []*models.Tool
	for rows.Next() {
		tool, err := t.scanTool(rows)
		if err != nil {
			logger.DBTools().Warn("Failed to scan tool: %v", err)
			return nil
		}
		tools = append(tools, tool)
	}

	if err := rows.Err(); err != nil {
		logger.DBTools().Error("Error iterating over tool rows: %v", err)
		return nil
	}
	return tools
}

func (t *Tool) GetByPress(pressNumber *models.PressNumber) ([]*models.Tool, error) {
	if pressNumber != nil && !models.IsValidPressNumber(pressNumber) {
		return nil, fmt.Errorf("invalid press number: %d (must be 0-5)", *pressNumber)
	}

	if pressNumber == nil {
		logger.DBTools().Debug("Getting inactive tools")
	} else {
		logger.DBTools().Debug("Getting active tools for press: %d", *pressNumber)
	}

	const query = `
		SELECT id, position, format, type, code, regenerating, press, notes, mods
		FROM tools WHERE press = $1 AND regenerating = 0
	`
	rows, err := t.db.Query(query, pressNumber)
	if err != nil {
		logger.DBTools().Error("Failed to query tools for press %v: %v", pressNumber, err)
		return nil, utils.NewDatabaseError("select", "tools", err)
	}
	defer rows.Close()

	var tools []*models.Tool
	for rows.Next() {
		tool, err := t.scanTool(rows)
		if err != nil {
			return nil, utils.NewDatabaseError("scan", "tools", err)
		}
		tools = append(tools, tool)
	}

	if err := rows.Err(); err != nil {
		logger.DBTools().Error("Error iterating over tool rows: %v", err)
		return nil, utils.NewDatabaseError("select", "tools", err)
	}

	logger.DBTools().Debug("Found %d tools for press %v", len(tools), pressNumber)
	return tools, nil
}

// GetPressUtilization returns a complete overview of press utilization across all presses
func (t *Tool) GetPressUtilization() ([]models.PressUtilization, error) {
	logger.DBTools().Info("Generating press utilization map")

	var utilization []models.PressUtilization

	// Valid press numbers: 0, 2, 3, 4, 5
	validPresses := []models.PressNumber{0, 2, 3, 4, 5}

	for _, pressNum := range validPresses {
		tools := t.GetActiveToolsForPress(pressNum)
		count := len(tools)

		utilization = append(utilization, models.PressUtilization{
			PressNumber: pressNum,
			Tools:       tools,
			Count:       count,
			Available:   count == 0,
		})
	}

	return utilization, nil
}

func (t *Tool) GetWithNotes(id int64) (*models.ToolWithNotes, error) {
	logger.DBTools().Debug("Getting tool with notes, ID: %d", id)

	tool, err := t.Get(id)
	if err != nil {
		return nil, err
	}

	notes, err := t.notes.GetByIDs(tool.LinkedNotes)
	if err != nil {
		return nil, fmt.Errorf("failed to load notes for tool")
	}

	logger.DBTools().Debug("Successfully retrieved tool with %d notes", len(notes))
	return &models.ToolWithNotes{Tool: tool, LoadedNotes: notes}, nil
}

func (t *Tool) List() ([]*models.Tool, error) {
	logger.DBTools().Debug("Listing all tools")

	const query = `SELECT id, position, format, type, code, regenerating, press, notes, mods FROM tools`
	rows, err := t.db.Query(query)
	if err != nil {
		return nil, utils.NewDatabaseError("select", "tools", err)
	}
	defer rows.Close()

	var tools []*models.Tool
	for rows.Next() {
		tool, err := t.scanTool(rows)
		if err != nil {
			return nil, utils.NewDatabaseError("scan", "tools", err)
		}
		tools = append(tools, tool)
	}

	if err := rows.Err(); err != nil {
		logger.DBTools().Error("Error iterating over tool rows: %v", err)
		return nil, utils.NewDatabaseError("select", "tools", err)
	}

	logger.DBTools().Debug("Successfully listed %d tools", len(tools))
	return tools, nil
}

func (t *Tool) ListWithNotes() ([]*models.ToolWithNotes, error) {
	logger.DBTools().Debug("Listing all tools with notes")

	tools, err := t.List()
	if err != nil {
		return nil, err
	}

	var result []*models.ToolWithNotes
	for _, tool := range tools {
		notes, err := t.notes.GetByIDs(tool.LinkedNotes)
		if err != nil {
			return nil, fmt.Errorf("failed to load notes for tool %d: %w", tool.ID, err)
		}

		result = append(result, &models.ToolWithNotes{Tool: tool, LoadedNotes: notes})
	}

	logger.DBTools().Debug("Successfully listed %d tools with notes", len(result))
	return result, nil
}

func (t *Tool) Update(tool *models.Tool, user *models.User) error {
	logger.DBTools().Info("Updating tool ID %d: %s (user: %s)", tool.ID, tool.String(), user.Name)

	if err := t.validateToolUniqueness(tool, tool.ID); err != nil {
		logger.DBTools().Warn("Tool update validation failed: %v", err)
		return err
	}

	formatBytes, notesBytes, modsBytes, err := t.marshalToolData(tool, user)
	if err != nil {
		logger.DBTools().Error("Failed to marshal tool data for update: %v", err)
		return err
	}

	const updateQuery = `
		UPDATE tools SET position = $1, format = $2, type = $3, code = $4,
		regenerating = $5, press = $6, notes = $7, mods = $8 WHERE id = $9
	`
	_, err = t.db.Exec(updateQuery, tool.Position, formatBytes, tool.Type, tool.Code,
		tool.Regenerating, tool.Press, notesBytes, modsBytes, tool.ID)
	if err != nil {
		logger.DBTools().Error("Failed to update tool %d: %v", tool.ID, err)
		return utils.NewDatabaseError("update", "tools", err)
	}

	t.createFeedUpdate("Werkzeug aktualisiert",
		fmt.Sprintf("Benutzer %s hat das Werkzeug %s aktualisiert.", user.Name, tool.String()), user)

	logger.DBTools().Info("Successfully updated tool ID: %d", tool.ID)
	return nil
}

func (t *Tool) UpdatePress(toolID int64, press *models.PressNumber, user *models.User) error {
	logger.DBTools().Info("Updating press assignment for tool %d (user: %s)", toolID, user.Name)

	tool, err := t.Get(toolID)
	if err != nil {
		return fmt.Errorf("failed to get tool for press update: %w", err)
	}

	if equalPressNumbers(tool.Press, press) {
		logger.DBTools().Debug("Tool %d press assignment unchanged", toolID)
		return nil
	}

	if err := tool.SetPress(press); err != nil {
		return fmt.Errorf("failed to set press for tool %d: %w", toolID, err)
	}

	t.updateMods(user, tool)

	modsBytes, err := json.Marshal(tool.Mods)
	if err != nil {
		logger.DBTools().Error("Failed to marshal mods: %v", err)
		return utils.NewDatabaseError("update", "tools", err)
	}

	const query = `UPDATE tools SET press = ?, mods = ? WHERE id = ?`
	_, err = t.db.Exec(query, press, modsBytes, toolID)
	if err != nil {
		logger.DBTools().Error("Failed to update press assignment: %v", err)
		return utils.NewDatabaseError("update", "tools", err)
	}

	tool.Press = press
	t.createFeedUpdate("Werkzeug Pressendaten aktualisiert",
		fmt.Sprintf("Benutzer %s hat die Pressendaten für Werkzeug %s aktualisiert.", user.Name, tool.String()), user)

	logger.DBTools().Info("Successfully updated press assignment for tool %d", toolID)
	return nil
}

func (t *Tool) UpdateRegenerating(toolID int64, regenerating bool, user *models.User) error {
	logger.DBTools().Info("Updating regenerating status for tool %d to %v (user: %s)", toolID, regenerating, user.Name)

	tool, err := t.Get(toolID)
	if err != nil {
		return fmt.Errorf("failed to get tool for regenerating status update: %w", err)
	}

	if tool.Regenerating == regenerating {
		logger.DBTools().Debug("Tool %d regenerating status unchanged: %v", toolID, regenerating)
		return nil
	}

	tool.Regenerating = regenerating
	t.updateMods(user, tool)

	modsBytes, err := json.Marshal(tool.Mods)
	if err != nil {
		logger.DBTools().Error("Failed to marshal mods: %v", err)
		return utils.NewDatabaseError("update", "tools", err)
	}

	const query = `UPDATE tools SET regenerating = ?, mods = ? WHERE id = ?`
	_, err = t.db.Exec(query, tool.Regenerating, modsBytes, tool.ID)
	if err != nil {
		logger.DBTools().Error("Failed to update regenerating status: %v", err)
		return utils.NewDatabaseError("update", "tools", err)
	}

	t.createFeedUpdate("Werkzeug Regenerierung aktualisiert",
		fmt.Sprintf("Benutzer %s hat die Regenerierung für Werkzeug %s aktualisiert.", user.Name, tool.String()), user)

	logger.DBTools().Info("Successfully updated regenerating status for tool %d", toolID)
	return nil
}

func (t *Tool) createFeedUpdate(title, message string, user *models.User) {
	if t.feeds != nil {
		feed := models.NewFeed(title, message, user.TelegramID)
		if err := t.feeds.Add(feed); err != nil {
			logger.DBTools().Warn("Failed to create feed update: %v", err)
		}
	}
}

func (t *Tool) marshalToolData(tool *models.Tool, user *models.User) ([]byte, []byte, []byte, error) {
	t.updateMods(user, tool)

	formatBytes, err := json.Marshal(tool.Format)
	if err != nil {
		return nil, nil, nil, utils.NewDatabaseError("marshal", "tools", err)
	}

	notesBytes, err := json.Marshal(tool.LinkedNotes)
	if err != nil {
		return nil, nil, nil, utils.NewDatabaseError("marshal", "tools", err)
	}

	modsBytes, err := json.Marshal(tool.Mods)
	if err != nil {
		return nil, nil, nil, utils.NewDatabaseError("marshal", "tools", err)
	}

	return formatBytes, notesBytes, modsBytes, nil
}

func (t *Tool) scanTool(scanner interfaces.Scannable) (*models.Tool, error) {
	tool := &models.Tool{}
	var format, linkedNotes, mods []byte

	if err := scanner.Scan(&tool.ID, &tool.Position, &format, &tool.Type,
		&tool.Code, &tool.Regenerating, &tool.Press, &linkedNotes, &mods); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(format, &tool.Format); err != nil {
		return nil, fmt.Errorf("failed to unmarshal format data")
	}

	if err := json.Unmarshal(linkedNotes, &tool.LinkedNotes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal linked notes data")
	}

	if err := json.Unmarshal(mods, &tool.Mods); err != nil {
		return nil, fmt.Errorf("failed to unmarshal mods data")
	}

	return tool, nil
}

func (t *Tool) updateMods(user *models.User, tool *models.Tool) {
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

func (t *Tool) validateToolUniqueness(tool *models.Tool, excludeID int64) error {
	formatBytes, err := json.Marshal(tool.Format)
	if err != nil {
		return fmt.Errorf("failed to marshal format data: %w", err)
	}

	var count int
	const query = `
		SELECT COUNT(*) FROM tools
		WHERE id != $1 AND position = $2 AND format = $3 AND type = $4 AND code = $5
	`
	err = t.db.QueryRow(query, excludeID, tool.Position, formatBytes, tool.Type, tool.Code).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check for existing tool: %w", err)
	}

	if count > 0 {
		return utils.NewAlreadyExistsError(
			fmt.Sprintf(
				"tool with position %d, format %s, type %s, and code %s already exists",
				tool.Position, tool.Format, tool.Type, tool.Code,
			),
		)
	}

	return nil
}

func equalPressNumbers(a, b *models.PressNumber) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
