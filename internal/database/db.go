package database

import (
	"database/sql"
)

// DB represents the main database connection and provides access to all data access objects.
type DB struct {
	Users                *Users
	Cookies              *Cookies
	Attachments          *Attachments
	TroubleReports       *TroubleReports
	TroubleReportsHelper *TroubleReportsHelper
	Notes                *Notes
	Tools                *Tools
	ToolsHelper          *ToolsHelper
	MetalSheets          *MetalSheets
	Presses              *Presses
	ToolRegenerations    *ToolRegenerations
	ToolCyclesHelper     *ToolCyclesHelper
	Feeds                *Feeds
	db                   *sql.DB
}

// New creates a new DB instance with all necessary table handlers initialized.
// Feeds must be created before Users and TroubleReports as they generate feed entries.
func New(db *sql.DB) *DB {
	feeds := NewFeeds(db)

	attachments := NewAttachments(db)
	troubleReports := NewTroubleReports(db, feeds)
	troubleReportsHelper := NewTroubleReportsHelper(troubleReports, attachments)

	notes := NewNotes(db)
	tools := NewTools(db, feeds)
	toolsHelper := NewToolsHelper(tools, notes)

	metalSheets := NewMetalSheets(db, feeds, notes)
	presses := NewPresses(db, feeds)
	toolRegenerations := NewToolRegenerations(db, feeds, presses)

	dbInstance := &DB{
		Users:                NewUsers(db, feeds),
		Cookies:              NewCookies(db),
		Attachments:          attachments,
		TroubleReports:       troubleReports,
		TroubleReportsHelper: troubleReportsHelper,
		Notes:                notes,
		Tools:                tools,
		ToolsHelper:          toolsHelper,
		MetalSheets:          metalSheets,
		Presses:              presses,
		ToolRegenerations:    toolRegenerations,
		Feeds:                feeds,
		db:                   db,
	}

	// Initialize ToolCyclesHelper after DB instance is created
	dbInstance.ToolCyclesHelper = NewToolCyclesHelper(dbInstance)

	return dbInstance
}
