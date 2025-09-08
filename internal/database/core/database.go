package database

import (
	"database/sql"

	"github.com/knackwurstking/pgpress/internal/database/services"
)

// DB represents the main database connection and provides access to all data access objects.
type DB struct {
	db *sql.DB

	// Kind of DataOperations
	PressCycles    *services.PressCycle
	Users          *services.User
	Attachments    *services.Attachment
	TroubleReports *services.TroubleReport
	Notes          *services.Note
	Tools          *services.Tool
	MetalSheets    *services.MetalSheet

	// TODO: Still need to make this services fit the `interfaces.DataOperations` interface
	Cookies           *services.Cookie
	ToolRegenerations *services.Regeneration
	Feeds             *services.Feed

	// Helper TODO: Merge helper with services like the PressCycles service
	UsersHelper          *services.UserHelper
	TroubleReportsHelper *services.TroubleReportHelper
}

// New creates a new DB instance with all necessary table handlers initialized.
// Feeds must be created before Users and TroubleReports as they generate feed entries.
func New(db *sql.DB) *DB {
	feeds := services.NewFeed(db)

	attachments := services.NewAttachment(db)
	troubleReports := services.NewTroubleReport(db, feeds)
	troubleReportsHelper := services.NewTroubleReportHelper(troubleReports, attachments)

	pressCycles := services.NewPressCycle(db, feeds)

	notes := services.NewNote(db)
	tools := services.NewTool(db, notes, feeds)

	metalSheets := services.NewMetalSheet(db, feeds, notes)
	toolRegenerations := services.NewRegeneration(db, feeds, pressCycles)
	usersHelper := services.NewUserHelper(db)

	dbInstance := &DB{
		Users:                services.NewUser(db, feeds),
		UsersHelper:          usersHelper,
		Cookies:              services.NewCookie(db),
		Attachments:          attachments,
		TroubleReports:       troubleReports,
		TroubleReportsHelper: troubleReportsHelper,
		Notes:                notes,
		Tools:                tools,
		MetalSheets:          metalSheets,
		PressCycles:          pressCycles,
		ToolRegenerations:    toolRegenerations,

		Feeds: feeds,
		db:    db,
	}

	return dbInstance
}
