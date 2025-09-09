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

// PressUtilization represents the utilization of a single press
type PressUtilization struct {
	PressNumber toolmodels.PressNumber `json:"press_number"`
	Tools       []*toolmodels.Tool     `json:"tools"`
	Count       int                    `json:"count"`
	Available   bool                   `json:"available"`
}

// PressUtilizationMap represents the complete utilization across all presses
type PressUtilizationMap struct {
	ActiveTools   map[toolmodels.PressNumber][]*toolmodels.Tool `json:"active_tools"`
	InactiveTools []*toolmodels.Tool                            `json:"inactive_tools"`
	Utilization   []PressUtilization                            `json:"utilization"`
	TotalActive   int                                           `json:"total_active"`
	TotalInactive int                                           `json:"total_inactive"`
}

type Service struct {
	db    *sql.DB
	notes *note.Service
	feeds *feed.Service
}

func New(db *sql.DB, notes *note.Service, feeds *feed.Service) *Service {
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
		panic(dberror.NewDatabaseError("create_table", "tools", "failed to create tools table", err))
	}

	logger.DBTools().Info("Tool service initialized")
	return &Service{db: db, notes: notes, feeds: feeds}
}

func (s *Service) List() ([]*toolmodels.Tool, error) {
	logger.DBTools().Debug("Listing all tools")

	const query = `SELECT id, position, format, type, code, regenerating, press, notes, mods FROM tools`
	rows, err := s.db.Query(query)
	if err != nil {
		logger.DBTools().Error("Failed to query tools: %v", err)
		return nil, dberror.NewDatabaseError("select", "tools", "failed to query tools", err)
	}
	defer rows.Close()

	var tools []*toolmodels.Tool
	for rows.Next() {
		tool, err := s.scanTool(rows)
		if err != nil {
			logger.DBTools().Error("Failed to scan tool: %v", err)
			return nil, dberror.WrapError(err, "failed to scan tool")
		}
		tools = append(tools, tool)
	}

	if err := rows.Err(); err != nil {
		logger.DBTools().Error("Error iterating over tool rows: %v", err)
		return nil, dberror.NewDatabaseError("select", "tools", "error iterating over rows", err)
	}

	logger.DBTools().Debug("Successfully listed %d tools", len(tools))
	return tools, nil
}

func (s *Service) Get(id int64) (*toolmodels.Tool, error) {
	logger.DBTools().Debug("Getting tool with ID: %d", id)

	const query = `SELECT id, position, format, type, code, regenerating, press, notes, mods FROM tools WHERE id = $1`
	row := s.db.QueryRow(query, id)

	tool, err := s.scanTool(row)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.DBTools().Debug("Tool not found: %d", id)
			return nil, dberror.ErrNotFound
		}
		logger.DBTools().Error("Failed to get tool %d: %v", id, err)
		return nil, dberror.NewDatabaseError("select", "tools", fmt.Sprintf("failed to get tool with ID %d", id), err)
	}

	logger.DBTools().Debug("Successfully retrieved tool: %s", tool.String())
	return tool, nil
}

func (s *Service) Add(tool *toolmodels.Tool, user *usermodels.User) (int64, error) {
	logger.DBTools().Info("Adding new tool: %s (user: %s)", tool.String(), user.Name)

	if err := s.validateToolUniqueness(tool, 0); err != nil {
		logger.DBTools().Warn("Tool validation failed: %v", err)
		return 0, err
	}

	formatBytes, notesBytes, modsBytes, err := s.marshalToolData(tool, user)
	if err != nil {
		logger.DBTools().Error("Failed to marshal tool data: %v", err)
		return 0, err
	}

	const insertQuery = `
		INSERT INTO tools (position, format, type, code, regenerating, press, notes, mods)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	result, err := s.db.Exec(insertQuery, tool.Position, formatBytes, tool.Type, tool.Code,
		tool.Regenerating, tool.Press, notesBytes, modsBytes)
	if err != nil {
		logger.DBTools().Error("Failed to insert tool: %v", err)
		return 0, dberror.NewDatabaseError("insert", "tools", "failed to insert tool", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		logger.DBTools().Error("Failed to get last insert ID: %v", err)
		return 0, dberror.NewDatabaseError("insert", "tools", "failed to get last insert ID", err)
	}

	tool.ID = id
	s.createFeedUpdate("Neues Werkzeug hinzugefügt",
		fmt.Sprintf("Benutzer %s hat ein neues Werkzeug %s zur Werkzeugliste hinzugefügt.", user.Name, tool.String()), user)

	logger.DBTools().Info("Successfully added tool with ID: %d", id)
	return id, nil
}

func (s *Service) Update(tool *toolmodels.Tool, user *usermodels.User) error {
	logger.DBTools().Info("Updating tool ID %d: %s (user: %s)", tool.ID, tool.String(), user.Name)

	if err := s.validateToolUniqueness(tool, tool.ID); err != nil {
		logger.DBTools().Warn("Tool update validation failed: %v", err)
		return err
	}

	formatBytes, notesBytes, modsBytes, err := s.marshalToolData(tool, user)
	if err != nil {
		logger.DBTools().Error("Failed to marshal tool data for update: %v", err)
		return err
	}

	const updateQuery = `
		UPDATE tools SET position = $1, format = $2, type = $3, code = $4,
		regenerating = $5, press = $6, notes = $7, mods = $8 WHERE id = $9
	`
	_, err = s.db.Exec(updateQuery, tool.Position, formatBytes, tool.Type, tool.Code,
		tool.Regenerating, tool.Press, notesBytes, modsBytes, tool.ID)
	if err != nil {
		logger.DBTools().Error("Failed to update tool %d: %v", tool.ID, err)
		return dberror.NewDatabaseError("update", "tools", fmt.Sprintf("failed to update tool with ID %d", tool.ID), err)
	}

	s.createFeedUpdate("Werkzeug aktualisiert",
		fmt.Sprintf("Benutzer %s hat das Werkzeug %s aktualisiert.", user.Name, tool.String()), user)

	logger.DBTools().Info("Successfully updated tool ID: %d", tool.ID)
	return nil
}

func (s *Service) Delete(id int64, user *usermodels.User) error {
	logger.DBTools().Info("Deleting tool ID %d (user: %s)", id, user.Name)

	tool, err := s.Get(id)
	if err != nil {
		logger.DBTools().Error("Failed to get tool before deletion: %v", err)
		return dberror.WrapError(err, "failed to get tool before deletion")
	}

	const deleteQuery = `DELETE FROM tools WHERE id = $1`
	_, err = s.db.Exec(deleteQuery, id)
	if err != nil {
		logger.DBTools().Error("Failed to delete tool %d: %v", id, err)
		return dberror.NewDatabaseError("delete", "tools", fmt.Sprintf("failed to delete tool with ID %d", id), err)
	}

	s.createFeedUpdate("Werkzeug entfernt",
		fmt.Sprintf("Benutzer %s hat das Werkzeug %s entfernt.", user.Name, tool.String()), user)

	logger.DBTools().Info("Successfully deleted tool ID: %d", id)
	return nil
}

func (s *Service) GetWithNotes(id int64) (*toolmodels.ToolWithNotes, error) {
	logger.DBTools().Debug("Getting tool with notes, ID: %d", id)

	tool, err := s.Get(id)
	if err != nil {
		return nil, err
	}

	notes, err := s.notes.GetByIDs(tool.LinkedNotes)
	if err != nil {
		logger.DBTools().Error("Failed to load notes for tool %d: %v", id, err)
		return nil, dberror.WrapError(err, "failed to load notes for tool")
	}

	logger.DBTools().Debug("Successfully retrieved tool with %d notes", len(notes))
	return &toolmodels.ToolWithNotes{Tool: tool, LoadedNotes: notes}, nil
}

func (s *Service) ListWithNotes() ([]*toolmodels.ToolWithNotes, error) {
	logger.DBTools().Debug("Listing all tools with notes")

	tools, err := s.List()
	if err != nil {
		return nil, err
	}

	var result []*toolmodels.ToolWithNotes
	for _, tool := range tools {
		notes, err := s.notes.GetByIDs(tool.LinkedNotes)
		if err != nil {
			logger.DBTools().Error("Failed to load notes for tool %d: %v", tool.ID, err)
			return nil, dberror.WrapError(err, fmt.Sprintf("failed to load notes for tool %d", tool.ID))
		}

		result = append(result, &toolmodels.ToolWithNotes{Tool: tool, LoadedNotes: notes})
	}

	logger.DBTools().Debug("Successfully listed %d tools with notes", len(result))
	return result, nil
}

func (s *Service) AddWithNotes(tool *toolmodels.Tool, user *usermodels.User, notes ...*notemodels.Note) (*toolmodels.ToolWithNotes, error) {
	logger.DBTools().Debug("Adding tool with %d notes (user: %s)", len(notes), user.Name)

	var noteIDs []int64
	for _, note := range notes {
		noteID, err := s.notes.Add(note, user)
		if err != nil {
			logger.DBTools().Error("Failed to add note: %v", err)
			return nil, dberror.WrapError(err, "failed to add note")
		}
		noteIDs = append(noteIDs, noteID)
	}

	tool.LinkedNotes = noteIDs
	toolID, err := s.Add(tool, user)
	if err != nil {
		return nil, dberror.WrapError(err, "failed to add tool")
	}

	tool.ID = toolID
	logger.DBTools().Debug("Successfully added tool with notes, ID: %d", toolID)
	return &toolmodels.ToolWithNotes{Tool: tool, LoadedNotes: notes}, nil
}

func (s *Service) GetByPress(pressNumber *toolmodels.PressNumber) ([]*toolmodels.Tool, error) {
	if pressNumber != nil && !toolmodels.IsValidPressNumber(pressNumber) {
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
	rows, err := s.db.Query(query, pressNumber)
	if err != nil {
		logger.DBTools().Error("Failed to query tools for press %v: %v", pressNumber, err)
		return nil, dberror.NewDatabaseError("select", "tools",
			fmt.Sprintf("failed to query tools for press %v", pressNumber), err)
	}
	defer rows.Close()

	var tools []*toolmodels.Tool
	for rows.Next() {
		tool, err := s.scanTool(rows)
		if err != nil {
			logger.DBTools().Error("Failed to scan tool: %v", err)
			return nil, dberror.WrapError(err, "failed to scan tool")
		}
		tools = append(tools, tool)
	}

	if err := rows.Err(); err != nil {
		logger.DBTools().Error("Error iterating over tool rows: %v", err)
		return nil, dberror.NewDatabaseError("select", "tools", "error iterating over rows", err)
	}

	logger.DBTools().Debug("Found %d tools for press %v", len(tools), pressNumber)
	return tools, nil
}

func (s *Service) UpdateRegenerating(toolID int64, regenerating bool, user *usermodels.User) error {
	logger.DBTools().Info("Updating regenerating status for tool %d to %v (user: %s)", toolID, regenerating, user.Name)

	tool, err := s.Get(toolID)
	if err != nil {
		return fmt.Errorf("failed to get tool for regenerating status update: %w", err)
	}

	if tool.Regenerating == regenerating {
		logger.DBTools().Debug("Tool %d regenerating status unchanged: %v", toolID, regenerating)
		return nil
	}

	tool.Regenerating = regenerating
	s.updateMods(user, tool)

	modsBytes, err := json.Marshal(tool.Mods)
	if err != nil {
		logger.DBTools().Error("Failed to marshal mods: %v", err)
		return dberror.NewDatabaseError("update", "tools", "failed to marshal mods", err)
	}

	const query = `UPDATE tools SET regenerating = ?, mods = ? WHERE id = ?`
	_, err = s.db.Exec(query, tool.Regenerating, modsBytes, tool.ID)
	if err != nil {
		logger.DBTools().Error("Failed to update regenerating status: %v", err)
		return dberror.NewDatabaseError("update", "tools",
			fmt.Sprintf("failed to update regenerating for tool %d", tool.ID), err)
	}

	s.createFeedUpdate("Werkzeug Regenerierung aktualisiert",
		fmt.Sprintf("Benutzer %s hat die Regenerierung für Werkzeug %s aktualisiert.", user.Name, tool.String()), user)

	logger.DBTools().Info("Successfully updated regenerating status for tool %d", toolID)
	return nil
}

func (s *Service) UpdatePress(toolID int64, press *toolmodels.PressNumber, user *usermodels.User) error {
	logger.DBTools().Info("Updating press assignment for tool %d (user: %s)", toolID, user.Name)

	tool, err := s.Get(toolID)
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

	s.updateMods(user, tool)

	modsBytes, err := json.Marshal(tool.Mods)
	if err != nil {
		logger.DBTools().Error("Failed to marshal mods: %v", err)
		return dberror.NewDatabaseError("update", "tools", "failed to marshal mods", err)
	}

	const query = `UPDATE tools SET press = ?, mods = ? WHERE id = ?`
	_, err = s.db.Exec(query, press, modsBytes, toolID)
	if err != nil {
		logger.DBTools().Error("Failed to update press assignment: %v", err)
		return dberror.NewDatabaseError("update", "tools",
			fmt.Sprintf("failed to update press for tool %d", toolID), err)
	}

	tool.Press = press
	s.createFeedUpdate("Werkzeug Pressendaten aktualisiert",
		fmt.Sprintf("Benutzer %s hat die Pressendaten für Werkzeug %s aktualisiert.", user.Name, tool.String()), user)

	logger.DBTools().Info("Successfully updated press assignment for tool %d", toolID)
	return nil
}

// GetPressUtilizationMap returns a complete overview of press utilization across all presses
func (s *Service) GetPressUtilizationMap() (*PressUtilizationMap, error) {
	logger.DBTools().Info("Generating press utilization map")

	// Get all active tools (non-regenerating)
	const activeQuery = `
		SELECT id, position, format, type, code, regenerating, press, notes, mods
		FROM tools WHERE regenerating = 0
	`
	rows, err := s.db.Query(activeQuery)
	if err != nil {
		logger.DBTools().Error("Failed to query active tools: %v", err)
		return nil, dberror.NewDatabaseError("select", "tools", "failed to query active tools", err)
	}
	defer rows.Close()

	activeTools := make(map[toolmodels.PressNumber][]*toolmodels.Tool)
	var inactiveTools []*toolmodels.Tool

	for rows.Next() {
		tool, err := s.scanTool(rows)
		if err != nil {
			logger.DBTools().Error("Failed to scan tool: %v", err)
			return nil, dberror.WrapError(err, "failed to scan tool")
		}

		if tool.Press == nil {
			inactiveTools = append(inactiveTools, tool)
		} else {
			activeTools[*tool.Press] = append(activeTools[*tool.Press], tool)
		}
	}

	if err := rows.Err(); err != nil {
		logger.DBTools().Error("Error iterating over tool rows: %v", err)
		return nil, dberror.NewDatabaseError("select", "tools", "error iterating over rows", err)
	}

	// Build utilization overview
	var utilization []PressUtilization
	totalActive := 0

	// Valid press numbers: 0, 2, 3, 4, 5
	validPresses := []toolmodels.PressNumber{0, 2, 3, 4, 5}

	for _, pressNum := range validPresses {
		tools := activeTools[pressNum]
		count := len(tools)
		totalActive += count

		utilization = append(utilization, PressUtilization{
			PressNumber: pressNum,
			Tools:       tools,
			Count:       count,
			Available:   count == 0,
		})
	}

	utilizationMap := &PressUtilizationMap{
		ActiveTools:   activeTools,
		InactiveTools: inactiveTools,
		Utilization:   utilization,
		TotalActive:   totalActive,
		TotalInactive: len(inactiveTools),
	}

	logger.DBTools().Info("Press utilization map generated - Active: %d, Inactive: %d",
		utilizationMap.TotalActive, utilizationMap.TotalInactive)

	return utilizationMap, nil
}

// Helper methods

func (s *Service) scanTool(scanner interfaces.Scannable) (*toolmodels.Tool, error) {
	tool := &toolmodels.Tool{}
	var format, linkedNotes, mods []byte

	if err := scanner.Scan(&tool.ID, &tool.Position, &format, &tool.Type,
		&tool.Code, &tool.Regenerating, &tool.Press, &linkedNotes, &mods); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(format, &tool.Format); err != nil {
		return nil, dberror.NewDatabaseError("scan", "tools", "failed to unmarshal format", err)
	}

	if err := json.Unmarshal(linkedNotes, &tool.LinkedNotes); err != nil {
		return nil, dberror.NewDatabaseError("scan", "tools", "failed to unmarshal notes", err)
	}

	if err := json.Unmarshal(mods, &tool.Mods); err != nil {
		return nil, dberror.WrapError(err, "failed to unmarshal mods data")
	}

	return tool, nil
}

func (s *Service) validateToolUniqueness(tool *toolmodels.Tool, excludeID int64) error {
	formatBytes, err := json.Marshal(tool.Format)
	if err != nil {
		return dberror.NewDatabaseError("validation", "tools", "failed to marshal format", err)
	}

	var count int
	const query = `
		SELECT COUNT(*) FROM tools
		WHERE id != $1 AND position = $2 AND format = $3 AND type = $4 AND code = $5
	`
	err = s.db.QueryRow(query, excludeID, tool.Position, formatBytes, tool.Type, tool.Code).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check for existing tool: %w", err)
	}

	if count > 0 {
		return dberror.ErrAlreadyExists
	}

	return nil
}

func (s *Service) marshalToolData(tool *toolmodels.Tool, user *usermodels.User) ([]byte, []byte, []byte, error) {
	s.updateMods(user, tool)

	formatBytes, err := json.Marshal(tool.Format)
	if err != nil {
		return nil, nil, nil, dberror.NewDatabaseError("marshal", "tools", "failed to marshal format", err)
	}

	notesBytes, err := json.Marshal(tool.LinkedNotes)
	if err != nil {
		return nil, nil, nil, dberror.NewDatabaseError("marshal", "tools", "failed to marshal notes", err)
	}

	modsBytes, err := json.Marshal(tool.Mods)
	if err != nil {
		return nil, nil, nil, dberror.NewDatabaseError("marshal", "tools", "failed to marshal mods", err)
	}

	return formatBytes, notesBytes, modsBytes, nil
}

func (s *Service) updateMods(user *usermodels.User, tool *toolmodels.Tool) {
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

func (s *Service) createFeedUpdate(title, message string, user *usermodels.User) {
	if s.feeds != nil {
		feed := feedmodels.New(title, message, user.TelegramID)
		if err := s.feeds.Add(feed); err != nil {
			logger.DBTools().Warn("Failed to create feed update: %v", err)
		}
	}
}

func equalPressNumbers(a, b *toolmodels.PressNumber) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
