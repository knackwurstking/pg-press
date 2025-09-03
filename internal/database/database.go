package database

import (
	"database/sql"

	"github.com/knackwurstking/pgpress/internal/attachment"
	"github.com/knackwurstking/pgpress/internal/cookie"
	"github.com/knackwurstking/pgpress/internal/feed"
	"github.com/knackwurstking/pgpress/internal/interfaces"
	"github.com/knackwurstking/pgpress/internal/metalsheet"
	"github.com/knackwurstking/pgpress/internal/models"
	"github.com/knackwurstking/pgpress/internal/note"
	"github.com/knackwurstking/pgpress/internal/presscycle"
	"github.com/knackwurstking/pgpress/internal/regeneration"
	"github.com/knackwurstking/pgpress/internal/tool"
	"github.com/knackwurstking/pgpress/internal/troublereport"
	"github.com/knackwurstking/pgpress/internal/user"
)

// DB represents the main database connection and provides access to all data access objects.
type DB struct {
	db *sql.DB

	// Kind of DataOperations, TODO: Need to change this
	Cookies           *cookie.Service
	ToolRegenerations *regeneration.Service
	Feeds             *feed.Service

	// DataOperations
	Users          interfaces.DataOperations[*models.User]
	Attachments    interfaces.DataOperations[*models.Attachment]
	TroubleReports interfaces.DataOperations[*models.TroubleReport]
	Notes          interfaces.DataOperations[*models.Note]
	Tools          interfaces.DataOperations[*models.Tool]
	MetalSheets    interfaces.DataOperations[*models.MetalSheet]
	PressCycles    interfaces.DataOperations[*models.PressCycle]

	// Helper
	UsersHelper          *user.UsersHelper
	TroubleReportsHelper *troublereport.TroubleReportsHelper
	ToolsHelper          *tool.ToolsHelper
	PressCyclesHelper    *presscycle.PressCyclesHelper
}

// New creates a new DB instance with all necessary table handlers initialized.
// Feeds must be created before Users and TroubleReports as they generate feed entries.
func New(db *sql.DB) *DB {
	feeds := feed.New(db)

	attachments := attachment.New(db)
	troubleReports := troublereport.New(db, feeds)
	troubleReportsHelper := troublereport.NewTroubleReportsHelper(troubleReports, attachments)

	pressCycles := presscycle.New(db, feeds)
	pressCyclesHelper := presscycle.NewPressCyclesHelper(pressCycles)

	notes := note.New(db)
	tools := tool.New(db, feeds)
	toolsHelper := tool.NewToolsHelper(tools, notes, pressCycles)

	metalSheets := metalsheet.New(db, feeds, notes)
	toolRegenerations := regeneration.New(db, feeds, pressCyclesHelper)
	usersHelper := user.NewUsersHelper(db)

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
		PressCyclesHelper:    pressCyclesHelper,
		ToolRegenerations:    toolRegenerations,

		Feeds: feeds,
		db:    db,
	}

	return dbInstance
}
