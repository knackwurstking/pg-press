// Package services provides convenient access to all database service packages.
// It uses type aliases to expose the service types from their respective
// sub-packages, allowing them to be accessed directly from this package.
package services

import (
	"database/sql"

	"github.com/knackwurstking/pgpress/internal/database/interfaces"
	"github.com/knackwurstking/pgpress/internal/database/models"
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

// Type Aliases to expose service types directly
type (
	Attachment          = attachment.Service
	Cookie              = cookie.Service
	Feed                = feed.Service
	MetalSheet          = metalsheet.Service
	Note                = note.Service
	PressCycle          = presscycle.Service
	Regeneration        = regeneration.Service
	Tool                = tool.Service
	ToolHelper          = tool.Helper
	TroubleReport       = troublereport.Service
	TroubleReportHelper = troublereport.Helper
	User                = user.Service
	UserHelper          = user.Helper
)

// NewAttachment creates a new attachment service.
func NewAttachment(db *sql.DB) *Attachment {
	return attachment.New(db)
}

// NewCookie creates a new cookie service.
func NewCookie(db *sql.DB) *Cookie {
	return cookie.New(db)
}

// NewFeed creates a new feed service.
func NewFeed(db *sql.DB) *Feed {
	return feed.New(db)
}

// NewMetalSheet creates a new metal sheet service.
func NewMetalSheet(db *sql.DB, feeds *Feed, notes *Note) *MetalSheet {
	return metalsheet.New(db, feeds, notes)
}

// NewNote creates a new note service.
func NewNote(db *sql.DB) *Note {
	return note.New(db)
}

// NewPressCycle creates a new press cycle service.
func NewPressCycle(db *sql.DB, feeds *Feed) *PressCycle {
	return presscycle.New(db, feeds)
}

// NewRegeneration creates a new regeneration service.
func NewRegeneration(db *sql.DB, feeds *Feed, pressCycles *PressCycle) *Regeneration {
	return regeneration.New(db, feeds, pressCycles)
}

// NewTool creates a new tool service.
func NewTool(db *sql.DB, feeds *Feed) *Tool {
	return tool.New(db, feeds)
}

// NewToolHelper creates a new tool helper.
func NewToolHelper(tools *Tool, notes *Note, pressCycles *PressCycle) *ToolHelper {
	return tool.NewHelper(tools, notes, pressCycles)
}

// NewTroubleReport creates a new trouble report service.
func NewTroubleReport(db *sql.DB, feeds *Feed) *TroubleReport {
	return troublereport.New(db, feeds)
}

// NewTroubleReportHelper creates a new trouble report helper.
func NewTroubleReportHelper(tr interfaces.DataOperations[*models.TroubleReport], a *Attachment) *TroubleReportHelper {
	return troublereport.NewTroubleReportsHelper(tr, a)
}

// NewUser creates a new user service.
func NewUser(db *sql.DB, feeds *Feed) *User {
	return user.New(db, feeds)
}

// NewUserHelper creates a new user helper.
func NewUserHelper(db *sql.DB) *UserHelper {
	return user.NewHelper(db)
}
