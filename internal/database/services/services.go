// Package services provides convenient access to all database service packages.
// It uses type aliases to expose the service types from their respective
// sub-packages, allowing them to be accessed directly from this package.
package services

import (
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

var (
	NewAttachment    = attachment.New
	NewCookie        = cookie.New
	NewFeed          = feed.New
	NewMetalSheet    = metalsheet.New
	NewNote          = note.New
	NewPressCycle    = presscycle.New
	NewRegeneration  = regeneration.New
	NewTool          = tool.New
	NewTroubleReport = troublereport.New
	NewUser          = user.New
)
