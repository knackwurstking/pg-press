package database

import (
	"database/sql"

	"github.com/knackwurstking/pgpress/internal/models"
)

type Broadcaster interface {
	Broadcast()
}

// DataOperations defines a basic generic interface for handling data models in the database.
// It standardizes the Add, Update, and Delete operations.
// This interface can be implemented by database handlers to provide a consistent API.
//
// T is the type of the data model (e.g., *Tool).
type DataOperations[T any] interface {
	Get(id int64) (T, error)
	List() ([]T, error)
	// Add creates a new record for the given model.
	// It may take a user for auditing purposes and may return the ID of the new record.
	Add(model T, user *models.User) (int64, error)

	// Update modifies an existing record.
	// It may take a user for auditing purposes. The model should contain its ID.
	Update(model T, user *models.User) error

	// Delete removes a record from the database by its ID.
	// It may take a user for auditing purposes.
	Delete(id int64, user *models.User) error
}

// DB represents the main database connection and provides access to all data access objects.
type DB struct {
	Users                DataOperations[*models.User]
	UsersHelper          *UsersHelper
	Cookies              *Cookies
	Attachments          *Attachments
	TroubleReports       DataOperations[*TroubleReport]
	TroubleReportsHelper *TroubleReportsHelper
	Notes                *Notes
	Tools                DataOperations[*models.Tool]
	ToolsHelper          *ToolsHelper
	MetalSheets          DataOperations[*models.MetalSheet]
	PressCycles          DataOperations[*models.PressCycle]
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
