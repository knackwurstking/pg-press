// Package services provides convenient access to all database service packages.
// It uses type aliases to expose the service types from their respective
// sub-packages, allowing them to be accessed directly from this package.
package services

import (
	"github.com/knackwurstking/pgpress/internal/services/cookie"
	"github.com/knackwurstking/pgpress/internal/services/feed"
	"github.com/knackwurstking/pgpress/internal/services/metalsheet"
	"github.com/knackwurstking/pgpress/internal/services/note"
	"github.com/knackwurstking/pgpress/internal/services/presscycle"
	"github.com/knackwurstking/pgpress/internal/services/regeneration"
	"github.com/knackwurstking/pgpress/internal/services/tool"
	"github.com/knackwurstking/pgpress/internal/services/troublereport"
	"github.com/knackwurstking/pgpress/internal/services/user"
)

// Type Aliases to expose service types directly
type (
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
