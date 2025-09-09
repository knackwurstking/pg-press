package models

// Re-export all model types from their respective packages
import (
	"github.com/knackwurstking/pgpress/internal/database/models/attachment"
	"github.com/knackwurstking/pgpress/internal/database/models/cookie"
	"github.com/knackwurstking/pgpress/internal/database/models/cycle"
	"github.com/knackwurstking/pgpress/internal/database/models/feed"
	"github.com/knackwurstking/pgpress/internal/database/models/metalsheet"
	"github.com/knackwurstking/pgpress/internal/database/models/note"
	"github.com/knackwurstking/pgpress/internal/database/models/tool"
	"github.com/knackwurstking/pgpress/internal/database/models/troublereport"
	"github.com/knackwurstking/pgpress/internal/database/models/user"
)

// User model types
type User = user.User

// Feed model types
type Feed = feed.Feed

// Tool model types
type (
	Tool          = tool.Tool
	ToolWithNotes = tool.ToolWithNotes
	ToolMod       = tool.ToolMod
	Format        = tool.Format
	Position      = tool.Position
	Status        = tool.Status
	PressNumber   = tool.PressNumber
)

// Tool constants
const (
	PositionTop         = tool.PositionTop
	PositionTopCassette = tool.PositionTopCassette
	PositionBottom      = tool.PositionBottom

	StatusActive       = tool.StatusActive
	StatusAvailable    = tool.StatusAvailable
	StatusRegenerating = tool.StatusRegenerating

	ToolCycleWarning = tool.ToolCycleWarning
	ToolCycleError   = tool.ToolCycleError
)

// Note model types
type (
	Note  = note.Note
	Level = note.Level
)

// Note constants
const (
	INFO      = note.INFO
	ATTENTION = note.ATTENTION
	BROKEN    = note.BROKEN
)

// Cookie model types
type Cookie = cookie.Cookie

// Cycle model types
type Cycle = cycle.Cycle

// Attachment model types
type Attachment = attachment.Attachment

// Metalsheet model types
type (
	MetalSheet          = metalsheet.MetalSheet
	MetalSheetWithNotes = metalsheet.MetalSheetWithNotes
	MetalSheetMod       = metalsheet.MetalSheetMod
)

// TroubleReport model types
type (
	TroubleReport                = troublereport.TroubleReport
	TroubleReportWithAttachments = troublereport.TroubleReportWithAttachments
	TroubleReportMod             = troublereport.TroubleReportMod
)

// Attachment constants
const (
	MinIDLength = attachment.MinIDLength
	MaxIDLength = attachment.MaxIDLength
	MaxDataSize = attachment.MaxDataSize
)

// TroubleReport constants
const (
	MinTitleLength   = troublereport.MinTitleLength
	MaxTitleLength   = troublereport.MaxTitleLength
	MinContentLength = troublereport.MinContentLength
	MaxContentLength = troublereport.MaxContentLength
)

// Cookie constants
const (
	DefaultExpiration  = cookie.DefaultExpiration
	MinValueLength     = cookie.MinValueLength
	MaxUserAgentLength = cookie.MaxUserAgentLength
)

// User constants
const (
	MinNameLength = user.MinNameLength
	MaxNameLength = user.MaxNameLength
)

// Constructor functions - User
var NewUser = user.NewUser
var NewUserFromInterfaceMap = user.NewUserFromInterfaceMap

// Constructor functions - Feed
var NewFeed = feed.New

// Constructor functions - Tool
var NewTool = tool.New
var IsValidPressNumber = tool.IsValidPressNumber

// Constructor functions - Note
var NewNote = note.New

// Constructor functions - Mod (generic functions cannot be re-exported as variables)
// Use mod.NewMods[T]() and mod.NewMod[T]() directly from "github.com/knackwurstking/pgpress/internal/mod"

// Constructor functions - Cookie
var NewCookie = cookie.New

// Constructor functions - Cycle
var NewCycle = cycle.NewCycle
var NewPressCycleWithID = cycle.NewPressCycleWithID

// Cycle utility functions
var FilterByTool = cycle.FilterByTool
var FilterByToolPosition = cycle.FilterByToolPosition

// Constructor functions - Metalsheet
var NewMetalSheet = metalsheet.New

// Constructor functions - TroubleReport
var NewTroubleReport = troublereport.New

// Mod constructor functions (generic functions cannot be re-exported as variables)
// Use mod.NewMod[T]() directly from "github.com/knackwurstking/pgpress/internal/mod"

// Cookie utility functions
var SortCookies = cookie.Sort
