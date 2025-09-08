// Package services provides convenient access to all database service packages.
// It uses type aliases to expose the service types from their respective
// sub-packages, allowing them to be accessed directly from this package.
package services

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

// Type Aliases to expose service types directly
type (
	Attachment    = attachment.Service
	Cookie        = cookie.Service
	Feed          = feed.Service
	MetalSheet    = metalsheet.Service
	Note          = note.Service
	PressCycle    = presscycle.Service
	Regeneration  = regeneration.Service
	Tool          = tool.Service
	TroubleReport = troublereport.Service
	User          = user.Service
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
func NewRegeneration(db *sql.DB, tools *Tool, feeds *Feed) *Regeneration {
	return regeneration.New(db, tools, feeds)
}

// NewTool creates a new tool service.
func NewTool(db *sql.DB, notes *Note, feeds *Feed) *Tool {
	return tool.New(db, notes, feeds)
}

// NewTroubleReport creates a new trouble report service.
func NewTroubleReport(db *sql.DB, a *Attachment, feeds *Feed) *TroubleReport {
	return troublereport.New(db, a, feeds)
}

// NewUser creates a new user service.
func NewUser(db *sql.DB, feeds *Feed) *User {
	return user.New(db, feeds)
}
