package database

import (
	"database/sql"
)

// DB represents the main database connection and provides access to all data access objects.
type DB struct {
	Users                DataOperations[*User]
	UsersHelper          *UsersHelper
	Cookies              *Cookies
	Attachments          *Attachments
	TroubleReports       DataOperations[*TroubleReport]
	TroubleReportsHelper *TroubleReportsHelper
	Notes                *Notes
	Tools                DataOperations[*Tool]
	ToolsHelper          *ToolsHelper
	MetalSheets          DataOperations[*MetalSheet]
	PressCycles          DataOperations[*PressCycle]
	PressCyclesHelper    *PressCyclesHelper
	ToolRegenerations    *ToolRegenerations

	Feeds *Feeds
	db    *sql.DB
}

// New creates a new DB instance with all necessary table handlers initialized.
// Feeds must be created before Users and TroubleReports as they generate feed entries.
func New(db *sql.DB) *DB {
	feeds := NewFeeds(db)

	attachments := NewAttachments(db)
	troubleReports := NewTroubleReports(db, feeds)
	troubleReportsHelper := NewTroubleReportsHelper(troubleReports, attachments)

	pressCycles := NewPressCycles(db, feeds)
	pressCyclesHelper := NewPressCyclesHelper(pressCycles)

	notes := NewNotes(db)
	tools := NewTools(db, feeds)
	toolsHelper := NewToolsHelper(tools, notes, pressCycles)

	metalSheets := NewMetalSheets(db, feeds, notes)
	toolRegenerations := NewToolRegenerations(db, feeds, pressCyclesHelper)
	usersHelper := NewUsersHelper(db)

	dbInstance := &DB{
		Users:                NewUsers(db, feeds),
		UsersHelper:          usersHelper,
		Cookies:              NewCookies(db),
		Attachments:          attachments,
		TroubleReports:       troubleReports,
		TroubleReportsHelper: troubleReportsHelper,
		Notes:                notes,
		Tools:                tools,
		ToolsHelper:          toolsHelper,
		MetalSheets:          metalSheets,
		PressCycles:          pressCycles,
		PressCyclesHelper:    pressCyclesHelper,
		ToolRegenerations:    toolRegenerations,

		Feeds: feeds,
		db:    db,
	}

	return dbInstance
}
