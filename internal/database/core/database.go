package database

import (
	"database/sql"

	"github.com/knackwurstking/pgpress/internal/database/services/attachment"
	"github.com/knackwurstking/pgpress/internal/database/services/cookie"
	"github.com/knackwurstking/pgpress/internal/database/services/feed"
	"github.com/knackwurstking/pgpress/internal/database/services/metalsheet"
	"github.com/knackwurstking/pgpress/internal/database/services/note"
	"github.com/knackwurstking/pgpress/internal/database/services/presscycle"
	"github.com/knackwurstking/pgpress/internal/database/services/regeneration"
	"github.com/knackwurstking/pgpress/internal/database/services/tool"
	"github.com/knackwurstking/pgpress/internal/database/services/troublereport"
	"github.com/knackwurstking/pgpress/internal/database/services/user"
)

// DB represents the main database connection and provides access to all data access objects.
type DB struct {
	db *sql.DB

	// Kind of DataOperations
	PressCycles    *presscycle.Service
	Users          *user.Service
	Attachments    *attachment.Service
	TroubleReports *troublereport.Service
	Notes          *note.Service
	Tools          *tool.Service
	MetalSheets    *metalsheet.Service

	// TODO: Still need to make this services fit the `interfaces.DataOperations` interface
	Cookies           *cookie.Service
	ToolRegenerations *regeneration.Service
	Feeds             *feed.Service

	// Helper TODO: Merge helper with services like the PressCycles service
	UsersHelper          *user.Helper
	TroubleReportsHelper *troublereport.Helper
	ToolsHelper          *tool.Helper
}

// New creates a new DB instance with all necessary table handlers initialized.
// Feeds must be created before Users and TroubleReports as they generate feed entries.
func New(db *sql.DB) *DB {
	feeds := feed.New(db)

	attachments := attachment.New(db)
	troubleReports := troublereport.New(db, feeds)
	troubleReportsHelper := troublereport.NewTroubleReportsHelper(troubleReports, attachments)

	pressCycles := presscycle.New(db, feeds)

	notes := note.New(db)
	tools := tool.New(db, feeds)
	toolsHelper := tool.NewHelper(tools, notes, pressCycles)

	metalSheets := metalsheet.New(db, feeds, notes)
	toolRegenerations := regeneration.New(db, feeds, pressCycles)
	usersHelper := user.NewHelper(db)

	dbInstance := &DB{
		Users:                user.New(db, feeds),
		UsersHelper:          usersHelper,
		Cookies:              cookie.New(db),
		Attachments:          attachments,
		TroubleReports:       troubleReports,
		TroubleReportsHelper: troubleReportsHelper,
		Notes:                notes,
		Tools:                tools,
		ToolsHelper:          toolsHelper,
		MetalSheets:          metalSheets,
		PressCycles:          pressCycles,
		ToolRegenerations:    toolRegenerations,

		Feeds: feeds,
		db:    db,
	}

	return dbInstance
}
