package database

import (
	"database/sql"

	"github.com/knackwurstking/pgpress/internal/services"
)

// DB represents the main database connection and provides access to all data access objects.
type DB struct {
	db *sql.DB

	// Kind of DataOperations
	PressCycles    *services.PressCycle
	Users          *services.User
	TroubleReports *services.TroubleReport
	Notes          *services.Note
	Tools          *services.Tool
	MetalSheets    *services.MetalSheet

	Attachments       *services.Attachment
	Cookies           *services.Cookie
	ToolRegenerations *services.Regeneration
	Feeds             *services.Feed
}

// New creates a new DB instance with all necessary table handlers initialized.
// Feeds must be created before Users and TroubleReports as they generate feed entries.
func New(db *sql.DB) *DB {
	feeds := services.NewFeed(db)

	attachments := services.NewAttachment(db)
	troubleReports := services.NewTroubleReport(db, attachments, feeds)

	pressCycles := services.NewPressCycle(db, feeds)

	notes := services.NewNote(db)
	tools := services.NewTool(db, notes, feeds)

	metalSheets := services.NewMetalSheet(db, feeds, notes)
	toolRegenerations := services.NewRegeneration(db, tools, feeds)

	dbInstance := &DB{
		Users:             services.NewUser(db, feeds),
		Cookies:           services.NewCookie(db),
		Attachments:       attachments,
		TroubleReports:    troubleReports,
		Notes:             notes,
		Tools:             tools,
		MetalSheets:       metalSheets,
		PressCycles:       pressCycles,
		ToolRegenerations: toolRegenerations,

		Feeds: feeds,
		db:    db,
	}

	return dbInstance
}
