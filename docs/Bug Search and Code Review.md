## User

Search for possible bugs and stuff

## Assistant

<thinking>
The user wants me to search for possible bugs and issues in the project. I should start by exploring the project structure to understand what we're working with, then look for common bug patterns, potential issues, and other problems.

Let me start by:
1. Listing the root directory to understand the project structure
2. Looking for common file types and patterns
3. Running diagnostics to see if there are any immediate errors
4. Then searching for common bug patterns in the code

Let me begin with exploring the project structure.
</thinking>

I'll help you search for potential bugs and issues in your project. Let me start by exploring the project structure and running some diagnostics.

**Tool Call: List the `pg-press` directory's contents**
Status: Completed

# Folders:
pg-press/bin
pg-press/cmd
pg-press/docs
pg-press/internal
pg-press/pkg
pg-press/scripts

# Files:
pg-press/.gitignore
pg-press/.gitmodules
pg-press/Makefile
pg-press/README.md
pg-press/go.mod
pg-press/go.sum


## Assistant

This looks like a Go project! Let me check for any immediate compilation issues and then explore the code structure:

**Tool Call: Check project diagnostics**
Status: Completed

No errors or warnings found in the project.

## Assistant

Good! No immediate compilation errors. Let me explore the main source directories to understand the project better:

**Tool Call: List the `pg-press/cmd` directory's contents**
Status: Completed

# Folders:
pg-press/cmd/pg-press


**Tool Call: List the `pg-press/internal` directory's contents**
Status: Completed

# Folders:
pg-press/internal/constants
pg-press/internal/database
pg-press/internal/env
pg-press/internal/interfaces
pg-press/internal/pdf
pg-press/internal/services
pg-press/internal/web


**Tool Call: List the `pg-press/pkg` directory's contents**
Status: Completed

# Folders:
pg-press/pkg/constants
pg-press/pkg/logger
pg-press/pkg/models
pg-press/pkg/modification
pg-press/pkg/utils


## Assistant

Now let me look at the main entry point and then search for common bug patterns:

**Tool Call: Read file `cmd/pg-press/main.go`**
Status: Completed

```pg-press/cmd/pg-press/main.go
package main

import (
	"os"
	"path/filepath"

	"github.com/knackwurstking/pgpress/pkg/logger"

	"github.com/SuperPaintman/nice/cli"
)

const (
	appName      = "pg-press"
	version      = "v0.0.1"
	databaseFile = "pg-press.db"

	exitCodeGeneric     = 1
	exitCodeNotFound    = 10
	exitCodeServerStart = 20 // exitCodeServerStart failed
)

var (
	configPath       string
	serverPathPrefix = os.Getenv("SERVER_PATH_PREFIX")
	serverAddress    = os.Getenv("SERVER_ADDR")
)

func init() {
	p, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}
	configPath = filepath.Join(p, appName)
	if err := os.MkdirAll(configPath, 0700); err != nil {
		panic(err)
	}
}

func main() {
	// Initialize colored logger
	logger.Initialize()

	a := cli.App{
		Name: appName,
		Usage: cli.Usage(`Exit Codes:
  1   Generic
  10  Not Found
`),
		Commands: []cli.Command{
			apiKeyCommand(),

			{
				Name: "user",
				Usage: cli.Usage(
					"Handle users database table, add, remove or modify user data"),
				Commands: []cli.Command{
					listUserCommand(),
					showUserCommand(),
					addUserCommand(),
					removeUserCommand(),
					modUserCommand(),
				},
			},

			{
				Name: "cookies",
				Usage: cli.Usage(
					"Handle cookies database table, remove cookies data"),
				Commands: []cli.Command{
					removeCookiesCommand(),
					autoCleanCookiesCommand(),
				},
			},

			{
				Name: "feeds",
				Usage: cli.Usage(
					"Handle feeds database table, remove, list, add, show, modify feeds data"),
				Commands: []cli.Command{
					listFeedsCommand(),
					removeFeedsCommand(),
				},
			},

			serverCommand(),

			cli.CompletionCommand(),
		},
		CommandFlags: []cli.CommandFlag{
			cli.HelpCommandFlag(),
			cli.VersionCommandFlag(version),
		},
	}

	a.HandleError(a.Run())
}

```


## Assistant

Now let me search for common bug patterns in Go code. I'll look for several potential issues:

**Tool Call: Search files for regex `panic\(`**
Status: Completed

Found 15 matches:

## Matches in pg-press/cmd/pg-press/main.go

### func init › L30-32
```
	if err != nil {
		panic(err)
	}
```

### func init › L34-36
```
	if err := os.MkdirAll(configPath, 0700); err != nil {
		panic(err)
	}
```

## Matches in pg-press/internal/services/attachments.go

### func NewAttachments › L31-33
```
	if _, err := db.Exec(query); err != nil {
		panic(fmt.Errorf("failed to create attachments table: %v", err))
	}
```

## Matches in pg-press/internal/services/cookies.go

### func NewCookies › L31-33
```
	if _, err := db.Exec(query); err != nil {
		panic(fmt.Errorf("failed to create cookies table: %v", err))
	}
```

## Matches in pg-press/internal/services/feeds.go

### func NewFeeds › L36-38
```
	if _, err := db.Exec(query); err != nil {
		panic(fmt.Errorf("failed to create feeds table: %v", err))
	}
```

## Matches in pg-press/internal/services/metal-sheets.go

### func NewMetalSheets › L27-29
```
	if err := metalSheet.createTable(); err != nil {
		panic(err)
	}
```

## Matches in pg-press/internal/services/modifications.go

### func (s *Modifications) createTable › L68-70
```
	if _, err := s.db.Exec(query); err != nil {
		panic(fmt.Errorf("failed to create modifications table: %v", err))
	}
```

## Matches in pg-press/internal/services/notes.go

### func NewNotes › L31-33
```
	if _, err := db.Exec(query); err != nil {
		panic(fmt.Errorf("failed to create notes table: %v", err))
	}
```

## Matches in pg-press/internal/services/press-cycles.go

### func NewPressCycles › L38-40
```
	if _, err := db.Exec(query); err != nil {
		panic(fmt.Errorf("failed to create press_cycles table: %v", err))
	}
```

## Matches in pg-press/internal/services/tool-regenerations.go

### func NewToolRegenerations › L36-38
```
	if _, err := db.Exec(query); err != nil {
		panic(fmt.Errorf("failed to create tool_regenerations table: %v", err))
	}
```

## Matches in pg-press/internal/services/tools.go

### func NewTools › L27-29
```
	if err := t.createTable(); err != nil {
		panic(err)
	}
```

## Matches in pg-press/internal/services/trouble-reports.go

### func NewTroubleReports › L29-31
```
	if err := troubleReport.createTable(db); err != nil {
		panic(err)
	}
```

## Matches in pg-press/internal/services/users.go

### func NewUsers › L36-38
```
	if _, err := db.Exec(query); err != nil {
		panic(fmt.Errorf("failed to create users table: %v", err))
	}
```

## Matches in pg-press/internal/web/shared/helpers/routes.go

### func RegisterEchoRoutes › L48-50
```
		default:
			panic("unhandled method: " + route.Method)
		}
```

## Matches in pg-press/pkg/models/metalsheet.go

### func MustParseMachineType › L49-51
```
	if err != nil {
		panic(err)
	}
```


**Tool Call: Search files for regex `defer.*Close\(\)`**
Status: Completed

Showing matches 1-20 (there were more matches found; use offset: 20 to see next page):

## Matches in pg-press/internal/services/attachments.go

### func (a *Attachments) List › L42-52
```
func (a *Attachments) List() ([]*models.Attachment, error) {
	a.log.Debug("Listing all attachments")

	query := `SELECT id, mime_type, data FROM attachments ORDER BY id ASC`
	rows, err := a.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("select attachments: %v", err)
	}
	defer rows.Close()

	var attachments []*models.Attachment
```

15 lines remaining in ancestor node. Read the file to see all.

### func (s *Attachments) GetByIDs › L114-118
```
		return nil, fmt.Errorf("select attachments: %v", err)
	}
	defer rows.Close()

	// Store attachments in a map for efficient lookup
```

## Matches in pg-press/internal/services/cookies.go

### func (c *Cookies) List › L39-49
```
func (c *Cookies) List() ([]*models.Cookie, error) {
	c.log.Info("Listing all cookies")

	query := `SELECT * FROM cookies ORDER BY last_login DESC`
	rows, err := c.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("select error: cookies: %v", err)
	}
	defer rows.Close()

	var cookies []*models.Cookie
```

15 lines remaining in ancestor node. Read the file to see all.

### func (c *Cookies) ListApiKey › L79-83
```
		return nil, fmt.Errorf("select error: cookies: %v", err)
	}
	defer rows.Close()

	var cookies []*models.Cookie
```

## Matches in pg-press/internal/services/feeds.go

### func (f *Feeds) List › L49-59
```
func (f *Feeds) List() ([]*models.Feed, error) {
	start := time.Now()

	query := `SELECT id, title, content, user_id, created_at FROM feeds ORDER BY created_at DESC`
	rows, err := f.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("select error: feeds: %v", err)
	}
	defer rows.Close()

	feeds, err := f.scanAllRows(rows)
```

12 lines remaining in ancestor node. Read the file to see all.

### func (f *Feeds) ListRange › L91-95
```
		return nil, fmt.Errorf("select error: feeds: %v", err)
	}
	defer rows.Close()

	feeds, err := f.scanAllRows(rows)
```

### func (f *Feeds) ListByUser › L130-134
```
		return nil, fmt.Errorf("select error: feeds: %v", err)
	}
	defer rows.Close()

	feeds, err := f.scanAllRows(rows)
```

## Matches in pg-press/internal/services/metal-sheets.go

### func (s *MetalSheets) List › L69-73
```
		return nil, fmt.Errorf("select error: metal_sheets: %v", err)
	}
	defer rows.Close()

	var sheets []*models.MetalSheet
```

### func (s *MetalSheets) GetByToolID › L142-146
```
		return nil, fmt.Errorf("select error: metal_sheets: %v", err)
	}
	defer rows.Close()

	var sheets []*models.MetalSheet
```

### func (s *MetalSheets) GetByMachineType › L180-184
```
		return nil, fmt.Errorf("select error: metal_sheets: %v", err)
	}
	defer rows.Close()

	var sheets []*models.MetalSheet
```

## Matches in pg-press/internal/services/modifications.go

### func (s *Modifications) List › L151-155
```
		return nil, fmt.Errorf("failed to query modifications: %v", err)
	}
	defer rows.Close()

	var modifications []*models.Modification[interface{}]
```

### func (s *Modifications) GetByUser › L313-317
```
		return nil, fmt.Errorf("failed to query modifications by user: %v", err)
	}
	defer rows.Close()

	var modifications []*models.Modification[interface{}]
```

### func (s *Modifications) GetByDateRange › L350-354
```
		return nil, fmt.Errorf("failed to query modifications by date range: %v", err)
	}
	defer rows.Close()

	var modifications []*models.Modification[any]
```

### func (s *Modifications) ListWithUser › L452-456
```
		return nil, fmt.Errorf("failed to query modifications with user: %v", err)
	}
	defer rows.Close()

	var modifications []*ModificationWithUser
```

## Matches in pg-press/internal/services/notes.go

### func (n *Notes) List › L50-54
```
		return nil, fmt.Errorf("select error: notes: %v", err)
	}
	defer rows.Close()

	var notes []*models.Note
```

### func (n *Notes) GetByIDs › L113-117
```
		return nil, fmt.Errorf("select error: notes: %v", err)
	}
	defer rows.Close()

	// Store attachments in a map for efficient lookup
```

### func (n *Notes) GetByPress › L155-159
```
		return nil, fmt.Errorf("select error: notes by press: %v", err)
	}
	defer rows.Close()

	var notes []*models.Note
```

### func (n *Notes) GetByTool › L188-192
```
		return nil, fmt.Errorf("select error: notes by tool: %v", err)
	}
	defer rows.Close()

	var notes []*models.Note
```

## Matches in pg-press/internal/services/press-cycles.go

### func (p *PressCycles) List › L126-130
```
		return nil, fmt.Errorf("select error: press_cycles: %v", err)
	}
	defer rows.Close()

	return p.scanPressCyclesRows(rows)
```

### func (s *PressCycles) GetPressCyclesForTool › L248-252
```
		return nil, fmt.Errorf("select error: press_cycles: %v", err)
	}
	defer rows.Close()

	s.log.Debug("Query executed successfully")
```


**Tool Call: Search files for regex `TODO|FIXME|XXX|HACK`**
Status: Completed

Found 1 matches:

## Matches in pg-press/internal/web/assets/css/bootstrap-icons.min.css

### .bi-mastodon::before › L1-6
```
/*!
 * Bootstrap Icons v1.13.1 (https://icons.getbootstrap.com/)
 * Copyright 2019-2024 The Bootstrap Authors
 * Licensed under MIT (https://github.com/twbs/icons/blob/main/LICENSE)
 */@font-face{font-display:block;font-family:bootstrap-icons;src:url("../bootstrap-icons.woff2") format("woff2"),url("../bootstrap-icons.woff") format("woff")}.bi::before,[class*=" bi-"]::before,[class^=bi-]::before{display:inline-block;font-family:bootstrap-icons!important;font-style:normal;font-weight:400!important;font-variant:normal;text-transform:none;line-height:1;vertical-align:-.125em;-webkit-font-smoothing:antialiased;-moz-osx-font-smoothing:grayscale}.bi-123::before{content:"\f67f"}.bi-alarm-fill::before{content:"\f101"}.bi-alarm::before{content:"\f102"}.bi-align-bottom::before{content:"\f103"}.bi-align-center::before{content:"\f104"}.bi-align-end::before{content:"\f105"}.bi-align-middle::before{content:"\f106"}.bi-align-start::before{content:"\f107"}.bi-align-top::before{content:"\f108"}.bi-alt::before{content:"\f109"}.bi-app-indicator::before{content:"\f10a"}.bi-app::before{content:"\f10b"}.bi-archive-fill::before{content:"\f10c"}.bi-archive::before{content:"\f10d"}.bi-arrow-90deg-down::before{content:"\f10e"}.bi-arrow-90deg-left::before{content:"\f10f"}.bi-arrow-90deg-right::before{content:"\f110"}.bi-arrow-90deg-up::before{content:"\f111"}.bi-arrow-bar-down::before{content:"\f112"}.bi-arrow-bar-left::before{content:"\f113"}.bi-arrow-bar-right::before{content:"\f114"}.bi-arrow-bar-up::before{content:"\f115"}.bi-arrow-clockwise::before{content:"\f116"}.bi-arrow-counterclockwise::before{content:"\f117"}.bi-arrow-down-circle-fill::before{content:"\f118"}.bi-arrow-down-circle::before{content:"\f119"}.bi-arrow-down-left-circle-fill::before{content:"\f11a"}.bi-arrow-down-left-circle::before{content:"\f11b"}.bi-arrow-down-left-square-fill::before{content:"\f11c"}.bi-arrow-down-left-square::before{content:"\f11d"}.bi-arrow-down-left::before{content:"\f11e"}.bi-arrow-down-right-circle-fill::before{content:"\f11f"}.bi-arrow-down-right-circle::before{content:"\f120"}.bi-arrow-down-right-square-fill::before{content:"\f121"}.bi-arrow-down-right-square::before{content:"\f122"}.bi-arrow-down-right::before{content:"\f123"}.bi-arrow-down-short::before{content:"\f124"}.bi-arrow-down-square-fill::before{content:"\f125"}.bi-arrow-down-square::before{content:"\f126"}.bi-arrow-down-up::before{content:"\f127"}.bi-arrow-down::before{content:"\f128"}.bi-arrow-left-circle-fill::before{content:"\f129"}.bi-arrow-left-circle::before{content:"\f12a"}.bi-arrow-left-right::before{content:"\f12b"}.bi-arrow-left-short::before{content:"\f12c"}.bi-arrow-left-square-fill::before{content:"\f12d"}.bi-arrow-left-square::before{content:"\f12e"}.bi-arrow-left::before{content:"\f12f"}.bi-arrow-repeat::before{content:"\f130"}.bi-arrow-return-left::before{content:"\f131"}.bi-arrow-return-right::before{content:"\f132"}.bi-arrow-right-circle-fill::before{content:"\f133"}.bi-arrow-right-circle::before{content:"\f134"}.bi-arrow-right-short::before{content:"\f135"}.bi-arrow-right-square-fill::before{content:"\f136"}.bi-arrow-right-square::before{content:"\f137"}.bi-arrow-right::before{content:"\f138"}.bi-arrow-up-circle-fill::before{content:"\f139"}.bi-arrow-up-circle::before{content:"\f13a"}.bi-arrow-up-left-circle-fill::before{content:"\f13b"}.bi-arrow-up-left-circle::before{content:"\f13c"}.bi-arrow-up-left-square-fill::before{content:"\f13d"}.bi-arrow-up-left-square::before{content:"\f13e"}.bi-arrow-up-left::before{content:"\f13f"}.bi-arrow-up-right-circle-fill::before{content:"\f140"}.bi-arrow-up-right-circle::before{content:"\f141"}.bi-arrow-up-right-square-fill::before{content:"\f142"}.bi-arrow-up-right-square::before{content:"\f143"}.bi-arrow-up-right::before{content:"\f144"}.bi-arrow-up-short::before{content:"\f145"}.bi-arrow-up-square-fill::before{content:"\f146"}.bi-arrow-up-square::before{content:"\f147"}.bi-arrow-up::before{content:"\f148"}.bi-arrows-angle-contract::before{content:"\f149"}.bi-arrows-angle-expand::before{content:"\f14a"}.bi-arrows-collapse::before{content:"\f14b"}.bi-arrows-expand::before{content:"\f14c"}.bi-arrows-fullscreen::before{content:"\f14d"}.bi-arrows-move::before{content:"\f14e"}.bi-aspect-ratio-fill::before{content:"\f14f"}.bi-aspect-ratio::before{content:"\f150"}.bi-asterisk::before{content:"\f151"}.bi-at::before{content:"\f152"}.bi-award-fill::before{content:"\f153"}.bi-award::before{content:"\f154"}.bi-back::before{content:"\f155"}.bi-backspace-fill::before{content:"\f156"}.bi-backspace-reverse-fill::before{content:"\f157"}.bi-backspace-reverse::before{content:"\f158"}.bi-backspace::before{content:"\f159"}.bi-badge-3d-fill::before{content:"\f15a"}.bi-badge-3d::before{content:"\f15b"}.bi-badge-4k-fill::before{content:"\f15c"}.bi-badge-4k::before{content:"\f15d"}.bi-badge-8k-fill::before{content:"\f15e"}.bi-badge-8k::before{content:"\f15f"}.bi-badge-ad-fill::before{content:"\f160"}.bi-badge-ad::before{content:"\f161"}.bi-badge-ar-fill::before{content:"\f162"}.bi-badge-ar::before{content:"\f163"}.bi-badge-cc-fill::before{content:"\f164"}.bi-badge-cc::before{content:"\f165"}.bi-badge-hd-fill::before{content:"\f166"}.bi-badge-hd::before{content:"\f167"}.bi-badge-tm-fill::before{content:"\f168"}.bi-badge-tm::before{content:"\f169"}.bi-badge-vo-fill::before{content:"\f16a"}.bi-badge-vo::before{content:"\f16b"}.bi-badge-vr-fill::before{content:"\f16c"}.bi-badge-vr::before{content:"\f16d"}.bi-badge-wc-fill::before{content:"\f16e"}.bi-badge-wc::before{content:"\f16f"}.bi-bag-check-fill::before{content:"\f170"}.bi-bag-check::before{content:"\f171"}.bi-bag-dash-fill::before{content:"\f172"}.bi-bag-dash::before{content:"\f173"}.bi-bag-fill::before{content:"\f174"}.bi-bag-plus-fill::before{content:"\f175"}.bi-bag-plus::before{content:"\f176"}.bi-bag-x-fill::before{content:"\f177"}.bi-bag-x::before{content:"\f178"}.bi-bag::before{content:"\f179"}.bi-bar-chart-fill::before{content:"\f17a"}.bi-bar-chart-line-fill::before{content:"\f17b"}.bi-bar-chart-line::before{content:"\f17c"}.bi-bar-chart-steps::before{content:"\f17d"}.bi-bar-chart::before{content:"\f17e"}.bi-basket-fill::before{content:"\f17f"}.bi-basket::before{content:"\f180"}.bi-basket2-fill::before{content:"\f181"}.bi-basket2::before{content:"\f182"}.bi-basket3-fill::before{content:"\f183"}.bi-basket3::before{content:"\f184"}.bi-battery-charging::before{content:"\f185"}.bi-battery-full::before{content:"\f186"}.bi-battery-half::before{content:"\f187"}.bi-battery::before{content:"\f188"}.bi-bell-fill::before{content:"\f189"}.bi-bell::before{content:"\f18a"}.bi-bezier::before{content:"\f18b"}.bi-bezier2::before{content:"\f18c"}.bi-bicycle::before{content:"\f18d"}.bi-binoculars-fill::before{content:"\f18e"}.bi-binoculars::before{content:"\f18f"}.bi-blockquote-left::before{content:"\f190"}.bi-blockquote-right::before{content:"\f191"}.bi-book-fill::before{content:"\f192"}.bi-book-half::before{content:"\f193"}.bi-book::before{content:"\f194"}.bi-bookmark-check-fill::before{content:"\f195"}.bi-bookmark-check::before{content:"\f196"}.bi-bookmark-dash-fill::before{content:"\f197"}.bi-bookmark-dash::before{content:"\f198"}.bi-bookmark-fill::before{content:"\f199"}.bi-bookmark-heart-fill::before{content:"\f19a"}.bi-bookmark-heart::before{content:"\f19b"}.bi-bookmark-plus-fill::before{content:"\f19c"}.bi-bookmark-plus::before{content:"\f19d"}.bi-bookmark-star-fill::before{content:"\f19e"}.bi-bookmark-star::before{content:"\f19f"}.bi-bookmark-x-fill::before{content:"\f1a0"}.bi-bookmark-x::before{content:"\f1a1"}.bi-bookmark::before{content:"\f1a2"}.bi-bookmarks-fill::before{content:"\f1a3"}.bi-bookmarks::before{content:"\f1a4"}.bi-bookshelf::before{content:"\f1a5"}.bi-bootstrap-fill::before{content:"\f1a6"}.bi-bootstrap-reboot::before{content:"\f1a7"}.bi-bootstrap::before{content:"\f1a8"}.bi-border-all::before{content:"\f1a9"}.bi-border-bottom::before{content:"\f1aa"}.bi-border-center::before{content:"\f1ab"}.bi-border-inner::before{content:"\f1ac"}.bi-border-left::before{content:"\f1ad"}.bi-border-middle::before{content:"\f1ae"}.bi-border-outer::before{content:"\f1af"}.bi-border-right::before{content:"\f1b0"}.bi-border-style::before{content:"\f1b1"}.bi-border-top::before{content:"\f1b2"}.bi-border-width::before{content:"\f1b3"}.bi-border::before{content:"\f1b4"}.bi-bounding-box-circles::before{content:"\f1b5"}.bi-bounding-box::before{content:"\f1b6"}.bi-box-arrow-down-left::before{content:"\f1b7"}.bi-box-arrow-down-right::before{content:"\f1b8"}.bi-box-arrow-down::before{content:"\f1b9"}.bi-box-arrow-in-down-left::before{content:"\f1ba"}.bi-box-arrow-in-down-right::before{content:"\f1bb"}.bi-box-arrow-in-down::before{content:"\f1bc"}.bi-box-arrow-in-left::before{content:"\f1bd"}.bi-box-arrow-in-right::before{content:"\f1be"}.bi-box-arrow-in-up-left::before{content:"\f1bf"}.bi-box-arrow-in-up-right::before{content:"\f1c0"}.bi-box-arrow-in-up::before{content:"\f1c1"}.bi-box-arrow-left::before{content:"\f1c2"}.bi-box-arrow-right::before{content:"\f1c3"}.bi-box-arrow-up-left::before{content:"\f1c4"}.bi-box-arrow-up-right::before{content:"\f1c5"}.bi-box-arrow-up::before{content:"\f1c6"}.bi-box-seam::before{content:"\f1c7"}.bi-box::before{content:"\f1c8"}.bi-braces::before{content:"\f1c9"}.bi-bricks::before{content:"\f1ca"}.bi-briefcase-fill::before{content:"\f1cb"}.bi-briefcase::before{content:"\f1cc"}.bi-brightness-alt-high-fill::before{content:"\f1cd"}.bi-brightness-alt-high::before{content:"\f1ce"}.bi-brightness-alt-low-fill::before{content:"\f1cf"}.bi-brightness-alt-low::before{content:"\f1d0"}.bi-brightness-high-fill::before{content:"\f1d1"}.bi-brightness-high::before{content:"\f1d2"}.bi-brightness-low-fill::before{content:"\f1d3"}.bi-brightness-low::before{content:"\f1d4"}.bi-broadcast-pin::before{content:"\f1d5"}.bi-broadcast::before{content:"\f1d6"}.bi-brush-fill::before{content:"\f1d7"}.bi-brush::before{content:"\f1d8"}.bi-bucket-fill::before{content:"\f1d9"}.bi-bucket::before{content:"\f1da"}.bi-bug-fill::before{content:"\f1db"}.bi-bug::before{content:"\f1dc"}.bi-building::before{content:"\f1dd"}.bi-bullseye::before{content:"\f1de"}.bi-calculator-fill::before{content:"\f1df"}.bi-calculator::before{content:"\f1e0"}.bi-calendar-check-fill::before{content:"\f1e1"}.bi-calendar-check::before{content:"\f1e2"}.bi-calendar-date-fill::before{content:"\f1e3"}.bi-calendar-date::before{content:"\f1e4"}.bi-calendar-day-fill::before{content:"\f1e5"}.bi-calendar-day::before{content:"\f1e6"}.bi-calendar-event-fill::before{content:"\f1e7"}.bi-calendar-event::before{content:"\f1e8"}.bi-calendar-fill::before{content:"\f1e9"}.bi-calendar-minus-fill::before{content:"\f1ea"}.bi-calendar-minus::before{content:"\f1eb"}.bi-calendar-month-fill::before{content:"\f1ec"}.bi-calendar-month::before{content:"\f1ed"}.bi-calendar-plus-fill::before{content:"\f1ee"}.bi-calendar-plus::before{content:"\f1ef"}.bi-calendar-range-fill::before{content:"\f1f0"}.bi-calendar-range::before{content:"\f1f1"}.bi-calendar-week-fill::before{content:"\f1f2"}.bi-calendar-week::before{content:"\f1f3"}.bi-calendar-x-fill::before{content:"\f1f4"}.bi-calendar-x::before{content:"\f1f5"}.bi-calendar::before{content:"\f1f6"}.bi-calendar2-check-fill::before{content:"\f1f7"}.bi-calendar2-check::before{content:"\f1f8"}.bi-calendar2-date-fill::before{content:"\f1f9"}.bi-calendar2-date::before{content:"\f1fa"}.bi-calendar2-day-fill::before{content:"\f1fb"}.bi-calendar2-day::before{content:"\f1fc"}.bi-calendar2-event-fill::before{content:"\f1fd"}.bi-calendar2-event::before{content:"\f1fe"}.bi-calendar2-fill::before{content:"\f1ff"}.bi-calendar2-minus-fill::before{content:"\f200"}.bi-calendar2-minus::before{content:"\f201"}.bi-calendar2-month-fill::before{content:"\f202"}.bi-calendar2-month::before{content:"\f203"}.bi-calendar2-plus-fill::before{content:"\f204"}.bi-calendar2-plus::before{content:"\f205"}.bi-calendar2-range-fill::before{content:"\f206"}.bi-calendar2-range::before{content:"\f207"}.bi-calendar2-week-fill::before{content:"\f208"}.bi-calendar2-week::before{content:"\f209"}.bi-calendar2-x-fill::before{content:"\f20a"}.bi-calendar2-x::before{content:"\f20b"}.bi-calendar2::before{content:"\f20c"}.bi-calendar3-event-fill::before{content:"\f20d"}.bi-calendar3-event::before{content:"\f20e"}.bi-calendar3-fill::before{content:"\f20f"}.bi-calendar3-range-fill::before{content:"\f210"}.bi-calendar3-range::before{content:"\f211"}.bi-calendar3-week-fill::before{content:"\f212"}.bi-calendar3-week::before{content:"\f213"}.bi-calendar3::before{content:"\f214"}.bi-calendar4-event::before{content:"\f215"}.bi-calendar4-range::before{content:"\f216"}.bi-calendar4-week::before{content:"\f217"}.bi-calendar4::before{content:"\f218"}.bi-camera-fill::before{content:"\f219"}.bi-camera-reels-fill::before{content:"\f21a"}.bi-camera-reels::before{content:"\f21b"}.bi-camera-video-fill::before{content:"\f21c"}.bi-camera-video-off-fill::before{content:"\f21d"}.bi-camera-video-off::before{content:"\f21e"}.bi-camera-video::before{content:"\f21f"}.bi-camera::before{content:"\f220"}.bi-camera2::before{content:"\f221"}.bi-capslock-fill::before{content:"\f222"}.bi-capslock::before{content:"\f223"}.bi-card-checklist::before{content:"\f224"}.bi-card-heading::before{content:"\f225"}.bi-card-image::before{content:"\f226"}.bi-card-list::before{content:"\f227"}.bi-card-text::before{content:"\f228"}.bi-caret-down-fill::before{content:"\f229"}.bi-caret-down-square-fill::before{content:"\f22a"}.bi-caret-down-square::before{content:"\f22b"}.bi-caret-down::before{content:"\f22c"}.bi-caret-left-fill::before{content:"\f22d"}.bi-caret-left-square-fill::before{content:"\f22e"}.bi-caret-left-square::before{content:"\f22f"}.bi-caret-left::before{content:"\f230"}.bi-caret-right-fill::before{content:"\f231"}.bi-caret-right-square-fill::before{content:"\f232"}.bi-caret-right-square::before{content:"\f233"}.bi-caret-right::before{content:"\f234"}.bi-caret-up-fill::before{content:"\f235"}.bi-caret-up-square-fill::before{content:"\f236"}.bi-caret-up-square::before{content:"\f237"}.bi-caret-up::before{content:"\f238"}.bi-cart-check-fill::before{content:"\f239"}.bi-cart-check::before{content:"\f23a"}.bi-cart-dash-fill::before{content:"\f23b"}.bi-cart-dash::before{content:"\f23c"}.bi-cart-fill::before{content:"\f23d"}.bi-cart-plus-fill::before{content:"\f23e"}.bi-cart-plus::before{content:"\f23f"}.bi-cart-x-fill::before{content:"\f240"}.bi-cart-x::before{content:"\f241"}.bi-cart::before{content:"\f242"}.bi-cart2::before{content:"\f243"}.bi-cart3::before{content:"\f244"}.bi-cart4::before{content:"\f245"}.bi-cash-stack::before{content:"\f246"}.bi-cash::before{content:"\f247"}.bi-cast::before{content:"\f248"}.bi-chat-dots-fill::before{content:"\f249"}.bi-chat-dots::before{content:"\f24a"}.bi-chat-fill::before{content:"\f24b"}.bi-chat-left-dots-fill::before{content:"\f24c"}.bi-chat-left-dots::before{content:"\f24d"}.bi-chat-left-fill::before{content:"\f24e"}.bi-chat-left-quote-fill::before{content:"\f24f"}.bi-chat-left-quote::before{content:"\f250"}.bi-chat-left-text-fill::before{content:"\f251"}.bi-chat-left-text::before{content:"\f252"}.bi-chat-left::before{content:"\f253"}.bi-chat-quote-fill::before{content:"\f254"}.bi-chat-quote::before{content:"\f255"}.bi-chat-right-dots-fill::before{content:"\f256"}.bi-chat-right-dots::before{content:"\f257"}.bi-chat-right-fill::before{content:"\f258"}.bi-chat-right-quote-fill::before{content:"\f259"}.bi-chat-right-quote::before{content:"\f25a"}.bi-chat-right-text-fill::before{content:"\f25b"}.bi-chat-right-text::before{content:"\f25c"}.bi-chat-right::before{content:"\f25d"}.bi-chat-square-dots-fill::before{content:"\f25e"}.bi-chat-square-dots::before{content:"\f25f"}.bi-chat-square-fill::before{content:"\f260"}.bi-chat-square-quote-fill::before{content:"\f261"}.bi-chat-square-quote::before{content:"\f262"}.bi-chat-square-text-fill::before{content:"\f263"}.bi-chat-square-text::before{content:"\f264"}.bi-chat-square::before{content:"\f265"}.bi-chat-text-fill::before{content:"\f266"}.bi-chat-text::before{content:"\f267"}.bi-chat::before{content:"\f268"}.bi-check-all::before{content:"\f269"}.bi-check-circle-fill::before{content:"\f26a"}.bi-check-circle::before{content:"\f26b"}.bi-check-square-fill::before{content:"\f26c"}.bi-check-square::before{content:"\f26d"}.bi-check::before{content:"\f26e"}.bi-check2-all::before{content:"\f26f"}.bi-check2-circle::before{content:"\f270"}.bi-check2-square::before{content:"\f271"}.bi-check2::before{content:"\f272"}.bi-chevron-bar-contract::before{content:"\f273"}.bi-chevron-bar-down::before{content:"\f274"}.bi-chevron-bar-expand::before{content:"\f275"}.bi-chevron-bar-left::before{content:"\f276"}.bi-chevron-bar-right::before{content:"\f277"}.bi-chevron-bar-up::before{content:"\f278"}.bi-chevron-compact-down::before{content:"\f279"}.bi-chevron-compact-left::before{content:"\f27a"}.bi-chevron-compact-right::before{content:"\f27b"}.bi-chevron-compact-up::before{content:"\f27c"}.bi-chevron-contract::before{content:"\f27d"}.bi-chevron-double-down::before{content:"\f27e"}.bi-chevron-double-left::before{content:"\f27f"}.bi-chevron-double-right::before{content:"\f280"}.bi-chevron-double-up::before{content:"\f281"}.bi-chevron-down::before{content:"\f282"}.bi-chevron-expand::before{content:"\f283"}.bi-chevron-left::before{content:"\f284"}.bi-chevron-right::before{content:"\f285"}.bi-chevron-up::before{content:"\f286"}.bi-circle-fill::before{content:"\f287"}.bi-circle-half::before{content:"\f288"}.bi-circle-square::before{content:"\f289"}.bi-circle::before{content:"\f28a"}.bi-clipboard-check::before{content:"\f28b"}.bi-clipboard-data::before{content:"\f28c"}.bi-clipboard-minus::before{content:"\f28d"}.bi-clipboard-plus::before{content:"\f28e"}.bi-clipboard-x::before{content:"\f28f"}.bi-clipboard::before{content:"\f290"}.bi-clock-fill::before{content:"\f291"}.bi-clock-history::before{content:"\f292"}.bi-clock::before{content:"\f293"}.bi-cloud-arrow-down-fill::before{content:"\f294"}.bi-cloud-arrow-down::before{content:"\f295"}.bi-cloud-arrow-up-fill::before{content:"\f296"}.bi-cloud-arrow-up::before{content:"\f297"}.bi-cloud-check-fill::before{content:"\f298"}.bi-cloud-check::before{content:"\f299"}.bi-cloud-download-fill::before{content:"\f29a"}.bi-cloud-download::before{content:"\f29b"}.bi-cloud-drizzle-fill::before{content:"\f29c"}.bi-cloud-drizzle::before{content:"\f29d"}.bi-cloud-fill::before{content:"\f29e"}.bi-cloud-fog-fill::before{content:"\f29f"}.bi-cloud-fog::before{content:"\f2a0"}.bi-cloud-fog2-fill::before{content:"\f2a1"}.bi-cloud-fog2::before{content:"\f2a2"}.bi-cloud-hail-fill::before{content:"\f2a3"}.bi-cloud-hail::before{content:"\f2a4"}.bi-cloud-haze-fill::before{content:"\f2a6"}.bi-cloud-haze::before{content:"\f2a7"}.bi-cloud-haze2-fill::before{content:"\f2a8"}.bi-cloud-lightning-fill::before{content:"\f2a9"}.bi-cloud-lightning-rain-fill::before{content:"\f2aa"}.bi-cloud-lightning-rain::before{content:"\f2ab"}.bi-cloud-lightning::before{content:"\f2ac"}.bi-cloud-minus-fill::before{content:"\f2ad"}.bi-cloud-minus::before{content:"\f2ae"}.bi-cloud-moon-fill::before{content:"\f2af"}.bi-cloud-moon::before{content:"\f2b0"}.bi-cloud-plus-fill::before{content:"\f2b1"}.bi-cloud-plus::before{content:"\f2b2"}.bi-cloud-rain-fill::before{content:"\f2b3"}.bi-cloud-rain-heavy-fill::before{content:"\f2b4"}.bi-cloud-rain-heavy::before{content:"\f2b5"}.bi-cloud-rain::before{content:"\f2b6"}.bi-cloud-slash-fill::before{content:"\f2b7"}.bi-cloud-slash::before{content:"\f2b8"}.bi-cloud-sleet-fill::before{content:"\f2b9"}.bi-cloud-sleet::before{content:"\f2ba"}.bi-cloud-snow-fill::before{content:"\f2bb"}.bi-cloud-snow::before{content:"\f2bc"}.bi-cloud-sun-fill::before{content:"\f2bd"}.bi-cloud-sun::before{content:"\f2be"}.bi-cloud-upload-fill::before{content:"\f2bf"}.bi-cloud-upload::before{content:"\f2c0"}.bi-cloud::before{content:"\f2c1"}.bi-clouds-fill::before{content:"\f2c2"}.bi-clouds::before{content:"\f2c3"}.bi-cloudy-fill::before{content:"\f2c4"}.bi-cloudy::before{content:"\f2c5"}.bi-code-slash::before{content:"\f2c6"}.bi-code-square::before{content:"\f2c7"}.bi-code::before{content:"\f2c8"}.bi-collection-fill::before{content:"\f2c9"}.bi-collection-play-fill::before{content:"\f2ca"}.bi-collection-play::before{content:"\f2cb"}.bi-collection::before{content:"\f2cc"}.bi-columns-gap::before{content:"\f2cd"}.bi-columns::before{content:"\f2ce"}.bi-command::before{content:"\f2cf"}.bi-compass-fill::before{content:"\f2d0"}.bi-compass::before{content:"\f2d1"}.bi-cone-striped::before{content:"\f2d2"}.bi-cone::before{content:"\f2d3"}.bi-controller::before{content:"\f2d4"}.bi-cpu-fill::before{content:"\f2d5"}.bi-cpu::before{content:"\f2d6"}.bi-credit-card-2-back-fill::before{content:"\f2d7"}.bi-credit-card-2-back::before{content:"\f2d8"}.bi-credit-card-2-front-fill::before{content:"\f2d9"}.bi-credit-card-2-front::before{content:"\f2da"}.bi-credit-card-fill::before{content:"\f2db"}.bi-credit-card::before{content:"\f2dc"}.bi-crop::before{content:"\f2dd"}.bi-cup-fill::before{content:"\f2de"}.bi-cup-straw::before{content:"\f2df"}.bi-cup::before{content:"\f2e0"}.bi-cursor-fill::before{content:"\f2e1"}.bi-cursor-text::before{content:"\f2e2"}.bi-cursor::before{content:"\f2e3"}.bi-dash-circle-dotted::before{content:"\f2e4"}.bi-dash-circle-fill::before{content:"\f2e5"}.bi-dash-circle::before{content:"\f2e6"}.bi-dash-square-dotted::before{content:"\f2e7"}.bi-dash-square-fill::before{content:"\f2e8"}.bi-dash-square::before{content:"\f2e9"}.bi-dash::before{content:"\f2ea"}.bi-diagram-2-fill::before{content:"\f2eb"}.bi-diagram-2::before{content:"\f2ec"}.bi-diagram-3-fill::before{content:"\f2ed"}.bi-diagram-3::before{content:"\f2ee"}.bi-diamond-fill::before{content:"\f2ef"}.bi-diamond-half::before{content:"\f2f0"}.bi-diamond::before{content:"\f2f1"}.bi-dice-1-fill::before{content:"\f2f2"}.bi-dice-1::before{content:"\f2f3"}.bi-dice-2-fill::before{content:"\f2f4"}.bi-dice-2::before{content:"\f2f5"}.bi-dice-3-fill::before{content:"\f2f6"}.bi-dice-3::before{content:"\f2f7"}.bi-dice-4-fill::before{content:"\f2f8"}.bi-dice-4::before{content:"\f2f9"}.bi-dice-5-fill::before{content:"\f2fa"}.bi-dice-5::before{content:"\f2fb"}.bi-dice-6-fill::before{content:"\f2fc"}.bi-dice-6::before{content:"\f2fd"}.bi-disc-fill::before{content:"\f2fe"}.bi-disc::before{content:"\f2ff"}.bi-discord::before{content:"\f300"}.bi-display-fill::before{content:"\f301"}.bi-display::before{content:"\f302"}.bi-distribute-horizontal::before{content:"\f303"}.bi-distribute-vertical::before{content:"\f304"}.bi-door-closed-fill::before{content:"\f305"}.bi-door-closed::before{content:"\f306"}.bi-door-open-fill::before{content:"\f307"}.bi-door-open::before{content:"\f308"}.bi-dot::before{content:"\f309"}.bi-download::before{content:"\f30a"}.bi-droplet-fill::before{content:"\f30b"}.bi-droplet-half::before{content:"\f30c"}.bi-droplet::before{content:"\f30d"}.bi-earbuds::before{content:"\f30e"}.bi-easel-fill::before{content:"\f30f"}.bi-easel::before{content:"\f310"}.bi-egg-fill::before{content:"\f311"}.bi-egg-fried::before{content:"\f312"}.bi-egg::before{content:"\f313"}.bi-eject-fill::before{content:"\f314"}.bi-eject::before{content:"\f315"}.bi-emoji-angry-fill::before{content:"\f316"}.bi-emoji-angry::before{content:"\f317"}.bi-emoji-dizzy-fill::before{content:"\f318"}.bi-emoji-dizzy::before{content:"\f319"}.bi-emoji-expressionless-fill::before{content:"\f31a"}.bi-emoji-expressionless::before{content:"\f31b"}.bi-emoji-frown-fill::before{content:"\f31c"}.bi-emoji-frown::before{content:"\f31d"}.bi-emoji-heart-eyes-fill::before{content:"\f31e"}.bi-emoji-heart-eyes::before{content:"\f31f"}.bi-emoji-laughing-fill::before{content:"\f320"}.bi-emoji-laughing::before{content:"\f321"}.bi-emoji-neutral-fill::before{content:"\f322"}.bi-emoji-neutral::before{content:"\f323"}.bi-emoji-smile-fill::before{content:"\f324"}.bi-emoji-smile-upside-down-fill::before{content:"\f325"}.bi-emoji-smile-upside-down::before{content:"\f326"}.bi-emoji-smile::before{content:"\f327"}.bi-emoji-sunglasses-fill::before{content:"\f328"}.bi-emoji-sunglasses::before{content:"\f329"}.bi-emoji-wink-fill::before{content:"\f32a"}.bi-emoji-wink::before{content:"\f32b"}.bi-envelope-fill::before{content:"\f32c"}.bi-envelope-open-fill::before{content:"\f32d"}.bi-envelope-open::before{content:"\f32e"}.bi-envelope::before{content:"\f32f"}.bi-eraser-fill::before{content:"\f330"}.bi-eraser::before{content:"\f331"}.bi-exclamation-circle-fill::before{content:"\f332"}.bi-exclamation-circle::before{content:"\f333"}.bi-exclamation-diamond-fill::before{content:"\f334"}.bi-exclamation-diamond::before{content:"\f335"}.bi-exclamation-octagon-fill::before{content:"\f336"}.bi-exclamation-octagon::before{content:"\f337"}.bi-exclamation-square-fill::before{content:"\f338"}.bi-exclamation-square::before{content:"\f339"}.bi-exclamation-triangle-fill::before{content:"\f33a"}.bi-exclamation-triangle::before{content:"\f33b"}.bi-exclamation::before{content:"\f33c"}.bi-exclude::before{content:"\f33d"}.bi-eye-fill::before{content:"\f33e"}.bi-eye-slash-fill::before{content:"\f33f"}.bi-eye-slash::before{content:"\f340"}.bi-eye::before{content:"\f341"}.bi-eyedropper::before{content:"\f342"}.bi-eyeglasses::before{content:"\f343"}.bi-facebook::before{content:"\f344"}.bi-file-arrow-down-fill::before{content:"\f345"}.bi-file-arrow-down::before{content:"\f346"}.bi-file-arrow-up-fill::before{content:"\f347"}.bi-file-arrow-up::before{content:"\f348"}.bi-file-bar-graph-fill::before{content:"\f349"}.bi-file-bar-graph::before{content:"\f34a"}.bi-file-binary-fill::before{content:"\f34b"}.bi-file-binary::before{content:"\f34c"}.bi-file-break-fill::before{content:"\f34d"}.bi-file-break::before{content:"\f34e"}.bi-file-check-fill::before{content:"\f34f"}.bi-file-check::before{content:"\f350"}.bi-file-code-fill::before{content:"\f351"}.bi-file-code::before{content:"\f352"}.bi-file-diff-fill::before{content:"\f353"}.bi-file-diff::before{content:"\f354"}.bi-file-earmark-arrow-down-fill::before{content:"\f355"}.bi-file-earmark-arrow-down::before{content:"\f356"}.bi-file-earmark-arrow-up-fill::before{content:"\f357"}.bi-file-earmark-arrow-up::before{content:"\f358"}.bi-file-earmark-bar-graph-fill::before{content:"\f359"}.bi-file-earmark-bar-graph::before{content:"\f35a"}.bi-file-earmark-binary-fill::before{content:"\f35b"}.bi-file-earmark-binary::before{content:"\f35c"}.bi-file-earmark-break-fill::before{content:"\f35d"}.bi-file-earmark-break::before{content:"\f35e"}.bi-file-earmark-check-fill::before{content:"\f35f"}.bi-file-earmark-check::before{content:"\f360"}.bi-file-earmark-code-fill::before{content:"\f361"}.bi-file-earmark-code::before{content:"\f362"}.bi-file-earmark-diff-fill::before{content:"\f363"}.bi-file-earmark-diff::before{content:"\f364"}.bi-file-earmark-easel-fill::before{content:"\f365"}.bi-file-earmark-easel::before{content:"\f366"}.bi-file-earmark-excel-fill::before{content:"\f367"}.bi-file-earmark-excel::before{content:"\f368"}.bi-file-earmark-fill::before{content:"\f369"}.bi-file-earmark-font-fill::before{content:"\f36a"}.bi-file-earmark-font::before{content:"\f36b"}.bi-file-earmark-image-fill::before{content:"\f36c"}.bi-file-earmark-image::before{content:"\f36d"}.bi-file-earmark-lock-fill::before{content:"\f36e"}.bi-file-earmark-lock::before{content:"\f36f"}.bi-file-earmark-lock2-fill::before{content:"\f370"}.bi-file-earmark-lock2::before{content:"\f371"}.bi-file-earmark-medical-fill::before{content:"\f372"}.bi-file-earmark-medical::before{content:"\f373"}.bi-file-earmark-minus-fill::before{content:"\f374"}.bi-file-earmark-minus::before{content:"\f375"}.bi-file-earmark-music-fill::before{content:"\f376"}.bi-file-earmark-music::before{content:"\f377"}.bi-file-earmark-person-fill::before{content:"\f378"}.bi-file-earmark-person::before{content:"\f379"}.bi-file-earmark-play-fill::before{content:"\f37a"}.bi-file-earmark-play::before{content:"\f37b"}.bi-file-earmark-plus-fill::before{content:"\f37c"}.bi-file-earmark-plus::before{content:"\f37d"}.bi-file-earmark-post-fill::before{content:"\f37e"}.bi-file-earmark-post::before{content:"\f37f"}.bi-file-earmark-ppt-fill::before{content:"\f380"}.bi-file-earmark-ppt::before{content:"\f381"}.bi-file-earmark-richtext-fill::before{content:"\f382"}.bi-file-earmark-richtext::before{content:"\f383"}.bi-file-earmark-ruled-fill::before{content:"\f384"}.bi-file-earmark-ruled::before{content:"\f385"}.bi-file-earmark-slides-fill::before{content:"\f386"}.bi-file-earmark-slides::before{content:"\f387"}.bi-file-earmark-spreadsheet-fill::before{content:"\f388"}.bi-file-earmark-spreadsheet::before{content:"\f389"}.bi-file-earmark-text-fill::before{content:"\f38a"}.bi-file-earmark-text::before{content:"\f38b"}.bi-file-earmark-word-fill::before{content:"\f38c"}.bi-file-earmark-word::before{content:"\f38d"}.bi-file-earmark-x-fill::before{content:"\f38e"}.bi-file-earmark-x::before{content:"\f38f"}.bi-file-earmark-zip-fill::before{content:"\f390"}.bi-file-earmark-zip::before{content:"\f391"}.bi-file-earmark::before{content:"\f392"}.bi-file-easel-fill::before{content:"\f393"}.bi-file-easel::before{content:"\f394"}.bi-file-excel-fill::before{content:"\f395"}.bi-file-excel::before{content:"\f396"}.bi-file-fill::before{content:"\f397"}.bi-file-font-fill::before{content:"\f398"}.bi-file-font::before{content:"\f399"}.bi-file-image-fill::before{content:"\f39a"}.bi-file-image::before{content:"\f39b"}.bi-file-lock-fill::before{content:"\f39c"}.bi-file-lock::before{content:"\f39d"}.bi-file-lock2-fill::before{content:"\f39e"}.bi-file-lock2::before{content:"\f39f"}.bi-file-medical-fill::before{content:"\f3a0"}.bi-file-medical::before{content:"\f3a1"}.bi-file-minus-fill::before{content:"\f3a2"}.bi-file-minus::before{content:"\f3a3"}.bi-file-music-fill::before{content:"\f3a4"}.bi-file-music::before{content:"\f3a5"}.bi-file-person-fill::before{content:"\f3a6"}.bi-file-person::before{content:"\f3a7"}.bi-file-play-fill::before{content:"\f3a8"}.bi-file-play::before{content:"\f3a9"}.bi-file-plus-fill::before{content:"\f3aa"}.bi-file-plus::before{content:"\f3ab"}.bi-file-post-fill::before{content:"\f3ac"}.bi-file-post::before{content:"\f3ad"}.bi-file-ppt-fill::before{content:"\f3ae"}.bi-file-ppt::before{content:"\f3af"}.bi-file-richtext-fill::before{content:"\f3b0"}.bi-file-richtext::before{content:"\f3b1"}.bi-file-ruled-fill::before{content:"\f3b2"}.bi-file-ruled::before{content:"\f3b3"}.bi-file-slides-fill::before{content:"\f3b4"}.bi-file-slides::before{content:"\f3b5"}.bi-file-spreadsheet-fill::before{content:"\f3b6"}.bi-file-spreadsheet::before{content:"\f3b7"}.bi-file-text-fill::before{content:"\f3b8"}.bi-file-text::before{content:"\f3b9"}.bi-file-word-fill::before{content:"\f3ba"}.bi-file-word::before{content:"\f3bb"}.bi-file-x-fill::before{content:"\f3bc"}.bi-file-x::before{content:"\f3bd"}.bi-file-zip-fill::before{content:"\f3be"}.bi-file-zip::before{content:"\f3bf"}.bi-file::before{content:"\f3c0"}.bi-files-alt::before{content:"\f3c1"}.bi-files::before{content:"\f3c2"}.bi-film::before{content:"\f3c3"}.bi-filter-circle-fill::before{content:"\f3c4"}.bi-filter-circle::before{content:"\f3c5"}.bi-filter-left::before{content:"\f3c6"}.bi-filter-right::before{content:"\f3c7"}.bi-filter-square-fill::before{content:"\f3c8"}.bi-filter-square::before{content:"\f3c9"}.bi-filter::before{content:"\f3ca"}.bi-flag-fill::before{content:"\f3cb"}.bi-flag::before{content:"\f3cc"}.bi-flower1::before{content:"\f3cd"}.bi-flower2::before{content:"\f3ce"}.bi-flower3::before{content:"\f3cf"}.bi-folder-check::before{content:"\f3d0"}.bi-folder-fill::before{content:"\f3d1"}.bi-folder-minus::before{content:"\f3d2"}.bi-folder-plus::before{content:"\f3d3"}.bi-folder-symlink-fill::before{content:"\f3d4"}.bi-folder-symlink::before{content:"\f3d5"}.bi-folder-x::before{content:"\f3d6"}.bi-folder::before{content:"\f3d7"}.bi-folder2-open::before{content:"\f3d8"}.bi-folder2::before{content:"\f3d9"}.bi-fonts::before{content:"\f3da"}.bi-forward-fill::before{content:"\f3db"}.bi-forward::before{content:"\f3dc"}.bi-front::before{content:"\f3dd"}.bi-fullscreen-exit::before{content:"\f3de"}.bi-fullscreen::before{content:"\f3df"}.bi-funnel-fill::before{content:"\f3e0"}.bi-funnel::before{content:"\f3e1"}.bi-gear-fill::before{content:"\f3e2"}.bi-gear-wide-connected::before{content:"\f3e3"}.bi-gear-wide::before{content:"\f3e4"}.bi-gear::before{content:"\f3e5"}.bi-gem::before{content:"\f3e6"}.bi-geo-alt-fill::before{content:"\f3e7"}.bi-geo-alt::before{content:"\f3e8"}.bi-geo-fill::before{content:"\f3e9"}.bi-geo::before{content:"\f3ea"}.bi-gift-fill::before{content:"\f3eb"}.bi-gift::before{content:"\f3ec"}.bi-github::before{content:"\f3ed"}.bi-globe::before{content:"\f3ee"}.bi-globe2::before{content:"\f3ef"}.bi-google::before{content:"\f3f0"}.bi-graph-down::before{content:"\f3f1"}.bi-graph-up::before{content:"\f3f2"}.bi-grid-1x2-fill::before{content:"\f3f3"}.bi-grid-1x2::before{content:"\f3f4"}.bi-grid-3x2-gap-fill::before{content:"\f3f5"}.bi-grid-3x2-gap::before{content:"\f3f6"}.bi-grid-3x2::before{content:"\f3f7"}.bi-grid-3x3-gap-fill::before{content:"\f3f8"}.bi-grid-3x3-gap::before{content:"\f3f9"}.bi-grid-3x3::before{content:"\f3fa"}.bi-grid-fill::before{content:"\f3fb"}.bi-grid::before{content:"\f3fc"}.bi-grip-horizontal::before{content:"\f3fd"}.bi-grip-vertical::before{content:"\f3fe"}.bi-hammer::before{content:"\f3ff"}.bi-hand-index-fill::before{content:"\f400"}.bi-hand-index-thumb-fill::before{content:"\f401"}.bi-hand-index-thumb::before{content:"\f402"}.bi-hand-index::before{content:"\f403"}.bi-hand-thumbs-down-fill::before{content:"\f404"}.bi-hand-thumbs-down::before{content:"\f405"}.bi-hand-thumbs-up-fill::before{content:"\f406"}.bi-hand-thumbs-up::before{content:"\f407"}.bi-handbag-fill::before{content:"\f408"}.bi-handbag::before{content:"\f409"}.bi-hash::before{content:"\f40a"}.bi-hdd-fill::before{content:"\f40b"}.bi-hdd-network-fill::before{content:"\f40c"}.bi-hdd-network::before{content:"\f40d"}.bi-hdd-rack-fill::before{content:"\f40e"}.bi-hdd-rack::before{content:"\f40f"}.bi-hdd-stack-fill::before{content:"\f410"}.bi-hdd-stack::before{content:"\f411"}.bi-hdd::before{content:"\f412"}.bi-headphones::before{content:"\f413"}.bi-headset::before{content:"\f414"}.bi-heart-fill::before{content:"\f415"}.bi-heart-half::before{content:"\f416"}.bi-heart::before{content:"\f417"}.bi-heptagon-fill::before{content:"\f418"}.bi-heptagon-half::before{content:"\f419"}.bi-heptagon::before{content:"\f41a"}.bi-hexagon-fill::before{content:"\f41b"}.bi-hexagon-half::before{content:"\f41c"}.bi-hexagon::before{content:"\f41d"}.bi-hourglass-bottom::before{content:"\f41e"}.bi-hourglass-split::before{content:"\f41f"}.bi-hourglass-top::before{content:"\f420"}.bi-hourglass::before{content:"\f421"}.bi-house-door-fill::before{content:"\f422"}.bi-house-door::before{content:"\f423"}.bi-house-fill::before{content:"\f424"}.bi-house::before{content:"\f425"}.bi-hr::before{content:"\f426"}.bi-hurricane::before{content:"\f427"}.bi-image-alt::before{content:"\f428"}.bi-image-fill::before{content:"\f429"}.bi-image::before{content:"\f42a"}.bi-images::before{content:"\f42b"}.bi-inbox-fill::before{content:"\f42c"}.bi-inbox::before{content:"\f42d"}.bi-inboxes-fill::before{content:"\f42e"}.bi-inboxes::before{content:"\f42f"}.bi-info-circle-fill::before{content:"\f430"}.bi-info-circle::before{content:"\f431"}.bi-info-square-fill::before{content:"\f432"}.bi-info-square::before{content:"\f433"}.bi-info::before{content:"\f434"}.bi-input-cursor-text::before{content:"\f435"}.bi-input-cursor::before{content:"\f436"}.bi-instagram::before{content:"\f437"}.bi-intersect::before{content:"\f438"}.bi-journal-album::before{content:"\f439"}.bi-journal-arrow-down::before{content:"\f43a"}.bi-journal-arrow-up::before{content:"\f43b"}.bi-journal-bookmark-fill::before{content:"\f43c"}.bi-journal-bookmark::before{content:"\f43d"}.bi-journal-check::before{content:"\f43e"}.bi-journal-code::before{content:"\f43f"}.bi-journal-medical::before{content:"\f440"}.bi-journal-minus::before{content:"\f441"}.bi-journal-plus::before{content:"\f442"}.bi-journal-richtext::before{content:"\f443"}.bi-journal-text::before{content:"\f444"}.bi-journal-x::before{content:"\f445"}.bi-journal::before{content:"\f446"}.bi-journals::before{content:"\f447"}.bi-joystick::before{content:"\f448"}.bi-justify-left::before{content:"\f449"}.bi-justify-right::before{content:"\f44a"}.bi-justify::before{content:"\f44b"}.bi-kanban-fill::before{content:"\f44c"}.bi-kanban::before{content:"\f44d"}.bi-key-fill::before{content:"\f44e"}.bi-key::before{content:"\f44f"}.bi-keyboard-fill::before{content:"\f450"}.bi-keyboard::before{content:"\f451"}.bi-ladder::before{content:"\f452"}.bi-lamp-fill::before{content:"\f453"}.bi-lamp::before{content:"\f454"}.bi-laptop-fill::before{content:"\f455"}.bi-laptop::before{content:"\f456"}.bi-layer-backward::before{content:"\f457"}.bi-layer-forward::before{content:"\f458"}.bi-layers-fill::before{content:"\f459"}.bi-layers-half::before{content:"\f45a"}.bi-layers::before{content:"\f45b"}.bi-layout-sidebar-inset-reverse::before{content:"\f45c"}.bi-layout-sidebar-inset::before{content:"\f45d"}.bi-layout-sidebar-reverse::before{content:"\f45e"}.bi-layout-sidebar::before{content:"\f45f"}.bi-layout-split::before{content:"\f460"}.bi-layout-text-sidebar-reverse::before{content:"\f461"}.bi-layout-text-sidebar::before{content:"\f462"}.bi-layout-text-window-reverse::before{content:"\f463"}.bi-layout-text-window::before{content:"\f464"}.bi-layout-three-columns::before{content:"\f465"}.bi-layout-wtf::before{content:"\f466"}.bi-life-preserver::before{content:"\f467"}.bi-lightbulb-fill::before{content:"\f468"}.bi-lightbulb-off-fill::before{content:"\f469"}.bi-lightbulb-off::before{content:"\f46a"}.bi-lightbulb::before{content:"\f46b"}.bi-lightning-charge-fill::before{content:"\f46c"}.bi-lightning-charge::before{content:"\f46d"}.bi-lightning-fill::before{content:"\f46e"}.bi-lightning::before{content:"\f46f"}.bi-link-45deg::before{content:"\f470"}.bi-link::before{content:"\f471"}.bi-linkedin::before{content:"\f472"}.bi-list-check::before{content:"\f473"}.bi-list-nested::before{content:"\f474"}.bi-list-ol::before{content:"\f475"}.bi-list-stars::before{content:"\f476"}.bi-list-task::before{content:"\f477"}.bi-list-ul::before{content:"\f478"}.bi-list::before{content:"\f479"}.bi-lock-fill::before{content:"\f47a"}.bi-lock::before{content:"\f47b"}.bi-mailbox::before{content:"\f47c"}.bi-mailbox2::before{content:"\f47d"}.bi-map-fill::before{content:"\f47e"}.bi-map::before{content:"\f47f"}.bi-markdown-fill::before{content:"\f480"}.bi-markdown::before{content:"\f481"}.bi-mask::before{content:"\f482"}.bi-megaphone-fill::before{content:"\f483"}.bi-megaphone::before{content:"\f484"}.bi-menu-app-fill::before{content:"\f485"}.bi-menu-app::before{content:"\f486"}.bi-menu-button-fill::before{content:"\f487"}.bi-menu-button-wide-fill::before{content:"\f488"}.bi-menu-button-wide::before{content:"\f489"}.bi-menu-button::before{content:"\f48a"}.bi-menu-down::before{content:"\f48b"}.bi-menu-up::before{content:"\f48c"}.bi-mic-fill::before{content:"\f48d"}.bi-mic-mute-fill::before{content:"\f48e"}.bi-mic-mute::before{content:"\f48f"}.bi-mic::before{content:"\f490"}.bi-minecart-loaded::before{content:"\f491"}.bi-minecart::before{content:"\f492"}.bi-moisture::before{content:"\f493"}.bi-moon-fill::before{content:"\f494"}.bi-moon-stars-fill::before{content:"\f495"}.bi-moon-stars::before{content:"\f496"}.bi-moon::before{content:"\f497"}.bi-mouse-fill::before{content:"\f498"}.bi-mouse::before{content:"\f499"}.bi-mouse2-fill::before{content:"\f49a"}.bi-mouse2::before{content:"\f49b"}.bi-mouse3-fill::before{content:"\f49c"}.bi-mouse3::before{content:"\f49d"}.bi-music-note-beamed::before{content:"\f49e"}.bi-music-note-list::before{content:"\f49f"}.bi-music-note::before{content:"\f4a0"}.bi-music-player-fill::before{content:"\f4a1"}.bi-music-player::before{content:"\f4a2"}.bi-newspaper::before{content:"\f4a3"}.bi-node-minus-fill::before{content:"\f4a4"}.bi-node-minus::before{content:"\f4a5"}.bi-node-plus-fill::before{content:"\f4a6"}.bi-node-plus::before{content:"\f4a7"}.bi-nut-fill::before{content:"\f4a8"}.bi-nut::before{content:"\f4a9"}.bi-octagon-fill::before{content:"\f4aa"}.bi-octagon-half::before{content:"\f4ab"}.bi-octagon::before{content:"\f4ac"}.bi-option::before{content:"\f4ad"}.bi-outlet::before{content:"\f4ae"}.bi-paint-bucket::before{content:"\f4af"}.bi-palette-fill::before{content:"\f4b0"}.bi-palette::before{content:"\f4b1"}.bi-palette2::before{content:"\f4b2"}.bi-paperclip::before{content:"\f4b3"}.bi-paragraph::before{content:"\f4b4"}.bi-patch-check-fill::before{content:"\f4b5"}.bi-patch-check::before{content:"\f4b6"}.bi-patch-exclamation-fill::before{content:"\f4b7"}.bi-patch-exclamation::before{content:"\f4b8"}.bi-patch-minus-fill::before{content:"\f4b9"}.bi-patch-minus::before{content:"\f4ba"}.bi-patch-plus-fill::before{content:"\f4bb"}.bi-patch-plus::before{content:"\f4bc"}.bi-patch-question-fill::before{content:"\f4bd"}.bi-patch-question::before{content:"\f4be"}.bi-pause-btn-fill::before{content:"\f4bf"}.bi-pause-btn::before{content:"\f4c0"}.bi-pause-circle-fill::before{content:"\f4c1"}.bi-pause-circle::before{content:"\f4c2"}.bi-pause-fill::before{content:"\f4c3"}.bi-pause::before{content:"\f4c4"}.bi-peace-fill::before{content:"\f4c5"}.bi-peace::before{content:"\f4c6"}.bi-pen-fill::before{content:"\f4c7"}.bi-pen::before{content:"\f4c8"}.bi-pencil-fill::before{content:"\f4c9"}.bi-pencil-square::before{content:"\f4ca"}.bi-pencil::before{content:"\f4cb"}.bi-pentagon-fill::before{content:"\f4cc"}.bi-pentagon-half::before{content:"\f4cd"}.bi-pentagon::before{content:"\f4ce"}.bi-people-fill::before{content:"\f4cf"}.bi-people::before{content:"\f4d0"}.bi-percent::before{content:"\f4d1"}.bi-person-badge-fill::before{content:"\f4d2"}.bi-person-badge::before{content:"\f4d3"}.bi-person-bounding-box::before{content:"\f4d4"}.bi-person-check-fill::before{content:"\f4d5"}.bi-person-check::before{content:"\f4d6"}.bi-person-circle::before{content:"\f4d7"}.bi-person-dash-fill::before{content:"\f4d8"}.bi-person-dash::before{content:"\f4d9"}.bi-person-fill::before{content:"\f4da"}.bi-person-lines-fill::before{content:"\f4db"}.bi-person-plus-fill::before{content:"\f4dc"}.bi-person-plus::before{content:"\f4dd"}.bi-person-square::before{content:"\f4de"}.bi-person-x-fill::before{content:"\f4df"}.bi-person-x::before{content:"\f4e0"}.bi-person::before{content:"\f4e1"}.bi-phone-fill::before{content:"\f4e2"}.bi-phone-landscape-fill::before{content:"\f4e3"}.bi-phone-landscape::before{content:"\f4e4"}.bi-phone-vibrate-fill::before{content:"\f4e5"}.bi-phone-vibrate::before{content:"\f4e6"}.bi-phone::before{content:"\f4e7"}.bi-pie-chart-fill::before{content:"\f4e8"}.bi-pie-chart::before{content:"\f4e9"}.bi-pin-angle-fill::before{content:"\f4ea"}.bi-pin-angle::before{content:"\f4eb"}.bi-pin-fill::before{content:"\f4ec"}.bi-pin::before{content:"\f4ed"}.bi-pip-fill::before{content:"\f4ee"}.bi-pip::before{content:"\f4ef"}.bi-play-btn-fill::before{content:"\f4f0"}.bi-play-btn::before{content:"\f4f1"}.bi-play-circle-fill::before{content:"\f4f2"}.bi-play-circle::before{content:"\f4f3"}.bi-play-fill::before{content:"\f4f4"}.bi-play::before{content:"\f4f5"}.bi-plug-fill::before{content:"\f4f6"}.bi-plug::before{content:"\f4f7"}.bi-plus-circle-dotted::before{content:"\f4f8"}.bi-plus-circle-fill::before{content:"\f4f9"}.bi-plus-circle::before{content:"\f4fa"}.bi-plus-square-dotted::before{content:"\f4fb"}.bi-plus-square-fill::before{content:"\f4fc"}.bi-plus-square::before{content:"\f4fd"}.bi-plus::before{content:"\f4fe"}.bi-power::before{content:"\f4ff"}.bi-printer-fill::before{content:"\f500"}.bi-printer::before{content:"\f501"}.bi-puzzle-fill::before{content:"\f502"}.bi-puzzle::before{content:"\f503"}.bi-question-circle-fill::before{content:"\f504"}.bi-question-circle::before{content:"\f505"}.bi-question-diamond-fill::before{content:"\f506"}.bi-question-diamond::before{content:"\f507"}.bi-question-octagon-fill::before{content:"\f508"}.bi-question-octagon::before{content:"\f509"}.bi-question-square-fill::before{content:"\f50a"}.bi-question-square::before{content:"\f50b"}.bi-question::before{content:"\f50c"}.bi-rainbow::before{content:"\f50d"}.bi-receipt-cutoff::before{content:"\f50e"}.bi-receipt::before{content:"\f50f"}.bi-reception-0::before{content:"\f510"}.bi-reception-1::before{content:"\f511"}.bi-reception-2::before{content:"\f512"}.bi-reception-3::before{content:"\f513"}.bi-reception-4::before{content:"\f514"}.bi-record-btn-fill::before{content:"\f515"}.bi-record-btn::before{content:"\f516"}.bi-record-circle-fill::before{content:"\f517"}.bi-record-circle::before{content:"\f518"}.bi-record-fill::before{content:"\f519"}.bi-record::before{content:"\f51a"}.bi-record2-fill::before{content:"\f51b"}.bi-record2::before{content:"\f51c"}.bi-reply-all-fill::before{content:"\f51d"}.bi-reply-all::before{content:"\f51e"}.bi-reply-fill::before{content:"\f51f"}.bi-reply::before{content:"\f520"}.bi-rss-fill::before{content:"\f521"}.bi-rss::before{content:"\f522"}.bi-rulers::before{content:"\f523"}.bi-save-fill::before{content:"\f524"}.bi-save::before{content:"\f525"}.bi-save2-fill::before{content:"\f526"}.bi-save2::before{content:"\f527"}.bi-scissors::before{content:"\f528"}.bi-screwdriver::before{content:"\f529"}.bi-search::before{content:"\f52a"}.bi-segmented-nav::before{content:"\f52b"}.bi-server::before{content:"\f52c"}.bi-share-fill::before{content:"\f52d"}.bi-share::before{content:"\f52e"}.bi-shield-check::before{content:"\f52f"}.bi-shield-exclamation::before{content:"\f530"}.bi-shield-fill-check::before{content:"\f531"}.bi-shield-fill-exclamation::before{content:"\f532"}.bi-shield-fill-minus::before{content:"\f533"}.bi-shield-fill-plus::before{content:"\f534"}.bi-shield-fill-x::before{content:"\f535"}.bi-shield-fill::before{content:"\f536"}.bi-shield-lock-fill::before{content:"\f537"}.bi-shield-lock::before{content:"\f538"}.bi-shield-minus::before{content:"\f539"}.bi-shield-plus::before{content:"\f53a"}.bi-shield-shaded::before{content:"\f53b"}.bi-shield-slash-fill::before{content:"\f53c"}.bi-shield-slash::before{content:"\f53d"}.bi-shield-x::before{content:"\f53e"}.bi-shield::before{content:"\f53f"}.bi-shift-fill::before{content:"\f540"}.bi-shift::before{content:"\f541"}.bi-shop-window::before{content:"\f542"}.bi-shop::before{content:"\f543"}.bi-shuffle::before{content:"\f544"}.bi-signpost-2-fill::before{content:"\f545"}.bi-signpost-2::before{content:"\f546"}.bi-signpost-fill::before{content:"\f547"}.bi-signpost-split-fill::before{content:"\f548"}.bi-signpost-split::before{content:"\f549"}.bi-signpost::before{content:"\f54a"}.bi-sim-fill::before{content:"\f54b"}.bi-sim::before{content:"\f54c"}.bi-skip-backward-btn-fill::before{content:"\f54d"}.bi-skip-backward-btn::before{content:"\f54e"}.bi-skip-backward-circle-fill::before{content:"\f54f"}.bi-skip-backward-circle::before{content:"\f550"}.bi-skip-backward-fill::before{content:"\f551"}.bi-skip-backward::before{content:"\f552"}.bi-skip-end-btn-fill::before{content:"\f553"}.bi-skip-end-btn::before{content:"\f554"}.bi-skip-end-circle-fill::before{content:"\f555"}.bi-skip-end-circle::before{content:"\f556"}.bi-skip-end-fill::before{content:"\f557"}.bi-skip-end::before{content:"\f558"}.bi-skip-forward-btn-fill::before{content:"\f559"}.bi-skip-forward-btn::before{content:"\f55a"}.bi-skip-forward-circle-fill::before{content:"\f55b"}.bi-skip-forward-circle::before{content:"\f55c"}.bi-skip-forward-fill::before{content:"\f55d"}.bi-skip-forward::before{content:"\f55e"}.bi-skip-start-btn-fill::before{content:"\f55f"}.bi-skip-start-btn::before{content:"\f560"}.bi-skip-start-circle-fill::before{content:"\f561"}.bi-skip-start-circle::before{content:"\f562"}.bi-skip-start-fill::before{content:"\f563"}.bi-skip-start::before{content:"\f564"}.bi-slack::before{content:"\f565"}.bi-slash-circle-fill::before{content:"\f566"}.bi-slash-circle::before{content:"\f567"}.bi-slash-square-fill::before{content:"\f568"}.bi-slash-square::before{content:"\f569"}.bi-slash::before{content:"\f56a"}.bi-sliders::before{content:"\f56b"}.bi-smartwatch::before{content:"\f56c"}.bi-snow::before{content:"\f56d"}.bi-snow2::before{content:"\f56e"}.bi-snow3::before{content:"\f56f"}.bi-sort-alpha-down-alt::before{content:"\f570"}.bi-sort-alpha-down::before{content:"\f571"}.bi-sort-alpha-up-alt::before{content:"\f572"}.bi-sort-alpha-up::before{content:"\f573"}.bi-sort-down-alt::before{content:"\f574"}.bi-sort-down::before{content:"\f575"}.bi-sort-numeric-down-alt::before{content:"\f576"}.bi-sort-numeric-down::before{content:"\f577"}.bi-sort-numeric-up-alt::before{content:"\f578"}.bi-sort-numeric-up::before{content:"\f579"}.bi-sort-up-alt::before{content:"\f57a"}.bi-sort-up::before{content:"\f57b"}.bi-soundwave::before{content:"\f57c"}.bi-speaker-fill::before{content:"\f57d"}.bi-speaker::before{content:"\f57e"}.bi-speedometer::before{content:"\f57f"}.bi-speedometer2::before{content:"\f580"}.bi-spellcheck::before{content:"\f581"}.bi-square-fill::before{content:"\f582"}.bi-square-half::before{content:"\f583"}.bi-square::before{content:"\f584"}.bi-stack::before{content:"\f585"}.bi-star-fill::before{content:"\f586"}.bi-star-half::before{content:"\f587"}.bi-star::before{content:"\f588"}.bi-stars::before{content:"\f589"}.bi-stickies-fill::before{content:"\f58a"}.bi-stickies::before{content:"\f58b"}.bi-sticky-fill::before{content:"\f58c"}.bi-sticky::before{content:"\f58d"}.bi-stop-btn-fill::before{content:"\f58e"}.bi-stop-btn::before{content:"\f58f"}.bi-stop-circle-fill::before{content:"\f590"}.bi-stop-circle::before{content:"\f591"}.bi-stop-fill::before{content:"\f592"}.bi-stop::before{content:"\f593"}.bi-stoplights-fill::before{content:"\f594"}.bi-stoplights::before{content:"\f595"}.bi-stopwatch-fill::before{content:"\f596"}.bi-stopwatch::before{content:"\f597"}.bi-subtract::before{content:"\f598"}.bi-suit-club-fill::before{content:"\f599"}.bi-suit-club::before{content:"\f59a"}.bi-suit-diamond-fill::before{content:"\f59b"}.bi-suit-diamond::before{content:"\f59c"}.bi-suit-heart-fill::before{content:"\f59d"}.bi-suit-heart::before{content:"\f59e"}.bi-suit-spade-fill::before{content:"\f59f"}.bi-suit-spade::before{content:"\f5a0"}.bi-sun-fill::before{content:"\f5a1"}.bi-sun::before{content:"\f5a2"}.bi-sunglasses::before{content:"\f5a3"}.bi-sunrise-fill::before{content:"\f5a4"}.bi-sunrise::before{content:"\f5a5"}.bi-sunset-fill::before{content:"\f5a6"}.bi-sunset::before{content:"\f5a7"}.bi-symmetry-horizontal::before{content:"\f5a8"}.bi-symmetry-vertical::before{content:"\f5a9"}.bi-table::before{content:"\f5aa"}.bi-tablet-fill::before{content:"\f5ab"}.bi-tablet-landscape-fill::before{content:"\f5ac"}.bi-tablet-landscape::before{content:"\f5ad"}.bi-tablet::before{content:"\f5ae"}.bi-tag-fill::before{content:"\f5af"}.bi-tag::before{content:"\f5b0"}.bi-tags-fill::before{content:"\f5b1"}.bi-tags::before{content:"\f5b2"}.bi-telegram::before{content:"\f5b3"}.bi-telephone-fill::before{content:"\f5b4"}.bi-telephone-forward-fill::before{content:"\f5b5"}.bi-telephone-forward::before{content:"\f5b6"}.bi-telephone-inbound-fill::before{content:"\f5b7"}.bi-telephone-inbound::before{content:"\f5b8"}.bi-telephone-minus-fill::before{content:"\f5b9"}.bi-telephone-minus::before{content:"\f5ba"}.bi-telephone-outbound-fill::before{content:"\f5bb"}.bi-telephone-outbound::before{content:"\f5bc"}.bi-telephone-plus-fill::before{content:"\f5bd"}.bi-telephone-plus::before{content:"\f5be"}.bi-telephone-x-fill::before{content:"\f5bf"}.bi-telephone-x::before{content:"\f5c0"}.bi-telephone::before{content:"\f5c1"}.bi-terminal-fill::before{content:"\f5c2"}.bi-terminal::before{content:"\f5c3"}.bi-text-center::before{content:"\f5c4"}.bi-text-indent-left::before{content:"\f5c5"}.bi-text-indent-right::before{content:"\f5c6"}.bi-text-left::before{content:"\f5c7"}.bi-text-paragraph::before{content:"\f5c8"}.bi-text-right::before{content:"\f5c9"}.bi-textarea-resize::before{content:"\f5ca"}.bi-textarea-t::before{content:"\f5cb"}.bi-textarea::before{content:"\f5cc"}.bi-thermometer-half::before{content:"\f5cd"}.bi-thermometer-high::before{content:"\f5ce"}.bi-thermometer-low::before{content:"\f5cf"}.bi-thermometer-snow::before{content:"\f5d0"}.bi-thermometer-sun::before{content:"\f5d1"}.bi-thermometer::before{content:"\f5d2"}.bi-three-dots-vertical::before{content:"\f5d3"}.bi-three-dots::before{content:"\f5d4"}.bi-toggle-off::before{content:"\f5d5"}.bi-toggle-on::before{content:"\f5d6"}.bi-toggle2-off::before{content:"\f5d7"}.bi-toggle2-on::before{content:"\f5d8"}.bi-toggles::before{content:"\f5d9"}.bi-toggles2::before{content:"\f5da"}.bi-tools::before{content:"\f5db"}.bi-tornado::before{content:"\f5dc"}.bi-trash-fill::before{content:"\f5dd"}.bi-trash::before{content:"\f5de"}.bi-trash2-fill::before{content:"\f5df"}.bi-trash2::before{content:"\f5e0"}.bi-tree-fill::before{content:"\f5e1"}.bi-tree::before{content:"\f5e2"}.bi-triangle-fill::before{content:"\f5e3"}.bi-triangle-half::before{content:"\f5e4"}.bi-triangle::before{content:"\f5e5"}.bi-trophy-fill::before{content:"\f5e6"}.bi-trophy::before{content:"\f5e7"}.bi-tropical-storm::before{content:"\f5e8"}.bi-truck-flatbed::before{content:"\f5e9"}.bi-truck::before{content:"\f5ea"}.bi-tsunami::before{content:"\f5eb"}.bi-tv-fill::before{content:"\f5ec"}.bi-tv::before{content:"\f5ed"}.bi-twitch::before{content:"\f5ee"}.bi-twitter::before{content:"\f5ef"}.bi-type-bold::before{content:"\f5f0"}.bi-type-h1::before{content:"\f5f1"}.bi-type-h2::before{content:"\f5f2"}.bi-type-h3::before{content:"\f5f3"}.bi-type-italic::before{content:"\f5f4"}.bi-type-strikethrough::before{content:"\f5f5"}.bi-type-underline::before{content:"\f5f6"}.bi-type::before{content:"\f5f7"}.bi-ui-checks-grid::before{content:"\f5f8"}.bi-ui-checks::before{content:"\f5f9"}.bi-ui-radios-grid::before{content:"\f5fa"}.bi-ui-radios::before{content:"\f5fb"}.bi-umbrella-fill::before{content:"\f5fc"}.bi-umbrella::before{content:"\f5fd"}.bi-union::before{content:"\f5fe"}.bi-unlock-fill::before{content:"\f5ff"}.bi-unlock::before{content:"\f600"}.bi-upc-scan::before{content:"\f601"}.bi-upc::before{content:"\f602"}.bi-upload::before{content:"\f603"}.bi-vector-pen::before{content:"\f604"}.bi-view-list::before{content:"\f605"}.bi-view-stacked::before{content:"\f606"}.bi-vinyl-fill::before{content:"\f607"}.bi-vinyl::before{content:"\f608"}.bi-voicemail::before{content:"\f609"}.bi-volume-down-fill::before{content:"\f60a"}.bi-volume-down::before{content:"\f60b"}.bi-volume-mute-fill::before{content:"\f60c"}.bi-volume-mute::before{content:"\f60d"}.bi-volume-off-fill::before{content:"\f60e"}.bi-volume-off::before{content:"\f60f"}.bi-volume-up-fill::before{content:"\f610"}.bi-volume-up::before{content:"\f611"}.bi-vr::before{content:"\f612"}.bi-wallet-fill::before{content:"\f613"}.bi-wallet::before{content:"\f614"}.bi-wallet2::before{content:"\f615"}.bi-watch::before{content:"\f616"}.bi-water::before{content:"\f617"}.bi-whatsapp::before{content:"\f618"}.bi-wifi-1::before{content:"\f619"}.bi-wifi-2::before{content:"\f61a"}.bi-wifi-off::before{content:"\f61b"}.bi-wifi::before{content:"\f61c"}.bi-wind::before{content:"\f61d"}.bi-window-dock::before{content:"\f61e"}.bi-window-sidebar::before{content:"\f61f"}.bi-window::before{content:"\f620"}.bi-wrench::before{content:"\f621"}.bi-x-circle-fill::before{content:"\f622"}.bi-x-circle::before{content:"\f623"}.bi-x-diamond-fill::before{content:"\f624"}.bi-x-diamond::before{content:"\f625"}.bi-x-octagon-fill::before{content:"\f626"}.bi-x-octagon::before{content:"\f627"}.bi-x-square-fill::before{content:"\f628"}.bi-x-square::before{content:"\f629"}.bi-x::before{content:"\f62a"}.bi-youtube::before{content:"\f62b"}.bi-zoom-in::before{content:"\f62c"}.bi-zoom-out::before{content:"\f62d"}.bi-bank::before{content:"\f62e"}.bi-bank2::before{content:"\f62f"}.bi-bell-slash-fill::before{content:"\f630"}.bi-bell-slash::before{content:"\f631"}.bi-cash-coin::before{content:"\f632"}.bi-check-lg::before{content:"\f633"}.bi-coin::before{content:"\f634"}.bi-currency-bitcoin::before{content:"\f635"}.bi-currency-dollar::before{content:"\f636"}.bi-currency-euro::before{content:"\f637"}.bi-currency-exchange::before{content:"\f638"}.bi-currency-pound::before{content:"\f639"}.bi-currency-yen::before{content:"\f63a"}.bi-dash-lg::before{content:"\f63b"}.bi-exclamation-lg::before{content:"\f63c"}.bi-file-earmark-pdf-fill::before{content:"\f63d"}.bi-file-earmark-pdf::before{content:"\f63e"}.bi-file-pdf-fill::before{content:"\f63f"}.bi-file-pdf::before{content:"\f640"}.bi-gender-ambiguous::before{content:"\f641"}.bi-gender-female::before{content:"\f642"}.bi-gender-male::before{content:"\f643"}.bi-gender-trans::before{content:"\f644"}.bi-headset-vr::before{content:"\f645"}.bi-info-lg::before{content:"\f646"}.bi-mastodon::before{content:"\f647"}.bi-messenger::before{content:"\f648"}.bi-piggy-bank-fill::before{content:"\f649"}.bi-piggy-bank::before{content:"\f64a"}.bi-pin-map-fill::before{content:"\f64b"}.bi-pin-map::before{content:"\f64c"}.bi-plus-lg::before{content:"\f64d"}.bi-question-lg::before{content:"\f64e"}.bi-recycle::before{content:"\f64f"}.bi-reddit::before{content:"\f650"}.bi-safe-fill::before{content:"\f651"}.bi-safe2-fill::before{content:"\f652"}.bi-safe2::before{content:"\f653"}.bi-sd-card-fill::before{content:"\f654"}.bi-sd-card::before{content:"\f655"}.bi-skype::before{content:"\f656"}.bi-slash-lg::before{content:"\f657"}.bi-translate::before{content:"\f658"}.bi-x-lg::before{content:"\f659"}.bi-safe::before{content:"\f65a"}.bi-apple::before{content:"\f65b"}.bi-microsoft::before{content:"\f65d"}.bi-windows::before{content:"\f65e"}.bi-behance::before{content:"\f65c"}.bi-dribbble::before{content:"\f65f"}.bi-line::before{content:"\f660"}.bi-medium::before{content:"\f661"}.bi-paypal::before{content:"\f662"}.bi-pinterest::before{content:"\f663"}.bi-signal::before{content:"\f664"}.bi-snapchat::before{content:"\f665"}.bi-spotify::before{content:"\f666"}.bi-stack-overflow::before{content:"\f667"}.bi-strava::before{content:"\f668"}.bi-wordpress::before{content:"\f669"}.bi-vimeo::before{content:"\f66a"}.bi-activity::before{content:"\f66b"}.bi-easel2-fill::before{content:"\f66c"}.bi-easel2::before{content:"\f66d"}.bi-easel3-fill::before{content:"\f66e"}.bi-easel3::before{content:"\f66f"}.bi-fan::before{content:"\f670"}.bi-fingerprint::before{content:"\f671"}.bi-graph-down-arrow::before{content:"\f672"}.bi-graph-up-arrow::before{content:"\f673"}.bi-hypnotize::before{content:"\f674"}.bi-magic::before{content:"\f675"}.bi-person-rolodex::before{content:"\f676"}.bi-person-video::before{content:"\f677"}.bi-person-video2::before{content:"\f678"}.bi-person-video3::before{content:"\f679"}.bi-person-workspace::before{content:"\f67a"}.bi-radioactive::before{content:"\f67b"}.bi-webcam-fill::before{content:"\f67c"}.bi-webcam::before{content:"\f67d"}.bi-yin-yang::before{content:"\f67e"}.bi-bandaid-fill::before{content:"\f680"}.bi-bandaid::before{content:"\f681"}.bi-bluetooth::before{content:"\f682"}.bi-body-text::before{content:"\f683"}.bi-boombox::before{content:"\f684"}.bi-boxes::before{content:"\f685"}.bi-dpad-fill::before{content:"\f686"}.bi-dpad::before{content:"\f687"}.bi-ear-fill::before{content:"\f688"}.bi-ear::before{content:"\f689"}.bi-envelope-check-fill::before{content:"\f68b"}.bi-envelope-check::before{content:"\f68c"}.bi-envelope-dash-fill::before{content:"\f68e"}.bi-envelope-dash::before{content:"\f68f"}.bi-envelope-exclamation-fill::before{content:"\f691"}.bi-envelope-exclamation::before{content:"\f692"}.bi-envelope-plus-fill::before{content:"\f693"}.bi-envelope-plus::before{content:"\f694"}.bi-envelope-slash-fill::before{content:"\f696"}.bi-envelope-slash::before{content:"\f697"}.bi-envelope-x-fill::before{content:"\f699"}.bi-envelope-x::before{content:"\f69a"}.bi-explicit-fill::before{content:"\f69b"}.bi-explicit::before{content:"\f69c"}.bi-git::before{content:"\f69d"}.bi-infinity::before{content:"\f69e"}.bi-list-columns-reverse::before{content:"\f69f"}.bi-list-columns::before{content:"\f6a0"}.bi-meta::before{content:"\f6a1"}.bi-nintendo-switch::before{content:"\f6a4"}.bi-pc-display-horizontal::before{content:"\f6a5"}.bi-pc-display::before{content:"\f6a6"}.bi-pc-horizontal::before{content:"\f6a7"}.bi-pc::before{content:"\f6a8"}.bi-playstation::before{content:"\f6a9"}.bi-plus-slash-minus::before{content:"\f6aa"}.bi-projector-fill::before{content:"\f6ab"}.bi-projector::before{content:"\f6ac"}.bi-qr-code-scan::before{content:"\f6ad"}.bi-qr-code::before{content:"\f6ae"}.bi-quora::before{content:"\f6af"}.bi-quote::before{content:"\f6b0"}.bi-robot::before{content:"\f6b1"}.bi-send-check-fill::before{content:"\f6b2"}.bi-send-check::before{content:"\f6b3"}.bi-send-dash-fill::before{content:"\f6b4"}.bi-send-dash::before{content:"\f6b5"}.bi-send-exclamation-fill::before{content:"\f6b7"}.bi-send-exclamation::before{content:"\f6b8"}.bi-send-fill::before{content:"\f6b9"}.bi-send-plus-fill::before{content:"\f6ba"}.bi-send-plus::before{content:"\f6bb"}.bi-send-slash-fill::before{content:"\f6bc"}.bi-send-slash::before{content:"\f6bd"}.bi-send-x-fill::before{content:"\f6be"}.bi-send-x::before{content:"\f6bf"}.bi-send::before{content:"\f6c0"}.bi-steam::before{content:"\f6c1"}.bi-terminal-dash::before{content:"\f6c3"}.bi-terminal-plus::before{content:"\f6c4"}.bi-terminal-split::before{content:"\f6c5"}.bi-ticket-detailed-fill::before{content:"\f6c6"}.bi-ticket-detailed::before{content:"\f6c7"}.bi-ticket-fill::before{content:"\f6c8"}.bi-ticket-perforated-fill::before{content:"\f6c9"}.bi-ticket-perforated::before{content:"\f6ca"}.bi-ticket::before{content:"\f6cb"}.bi-tiktok::before{content:"\f6cc"}.bi-window-dash::before{content:"\f6cd"}.bi-window-desktop::before{content:"\f6ce"}.bi-window-fullscreen::before{content:"\f6cf"}.bi-window-plus::before{content:"\f6d0"}.bi-window-split::before{content:"\f6d1"}.bi-window-stack::before{content:"\f6d2"}.bi-window-x::before{content:"\f6d3"}.bi-xbox::before{content:"\f6d4"}.bi-ethernet::before{content:"\f6d5"}.bi-hdmi-fill::before{content:"\f6d6"}.bi-hdmi::before{content:"\f6d7"}.bi-usb-c-fill::before{content:"\f6d8"}.bi-usb-c::before{content:"\f6d9"}.bi-usb-fill::before{content:"\f6da"}.bi-usb-plug-fill::before{content:"\f6db"}.bi-usb-plug::before{content:"\f6dc"}.bi-usb-symbol::before{content:"\f6dd"}.bi-usb::before{content:"\f6de"}.bi-boombox-fill::before{content:"\f6df"}.bi-displayport::before{content:"\f6e1"}.bi-gpu-card::before{content:"\f6e2"}.bi-memory::before{content:"\f6e3"}.bi-modem-fill::before{content:"\f6e4"}.bi-modem::before{content:"\f6e5"}.bi-motherboard-fill::before{content:"\f6e6"}.bi-motherboard::before{content:"\f6e7"}.bi-optical-audio-fill::before{content:"\f6e8"}.bi-optical-audio::before{content:"\f6e9"}.bi-pci-card::before{content:"\f6ea"}.bi-router-fill::before{content:"\f6eb"}.bi-router::before{content:"\f6ec"}.bi-thunderbolt-fill::before{content:"\f6ef"}.bi-thunderbolt::before{content:"\f6f0"}.bi-usb-drive-fill::before{content:"\f6f1"}.bi-usb-drive::before{content:"\f6f2"}.bi-usb-micro-fill::before{content:"\f6f3"}.bi-usb-micro::before{content:"\f6f4"}.bi-usb-mini-fill::before{content:"\f6f5"}.bi-usb-mini::before{content:"\f6f6"}.bi-cloud-haze2::before{content:"\f6f7"}.bi-device-hdd-fill::before{content:"\f6f8"}.bi-device-hdd::before{content:"\f6f9"}.bi-device-ssd-fill::before{content:"\f6fa"}.bi-device-ssd::before{content:"\f6fb"}.bi-displayport-fill::before{content:"\f6fc"}.bi-mortarboard-fill::before{content:"\f6fd"}.bi-mortarboard::before{content:"\f6fe"}.bi-terminal-x::before{content:"\f6ff"}.bi-arrow-through-heart-fill::before{content:"\f700"}.bi-arrow-through-heart::before{content:"\f701"}.bi-badge-sd-fill::before{content:"\f702"}.bi-badge-sd::before{content:"\f703"}.bi-bag-heart-fill::before{content:"\f704"}.bi-bag-heart::before{content:"\f705"}.bi-balloon-fill::before{content:"\f706"}.bi-balloon-heart-fill::before{content:"\f707"}.bi-balloon-heart::before{content:"\f708"}.bi-balloon::before{content:"\f709"}.bi-box2-fill::before{content:"\f70a"}.bi-box2-heart-fill::before{content:"\f70b"}.bi-box2-heart::before{content:"\f70c"}.bi-box2::before{content:"\f70d"}.bi-braces-asterisk::before{content:"\f70e"}.bi-calendar-heart-fill::before{content:"\f70f"}.bi-calendar-heart::before{content:"\f710"}.bi-calendar2-heart-fill::before{content:"\f711"}.bi-calendar2-heart::before{content:"\f712"}.bi-chat-heart-fill::before{content:"\f713"}.bi-chat-heart::before{content:"\f714"}.bi-chat-left-heart-fill::before{content:"\f715"}.bi-chat-left-heart::before{content:"\f716"}.bi-chat-right-heart-fill::before{content:"\f717"}.bi-chat-right-heart::before{content:"\f718"}.bi-chat-square-heart-fill::before{content:"\f719"}.bi-chat-square-heart::before{content:"\f71a"}.bi-clipboard-check-fill::before{content:"\f71b"}.bi-clipboard-data-fill::before{content:"\f71c"}.bi-clipboard-fill::before{content:"\f71d"}.bi-clipboard-heart-fill::before{content:"\f71e"}.bi-clipboard-heart::before{content:"\f71f"}.bi-clipboard-minus-fill::before{content:"\f720"}.bi-clipboard-plus-fill::before{content:"\f721"}.bi-clipboard-pulse::before{content:"\f722"}.bi-clipboard-x-fill::before{content:"\f723"}.bi-clipboard2-check-fill::before{content:"\f724"}.bi-clipboard2-check::before{content:"\f725"}.bi-clipboard2-data-fill::before{content:"\f726"}.bi-clipboard2-data::before{content:"\f727"}.bi-clipboard2-fill::before{content:"\f728"}.bi-clipboard2-heart-fill::before{content:"\f729"}.bi-clipboard2-heart::before{content:"\f72a"}.bi-clipboard2-minus-fill::before{content:"\f72b"}.bi-clipboard2-minus::before{content:"\f72c"}.bi-clipboard2-plus-fill::before{content:"\f72d"}.bi-clipboard2-plus::before{content:"\f72e"}.bi-clipboard2-pulse-fill::before{content:"\f72f"}.bi-clipboard2-pulse::before{content:"\f730"}.bi-clipboard2-x-fill::before{content:"\f731"}.bi-clipboard2-x::before{content:"\f732"}.bi-clipboard2::before{content:"\f733"}.bi-emoji-kiss-fill::before{content:"\f734"}.bi-emoji-kiss::before{content:"\f735"}.bi-envelope-heart-fill::before{content:"\f736"}.bi-envelope-heart::before{content:"\f737"}.bi-envelope-open-heart-fill::before{content:"\f738"}.bi-envelope-open-heart::before{content:"\f739"}.bi-envelope-paper-fill::before{content:"\f73a"}.bi-envelope-paper-heart-fill::before{content:"\f73b"}.bi-envelope-paper-heart::before{content:"\f73c"}.bi-envelope-paper::before{content:"\f73d"}.bi-filetype-aac::before{content:"\f73e"}.bi-filetype-ai::before{content:"\f73f"}.bi-filetype-bmp::before{content:"\f740"}.bi-filetype-cs::before{content:"\f741"}.bi-filetype-css::before{content:"\f742"}.bi-filetype-csv::before{content:"\f743"}.bi-filetype-doc::before{content:"\f744"}.bi-filetype-docx::before{content:"\f745"}.bi-filetype-exe::before{content:"\f746"}.bi-filetype-gif::before{content:"\f747"}.bi-filetype-heic::before{content:"\f748"}.bi-filetype-html::before{content:"\f749"}.bi-filetype-java::before{content:"\f74a"}.bi-filetype-jpg::before{content:"\f74b"}.bi-filetype-js::before{content:"\f74c"}.bi-filetype-jsx::before{content:"\f74d"}.bi-filetype-key::before{content:"\f74e"}.bi-filetype-m4p::before{content:"\f74f"}.bi-filetype-md::before{content:"\f750"}.bi-filetype-mdx::before{content:"\f751"}.bi-filetype-mov::before{content:"\f752"}.bi-filetype-mp3::before{content:"\f753"}.bi-filetype-mp4::before{content:"\f754"}.bi-filetype-otf::before{content:"\f755"}.bi-filetype-pdf::before{content:"\f756"}.bi-filetype-php::before{content:"\f757"}.bi-filetype-png::before{content:"\f758"}.bi-filetype-ppt::before{content:"\f75a"}.bi-filetype-psd::before{content:"\f75b"}.bi-filetype-py::before{content:"\f75c"}.bi-filetype-raw::before{content:"\f75d"}.bi-filetype-rb::before{content:"\f75e"}.bi-filetype-sass::before{content:"\f75f"}.bi-filetype-scss::before{content:"\f760"}.bi-filetype-sh::before{content:"\f761"}.bi-filetype-svg::before{content:"\f762"}.bi-filetype-tiff::before{content:"\f763"}.bi-filetype-tsx::before{content:"\f764"}.bi-filetype-ttf::before{content:"\f765"}.bi-filetype-txt::before{content:"\f766"}.bi-filetype-wav::before{content:"\f767"}.bi-filetype-woff::before{content:"\f768"}.bi-filetype-xls::before{content:"\f76a"}.bi-filetype-xml::before{content:"\f76b"}.bi-filetype-yml::before{content:"\f76c"}.bi-heart-arrow::before{content:"\f76d"}.bi-heart-pulse-fill::before{content:"\f76e"}.bi-heart-pulse::before{content:"\f76f"}.bi-heartbreak-fill::before{content:"\f770"}.bi-heartbreak::before{content:"\f771"}.bi-hearts::before{content:"\f772"}.bi-hospital-fill::before{content:"\f773"}.bi-hospital::before{content:"\f774"}.bi-house-heart-fill::before{content:"\f775"}.bi-house-heart::before{content:"\f776"}.bi-incognito::before{content:"\f777"}.bi-magnet-fill::before{content:"\f778"}.bi-magnet::before{content:"\f779"}.bi-person-heart::before{content:"\f77a"}.bi-person-hearts::before{content:"\f77b"}.bi-phone-flip::before{content:"\f77c"}.bi-plugin::before{content:"\f77d"}.bi-postage-fill::before{content:"\f77e"}.bi-postage-heart-fill::before{content:"\f77f"}.bi-postage-heart::before{content:"\f780"}.bi-postage::before{content:"\f781"}.bi-postcard-fill::before{content:"\f782"}.bi-postcard-heart-fill::before{content:"\f783"}.bi-postcard-heart::before{content:"\f784"}.bi-postcard::before{content:"\f785"}.bi-search-heart-fill::before{content:"\f786"}.bi-search-heart::before{content:"\f787"}.bi-sliders2-vertical::before{content:"\f788"}.bi-sliders2::before{content:"\f789"}.bi-trash3-fill::before{content:"\f78a"}.bi-trash3::before{content:"\f78b"}.bi-valentine::before{content:"\f78c"}.bi-valentine2::before{content:"\f78d"}.bi-wrench-adjustable-circle-fill::before{content:"\f78e"}.bi-wrench-adjustable-circle::before{content:"\f78f"}.bi-wrench-adjustable::before{content:"\f790"}.bi-filetype-json::before{content:"\f791"}.bi-filetype-pptx::before{content:"\f792"}.bi-filetype-xlsx::before{content:"\f793"}.bi-1-circle-fill::before{content:"\f796"}.bi-1-circle::before{content:"\f797"}.bi-1-square-fill::before{content:"\f798"}.bi-1-square::before{content:"\f799"}.bi-2-circle-fill::before{content:"\f79c"}.bi-2-circle::before{content:"\f79d"}.bi-2-square-fill::before{content:"\f79e"}.bi-2-square::before{content:"\f79f"}.bi-3-circle-fill::before{content:"\f7a2"}.bi-3-circle::before{content:"\f7a3"}.bi-3-square-fill::before{content:"\f7a4"}.bi-3-square::before{content:"\f7a5"}.bi-4-circle-fill::before{content:"\f7a8"}.bi-4-circle::before{content:"\f7a9"}.bi-4-square-fill::before{content:"\f7aa"}.bi-4-square::before{content:"\f7ab"}.bi-5-circle-fill::before{content:"\f7ae"}.bi-5-circle::before{content:"\f7af"}.bi-5-square-fill::before{content:"\f7b0"}.bi-5-square::before{content:"\f7b1"}.bi-6-circle-fill::before{content:"\f7b4"}.bi-6-circle::before{content:"\f7b5"}.bi-6-square-fill::before{content:"\f7b6"}.bi-6-square::before{content:"\f7b7"}.bi-7-circle-fill::before{content:"\f7ba"}.bi-7-circle::before{content:"\f7bb"}.bi-7-square-fill::before{content:"\f7bc"}.bi-7-square::before{content:"\f7bd"}.bi-8-circle-fill::before{content:"\f7c0"}.bi-8-circle::before{content:"\f7c1"}.bi-8-square-fill::before{content:"\f7c2"}.bi-8-square::before{content:"\f7c3"}.bi-9-circle-fill::before{content:"\f7c6"}.bi-9-circle::before{content:"\f7c7"}.bi-9-square-fill::before{content:"\f7c8"}.bi-9-square::before{content:"\f7c9"}.bi-airplane-engines-fill::before{content:"\f7ca"}.bi-airplane-engines::before{content:"\f7cb"}.bi-airplane-fill::before{content:"\f7cc"}.bi-airplane::before{content:"\f7cd"}.bi-alexa::before{content:"\f7ce"}.bi-alipay::before{content:"\f7cf"}.bi-android::before{content:"\f7d0"}.bi-android2::before{content:"\f7d1"}.bi-box-fill::before{content:"\f7d2"}.bi-box-seam-fill::before{content:"\f7d3"}.bi-browser-chrome::before{content:"\f7d4"}.bi-browser-edge::before{content:"\f7d5"}.bi-browser-firefox::before{content:"\f7d6"}.bi-browser-safari::before{content:"\f7d7"}.bi-c-circle-fill::before{content:"\f7da"}.bi-c-circle::before{content:"\f7db"}.bi-c-square-fill::before{content:"\f7dc"}.bi-c-square::before{content:"\f7dd"}.bi-capsule-pill::before{content:"\f7de"}.bi-capsule::before{content:"\f7df"}.bi-car-front-fill::before{content:"\f7e0"}.bi-car-front::before{content:"\f7e1"}.bi-cassette-fill::before{content:"\f7e2"}.bi-cassette::before{content:"\f7e3"}.bi-cc-circle-fill::before{content:"\f7e6"}.bi-cc-circle::before{content:"\f7e7"}.bi-cc-square-fill::before{content:"\f7e8"}.bi-cc-square::before{content:"\f7e9"}.bi-cup-hot-fill::before{content:"\f7ea"}.bi-cup-hot::before{content:"\f7eb"}.bi-currency-rupee::before{content:"\f7ec"}.bi-dropbox::before{content:"\f7ed"}.bi-escape::before{content:"\f7ee"}.bi-fast-forward-btn-fill::before{content:"\f7ef"}.bi-fast-forward-btn::before{content:"\f7f0"}.bi-fast-forward-circle-fill::before{content:"\f7f1"}.bi-fast-forward-circle::before{content:"\f7f2"}.bi-fast-forward-fill::before{content:"\f7f3"}.bi-fast-forward::before{content:"\f7f4"}.bi-filetype-sql::before{content:"\f7f5"}.bi-fire::before{content:"\f7f6"}.bi-google-play::before{content:"\f7f7"}.bi-h-circle-fill::before{content:"\f7fa"}.bi-h-circle::before{content:"\f7fb"}.bi-h-square-fill::before{content:"\f7fc"}.bi-h-square::before{content:"\f7fd"}.bi-indent::before{content:"\f7fe"}.bi-lungs-fill::before{content:"\f7ff"}.bi-lungs::before{content:"\f800"}.bi-microsoft-teams::before{content:"\f801"}.bi-p-circle-fill::before{content:"\f804"}.bi-p-circle::before{content:"\f805"}.bi-p-square-fill::before{content:"\f806"}.bi-p-square::before{content:"\f807"}.bi-pass-fill::before{content:"\f808"}.bi-pass::before{content:"\f809"}.bi-prescription::before{content:"\f80a"}.bi-prescription2::before{content:"\f80b"}.bi-r-circle-fill::before{content:"\f80e"}.bi-r-circle::before{content:"\f80f"}.bi-r-square-fill::before{content:"\f810"}.bi-r-square::before{content:"\f811"}.bi-repeat-1::before{content:"\f812"}.bi-repeat::before{content:"\f813"}.bi-rewind-btn-fill::before{content:"\f814"}.bi-rewind-btn::before{content:"\f815"}.bi-rewind-circle-fill::before{content:"\f816"}.bi-rewind-circle::before{content:"\f817"}.bi-rewind-fill::before{content:"\f818"}.bi-rewind::before{content:"\f819"}.bi-train-freight-front-fill::before{content:"\f81a"}.bi-train-freight-front::before{content:"\f81b"}.bi-train-front-fill::before{content:"\f81c"}.bi-train-front::before{content:"\f81d"}.bi-train-lightrail-front-fill::before{content:"\f81e"}.bi-train-lightrail-front::before{content:"\f81f"}.bi-truck-front-fill::before{content:"\f820"}.bi-truck-front::before{content:"\f821"}.bi-ubuntu::before{content:"\f822"}.bi-unindent::before{content:"\f823"}.bi-unity::before{content:"\f824"}.bi-universal-access-circle::before{content:"\f825"}.bi-universal-access::before{content:"\f826"}.bi-virus::before{content:"\f827"}.bi-virus2::before{content:"\f828"}.bi-wechat::before{content:"\f829"}.bi-yelp::before{content:"\f82a"}.bi-sign-stop-fill::before{content:"\f82b"}.bi-sign-stop-lights-fill::before{content:"\f82c"}.bi-sign-stop-lights::before{content:"\f82d"}.bi-sign-stop::before{content:"\f82e"}.bi-sign-turn-left-fill::before{content:"\f82f"}.bi-sign-turn-left::before{content:"\f830"}.bi-sign-turn-right-fill::before{content:"\f831"}.bi-sign-turn-right::before{content:"\f832"}.bi-sign-turn-slight-left-fill::before{content:"\f833"}.bi-sign-turn-slight-left::before{content:"\f834"}.bi-sign-turn-slight-right-fill::before{content:"\f835"}.bi-sign-turn-slight-right::before{content:"\f836"}.bi-sign-yield-fill::before{content:"\f837"}.bi-sign-yield::before{content:"\f838"}.bi-ev-station-fill::before{content:"\f839"}.bi-ev-station::before{content:"\f83a"}.bi-fuel-pump-diesel-fill::before{content:"\f83b"}.bi-fuel-pump-diesel::before{content:"\f83c"}.bi-fuel-pump-fill::before{content:"\f83d"}.bi-fuel-pump::before{content:"\f83e"}.bi-0-circle-fill::before{content:"\f83f"}.bi-0-circle::before{content:"\f840"}.bi-0-square-fill::before{content:"\f841"}.bi-0-square::before{content:"\f842"}.bi-rocket-fill::before{content:"\f843"}.bi-rocket-takeoff-fill::before{content:"\f844"}.bi-rocket-takeoff::before{content:"\f845"}.bi-rocket::before{content:"\f846"}.bi-stripe::before{content:"\f847"}.bi-subscript::before{content:"\f848"}.bi-superscript::before{content:"\f849"}.bi-trello::before{content:"\f84a"}.bi-envelope-at-fill::before{content:"\f84b"}.bi-envelope-at::before{content:"\f84c"}.bi-regex::before{content:"\f84d"}.bi-text-wrap::before{content:"\f84e"}.bi-sign-dead-end-fill::before{content:"\f84f"}.bi-sign-dead-end::before{content:"\f850"}.bi-sign-do-not-enter-fill::before{content:"\f851"}.bi-sign-do-not-enter::before{content:"\f852"}.bi-sign-intersection-fill::before{content:"\f853"}.bi-sign-intersection-side-fill::before{content:"\f854"}.bi-sign-intersection-side::before{content:"\f855"}.bi-sign-intersection-t-fill::before{content:"\f856"}.bi-sign-intersection-t::before{content:"\f857"}.bi-sign-intersection-y-fill::before{content:"\f858"}.bi-sign-intersection-y::before{content:"\f859"}.bi-sign-intersection::before{content:"\f85a"}.bi-sign-merge-left-fill::before{content:"\f85b"}.bi-sign-merge-left::before{content:"\f85c"}.bi-sign-merge-right-fill::before{content:"\f85d"}.bi-sign-merge-right::before{content:"\f85e"}.bi-sign-no-left-turn-fill::before{content:"\f85f"}.bi-sign-no-left-turn::before{content:"\f860"}.bi-sign-no-parking-fill::before{content:"\f861"}.bi-sign-no-parking::before{content:"\f862"}.bi-sign-no-right-turn-fill::before{content:"\f863"}.bi-sign-no-right-turn::before{content:"\f864"}.bi-sign-railroad-fill::before{content:"\f865"}.bi-sign-railroad::before{content:"\f866"}.bi-building-add::before{content:"\f867"}.bi-building-check::before{content:"\f868"}.bi-building-dash::before{content:"\f869"}.bi-building-down::before{content:"\f86a"}.bi-building-exclamation::before{content:"\f86b"}.bi-building-fill-add::before{content:"\f86c"}.bi-building-fill-check::before{content:"\f86d"}.bi-building-fill-dash::before{content:"\f86e"}.bi-building-fill-down::before{content:"\f86f"}.bi-building-fill-exclamation::before{content:"\f870"}.bi-building-fill-gear::before{content:"\f871"}.bi-building-fill-lock::before{content:"\f872"}.bi-building-fill-slash::before{content:"\f873"}.bi-building-fill-up::before{content:"\f874"}.bi-building-fill-x::before{content:"\f875"}.bi-building-fill::before{content:"\f876"}.bi-building-gear::before{content:"\f877"}.bi-building-lock::before{content:"\f878"}.bi-building-slash::before{content:"\f879"}.bi-building-up::before{content:"\f87a"}.bi-building-x::before{content:"\f87b"}.bi-buildings-fill::before{content:"\f87c"}.bi-buildings::before{content:"\f87d"}.bi-bus-front-fill::before{content:"\f87e"}.bi-bus-front::before{content:"\f87f"}.bi-ev-front-fill::before{content:"\f880"}.bi-ev-front::before{content:"\f881"}.bi-globe-americas::before{content:"\f882"}.bi-globe-asia-australia::before{content:"\f883"}.bi-globe-central-south-asia::before{content:"\f884"}.bi-globe-europe-africa::before{content:"\f885"}.bi-house-add-fill::before{content:"\f886"}.bi-house-add::before{content:"\f887"}.bi-house-check-fill::before{content:"\f888"}.bi-house-check::before{content:"\f889"}.bi-house-dash-fill::before{content:"\f88a"}.bi-house-dash::before{content:"\f88b"}.bi-house-down-fill::before{content:"\f88c"}.bi-house-down::before{content:"\f88d"}.bi-house-exclamation-fill::before{content:"\f88e"}.bi-house-exclamation::before{content:"\f88f"}.bi-house-gear-fill::before{content:"\f890"}.bi-house-gear::before{content:"\f891"}.bi-house-lock-fill::before{content:"\f892"}.bi-house-lock::before{content:"\f893"}.bi-house-slash-fill::before{content:"\f894"}.bi-house-slash::before{content:"\f895"}.bi-house-up-fill::before{content:"\f896"}.bi-house-up::before{content:"\f897"}.bi-house-x-fill::before{content:"\f898"}.bi-house-x::before{content:"\f899"}.bi-person-add::before{content:"\f89a"}.bi-person-down::before{content:"\f89b"}.bi-person-exclamation::before{content:"\f89c"}.bi-person-fill-add::before{content:"\f89d"}.bi-person-fill-check::before{content:"\f89e"}.bi-person-fill-dash::before{content:"\f89f"}.bi-person-fill-down::before{content:"\f8a0"}.bi-person-fill-exclamation::before{content:"\f8a1"}.bi-person-fill-gear::before{content:"\f8a2"}.bi-person-fill-lock::before{content:"\f8a3"}.bi-person-fill-slash::before{content:"\f8a4"}.bi-person-fill-up::before{content:"\f8a5"}.bi-person-fill-x::before{content:"\f8a6"}.bi-person-gear::before{content:"\f8a7"}.bi-person-lock::before{content:"\f8a8"}.bi-person-slash::before{content:"\f8a9"}.bi-person-up::before{content:"\f8aa"}.bi-scooter::before{content:"\f8ab"}.bi-taxi-front-fill::before{content:"\f8ac"}.bi-taxi-front::before{content:"\f8ad"}.bi-amd::before{content:"\f8ae"}.bi-database-add::before{content:"\f8af"}.bi-database-check::before{content:"\f8b0"}.bi-database-dash::before{content:"\f8b1"}.bi-database-down::before{content:"\f8b2"}.bi-database-exclamation::before{content:"\f8b3"}.bi-database-fill-add::before{content:"\f8b4"}.bi-database-fill-check::before{content:"\f8b5"}.bi-database-fill-dash::before{content:"\f8b6"}.bi-database-fill-down::before{content:"\f8b7"}.bi-database-fill-exclamation::before{content:"\f8b8"}.bi-database-fill-gear::before{content:"\f8b9"}.bi-database-fill-lock::before{content:"\f8ba"}.bi-database-fill-slash::before{content:"\f8bb"}.bi-database-fill-up::before{content:"\f8bc"}.bi-database-fill-x::before{content:"\f8bd"}.bi-database-fill::before{content:"\f8be"}.bi-database-gear::before{content:"\f8bf"}.bi-database-lock::before{content:"\f8c0"}.bi-database-slash::before{content:"\f8c1"}.bi-database-up::before{content:"\f8c2"}.bi-database-x::before{content:"\f8c3"}.bi-database::before{content:"\f8c4"}.bi-houses-fill::before{content:"\f8c5"}.bi-houses::before{content:"\f8c6"}.bi-nvidia::before{content:"\f8c7"}.bi-person-vcard-fill::before{content:"\f8c8"}.bi-person-vcard::before{content:"\f8c9"}.bi-sina-weibo::before{content:"\f8ca"}.bi-tencent-qq::before{content:"\f8cb"}.bi-wikipedia::before{content:"\f8cc"}.bi-alphabet-uppercase::before{content:"\f2a5"}.bi-alphabet::before{content:"\f68a"}.bi-amazon::before{content:"\f68d"}.bi-arrows-collapse-vertical::before{content:"\f690"}.bi-arrows-expand-vertical::before{content:"\f695"}.bi-arrows-vertical::before{content:"\f698"}.bi-arrows::before{content:"\f6a2"}.bi-ban-fill::before{content:"\f6a3"}.bi-ban::before{content:"\f6b6"}.bi-bing::before{content:"\f6c2"}.bi-cake::before{content:"\f6e0"}.bi-cake2::before{content:"\f6ed"}.bi-cookie::before{content:"\f6ee"}.bi-copy::before{content:"\f759"}.bi-crosshair::before{content:"\f769"}.bi-crosshair2::before{content:"\f794"}.bi-emoji-astonished-fill::before{content:"\f795"}.bi-emoji-astonished::before{content:"\f79a"}.bi-emoji-grimace-fill::before{content:"\f79b"}.bi-emoji-grimace::before{content:"\f7a0"}.bi-emoji-grin-fill::before{content:"\f7a1"}.bi-emoji-grin::before{content:"\f7a6"}.bi-emoji-surprise-fill::before{content:"\f7a7"}.bi-emoji-surprise::before{content:"\f7ac"}.bi-emoji-tear-fill::before{content:"\f7ad"}.bi-emoji-tear::before{content:"\f7b2"}.bi-envelope-arrow-down-fill::before{content:"\f7b3"}.bi-envelope-arrow-down::before{content:"\f7b8"}.bi-envelope-arrow-up-fill::before{content:"\f7b9"}.bi-envelope-arrow-up::before{content:"\f7be"}.bi-feather::before{content:"\f7bf"}.bi-feather2::before{content:"\f7c4"}.bi-floppy-fill::before{content:"\f7c5"}.bi-floppy::before{content:"\f7d8"}.bi-floppy2-fill::before{content:"\f7d9"}.bi-floppy2::before{content:"\f7e4"}.bi-gitlab::before{content:"\f7e5"}.bi-highlighter::before{content:"\f7f8"}.bi-marker-tip::before{content:"\f802"}.bi-nvme-fill::before{content:"\f803"}.bi-nvme::before{content:"\f80c"}.bi-opencollective::before{content:"\f80d"}.bi-pci-card-network::before{content:"\f8cd"}.bi-pci-card-sound::before{content:"\f8ce"}.bi-radar::before{content:"\f8cf"}.bi-send-arrow-down-fill::before{content:"\f8d0"}.bi-send-arrow-down::before{content:"\f8d1"}.bi-send-arrow-up-fill::before{content:"\f8d2"}.bi-send-arrow-up::before{content:"\f8d3"}.bi-sim-slash-fill::before{content:"\f8d4"}.bi-sim-slash::before{content:"\f8d5"}.bi-sourceforge::before{content:"\f8d6"}.bi-substack::before{content:"\f8d7"}.bi-threads-fill::before{content:"\f8d8"}.bi-threads::before{content:"\f8d9"}.bi-transparency::before{content:"\f8da"}.bi-twitter-x::before{content:"\f8db"}.bi-type-h4::before{content:"\f8dc"}.bi-type-h5::before{content:"\f8dd"}.bi-type-h6::before{content:"\f8de"}.bi-backpack-fill::before{content:"\f8df"}.bi-backpack::before{content:"\f8e0"}.bi-backpack2-fill::before{content:"\f8e1"}.bi-backpack2::before{content:"\f8e2"}.bi-backpack3-fill::before{content:"\f8e3"}.bi-backpack3::before{content:"\f8e4"}.bi-backpack4-fill::before{content:"\f8e5"}.bi-backpack4::before{content:"\f8e6"}.bi-brilliance::before{content:"\f8e7"}.bi-cake-fill::before{content:"\f8e8"}.bi-cake2-fill::before{content:"\f8e9"}.bi-duffle-fill::before{content:"\f8ea"}.bi-duffle::before{content:"\f8eb"}.bi-exposure::before{content:"\f8ec"}.bi-gender-neuter::before{content:"\f8ed"}.bi-highlights::before{content:"\f8ee"}.bi-luggage-fill::before{content:"\f8ef"}.bi-luggage::before{content:"\f8f0"}.bi-mailbox-flag::before{content:"\f8f1"}.bi-mailbox2-flag::before{content:"\f8f2"}.bi-noise-reduction::before{content:"\f8f3"}.bi-passport-fill::before{content:"\f8f4"}.bi-passport::before{content:"\f8f5"}.bi-person-arms-up::before{content:"\f8f6"}.bi-person-raised-hand::before{content:"\f8f7"}.bi-person-standing-dress::before{content:"\f8f8"}.bi-person-standing::before{content:"\f8f9"}.bi-person-walking::before{content:"\f8fa"}.bi-person-wheelchair::before{content:"\f8fb"}.bi-shadows::before{content:"\f8fc"}.bi-suitcase-fill::before{content:"\f8fd"}.bi-suitcase-lg-fill::before{content:"\f8fe"}.bi-suitcase-lg::before{content:"\f8ff"}.bi-suitcase::before{content:"\f900"}.bi-suitcase2-fill::before{content:"\f901"}.bi-suitcase2::before{content:"\f902"}.bi-vignette::before{content:"\f903"}.bi-bluesky::before{content:"\f7f9"}.bi-tux::before{content:"\f904"}.bi-beaker-fill::before{content:"\f905"}.bi-beaker::before{content:"\f906"}.bi-flask-fill::before{content:"\f907"}.bi-flask-florence-fill::before{content:"\f908"}.bi-flask-florence::before{content:"\f909"}.bi-flask::before{content:"\f90a"}.bi-leaf-fill::before{content:"\f90b"}.bi-leaf::before{content:"\f90c"}.bi-measuring-cup-fill::before{content:"\f90d"}.bi-measuring-cup::before{content:"\f90e"}.bi-unlock2-fill::before{content:"\f90f"}.bi-unlock2::before{content:"\f910"}.bi-battery-low::before{content:"\f911"}.bi-anthropic::before{content:"\f912"}.bi-apple-music::before{content:"\f913"}.bi-claude::before{content:"\f914"}.bi-openai::before{content:"\f915"}.bi-perplexity::before{content:"\f916"}.bi-css::before{content:"\f917"}.bi-javascript::before{content:"\f918"}.bi-typescript::before{content:"\f919"}.bi-fork-knife::before{content:"\f91a"}.bi-globe-americas-fill::before{content:"\f91b"}.bi-globe-asia-australia-fill::before{content:"\f91c"}.bi-globe-central-south-asia-fill::before{content:"\f91d"}.bi-globe-europe-africa-fill::before{content:"\f91e"}

```


**Tool Call: Search files for regex `if.*err.*!=.*nil`**
Status: Completed

Showing matches 1-20 (there were more matches found; use offset: 20 to see next page):

## Matches in pg-press/cmd/pg-press/commands-api-key.go

### func apiKeyCommand › L16-25
```
			return func(cmd *cli.Command) error {
				apiKey, err := keymaker.NewApiKey("pgp", 32)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Generating a new api key failed: %s\n", err.Error())
					return err
				}

				fmt.Print(apiKey) // Yes, no newline at the end
				return nil
			}
```

## Matches in pg-press/cmd/pg-press/commands-cookies.go

### func removeCookiesCommand › L32-46
```
			return func(cmd *cli.Command) error {
				db, err := openDB(*customDBPath)
				if err != nil {
					return err
				}

				if *useApiKey {
					err = db.Cookies.RemoveApiKey(*value)
				} else {
					err = db.Cookies.Remove(*value)
				}

				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
					os.Exit(exitCodeGeneric)
```

4 lines remaining in ancestor node. Read the file to see all.

### func autoCleanCookiesCommand › L69-79
```
			return func(cmd *cli.Command) error {
				db, err := openDB(*customDBPath)
				if err != nil {
					return err
				}

				t := time.Now().Add(0 - constants.CookieExpirationDuration).UnixMilli()
				isExpired := func(cookie *models.Cookie) bool {
					return t >= cookie.LastLogin
				}

```

48 lines remaining in ancestor node. Read the file to see all.

### func autoCleanCookiesCommand › L81-95
```
				if *telegramID != 0 {
					u, err := db.Users.Get(*telegramID)
					if err != nil {
						if utils.IsNotFoundError(err) {
							os.Exit(exitCodeNotFound)
						}

						fmt.Fprintf(os.Stderr, "Get user \"%d\" failed: %s\n", *telegramID, err.Error())
						os.Exit(exitCodeGeneric)
					}

					cookies, err := db.Cookies.ListApiKey(u.ApiKey)
					if err != nil {
						fmt.Fprintf(os.Stderr, "List cookies for user \"%d\" failed: %s\n", *telegramID, err.Error())
						os.Exit(exitCodeGeneric)
```

13 lines remaining in ancestor node. Read the file to see all.

### func autoCleanCookiesCommand › L99-104
```
						if isExpired(cookie) {
							if err = db.Cookies.Remove(cookie.Value); err != nil {
								// Print out error and continue
								fmt.Fprintf(os.Stderr, "Removing cookie with value \"%s\" failed: %s\n", cookie.Value, err.Error())
							}
						}
```

### func autoCleanCookiesCommand › L110-114
```
				// Clean up all cookies
				cookies, err := db.Cookies.List()
				if err != nil {
					fmt.Fprintf(os.Stderr, "List cookies from database failed: %s\n", err.Error())
					os.Exit(exitCodeGeneric)
```

### func autoCleanCookiesCommand › L118-123
```
					if isExpired(cookie) {
						if err = db.Cookies.Remove(cookie.Value); err != nil {
							// Print out error and continue
							fmt.Fprintf(os.Stderr, "Removing cookie with value \"%s\" failed: %s\n", cookie.Value, err.Error())
						}
					}
```

## Matches in pg-press/cmd/pg-press/commands-feeds.go

### func listFeedsCommand › L60-70
```
			return func(cmd *cli.Command) error {
				db, err := openDB(*customDBPath)
				if err != nil {
					return err
				}

				var feeds []*models.Feed

				// Get feeds based on parameters
				if *limit > 0 {
					feeds, err = db.Feeds.ListRange(*offset, *limit)
```

46 lines remaining in ancestor node. Read the file to see all.

### func listFeedsCommand › L73-77
```
				}

				if err != nil {
					return err
				}
```

### func removeFeedsCommand › L156-166
```
			return func(cmd *cli.Command) error {
				db, err := openDB(*customDBPath)
				if err != nil {
					return err
				}

				// Remove by IDs
				if *idsStr != "" {
					ids := strings.Split(*idsStr, ",")
					// Trim whitespace from each ID
					for i, id := range ids {
```

17 lines remaining in ancestor node. Read the file to see all.

### func filterFeedsByDate › L196-202
```
	if since != "" {
		sinceTime, err = parseDateTime(since)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid since date format: %s\n", err)
			return feeds
		}
	}
```

### func filterFeedsByDate › L204-210
```
	if before != "" {
		beforeTime, err = parseDateTime(before)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid before date format: %s\n", err)
			return feeds
		}
	}
```

### func removeFeedsByIDs › L250-260
```
	for _, idStr := range ids {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			failed = append(failed, fmt.Sprintf("invalid ID '%s': %s", idStr, err))
			continue
		}

		err = db.Feeds.Delete(int64(id))
		if err != nil {
			if utils.IsNotFoundError(err) {
				failed = append(failed, fmt.Sprintf("feed ID %d not found", id))
```

9 lines remaining in ancestor node. Read the file to see all.

### func removeFeedsByDuration › L287-297
```
func removeFeedsByDuration(db *database.DB, durationStr string) error {
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		// Try parsing as days if direct parsing fails
		if strings.HasSuffix(durationStr, "d") {
			days, parseErr := strconv.Atoi(strings.TrimSuffix(durationStr, "d"))
			if parseErr != nil {
				return fmt.Errorf("invalid duration format: %s (use format like '24h', '7d', '30d')", durationStr)
			}
			duration = time.Duration(days) * 24 * time.Hour
		} else {
```

17 lines remaining in ancestor node. Read the file to see all.

### func removeFeedsByDuration › L304-308
```

	rowsAffected, err := db.Feeds.DeleteBefore(timestamp)
	if err != nil {
		return fmt.Errorf("failed to remove feeds: %s", err)
	}
```

### func removeFeedsByDate › L316-326
```
func removeFeedsByDate(db *database.DB, dateStr string) error {
	cutoffTime, err := parseDateTime(dateStr)
	if err != nil {
		return fmt.Errorf("invalid date format: %s (use format '2006-01-02' or '2006-01-02 15:04:05')", dateStr)
	}

	timestamp := cutoffTime.UnixMilli()

	rowsAffected, err := db.Feeds.DeleteBefore(timestamp)
	if err != nil {
		return fmt.Errorf("failed to remove feeds: %s", err)
```

6 lines remaining in ancestor node. Read the file to see all.

## Matches in pg-press/cmd/pg-press/commands-server.go

### func serverCommand › L42-51
```
				if *logFile != "" {
					f, err := os.OpenFile(*logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
					if err != nil {
						log().Error("Failed to open log file %s: %v", *logFile, err)
						return err
					} else {
						log().SetOutput(f)
						log().Info("Redirected logs to file: %s", *logFile)
					}
				}
```

### func serverCommand › L52-56
```

				db, err := openDB(*customDBPath)
				if err != nil {
					log().Error("Failed to open database: %v", err)
					return err
```

### func serverCommand › L69-73
```

				log().Info("Starting HTTP server on %s", *addr)
				if err := e.Start(*addr); err != nil &&
					err != http.ErrServerClosed {
					log().Error("Server startup failed on %s: %v", *addr, err)
```

### func createHTTPErrorHandler › L159-163
```
			} else {
				if err := c.String(code, message); err != nil {
					log().Error("Failed to send string error response: %v", err)
				}
			}
```


**Tool Call: Search files for regex `sql\.Open|db\.Open|database\.Open`**
Status: Completed

Found 1 matches:

## Matches in pg-press/cmd/pg-press/utils.go

### func openDB › L33-37
```
	connectionString := path + "?_busy_timeout=30000&_journal_mode=WAL&_foreign_keys=on&_synchronous=NORMAL"

	db, err := sql.Open("sqlite3", connectionString)
	if err != nil {
		return nil, err
```


**Tool Call: Search files for regex `Close\(\).*err`**
Status: Completed

Found 1 matches:

## Matches in pg-press/internal/web/assets/js/htmx-ext-ws-v2.0.3.min.js

### L1
```
(function(){var v;htmx.defineExtension("ws",{init:function(e){v=e;if(!htmx.createWebSocket){htmx.createWebSocket=t}if(!htmx.config.wsReconnectDelay){htmx.config.wsReconnectDelay="full-jitter"}},onEvent:function(e,t){var n=t.target||t.detail.elt;switch(e){case"htmx:beforeCleanupElement":var r=v.getInternalData(n);if(r.webSocket){r.webSocket.close()}return;case"htmx:beforeProcessNode":h(l(n,"ws-connect"),function(e){s(e)});h(l(n,"ws-send"),function(e){a(e)})}}});function i(e){return e.trim().split(/\s+/)}function r(e){var t=v.getAttributeValue(e,"hx-ws");if(t){var n=i(t);for(var r=0;r<n.length;r++){var s=n[r].split(/:(.+)/);if(s[0]==="connect"){return s[1]}}}}function s(a){if(!v.bodyContains(a)){return}var e=v.getAttributeValue(a,"ws-connect");if(e==null||e===""){var t=r(a);if(t==null){return}else{e=t}}if(e.indexOf("/")===0){var n=location.hostname+(location.port?":"+location.port:"");if(location.protocol==="https:"){e="wss://"+n+e}else if(location.protocol==="http:"){e="ws://"+n+e}}var o=c(a,function(){return htmx.createWebSocket(e)});o.addEventListener("message",function(e){if(m(a)){return}var t=e.data;if(!v.triggerEvent(a,"htmx:wsBeforeMessage",{message:t,socketWrapper:o.publicInterface})){return}v.withExtensions(a,function(e){t=e.transformResponse(t,null,a)});var n=v.makeSettleInfo(a);var r=v.makeFragment(t);if(r.children.length){var s=Array.from(r.children);for(var i=0;i<s.length;i++){v.oobSwap(v.getAttributeValue(s[i],"hx-swap-oob")||"true",s[i],n)}}v.settleImmediately(n.tasks);v.triggerEvent(a,"htmx:wsAfterMessage",{message:t,socketWrapper:o.publicInterface})});v.getInternalData(a).webSocket=o}function c(r,t){var s={socket:null,messageQueue:[],retryCount:0,events:{},addEventListener:function(e,t){if(this.socket){this.socket.addEventListener(e,t)}if(!this.events[e]){this.events[e]=[]}this.events[e].push(t)},sendImmediately:function(e,t){if(!this.socket){v.triggerErrorEvent()}if(!t||v.triggerEvent(t,"htmx:wsBeforeSend",{message:e,socketWrapper:this.publicInterface})){this.socket.send(e);t&&v.triggerEvent(t,"htmx:wsAfterSend",{message:e,socketWrapper:this.publicInterface})}},send:function(e,t){if(this.socket.readyState!==this.socket.OPEN){this.messageQueue.push({message:e,sendElt:t})}else{this.sendImmediately(e,t)}},handleQueuedMessages:function(){while(this.messageQueue.length>0){var e=this.messageQueue[0];if(this.socket.readyState===this.socket.OPEN){this.sendImmediately(e.message,e.sendElt);this.messageQueue.shift()}else{break}}},init:function(){if(this.socket&&this.socket.readyState===this.socket.OPEN){this.socket.close()}var n=t();v.triggerEvent(r,"htmx:wsConnecting",{event:{type:"connecting"}});this.socket=n;n.onopen=function(e){s.retryCount=0;v.triggerEvent(r,"htmx:wsOpen",{event:e,socketWrapper:s.publicInterface});s.handleQueuedMessages()};n.onclose=function(e){if(!m(r)&&[1006,1012,1013].indexOf(e.code)>=0){var t=f(s.retryCount);setTimeout(function(){s.retryCount+=1;s.init()},t)}v.triggerEvent(r,"htmx:wsClose",{event:e,socketWrapper:s.publicInterface})};n.onerror=function(e){v.triggerErrorEvent(r,"htmx:wsError",{error:e,socketWrapper:s});m(r)};var e=this.events;Object.keys(e).forEach(function(t){e[t].forEach(function(e){n.addEventListener(t,e)})})},close:function(){this.socket.close()}};s.init();s.publicInterface={send:s.send.bind(s),sendImmediately:s.sendImmediately.bind(s),reconnect:s.init.bind(s),queue:s.messageQueue};return s}function a(e){var t=v.getAttributeValue(e,"hx-ws");if(t&&t!=="send"){return}var n=v.getClosestMatch(e,o);u(n,e)}function o(e){return v.getInternalData(e).webSocket!=null}function u(g,d){var t=v.getInternalData(d);var e=v.getTriggerSpecs(d);e.forEach(function(e){v.addTriggerHandler(d,e,t,function(e,t){if(m(g)){return}var n=v.getInternalData(g).webSocket;var r=v.getHeaders(d,v.getTarget(d));var s=v.getInputValues(d,"post");var i=s.errors;var a=Object.assign({},s.values);var o=v.getExpressionVars(d);var c=v.mergeObjects(a,o);var u=v.filterValues(c,d);var f={parameters:u,unfilteredParameters:c,headers:r,errors:i,triggeringEvent:t,messageBody:undefined,socketWrapper:n.publicInterface};if(!v.triggerEvent(e,"htmx:wsConfigSend",f)){return}if(i&&i.length>0){v.triggerEvent(e,"htmx:validation:halted",i);return}var l=f.messageBody;if(l===undefined){var h=Object.assign({},f.parameters);if(f.headers){h.HEADERS=r}l=JSON.stringify(h)}n.send(l,e);if(t&&v.shouldCancel(t,e)){t.preventDefault()}})})}function f(e){var t=htmx.config.wsReconnectDelay;if(typeof t==="function"){return t(e)}if(t==="full-jitter"){var n=Math.min(e,6);var r=1e3*Math.pow(2,n);return r*Math.random()}logError('htmx.config.wsReconnectDelay must either be a function or the string "full-jitter"')}function m(e){if(!v.bodyContains(e)){var t=v.getInternalData(e);if(t.webSocket){t.webSocket.close();return true}return false}return false}function t(e){var t=new WebSocket(e,[]);t.binaryType=htmx.config.wsBinaryType;return t}function l(e,t){var n=[];if(v.hasAttribute(e,t)||v.hasAttribute(e,"hx-ws")){n.push(e)}e.querySelectorAll("["+t+"], [data-"+t+"], [data-hx-ws], [hx-ws]").forEach(function(e){n.push(e)});return n}function h(e,t){if(e){for(var n=0;n<e.length;n++){t(e[n])}}}})();
```


**Tool Call: Search files for regex `err.*:=.*rows\.Scan`**
Status: Completed

Found 4 matches:

## Matches in pg-press/internal/services/modifications.go

### func (s *Modifications) List › L156-163
```
	for rows.Next() {
		mod := &models.Modification[interface{}]{}
		err := rows.Scan(&mod.ID, &mod.UserID, &mod.Data, &mod.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan modification: %v", err)
		}
		modifications = append(modifications, mod)
	}
```

### func (s *Modifications) GetByUser › L318-325
```
	for rows.Next() {
		mod := &models.Modification[interface{}]{}
		err := rows.Scan(&mod.ID, &mod.UserID, &mod.Data, &mod.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan modification: %v", err)
		}
		modifications = append(modifications, mod)
	}
```

### func (s *Modifications) GetByDateRange › L355-362
```
	for rows.Next() {
		mod := &models.Modification[any]{}
		err := rows.Scan(&mod.ID, &mod.UserID, &mod.Data, &mod.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan modification: %v", err)
		}
		modifications = append(modifications, mod)
	}
```

### func (s *Modifications) ListWithUser › L457-467
```
	for rows.Next() {
		modWithUser := &ModificationWithUser{}
		err := rows.Scan(
			&modWithUser.Modification.ID,
			&modWithUser.Modification.UserID,
			&modWithUser.Modification.Data,
			&modWithUser.Modification.CreatedAt,
			&modWithUser.User.Name,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan modification with user: %v", err)
```

4 lines remaining in ancestor node. Read the file to see all.


**Tool Call: Search files for regex `log\.(Fatal|Error|Warn|Print)`**
Status: Completed

Showing matches 1-20 (there were more matches found; use offset: 20 to see next page):

## Matches in pg-press/internal/services/feeds.go

### func (f *Feeds) List › L66-68
```
	if elapsed > 100*time.Millisecond {
		f.log.Warn("Slow feed list query took %v for %d feeds", elapsed, len(feeds))
	}
```

### func (f *Feeds) ListRange › L102-104
```
	if elapsed > 100*time.Millisecond {
		f.log.Warn("Slow feed range query took %v (offset=%d, limit=%d, returned=%d)", elapsed, offset, limit, len(feeds))
	}
```

### func (f *Feeds) ListByUser › L141-143
```
	if elapsed > 100*time.Millisecond {
		f.log.Warn("Slow user feeds query took %v (userID=%d, offset=%d, limit=%d, returned=%d)", elapsed, userID, offset, limit, len(feeds))
	}
```

### func (f *Feeds) Add › L179-181
```
	if elapsed > 50*time.Millisecond {
		f.log.Warn("Slow feed insert took %v for user %d", elapsed, feedData.UserID)
	}
```

### func (f *Feeds) Count › L198-200
```
	if elapsed > 50*time.Millisecond {
		f.log.Warn("Slow feed count query took %v (result: %d)", elapsed, count)
	}
```

### func (f *Feeds) CountByUser › L221-223
```
	if elapsed > 50*time.Millisecond {
		f.log.Warn("Slow user feed count query took %v (userID=%d, result=%d)", elapsed, userID, count)
	}
```

### func (f *Feeds) DeleteBefore › L250-252
```
	if elapsed > 100*time.Millisecond {
		f.log.Warn("Slow feed deletion took %v (timestamp=%d, deleted=%d)", elapsed, timestamp, rowsAffected)
	}
```

### func (f *Feeds) scanAllRows › L317-319
```
	if scanElapsed > 50*time.Millisecond {
		f.log.Warn("Slow feed row scanning took %v for %d rows", scanElapsed, len(feeds))
	}
```

## Matches in pg-press/internal/services/metal-sheets.go

### func (s *MetalSheets) GetForPress › L215-218
```
		if err != nil {
			s.log.Error("Failed to get metal sheets for tool %d: %v", toolID, err)
			continue
		}
```

### func (s *MetalSheets) GetForPress › L239-242
```
		default:
			otherCount++
			s.log.Warn("Found metal sheet %d with unexpected identifier: %s", sheet.ID, sheet.Identifier)
		}
```

## Matches in pg-press/internal/services/press-cycles.go

### func (s *PressCycles) GetPartialCycles › L77-82
```
		if err != sql.ErrNoRows {
			s.log.Error(
				"Failed to get previous total cycles for press %d, tool %d, position %s: %v",
				cycle.PressNumber, cycle.ToolID, cycle.ToolPosition, err,
			)
		}
```

## Matches in pg-press/internal/services/tool-regenerations.go

### func (r *ToolRegenerations) AddToolRegeneration › L143-148
```
	if err != nil {
		// Undo the tool's regeneration status
		r.log.Error("Failed to create new regeneration record: tool_id=%d", toolID)
		r.log.Debug("Undoing tool regeneration status: tool_id=%d", toolID)
		return nil, r.tools.UpdateRegenerating(toolID, false, user)
	}
```

### func (r *ToolRegenerations) AbortToolRegeneration › L181-184
```
	if err != nil {
		r.log.Warn("No regeneration record found to abort for tool_id=%d: %v", toolID, err)
		// Continue with status update even if no record found
	} else {
```

### func (r *ToolRegenerations) AbortToolRegeneration › L187-190
```
		if err := r.Delete(lastRegen.ID); err != nil {
			r.log.Error("Failed to delete regeneration record: id=%d, error=%v", lastRegen.ID, err)
			return fmt.Errorf("failed to delete regeneration record: %v", err)
		}
```

## Matches in pg-press/internal/services/tools.go

### func (t *Tools) GetActiveToolsForPress › L137-140
```
	if err != nil {
		t.log.Error("Failed to query active tools: %v", err)
		return nil
	}
```

### func (t *Tools) GetActiveToolsForPress › L152-155
```
	if err := rows.Err(); err != nil {
		t.log.Error("Error iterating over tool rows: %v", err)
		return nil
	}
```

## Matches in pg-press/internal/services/trouble-reports.go

### func (s *TroubleReports) List › L59-62
```
	if err != nil {
		s.log.Error("Failed to execute trouble reports list query: %v", err)
		return nil, fmt.Errorf("select error: trouble_reports: %v", err)
	}
```

### func (s *TroubleReports) List › L77-80
```
	if err := rows.Err(); err != nil {
		s.log.Error("Error iterating over %d trouble report rows: %v", reportCount, err)
		return nil, fmt.Errorf("select error: trouble_reports: %v", err)
	}
```

### func (s *TroubleReports) Add › L142-145
```
	if err != nil {
		s.log.Error("Failed to get last insert ID for trouble report by %s: %v", userInfo, err)
		return 0, fmt.Errorf("insert error: trouble_reports: %v", err)
	}
```

### func (s *TroubleReports) Add › L154-157
```
	if err := s.modifications.AddTroubleReportMod(u.TelegramID, id, modData); err != nil {
		s.log.Error("Failed to save initial modification for trouble report %d: %v", id, err)
		// Don't fail the entire operation for modification tracking
	}
```


**Tool Call: Search files for regex `nil.*pointer|panic.*nil`**
Status: Completed

No matches found

**Tool Call: Search files for regex `race.*condition|goroutine.*leak|deadlock`**
Status: Completed

No matches found

**Tool Call: Search files for regex `go.*func\(|go\s+\w+\(`**
Status: Completed

No matches found

## Assistant

Let me also check for SQL injection vulnerabilities and other security issues:

**Tool Call: Search files for regex `fmt\.Sprintf.*\+|".*\+.*"`**
Status: Completed

Showing matches 1-20 (there were more matches found; use offset: 20 to see next page):

## Matches in pg-press/internal/pdf/image-utils.go

### func addImageCaptions › L113-115
```
	o.PDF.CellFormat(position.ImageWidth, 4,
		o.Translator(fmt.Sprintf("Anhang %d", position.StartIndex+1)),
		"0", 0, "C", false, 0, "")
```

### func addImageCaptions › L120-124
```
		o.PDF.CellFormat(
			position.ImageWidth, 4,
			o.Translator(fmt.Sprintf("Anhang %d", position.StartIndex+2)),
			"0", 0, "C", false, 0, "",
		)
```

## Matches in pg-press/internal/services/cookies.go

### func (c *Cookies) Add › L124-134
```
func (c *Cookies) Add(cookie *models.Cookie) error {
	c.log.Info("Adding cookie: %+v", cookie)

	if cookie == nil {
		c.log.Debug("Validation failed: cookie is nil")
		return utils.NewValidationError("cookie: cookie cannot be nil")
	}

	if cookie.Value == "" {
		c.log.Debug("Validation failed: empty cookie value")
		return utils.NewValidationError("value: cookie value cannot be empty")
```

34 lines remaining in ancestor node. Read the file to see all.

### func (c *Cookies) Update › L171-181
```
func (c *Cookies) Update(value string, cookie *models.Cookie) error {
	c.log.Info("Updating cookie: %+v, value: %s", cookie, value)

	if value == "" {
		c.log.Debug("Validation failed: empty current cookie value")
		return utils.NewValidationError("value: current cookie value cannot be empty")
	}
	if cookie == nil {
		c.log.Debug("Validation failed: cookie is nil")
		return utils.NewValidationError("cookie: cookie cannot be nil")
	}
```

33 lines remaining in ancestor node. Read the file to see all.

## Matches in pg-press/internal/services/trouble-reports.go

### func (s *TroubleReports) UpdateWithAttachments › L442-446
```
	tr.ID = id

	s.log.Debug("Combined attachments for update %d: %d existing + %d new = %d total",
		id, originalAttachmentCount, len(newAttachmentIDs), len(allAttachmentIDs))

```

## Matches in pg-press/internal/web/assets/css/ui.min.css

### @import"https://fonts.googleapis.com/css2?family=Recursive:slnt,wght,CASL,CRSV,MONO@-15..0,300..1000,0..1,0..1,0..1&display=swap" › L1-2
```
@import"https://fonts.googleapis.com/css2?family=Recursive:slnt,wght,CASL,CRSV,MONO@-15..0,300..1000,0..1,0..1,0..1&display=swap";:root{--ui-font-family: "Recursive", system-ui, sans-serif;--ui-font-size: 1rem;--ui-line-height: 1.5;--ui-line-height-code: 1.25;--ui-spacing: .5rem;--ui-radius: .25rem;--ui-border-width: 1px;--ui-border-style: solid;--ui-app-bar-height: 3rem;--ui-hue: 225;--ui-saturation: 10%}:root,[data-theme=light]{--ui-color-50: hsl(var(--ui-hue), 0%, 98%);--ui-color-100: hsl(var(--ui-hue), var(--ui-saturation), 95%);--ui-color-200: hsl(var(--ui-hue), var(--ui-saturation), 88%);--ui-color-300: hsl(var(--ui-hue), var(--ui-saturation), 78%);--ui-color-400: hsl(var(--ui-hue), var(--ui-saturation), 60%);--ui-color-500: hsl(var(--ui-hue), var(--ui-saturation), 42%);--ui-color-600: hsl(var(--ui-hue), var(--ui-saturation), 32%);--ui-color-700: hsl(var(--ui-hue), var(--ui-saturation), 24%);--ui-color-800: hsl(var(--ui-hue), var(--ui-saturation), 15%);--ui-color-900: hsl(var(--ui-hue), var(--ui-saturation), 8%);--ui-color-950: hsl(var(--ui-hue), 0%, 3%);--ui-bg: var(--ui-color-50);--ui-text: var(--ui-color-950);--ui-text-input: hsla(var(--ui-hue), 0%, 0%, .85);--ui-muted-text: hsl(var(--ui-hue), 15%, 25%);--ui-border-color: hsla(var(--ui-hue), 20%, 20%, .2);--ui-backdrop-color: hsla(var(--ui-hue), 25%, 98%, .85);--ui-primary: hsl(120, 85%, 55%);--ui-primary-hover: hsl(120, 85%, 50%);--ui-primary-active: hsl(120, 85%, 40%);--ui-primary-text: var(--ui-bg);--ui-secondary: hsl(var(--ui-hue), 10%, 60%);--ui-secondary-hover: hsl(var(--ui-hue), 10%, 52%);--ui-secondary-active: hsl(var(--ui-hue), 10%, 40%);--ui-secondary-text: var(--ui-text);--ui-success: hsl(140, 85%, 55%);--ui-success-hover: hsl(140, 85%, 50%);--ui-success-active: hsl(140, 85%, 40%);--ui-success-text: hsl(0, 0%, 100%);--ui-warning: hsl(42, 95%, 58%);--ui-warning-hover: hsl(42, 95%, 53%);--ui-warning-active: hsl(42, 95%, 43%);--ui-warning-text: hsl(32, 95%, 12%);--ui-destructive: hsl(0, 85%, 60%);--ui-destructive-hover: hsl(0, 85%, 55%);--ui-destructive-active: hsl(0, 85%, 45%);--ui-destructive-text: hsl(0, 93%, 94%);--ui-error: var(--ui-destructive);--ui-error-text: var(--ui-destructive-text);--ui-info: hsl(200, 85%, 50%);--ui-info-hover: hsl(200, 85%, 45%);--ui-info-active: hsl(200, 85%, 35%);--ui-info-text: hsl(200, 95%, 95%);--ui-muted: hsl(var(--ui-hue), 12%, 85%);--ui-contrast: var(--ui-color-950);--ui-contrast-hover: hsl(var(--ui-hue), 0%, 10%);--ui-contrast-active: hsl(var(--ui-hue), 0%, 15%);--ui-contrast-text: var(--ui-color-50);color-scheme:light}@media (prefers-color-scheme: dark){:root:not([data-theme=light]){--ui-color-50: hsl(var(--ui-hue), 0%, 0%);--ui-color-100: hsl(var(--ui-hue), var(--ui-saturation), 8%);--ui-color-200: hsl(var(--ui-hue), var(--ui-saturation), 12%);--ui-color-300: hsl(var(--ui-hue), var(--ui-saturation), 18%);--ui-color-400: hsl(var(--ui-hue), var(--ui-saturation), 28%);--ui-color-500: hsl(var(--ui-hue), var(--ui-saturation), 42%);--ui-color-600: hsl(var(--ui-hue), var(--ui-saturation), 58%);--ui-color-700: hsl(var(--ui-hue), var(--ui-saturation), 72%);--ui-color-800: hsl(var(--ui-hue), var(--ui-saturation), 85%);--ui-color-900: hsl(var(--ui-hue), var(--ui-saturation), 92%);--ui-color-950: hsl(var(--ui-hue), 0%, 98%);--ui-text-input: hsla(var(--ui-hue), 0%, 100%, .85);--ui-muted-text: hsl(var(--ui-hue), 15%, 75%);--ui-border-color: hsla(var(--ui-hue), 20%, 80%, .25);--ui-backdrop-color: hsla(var(--ui-hue), 25%, 5%, .85);--ui-primary: hsl(120, 90%, 68%);--ui-primary-hover: hsl(120, 90%, 63%);--ui-primary-active: hsl(120, 90%, 53%);--ui-secondary: hsl(var(--ui-hue), 10%, 42%);--ui-secondary-hover: hsl(var(--ui-hue), 10%, 50%);--ui-secondary-active: hsl(var(--ui-hue), 10%, 62%);--ui-success: hsl(140, 75%, 65%);--ui-success-hover: hsl(140, 75%, 60%);--ui-success-active: hsl(140, 75%, 50%);--ui-success-text: hsl(140, 50%, 15%);--ui-warning: hsl(42, 90%, 65%);--ui-warning-hover: hsl(42, 90%, 60%);--ui-warning-active: hsl(42, 90%, 50%);--ui-warning-text: hsl(32, 85%, 15%);--ui-destructive: hsl(0, 85%, 72%);--ui-destructive-hover: hsl(0, 85%, 67%);--ui-destructive-active: hsl(0, 85%, 52%);--ui-destructive-text: hsl(0, 75%, 25%);--ui-info: hsl(200, 85%, 25%);--ui-info-hover: hsl(200, 85%, 30%);--ui-info-active: hsl(200, 85%, 40%);--ui-info-text: var(--ui-text);--ui-muted: hsl(var(--ui-hue), 18%, 15%);--ui-contrast: var(--ui-color-950);--ui-contrast-hover: hsl(var(--ui-hue), 0%, 90%);--ui-contrast-active: hsl(var(--ui-hue), 0%, 85%);color-scheme:dark}}[data-theme=dark]{--ui-color-50: hsl(var(--ui-hue), 0%, 0%);--ui-color-100: hsl(var(--ui-hue), var(--ui-saturation), 8%);--ui-color-200: hsl(var(--ui-hue), var(--ui-saturation), 12%);--ui-color-300: hsl(var(--ui-hue), var(--ui-saturation), 18%);--ui-color-400: hsl(var(--ui-hue), var(--ui-saturation), 28%);--ui-color-500: hsl(var(--ui-hue), var(--ui-saturation), 42%);--ui-color-600: hsl(var(--ui-hue), var(--ui-saturation), 58%);--ui-color-700: hsl(var(--ui-hue), var(--ui-saturation), 72%);--ui-color-800: hsl(var(--ui-hue), var(--ui-saturation), 85%);--ui-color-900: hsl(var(--ui-hue), var(--ui-saturation), 92%);--ui-color-950: hsl(var(--ui-hue), 0%, 98%);--ui-bg: var(--ui-color-50);--ui-text: var(--ui-color-950);--ui-text-input: hsla(var(--ui-hue), 0%, 100%, .85);--ui-muted-text: hsl(var(--ui-hue), 15%, 75%);--ui-border-color: hsla(var(--ui-hue), 20%, 80%, .25);--ui-backdrop-color: hsla(var(--ui-hue), 25%, 5%, .85);--ui-primary: hsl(120, 90%, 68%);--ui-primary-hover: hsl(120, 90%, 63%);--ui-primary-active: hsl(120, 90%, 53%);--ui-primary-text: var(--ui-bg);--ui-secondary: hsl(var(--ui-hue), 10%, 42%);--ui-secondary-hover: hsl(var(--ui-hue), 10%, 50%);--ui-secondary-active: hsl(var(--ui-hue), 10%, 62%);--ui-secondary-text: var(--ui-text);--ui-success: hsl(140, 75%, 65%);--ui-success-hover: hsl(140, 75%, 60%);--ui-success-active: hsl(140, 75%, 50%);--ui-success-text: hsl(140, 50%, 15%);--ui-warning: hsl(42, 90%, 65%);--ui-warning-hover: hsl(42, 90%, 60%);--ui-warning-active: hsl(42, 90%, 50%);--ui-warning-text: hsl(32, 85%, 15%);--ui-destructive: hsl(0, 85%, 72%);--ui-destructive-hover: hsl(0, 85%, 67%);--ui-destructive-active: hsl(0, 85%, 52%);--ui-destructive-text: hsl(0, 75%, 25%);--ui-error: var(--ui-destructive);--ui-error-text: var(--ui-destructive-text);--ui-info: hsl(200, 85%, 25%);--ui-info-hover: hsl(200, 85%, 30%);--ui-info-active: hsl(200, 85%, 40%);--ui-info-text: var(--ui-text);--ui-muted: hsl(var(--ui-hue), 18%, 15%);--ui-contrast: var(--ui-color-950);--ui-contrast-hover: hsl(var(--ui-hue), 0%, 90%);--ui-contrast-active: hsl(var(--ui-hue), 0%, 85%);--ui-contrast-text: var(--ui-color-50);color-scheme:dark}html{font-family:var(--ui-font-family);font-size:var(--ui-font-size);background-color:var(--ui-bg);color:var(--ui-text)}body{line-height:var(--ui-line-height)}:root{--mono: 0;--casl: 1;--wght: 400;--slnt: 0;--CRSV: .5}*{font-variation-settings:"MONO" var(--mono),"CASL" var(--casl),"wght" var(--wght),"slnt" var(--slnt),"CRSV" var(--CRSV)}h1,h2,h3,h4,h5,h6{--wght: 700;--slnt: -3}b,strong{--wght: 600}i,em{--slnt: -10}code,kbd,pre,samp{--mono: 1;--casl: 0}small{--wght: 300}blockquote{--wght: 350;--slnt: -10}*,*:before,*:after{position:relative;box-sizing:inherit;margin:0;padding:0}html{box-sizing:border-box;-webkit-tap-highlight-color:transparent;-webkit-font-smoothing:antialiased;-moz-osx-font-smoothing:grayscale}img,picture,video,canvas,svg{display:block;max-width:100%;height:auto}input,button,textarea,select{font:inherit}p,h1,h2,h3,h4,h5,h6{overflow-wrap:break-word}ul[role=list],ol[role=list]{list-style:none}button{background:none;border:none;cursor:pointer}a{color:inherit;text-decoration:none}button,[role=button],input[type=submit],input[type=button],label[for],select,summary,[tabindex]:not([tabindex="-1"]){cursor:pointer}[disabled],[aria-disabled=true]{cursor:not-allowed}html{box-sizing:border-box;font-size:var(--ui-font-size);color-scheme:light dark;padding:env(safe-area-inset-top) env(safe-area-inset-right) env(safe-area-inset-bottom) env(safe-area-inset-left)}body{line-height:var(--ui-line-height);background-color:var(--ui-bg);color:var(--ui-text)}::selection{background-color:var(--ui-primary);color:var(--ui-primary-text)}::backdrop{background-color:var(--ui-backdrop-color)}[disabled],:disabled{opacity:.45;cursor:not-allowed;pointer-events:none;-webkit-user-select:none;user-select:none}.draggable{-webkit-user-select:none;user-select:none;-webkit-user-drag:element}:focus{outline:none}:focus-visible{outline:2px solid var(--ui-primary);outline-offset:2px}@media (prefers-contrast: high){:focus-visible{outline-width:3px;outline-offset:3px}}@media (prefers-reduced-motion: reduce){*,*:before,*:after{animation-duration:.01ms!important;animation-iteration-count:1!important;transition-duration:.01ms!important;scroll-behavior:auto!important}}form>*{margin:var(--ui-spacing) 0}input:not([type]),input[type=""],input[type=text],input[type=email],input[type=password],input[type=tel],input[type=url],input[type=search],input[type=number],input[type=date],input[type=time],input[type=datetime-local],input[type=month],input[type=week],textarea{display:inline-block;padding:var(--ui-spacing) calc(var(--ui-spacing) * 2);background-color:transparent;color:var(--ui-text-input);border:var(--ui-border-width) solid var(--ui-border-color);border-radius:var(--ui-radius);font-size:1rem;accent-color:var(--ui-primary);transition:border-color .25s ease-in-out}input:not([type]):hover,input[type=""]:hover,input[type=text]:hover,input[type=email]:hover,input[type=password]:hover,input[type=tel]:hover,input[type=url]:hover,input[type=search]:hover,input[type=number]:hover,input[type=date]:hover,input[type=time]:hover,input[type=datetime-local]:hover,input[type=month]:hover,input[type=week]:hover,textarea:hover{border-color:var(--ui-primary)}input:not([type]):focus,input[type=""]:focus,input[type=text]:focus,input[type=email]:focus,input[type=password]:focus,input[type=tel]:focus,input[type=url]:focus,input[type=search]:focus,input[type=number]:focus,input[type=date]:focus,input[type=time]:focus,input[type=datetime-local]:focus,input[type=month]:focus,input[type=week]:focus,textarea:focus{outline:none;border-color:var(--ui-primary)}input:not([type]):focus-visible,input[type=""]:focus-visible,input[type=text]:focus-visible,input[type=email]:focus-visible,input[type=password]:focus-visible,input[type=tel]:focus-visible,input[type=url]:focus-visible,input[type=search]:focus-visible,input[type=number]:focus-visible,input[type=date]:focus-visible,input[type=time]:focus-visible,input[type=datetime-local]:focus-visible,input[type=month]:focus-visible,input[type=week]:focus-visible,textarea:focus-visible{outline:2px solid var(--ui-primary);outline-offset:2px}textarea{min-height:calc(var(--ui-spacing) * 8);resize:vertical}input[type=number]::-webkit-inner-spin-button,input[type=number]::-webkit-outer-spin-button{-webkit-appearance:none;margin:0}input[type=number]{-moz-appearance:textfield}input[type=checkbox],input[type=radio]{cursor:pointer;accent-color:var(--ui-primary)}input[type=checkbox]{display:inline-block;height:1.5rem;width:1.5rem}input[type=color]{cursor:pointer;height:calc(var(--ui-spacing) * 2 + 2rem);padding:var(--ui-spacing);border:var(--ui-border-width) solid var(--ui-border-color);border-radius:var(--ui-radius);background-color:transparent}input[type=file]{padding:var(--ui-spacing);font-size:1rem}input[readonly],textarea[readonly]{opacity:.7;background-color:var(--ui-muted, rgba(128, 128, 128, .05));cursor:not-allowed;-webkit-user-select:none;user-select:none;border-color:var(--ui-border-color)!important}input[readonly]:hover,textarea[readonly]:hover{border-color:var(--ui-border-color)!important}input[readonly]:focus,textarea[readonly]:focus,input[readonly]:focus-visible,textarea[readonly]:focus-visible{outline:none!important;border-color:var(--ui-border-color)!important;box-shadow:none!important}input:disabled,textarea:disabled,select:disabled{opacity:.5;background-color:var(--ui-muted, rgba(128, 128, 128, .1));cursor:not-allowed;-webkit-user-select:none;user-select:none;border-style:dashed;border-color:var(--ui-border-color)!important}input:disabled:hover,textarea:disabled:hover,select:disabled:hover{border-color:var(--ui-border-color)!important}input:disabled:focus,textarea:disabled:focus,select:disabled:focus,input:disabled:focus-visible,textarea:disabled:focus-visible,select:disabled:focus-visible{outline:none!important;border-color:var(--ui-border-color)!important;box-shadow:none!important}input[type=checkbox][readonly],input[type=radio][readonly]{pointer-events:none;opacity:.7}input[type=checkbox]:disabled,input[type=radio]:disabled{cursor:not-allowed;opacity:.5}input[type=color][readonly]{pointer-events:none;opacity:.7}input[type=color]:disabled{cursor:not-allowed;opacity:.5}input[type=file][readonly]{pointer-events:none;opacity:.7}input[type=file]:disabled{cursor:not-allowed;opacity:.5}input[aria-invalid=true],textarea[aria-invalid=true],select[aria-invalid=true]{border-color:var(--ui-destructive)!important}input[aria-invalid=false],textarea[aria-invalid=false],select[aria-invalid=false]{border-color:var(--ui-success)!important}input.primary,textarea.primary{--input-color: var(--ui-primary);--input-hover: var(--ui-primary-hover);border-color:var(--input-color);accent-color:var(--input-color)}input.secondary,textarea.secondary{--input-color: var(--ui-secondary);--input-hover: var(--ui-secondary-hover);border-color:var(--input-color);accent-color:var(--input-color)}input.success,textarea.success{--input-color: var(--ui-success);--input-hover: var(--ui-success-hover);border-color:var(--input-color);accent-color:var(--input-color)}input.warning,textarea.warning{--input-color: var(--ui-warning);--input-hover: var(--ui-warning-hover);border-color:var(--input-color);accent-color:var(--input-color)}input.destructive,textarea.destructive{--input-color: var(--ui-destructive);--input-hover: var(--ui-destructive-hover);border-color:var(--input-color);accent-color:var(--input-color)}input.info,textarea.info{--input-color: var(--ui-info);--input-hover: var(--ui-info-hover);border-color:var(--input-color);accent-color:var(--input-color)}input.contrast,textarea.contrast{--input-color: var(--ui-contrast);--input-hover: var(--ui-contrast-hover);border-color:var(--input-color);accent-color:var(--input-color)}input:hover,textarea:hover{border-color:var(--input-hover)}input:focus:not([readonly]):not(:disabled),input:focus-visible:not([readonly]):not(:disabled),textarea:focus:not([readonly]):not(:disabled),textarea:focus-visible:not([readonly]):not(:disabled){border-color:var(--input-color);outline-color:var(--input-color)}input[readonly],textarea[readonly]{border-color:var(--ui-border-color)!important;accent-color:var(--ui-border-color)!important}a{color:var(--ui-primary);text-decoration:underline;text-decoration-thickness:1px;text-underline-offset:.125em;cursor:pointer;transition:color .2s ease,text-decoration-color .2s ease,text-decoration-thickness .2s ease,opacity .2s ease;outline-offset:2px}a:visited:not([role=button]){color:var(--ui-primary);opacity:.85}a:hover:not([role=button]){color:var(--ui-primary--hover);text-decoration-thickness:2px}a:active:not([role=button]){color:var(--ui-primary--active);opacity:.9}a:focus:not([role=button]){outline:none}a:focus-visible:not([role=button]){outline:2px solid var(--ui-primary);outline-offset:2px;border-radius:var(--ui-radius)}a[href^=http]:not([href*=localhost]):not([role=button]):after,a[target=_blank]:not([role=button]):after{content:" ↗";font-size:.875rem;opacity:.7;display:inline-block;margin-left:.125em;transition:opacity .2s ease}a[href^=http]:not([href*=localhost]):not([role=button]):hover:after,a[target=_blank]:not([role=button]):hover:after{opacity:1}a[href^="mailto:"]:not([role=button]):before{content:"✉ ";opacity:.7;margin-right:.125em}a[href^="tel:"]:not([role=button]):before{content:"☎ ";opacity:.7;margin-right:.125em}a.anchor-link,a.header-link{text-decoration:none;opacity:0;transition:opacity .2s ease;margin-left:.5em}h1:hover>a.anchor-link,h2:hover>a.anchor-link,h3:hover>a.anchor-link,h4:hover>a.anchor-link,h5:hover>a.anchor-link,h6:hover>a.anchor-link,h1:hover>a.header-link,h2:hover>a.header-link,h3:hover>a.header-link,h4:hover>a.header-link,h5:hover>a.header-link,h6:hover>a.header-link{opacity:.7}a.anchor-link:hover,a.header-link:hover{opacity:1!important}a[aria-disabled=true]:not([role=button]),a.disabled:not([role=button]){color:var(--ui-muted-text);opacity:.5;cursor:not-allowed;pointer-events:none;text-decoration-color:var(--ui-muted-text)}a.secondary:not([role=button]){color:var(--ui-secondary)}a.secondary:not([role=button]):hover{color:var(--ui-secondary--hover)}a.secondary:not([role=button]):active{color:var(--ui-secondary--active)}a.success:not([role=button]){color:var(--ui-success)}a.success:not([role=button]):hover{color:var(--ui-success--hover)}a.success:not([role=button]):active{color:var(--ui-success--active)}a.warning:not([role=button]){color:var(--ui-warning)}a.warning:not([role=button]):hover{color:var(--ui-warning--hover)}a.warning:not([role=button]):active{color:var(--ui-warning--active)}a.destructive:not([role=button]){color:var(--ui-destructive)}a.destructive:not([role=button]):hover{color:var(--ui-destructive--hover)}a.destructive:not([role=button]):active{color:var(--ui-destructive--active)}a.info:not([role=button]){color:var(--ui-info)}a.info:not([role=button]):hover{color:var(--ui-info--hover)}a.info:not([role=button]):active{color:var(--ui-info--active)}a.muted:not([role=button]){color:var(--ui-muted-text);opacity:.85}a.muted:not([role=button]):hover{opacity:1}a.contrast:not([role=button]){color:var(--ui-contrast)}a.contrast:not([role=button]):hover{color:var(--ui-contrast--hover)}a.contrast:not([role=button]):active{color:var(--ui-contrast--active)}a.no-underline:not([role=button]){text-decoration:none}a.no-underline:not([role=button]):hover{text-decoration:underline}a.block:not([role=button]){display:block;width:100%}a.inline-block:not([role=button]){display:inline-block}nav a:not([role=button]),a.nav-link:not([role=button]){text-decoration:none;padding:calc(var(--ui-spacing) / 2) var(--ui-spacing);border-radius:var(--ui-radius);transition:background-color .2s ease,color .2s ease}nav a:not([role=button]):hover,a.nav-link:not([role=button]):hover{background-color:var(--ui-color-100)}nav a.active:not([role=button]),a.nav-link.active:not([role=button]){background-color:var(--ui-primary);color:var(--ui-primary-text)}.breadcrumb a:not([role=button]),a.breadcrumb-link:not([role=button]){text-decoration:none;color:var(--ui-muted-text)}.breadcrumb a:not([role=button]):hover,a.breadcrumb-link:not([role=button]):hover{color:var(--ui-primary);text-decoration:underline}a.card-link{position:absolute;top:0;left:0;right:0;bottom:0;z-index:1;text-indent:-9999px}@media (prefers-reduced-motion: reduce){a:not([role=button]){transition:none}}@media (prefers-contrast: high){a:not([role=button]){text-decoration-thickness:2px}a:not([role=button]):hover{text-decoration-thickness:3px}a:focus-visible:not([role=button]){outline-width:3px;outline-offset:3px}}@media print{a[href]:not([role=button]):after{content:" (" attr(href) ")";font-size:.875rem;font-weight:400}a[href^="javascript:"]:not([role=button]):after,a[href^="#"]:not([role=button]):after{content:""}}button,[role=button],input[type=button],input[type=submit],input[type=reset]{display:inline-flex;align-items:center;justify-content:center;padding:calc(var(--ui-spacing) / 2) var(--ui-spacing);font-size:1.1rem;font-weight:500;line-height:1.5;text-transform:capitalize;text-decoration:none;white-space:nowrap;vertical-align:middle;border:var(--ui-border-width) solid transparent;border-radius:var(--ui-radius);background-color:var(--ui-primary);color:var(--ui-primary-text);cursor:pointer;-webkit-user-select:none;user-select:none;transition:all .25s ease-in-out;position:relative;overflow:hidden}a[role=button]{text-underline-offset:initial}button:hover,[role=button]:hover,input[type=button]:hover,input[type=submit]:hover,input[type=reset]:hover{transform:translateY(-1px);box-shadow:0 2px 4px #0000001a}button:active,[role=button]:active,input[type=button]:active,input[type=submit]:active,input[type=reset]:active,button.active,[role=button].active{transform:translateY(1px);box-shadow:none;transition:none}button:focus-visible,[role=button]:focus-visible,input[type=button]:focus-visible,input[type=submit]:focus-visible,input[type=reset]:focus-visible{outline:2px solid var(--ui-primary);outline-offset:2px}button,[role=button]{--btn-bg: var(--ui-primary);--btn-bg-hover: var(--ui-primary-hover);--btn-bg-active: var(--ui-primary-active);--btn-text: var(--ui-primary-text);--btn-border: transparent}button.secondary,[role=button].secondary{--btn-bg: var(--ui-secondary);--btn-bg-hover: var(--ui-secondary-hover);--btn-bg-active: var(--ui-secondary-active);--btn-text: var(--ui-secondary-text)}button.success,[role=button].success{--btn-bg: var(--ui-success);--btn-bg-hover: var(--ui-success-hover);--btn-bg-active: var(--ui-success-active);--btn-text: var(--ui-success-text)}button.warning,[role=button].warning{--btn-bg: var(--ui-warning);--btn-bg-hover: var(--ui-warning-hover);--btn-bg-active: var(--ui-warning-active);--btn-text: var(--ui-warning-text)}button.destructive,[role=button].destructive,input[type=reset]{--btn-bg: var(--ui-destructive);--btn-bg-hover: var(--ui-destructive-hover);--btn-bg-active: var(--ui-destructive-active);--btn-text: var(--ui-destructive-text)}button.info,[role=button].info{--btn-bg: var(--ui-info);--btn-bg-hover: var(--ui-info-hover);--btn-bg-active: var(--ui-info-active);--btn-text: var(--ui-info-text)}button.contrast,[role=button].contrast{--btn-bg: var(--ui-contrast);--btn-bg-hover: var(--ui-contrast-hover);--btn-bg-active: var(--ui-contrast-active);--btn-text: var(--ui-contrast-text)}button.muted,[role=button].muted{--btn-bg: var(--ui-muted);--btn-bg-hover: var(--ui-color-200);--btn-bg-active: var(--ui-color-300);--btn-text: var(--ui-muted-text)}button:not(.outline):not(.ghost),[role=button]:not(.outline):not(.ghost),input[type=button]:not(.outline):not(.ghost),input[type=submit]:not(.outline):not(.ghost),input[type=reset]:not(.outline):not(.ghost){background-color:var(--btn-bg);color:var(--btn-text);border-color:var(--btn-border)}button:not(.outline):not(.ghost):hover,[role=button]:not(.outline):not(.ghost):hover,input[type=button]:not(.outline):not(.ghost):hover,input[type=submit]:not(.outline):not(.ghost):hover,input[type=reset]:not(.outline):not(.ghost):hover{background-color:var(--btn-bg-hover)}button:not(.outline):not(.ghost):active,[role=button]:not(.outline):not(.ghost):active,input[type=button]:not(.outline):not(.ghost):active,input[type=submit]:not(.outline):not(.ghost):active,input[type=reset]:not(.outline):not(.ghost):active{background-color:var(--btn-bg-active)}button.outline,[role=button].outline{background-color:transparent;color:var(--btn-bg);border-color:currentColor}button.outline:hover,[role=button].outline:hover{background-color:transparent;color:var(--btn-bg-hover)}button.outline:active,[role=button].outline:active{background-color:transparent;color:var(--btn-bg-active)}button.ghost,[role=button].ghost{background-color:transparent;color:var(--btn-bg);border-color:transparent}button.ghost:hover,[role=button].ghost:hover{background-color:var(--ui-color-100);color:var(--btn-bg-hover)}button.ghost:active,[role=button].ghost:active{background-color:transparent;color:var(--btn-bg-active)}button.small,[role=button].small{padding:calc(var(--ui-spacing) / 4) calc(var(--ui-spacing) / 1.5);font-size:.875rem}button.large,[role=button].large{padding:calc(var(--ui-spacing) * .75) calc(var(--ui-spacing) * 1.5);font-size:1.25rem}button.round,[role=button].round{border-radius:9999px}button.icon,[role=button].icon{padding:calc(var(--ui-spacing) / 2);aspect-ratio:1;min-width:2.5rem;min-height:2.5rem}button.small.icon,[role=button].small.icon{padding:calc(var(--ui-spacing) / 4);min-width:2rem;min-height:2rem}button.icon:focus,[role=button].icon:focus{outline:none}button.full,[role=button].full{width:100%}button.loading,[role=button].loading{color:transparent;pointer-events:none}button.loading:after,[role=button].loading:after{content:"";position:absolute;width:1rem;height:1rem;margin:auto;border:2px solid transparent;border-radius:50%;border-top-color:currentColor;animation:spinner .6s linear infinite}button>svg,[role=button]>svg{width:1.25em;height:1.25em;vertical-align:middle;fill:currentColor}button>svg:first-child:not(:last-child),[role=button]>svg:first-child:not(:last-child){margin-right:calc(var(--ui-spacing) / 2)}button>svg:last-child:not(:first-child),[role=button]>svg:last-child:not(:first-child){margin-left:calc(var(--ui-spacing) / 2)}.button-group{display:inline-flex;gap:0}.button-group>button,.button-group>[role=button]{border-radius:0;margin-left:calc(var(--ui-border-width) * -1)}.button-group>:first-child{border-radius:var(--ui-radius) 0 0 var(--ui-radius);margin-left:0}.button-group>:last-child{border-radius:0 var(--ui-radius) var(--ui-radius) 0}.button-group>:focus{z-index:1}@media (prefers-reduced-motion: reduce){button,[role=button],input[type=button],input[type=submit],input[type=reset]{transition:none}}@media (prefers-contrast: high){button:focus-visible,[role=button]:focus-visible,input[type=button]:focus-visible,input[type=submit]:focus-visible,input[type=reset]:focus-visible{outline-width:3px;outline-offset:3px}}@media (hover: none) and (pointer: coarse){button,[role=button],input[type=button],input[type=submit],input[type=reset]{min-height:44px;min-width:44px}}blockquote{position:relative;margin:calc(var(--ui-spacing) * 2) 0;padding:var(--ui-spacing) calc(var(--ui-spacing) * 3);padding-left:calc(var(--ui-spacing) * 4);background-color:var(--ui-color-50);border-left:4px solid var(--ui-primary);border-radius:var(--ui-radius);font-style:italic;color:var(--ui-text);line-height:1.6;overflow:hidden}blockquote:before{content:"“";position:absolute;left:calc(var(--ui-spacing) * 1.5);top:calc(var(--ui-spacing) * .5);font-size:3rem;color:var(--ui-primary);opacity:.25;font-style:normal;line-height:1;font-weight:700}blockquote blockquote{margin:var(--ui-spacing) 0;border-left-width:3px;opacity:.95}blockquote cite{display:block;margin-top:var(--ui-spacing);font-size:.875rem;font-style:normal;color:var(--ui-muted-text);text-align:right}blockquote cite:before{content:"— "}blockquote footer{margin-top:var(--ui-spacing);font-size:.875rem;font-style:normal;color:var(--ui-muted-text)}blockquote footer:before{content:"— "}blockquote p{margin:0;padding:0}blockquote p:not(:last-child){margin-bottom:var(--ui-spacing)}blockquote p:first-child{margin-top:0}blockquote.clean{background-color:transparent;padding-left:calc(var(--ui-spacing) * 3)}blockquote.clean:before{left:var(--ui-spacing)}blockquote.centered{text-align:center;border-left:none;border-top:2px solid var(--ui-primary);border-bottom:2px solid var(--ui-primary);padding:calc(var(--ui-spacing) * 2) var(--ui-spacing);max-width:40rem;margin-left:auto;margin-right:auto}blockquote.centered:before{left:50%;transform:translate(-50%)}blockquote.centered cite,blockquote.centered footer{text-align:center}blockquote.large{font-size:1.25rem;padding:calc(var(--ui-spacing) * 3) calc(var(--ui-spacing) * 4);padding-left:calc(var(--ui-spacing) * 5)}blockquote.large:before{font-size:4rem}blockquote.small{font-size:.875rem;padding:var(--ui-spacing) calc(var(--ui-spacing) * 2);padding-left:calc(var(--ui-spacing) * 3);border-left-width:3px}blockquote.small:before{font-size:2rem;left:var(--ui-spacing)}blockquote.callout{background-color:var(--ui-primary);color:var(--ui-primary-text);border-left-color:var(--ui-primary-text);font-style:normal;font-weight:500}blockquote.callout:before{color:var(--ui-primary-text);opacity:.15}blockquote.callout cite,blockquote.callout footer{color:var(--ui-primary-text);opacity:.85}blockquote.primary{border-left-color:var(--ui-primary);background-color:#2bee2b0d}blockquote.primary:before{color:var(--ui-primary)}blockquote.secondary{border-left-color:var(--ui-secondary);background-color:var(--ui-color-100)}blockquote.secondary:before{color:var(--ui-secondary)}blockquote.success{border-left-color:var(--ui-success);background-color:#2bee6c0d}blockquote.success:before{color:var(--ui-success)}blockquote.warning{border-left-color:var(--ui-warning);background-color:#fabd2e0d}blockquote.warning:before{color:var(--ui-warning)}blockquote.destructive{border-left-color:var(--ui-destructive);background-color:#f042420d}blockquote.destructive:before{color:var(--ui-destructive)}blockquote.info{border-left-color:var(--ui-info);background-color:#13a4ec0d}blockquote.info:before{color:var(--ui-info)}blockquote.muted{border-left-color:var(--ui-muted);background-color:var(--ui-muted);color:var(--ui-muted-text)}blockquote.muted:before{color:var(--ui-muted-text)}blockquote.pull-left,blockquote.pull-right{width:40%;margin:var(--ui-spacing);font-size:1.125rem;border-top:2px solid var(--ui-primary);border-bottom:2px solid var(--ui-primary);border-left:none;padding:calc(var(--ui-spacing) * 2) 0;background-color:transparent}blockquote.pull-left{float:left;margin-left:0;margin-right:calc(var(--ui-spacing) * 3);text-align:left}blockquote.pull-right{float:right;margin-right:0;margin-left:calc(var(--ui-spacing) * 3);text-align:right}blockquote.pull-left:before,blockquote.pull-right:before{display:none}blockquote.testimonial{text-align:center;border:none;background-color:var(--ui-color-50);padding:calc(var(--ui-spacing) * 3);border-radius:calc(var(--ui-radius) * 2);position:relative;margin:calc(var(--ui-spacing) * 4) auto;max-width:40rem}blockquote.testimonial:before{content:"“";position:absolute;top:calc(var(--ui-spacing) * -1);left:50%;transform:translate(-50%);font-size:5rem;color:var(--ui-primary);opacity:.15}blockquote.testimonial:after{content:"”";position:absolute;bottom:calc(var(--ui-spacing) * -2);right:50%;transform:translate(50%);font-size:5rem;color:var(--ui-primary);opacity:.15}blockquote.testimonial cite,blockquote.testimonial footer{margin-top:calc(var(--ui-spacing) * 2);text-align:center}blockquote.tweet{border-left-color:#1da1f2;background-color:#1ca0f20d;position:relative;padding-right:calc(var(--ui-spacing) * 5)}blockquote.tweet:before{content:"𝕏";color:#1da1f2;font-weight:700;font-style:normal}blockquote.code{font-family:monospace;font-style:normal;background-color:var(--ui-muted);border-left-color:var(--ui-secondary)}blockquote.code:before{content:"</>";font-family:monospace;color:var(--ui-secondary)}@media (prefers-reduced-motion: reduce){blockquote{transition:none}}@media (prefers-contrast: high){blockquote{border-left-width:6px;border-left-style:solid}blockquote.centered{border-top-width:3px;border-bottom-width:3px}}@media print{blockquote{page-break-inside:avoid;border-left-width:2pt;background-color:transparent}blockquote:before{display:none}}@media (max-width: 640px){blockquote.pull-left,blockquote.pull-right{float:none;width:100%;margin:calc(var(--ui-spacing) * 2) 0}blockquote.large{font-size:1.125rem;padding:calc(var(--ui-spacing) * 2) calc(var(--ui-spacing) * 3);padding-left:calc(var(--ui-spacing) * 4)}blockquote.testimonial{padding:calc(var(--ui-spacing) * 2)}}code,pre{background-color:var(--ui-muted);color:var(--ui-muted-text);padding:16px;line-height:var(--ui-line-height-code);border:none;border-radius:8px}code{display:inline-block;padding:.1rem .25rem}pre{display:block;overflow:auto}article{padding:calc(var(--ui-spacing) * 1.5)}section{margin:var(--ui-spacing) 0}details{padding:var(--ui-spacing);border-radius:var(--ui-radius)}details:not([open])>*:not(summary){display:none}summary{display:flex;cursor:pointer;align-items:center;justify-content:space-between;padding:var(--ui-spacing);overflow:hidden;list-style:none;border-radius:var(--ui-radius)}summary::-webkit-details-marker{display:none}summary:after{content:"+";padding-left:1rem;padding-right:var(--ui-spacing)}details[open] summary:after{content:"-"}details[open] summary{border-bottom-left-radius:0;border-bottom-right-radius:0;margin-bottom:var(--ui-spacing)}details.primary{border:var(--ui-border-width) var(--ui-border-style) var(--ui-primary)}details.primary summary{background-color:var(--ui-primary);color:var(--ui-primary-text)}details.secondary{border:var(--ui-border-width) var(--ui-border-style) var(--ui-secondary)}details.secondary summary{background-color:var(--ui-secondary);color:var(--ui-secondary-text)}details.info{border:var(--ui-border-width) var(--ui-border-style) var(--ui-info)}details.info summary{background-color:var(--ui-info);color:var(--ui-info-text)}details.success{border:var(--ui-border-width) var(--ui-border-style) var(--ui-success)}details.success summary{background-color:var(--ui-success);color:var(--ui-success-text)}details.warning{border:var(--ui-border-width) var(--ui-border-style) var(--ui-warning)}details.warning summary{background-color:var(--ui-warning);color:var(--ui-warning-text)}details.destructive{border:var(--ui-border-width) var(--ui-border-style) var(--ui-destructive)}details.destructive summary{background-color:var(--ui-destructive);color:var(--ui-destructive-text)}details.contrast{border:var(--ui-border-width) var(--ui-border-style) var(--ui-contrast)}details.contrast summary{background-color:var(--ui-contrast);color:var(--ui-contrast-text)}dialog{-ms-overflow-style:none;scrollbar-width:none;position:fixed;background-color:var(--ui-bg);color:var(--ui-text);padding:var(--ui-spacing);margin:auto;border:var(--ui-border-width) var(--ui-border-style) var(--ui-border-color);border-radius:var(--ui-radius)}dialog::backdrop{background-color:var(--ui-backdrop-color);backdrop-filter:blur(5px);-webkit-backdrop-filter:blur(5px)}dialog.clean{border:none;background-color:transparent}dialog.fullscreen{width:100%;height:100%;padding:env(safe-area-inset-top,var(--ui-spacing)) env(safe-area-inset-right,var(--ui-spacing)) env(safe-area-inset-bottom,var(--ui-spacing)) env(safe-area-inset-left,var(--ui-spacing))}dialog::-webkit-scrollbar{display:none}dialog.primary{border-color:var(--ui-primary);background-color:var(--ui-primary);color:var(--ui-primary-text)}dialog.secondary{border-color:var(--ui-secondary);background-color:var(--ui-secondary);color:var(--ui-secondary-text)}dialog.info{border-color:var(--ui-info);background-color:var(--ui-info);color:var(--ui-info-text)}dialog.success{border-color:var(--ui-success);background-color:var(--ui-success);color:var(--ui-success-text)}dialog.warning{border-color:var(--ui-warning);background-color:var(--ui-warning);color:var(--ui-warning-text)}dialog.destructive{border-color:var(--ui-destructive);background-color:var(--ui-destructive);color:var(--ui-destructive-text)}dialog.contrast{border-color:var(--ui-contrast);background-color:var(--ui-contrast);color:var(--ui-contrast-text)}fieldset{padding:var(--ui-spacing);padding-top:0;border:var(--ui-border-width) var(--ui-border-style) var(--ui-border-color);border-radius:var(--ui-radius);background-color:transparent}fieldset>legend{padding:0 var(--ui-spacing);font-weight:600;background-color:var(--ui-bg)}fieldset.primary{border-color:var(--ui-primary)}fieldset.primary>legend{color:var(--ui-primary)}fieldset.secondary{border-color:var(--ui-secondary)}fieldset.secondary>legend{color:var(--ui-secondary)}fieldset.info{border-color:var(--ui-info)}fieldset.info>legend{color:var(--ui-info)}fieldset.success{border-color:var(--ui-success)}fieldset.success>legend{color:var(--ui-success)}fieldset.warning{border-color:var(--ui-warning)}fieldset.warning>legend{color:var(--ui-warning)}fieldset.destructive{border-color:var(--ui-destructive)}fieldset.destructive>legend{color:var(--ui-destructive)}fieldset.contrast{border-color:var(--ui-contrast)}fieldset.contrast>legend{color:var(--ui-contrast)}fieldset.filled.primary{background-color:var(--ui-primary);color:var(--ui-primary-text)}fieldset.filled.primary>legend{color:var(--ui-primary-text)}fieldset.filled.secondary{background-color:var(--ui-secondary);color:var(--ui-secondary-text)}fieldset.filled.secondary>legend{color:var(--ui-secondary-text)}fieldset.filled.info{background-color:var(--ui-info);color:var(--ui-info-text)}fieldset.filled.info>legend{color:var(--ui-info-text)}fieldset.filled.success{background-color:var(--ui-success);color:var(--ui-success-text)}fieldset.filled.success>legend{color:var(--ui-success-text)}fieldset.filled.warning{background-color:var(--ui-warning);color:var(--ui-warning-text)}fieldset.filled.warning>legend{color:var(--ui-warning-text)}fieldset.filled.destructive{background-color:var(--ui-destructive);color:var(--ui-destructive-text)}fieldset.filled.destructive>legend{color:var(--ui-destructive-text)}fieldset.filled.contrast{background-color:var(--ui-contrast);color:var(--ui-contrast-text)}fieldset.filled.contrast>legend{color:var(--ui-contrast-text)}fieldset.muted{border-color:var(--ui-muted);background-color:var(--ui-muted);color:var(--ui-muted-text)}fieldset.muted>legend{color:var(--ui-muted-text)}ul{list-style:none;padding:var(--ui-spacing);overflow:hidden}ul>li{padding:var(--ui-spacing)}ol{padding:var(--ui-spacing);padding-left:calc(1.5rem + var(--ui-spacing));overflow:hidden}ol>li{padding:var(--ui-spacing)}dl{padding:var(--ui-spacing)}dl>dd{margin-left:24px}dl>dd:not(:last-child){margin-bottom:.25rem}select{display:inline-block;min-height:calc(var(--ui-spacing) * 2 + 1.5rem);padding:var(--ui-spacing) calc(var(--ui-spacing) * 2);background-color:var(--ui-bg);color:var(--ui-text);border:var(--ui-border-width) solid var(--ui-border-color);border-radius:var(--ui-radius);font-size:1rem;cursor:pointer;accent-color:var(--ui-primary);transition:border-color .25s ease-in-out}select:hover{border-color:var(--ui-primary)}select:focus{outline:none;border-color:var(--ui-primary)}select:focus-visible{outline:2px solid var(--ui-primary);outline-offset:2px}option{background-color:var(--ui-bg);color:var(--ui-text);padding:var(--ui-spacing)}select[aria-invalid=true]{border-color:var(--ui-destructive)!important}select[aria-invalid=false]{border-color:var(--ui-success)!important}select.primary{--select-color: var(--ui-primary)}select.secondary{--select-color: var(--ui-secondary)}select.success{--select-color: var(--ui-success)}select.warning{--select-color: var(--ui-warning)}select.destructive{--select-color: var(--ui-destructive)}select.info{--select-color: var(--ui-info)}select.contrast{--select-color: var(--ui-contrast)}select[class*=primary],select[class*=secondary],select[class*=success],select[class*=warning],select[class*=destructive],select[class*=info],select[class*=contrast]{border-color:var(--select-color);accent-color:var(--select-color)}select[class*=primary]:hover,select[class*=secondary]:hover,select[class*=success]:hover,select[class*=warning]:hover,select[class*=destructive]:hover,select[class*=info]:hover,select[class*=contrast]:hover{border-color:var(--select-color)}select[class*=primary]:focus,select[class*=secondary]:focus,select[class*=success]:focus,select[class*=warning]:focus,select[class*=destructive]:focus,select[class*=info]:focus,select[class*=contrast]:focus,select[class*=primary]:focus-visible,select[class*=secondary]:focus-visible,select[class*=success]:focus-visible,select[class*=warning]:focus-visible,select[class*=destructive]:focus-visible,select[class*=info]:focus-visible,select[class*=contrast]:focus-visible{outline-color:var(--select-color);border-color:var(--select-color)}select:disabled{opacity:.45;cursor:not-allowed}select.small{padding:calc(var(--ui-spacing) / 2) var(--ui-spacing);font-size:.875rem;min-height:calc(var(--ui-spacing) + 1.25rem)}select.large{padding:calc(var(--ui-spacing) * 1.5) calc(var(--ui-spacing) * 2.5);font-size:1.125rem;min-height:calc(var(--ui-spacing) * 3 + 1.75rem)}select[multiple]{min-height:auto;padding:var(--ui-spacing)}select[multiple] option{padding:calc(var(--ui-spacing) / 2) var(--ui-spacing)}@media (prefers-contrast: high){select:focus-visible{outline-width:3px;outline-offset:3px}}hr{width:100%;border:none;height:var(--ui-border-width);margin:var(--ui-spacing) 0;background-color:var(--ui-border-color)}input[type=range]{-webkit-appearance:none;-moz-appearance:none;appearance:none;width:100%;height:8px;margin:1.25em auto;background-color:transparent}input[type=range]::-webkit-slider-runnable-track{-webkit-appearance:none;-moz-appearance:none;appearance:none;height:100%;border:var(--ui-border-width) var(--ui-border-style) var(--ui-border-color);border-radius:var(--ui-radius);background:var(--ui-primary)}input[type=range]::-webkit-slider-thumb{-webkit-appearance:none;-moz-appearance:none;appearance:none;background:var(--ui-primary-text);transition:background .25s linear;height:2em;width:1em;transform:translateY(-40%);border:var(--ui-border-width) var(--ui-border-style) var(--ui-border-color);border-radius:var(--ui-radius);cursor:grab}input[type=range]::-webkit-slider-thumb:active{cursor:grabbing;background:var(--ui-primary)}input[type=range]::-moz-range-track{-moz-appearance:none;-webkit-appearance:none;appearance:none;height:100%;border:var(--ui-border-width) var(--ui-border-style) var(--ui-border-color);border-radius:var(--ui-radius);background:var(--ui-primary)}input[type=range]::-moz-range-thumb{-moz-appearance:none;-webkit-appearance:none;appearance:none;background:var(--ui-primary-text);transition:background .25s linear;height:2em;width:1em;border:var(--ui-border-width) var(--ui-border-style) var(--ui-border-color);border-radius:var(--ui-radius);cursor:grab}input[type=range]::-moz-range-thumb:active{cursor:grabbing;background:var(--ui-primary)}input[type=range]:focus{outline:none}input[type=range]:focus-visible{outline:2px solid var(--ui-primary);outline-offset:2px}input[type=range]:focus-visible::-webkit-slider-thumb{box-shadow:0 0 0 2px var(--ui-bg),0 0 0 4px var(--ui-primary)}input[type=range]:focus-visible::-moz-range-thumb{box-shadow:0 0 0 2px var(--ui-bg),0 0 0 4px var(--ui-primary)}table{table-layout:fixed 100%;border-collapse:separate;width:100%}table,thead,tbody,tr,th,td{border:var(--ui-border-width) var(--ui-border-style) var(--ui-border-color);border-radius:var(--ui-radius)}th{padding:var(--ui-spacing);color:var(--ui-secondary)}td{padding:var(--ui-spacing);background-color:transparent;text-align:center}tr.left>*,th.left,td.left{text-align:left}tr.right>*,th.right,td.right{text-align:right}table.hover tbody tr:hover{background-color:var(--ui-color-100)}table.striped tbody tr:nth-child(2n){background-color:var(--ui-color-50)}table.striped.hover tbody tr:nth-child(2n):hover{background-color:var(--ui-color-150)}table.striped.hover tbody tr:nth-child(odd):hover{background-color:var(--ui-color-100)}table.primary th{background-color:var(--ui-primary);color:var(--ui-primary-text);border-color:var(--ui-primary)}table.secondary th{background-color:var(--ui-secondary);color:var(--ui-secondary-text);border-color:var(--ui-secondary)}table.info th{background-color:var(--ui-info);color:var(--ui-info-text);border-color:var(--ui-info)}table.success th{background-color:var(--ui-success);color:var(--ui-success-text);border-color:var(--ui-success)}table.warning th{background-color:var(--ui-warning);color:var(--ui-warning-text);border-color:var(--ui-warning)}table.destructive th{background-color:var(--ui-destructive);color:var(--ui-destructive-text);border-color:var(--ui-destructive)}table.contrast th{background-color:var(--ui-contrast);color:var(--ui-contrast-text);border-color:var(--ui-contrast)}table.bordered{border-collapse:collapse}table.bordered th,table.bordered td{border:var(--ui-border-width) var(--ui-border-style) var(--ui-border-color)}table.borderless,table.borderless th,table.borderless td,table.borderless thead,table.borderless tbody,table.borderless tr{border:none}table.compact th,table.compact td{padding:8px}table.spacious th,table.spacious td{padding:24px}html{font-size:var(--ui-font-size)}body{font-size:100%}h1{font-size:2rem}h2{font-size:1.75rem}h3{font-size:1.5rem}h4{font-size:1.25rem}h5{font-size:1rem}h6{font-size:.75rem}p,label{padding:var(--ui-spacing)}label[for],label:has(input){cursor:pointer;font-size:.9rem}small{font-size:.75rem}summary{font-size:1.1rem}code,pre{font-size:.95rem}fieldset>legend{font-size:1.15rem}dl>dd{font-size:.95rem}textarea,input:not([role=button]):not([role=submit]):not([role=reset]),select{font-size:1rem}input[type=range]{font-size:1rem}button,[role=button],input[type=button],input[type=submit],input[type=reset]{font-size:1.1rem}td{font-size:.9rem}table.compact th,table.compact td{font-size:.85rem}table.spacious th,table.spacious td{font-size:1rem}.card-title{font-size:1.25rem}.card-subtitle{font-size:.875rem}.card-meta{font-size:.75rem}a[href^=http]:not([href*=localhost]):not([role=button]):after,a[target=_blank]:not([role=button]):after{font-size:.875rem}.app-bar{background-color:var(--ui-backdrop-color);backdrop-filter:blur(5px);-webkit-backdrop-filter:blur(5px);display:flex;flex-flow:row nowrap;align-items:center;justify-content:space-between;padding:calc(var(--ui-spacing) / 2);overflow:hidden;width:100%;gap:var(--ui-spacing)}.app-bar.top,.app-bar.bottom{position:absolute;left:0;right:0;height:var(--ui-app-bar-height)}.app-bar.fixed{position:fixed}.app-bar.top{top:0;border-bottom:var(--ui-border-width) var(--ui-border-style) var(--ui-border-color)}.app-bar.bottom{bottom:0;border-top:var(--ui-border-width) var(--ui-border-style) var(--ui-border-color)}.app-bar>.app-bar-center{overflow:hidden}.app-bar>.app-bar-left,.app-bar>.app-bar-center,.app-bar>.app-bar-right{display:flex;flex-flow:row nowrap;align-items:center;justify-content:center;height:100%;width:100%;gap:var(--ui-spacing)}.app-bar>.app-bar-left,.app-bar>.app-bar-right{width:fit-content}.app-bar>.app-bar-left{justify-content:flex-start}.app-bar>.app-bar-right{justify-content:flex-end}.card{--card-padding: calc(var(--ui-spacing) * 1.5);display:block;background-color:var(--ui-bg);border:var(--ui-border-width) solid var(--ui-border-color);border-radius:var(--ui-radius);overflow:hidden;position:relative;transition:transform .2s ease-in-out,box-shadow .2s ease-in-out,border-color .2s ease-in-out}.card:hover{box-shadow:0 4px 12px hsla(var(--ui-hue),var(--ui-saturation),20%,.1),0 2px 4px hsla(var(--ui-hue),var(--ui-saturation),20%,.06)}.card:focus-within{border-color:var(--ui-primary);outline:2px solid var(--ui-primary);outline-offset:2px}.card-header,.card-body,.card-footer{padding:var(--card-padding)}.card-header{border-bottom:var(--ui-border-width) solid var(--ui-border-color);background-color:var(--ui-color-50)}.card-footer{border-top:var(--ui-border-width) solid var(--ui-border-color);background-color:var(--ui-color-50)}.card.elevated{border:none;box-shadow:0 2px 8px hsla(var(--ui-hue),var(--ui-saturation),20%,.08),0 1px 2px hsla(var(--ui-hue),var(--ui-saturation),20%,.04)}.card.elevated:hover{box-shadow:0 8px 24px hsla(var(--ui-hue),var(--ui-saturation),20%,.12),0 4px 8px hsla(var(--ui-hue),var(--ui-saturation),20%,.08)}.card.flat{border:none;box-shadow:none;background-color:var(--ui-muted)}.card.flat:hover{background-color:var(--ui-color-200)}.card.outlined{background-color:transparent;border-width:calc(var(--ui-border-width) * 1.5)}.card.outlined:hover{border-color:var(--ui-primary);background-color:var(--ui-color-50)}.card.seamless .card-header,.card.seamless .card-footer{border:none;background-color:transparent}.card.interactive{cursor:pointer;-webkit-user-select:none;user-select:none}.card.interactive:active{transition:transform .1s ease-in-out}.card.interactive:focus{outline:2px solid var(--ui-primary);outline-offset:2px}.card.compact{--card-padding: calc(var(--ui-spacing) * 1)}.card.comfortable{--card-padding: calc(var(--ui-spacing) * 1.5)}.card.spacious{--card-padding: calc(var(--ui-spacing) * 2)}.card-image{width:100%;height:auto;display:block;object-fit:cover}.card-image.aspect-video{aspect-ratio:16 / 9}.card-image.aspect-square{aspect-ratio:1 / 1}.card-image.aspect-portrait{aspect-ratio:3 / 4}.card-title{margin:0 0 calc(var(--ui-spacing) / 2) 0;font-weight:600;font-size:1.25rem;line-height:1.3;color:var(--ui-text)}.card-subtitle{margin:0 0 var(--ui-spacing) 0;font-size:.875rem;color:var(--ui-muted-text);line-height:1.4}.card-description{margin:0 0 var(--ui-spacing) 0;color:var(--ui-text);line-height:1.5}.card-meta{margin:0;font-size:.75rem;color:var(--ui-muted-text)}.card-actions{display:flex;gap:calc(var(--ui-spacing) / 2);align-items:center;justify-content:flex-start;flex-wrap:wrap}.card-actions.center{justify-content:center}.card-actions.end{justify-content:flex-end}.card-actions.between{justify-content:space-between}.card-actions.stretch{justify-content:stretch}.card-actions.stretch>*{flex:1}.card.horizontal{display:flex;flex-direction:row;align-items:stretch}.card.horizontal .card-image{width:auto;min-width:120px;max-width:200px;flex-shrink:0}.card.horizontal .card-content{flex:1;display:flex;flex-direction:column}.card.horizontal .card-body{flex:1}.card.media{overflow:hidden}.card.media .card-image{transition:transform .3s ease-in-out}.card.media:hover .card-image{transform:scale(1.05)}.card-grid{display:grid;gap:var(--ui-spacing);grid-template-columns:repeat(auto-fill,minmax(280px,1fr))}.card-grid.compact{grid-template-columns:repeat(auto-fill,minmax(240px,1fr))}.card-grid.wide{grid-template-columns:repeat(auto-fill,minmax(320px,1fr))}.card.primary{--card-color: var(--ui-primary);--card-color-hover: var(--ui-primary-hover);--card-color-text: var(--ui-primary-text);border-color:var(--card-color)}.card.secondary{--card-color: var(--ui-secondary);--card-color-hover: var(--ui-secondary-hover);--card-color-text: var(--ui-secondary-text);border-color:var(--card-color)}.card.success{--card-color: var(--ui-success);--card-color-hover: var(--ui-success-hover);--card-color-text: var(--ui-success-text);border-color:var(--card-color)}.card.warning{--card-color: var(--ui-warning);--card-color-hover: var(--ui-warning-hover);--card-color-text: var(--ui-warning-text);border-color:var(--card-color)}.card.destructive{--card-color: var(--ui-destructive);--card-color-hover: var(--ui-destructive-hover);--card-color-text: var(--ui-destructive-text);border-color:var(--card-color)}.card.info{--card-color: var(--ui-info);--card-color-hover: var(--ui-info-hover);--card-color-text: var(--ui-info-text);border-color:var(--card-color)}.card.contrast{--card-color: var(--ui-contrast);--card-color-hover: var(--ui-contrast-hover);--card-color-text: var(--ui-contrast-text);border-color:var(--card-color)}.card.muted{--card-color: var(--ui-muted);--card-color-hover: var(--ui-color-300);--card-color-text: var(--ui-muted-text);border-color:var(--card-color)}.card.primary .card-header,.card.secondary .card-header,.card.success .card-header,.card.warning .card-header,.card.destructive .card-header,.card.info .card-header,.card.contrast .card-header,.card.muted .card-header{background-color:var(--card-color);color:var(--card-color-text);border-color:var(--card-color)}.card.outlined.primary .card-header,.card.outlined.secondary .card-header,.card.outlined.success .card-header,.card.outlined.warning .card-header,.card.outlined.destructive .card-header,.card.outlined.info .card-header,.card.outlined.contrast .card-header,.card.outlined.muted .card-header{background-color:transparent;color:var(--card-color);border-color:transparent}.card.primary:hover,.card.secondary:hover,.card.success:hover,.card.warning:hover,.card.destructive:hover,.card.info:hover,.card.contrast:hover,.card.muted:hover{border-color:var(--card-color-hover)}@media (max-width: 640px){.card.horizontal{flex-direction:column}.card.horizontal .card-image{width:100%;min-width:auto;max-width:none}.card-grid{grid-template-columns:1fr}}@media (prefers-reduced-motion: reduce){.card,.card.media .card-image{transition:none}}@media (prefers-contrast: high){.card{border-width:calc(var(--ui-border-width) * 2)}.card:focus-within{outline-width:3px;outline-offset:3px}}.container{display:block;width:100%;max-width:65rem;margin:0 auto;padding:var(--ui-spacing);overflow:auto}.container.fluid{width:100%;max-width:100%}.drawer{-ms-overflow-style:none;scrollbar-width:none;background-color:var(--ui-backdrop-color);position:absolute;top:0;left:-100%;overflow:auto;transition:left .25s ease-in-out,width .25s ease-in-out;width:var(--width, fit-content);max-width:calc(100% - 2.5rem);height:100%;border-right:var(--ui-border-width) var(--ui-border-style) var(--ui-border-color)}.drawer::-webkit-scrollbar{display:none}.drawer.open{left:0}.drawer.open:before{left:0}.drawer-backdrop{background-color:var(--ui-backdrop-color);position:absolute;top:0;left:0;width:100%;height:100%}.drawer-backdrop:has(~.drawer:not(.open)){display:none}@keyframes spinner{to{transform:rotate(360deg)}}.spinner{background-color:var(--ui-backdrop-color);position:absolute;top:0;left:0;width:100%;height:100%}.spinner:after{z-index:1;position:absolute;top:50%;left:50%;margin-top:-1.25rem;margin-left:-1.25rem;content:"";width:2.5rem;height:2.5rem;border:2px var(--ui-border-style) var(--ui-border-color);border-top-color:var(--ui-primary);border-radius:50%;animation:spinner .6s linear infinite}.badge{display:inline-flex;align-items:center;justify-content:center;font-size:.75rem;font-weight:600;line-height:1;padding:.25em .5em;border-radius:var(--ui-radius);white-space:nowrap;vertical-align:baseline;transition:all .2s ease-in-out}.badge-xs{font-size:.625rem;padding:.125em .375em}.badge-sm{font-size:.6875rem;padding:.1875em .4375em}.badge-md{font-size:.75rem;padding:.25em .5em}.badge-lg{font-size:.875rem;padding:.375em .625em}.badge-xl{font-size:1rem;padding:.5em .75em}.badge-pill{border-radius:9999px}.badge-square{border-radius:0}.badge-rounded{border-radius:var(--ui-radius)}.badge-dot{width:.5rem;height:.5rem;padding:0;border-radius:50%;display:inline-block}.badge-dot-sm{width:.375rem;height:.375rem}.badge-dot-lg{width:.625rem;height:.625rem}.badge-dot-xl{width:.75rem;height:.75rem}.badge-dismissible{padding-right:1.5em;position:relative}.badge-dismissible .badge-close{position:absolute;right:.25em;top:50%;transform:translateY(-50%);width:1em;height:1em;padding:0;background:none;border:none;color:inherit;opacity:.7;cursor:pointer;transition:opacity .2s}.badge-dismissible .badge-close:hover{opacity:1}.badge-container{position:relative;display:inline-block}.badge-top-right{position:absolute;top:0;right:0;transform:translate(50%,-50%)}.badge-top-left{position:absolute;top:0;left:0;transform:translate(-50%,-50%)}.badge-bottom-right{position:absolute;bottom:0;right:0;transform:translate(50%,50%)}.badge-bottom-left{position:absolute;bottom:0;left:0;transform:translate(-50%,50%)}.badge-counter{min-width:1.5em;height:1.5em;padding:.25em;border-radius:50%;font-size:.625rem;font-weight:700}.badge-counter-sm{min-width:1.25em;height:1.25em;font-size:.5625rem}.badge-counter-lg{min-width:1.75em;height:1.75em;font-size:.75rem}.badge-max-count{min-width:2em}.badge-group{display:inline-flex;gap:.25rem;flex-wrap:wrap;align-items:center}.badge-group-vertical{flex-direction:column;align-items:flex-start}.badge-list{display:flex;flex-wrap:wrap;gap:.5rem;padding:0;margin:0;list-style:none}@keyframes badge-pulse{0%{box-shadow:0 0 0 0 currentColor}70%{box-shadow:0 0 0 .5rem transparent}to{box-shadow:0 0 0 0 transparent}}@keyframes badge-ping{75%,to{transform:scale(2);opacity:0}}@keyframes badge-bounce{0%,to{transform:translateY(0)}50%{transform:translateY(-25%)}}.badge-pulse{animation:badge-pulse 2s infinite}.badge-ping{position:relative}.badge-ping:after{content:"";position:absolute;top:0;right:0;bottom:0;left:0;border-radius:inherit;background-color:currentColor;animation:badge-ping 1.5s cubic-bezier(0,0,.2,1) infinite}.badge-bounce{animation:badge-bounce 1s ease-in-out infinite}.badge-icon{display:inline-flex;align-items:center;gap:.25em}.badge-icon svg,.badge-icon img{width:1em;height:1em;flex-shrink:0}.badge-outline{background-color:transparent;border:1px solid currentColor}.badge-ghost{background-color:transparent;border:1px solid transparent}.badge-disabled,.badge:disabled{opacity:.5;cursor:not-allowed;pointer-events:none}.badge-clickable{cursor:pointer;transition:transform .1s ease-in-out}.badge-clickable:hover{transform:scale(1.05)}.badge-clickable:active{transform:scale(.95)}.badge-status{position:relative;padding-left:1.5em}.badge-status:before{content:"";position:absolute;left:.5em;top:50%;transform:translateY(-50%);width:.5em;height:.5em;border-radius:50%;background-color:currentColor}.badge-status-online:before{background-color:var(--ui-success, #10b981)}.badge-status-offline:before{background-color:var(--ui-muted, #6b7280)}.badge-status-busy:before{background-color:var(--ui-warning, #f59e0b)}.badge-status-away:before{background-color:var(--ui-info, #3b82f6)}.badge-loading{position:relative;color:transparent}.badge-loading:after{content:"";position:absolute;width:1em;height:1em;top:50%;left:50%;margin-left:-.5em;margin-top:-.5em;border:.125em solid currentColor;border-right-color:transparent;border-radius:50%;animation:spin .75s linear infinite}.badge-float{position:fixed;z-index:1000}.badge-float-top{top:1rem;left:50%;transform:translate(-50%)}.badge-float-bottom{bottom:1rem;left:50%;transform:translate(-50%)}.badge-gradient{background:linear-gradient(135deg,var(--ui-primary),var(--ui-secondary));color:#fff}@media (prefers-reduced-motion: reduce){.badge-pulse,.badge-ping:after,.badge-bounce,.badge-loading:after{animation:none}.badge-clickable:hover{transform:none}}@media (prefers-color-scheme: dark){.badge-outline{border-color:#fff3}}.flex{display:flex}.inline-flex{display:inline-flex}.flex-row{flex-direction:row}.flex-col{flex-direction:column}.flex-row-reverse{flex-direction:row-reverse}.flex-col-reverse{flex-direction:column-reverse}.flex-wrap{flex-wrap:wrap}.flex-nowrap{flex-wrap:nowrap}.flex-wrap-reverse{flex-wrap:wrap-reverse}.items-start{align-items:flex-start}.items-center{align-items:center}.items-end{align-items:flex-end}.items-baseline{align-items:baseline}.items-stretch{align-items:stretch}.justify-start{justify-content:flex-start}.justify-center{justify-content:center}.justify-end{justify-content:flex-end}.justify-between{justify-content:space-between}.justify-around{justify-content:space-around}.justify-evenly{justify-content:space-evenly}.gap{gap:var(--ui-spacing)}.gap-0{gap:0}.gap-sm{gap:calc(var(--ui-spacing) / 2)}.gap-lg{gap:calc(var(--ui-spacing) * 2)}.flex-1{flex:1 1 0%}.flex-auto{flex:1 1 auto}.flex-initial{flex:0 1 auto}.flex-none{flex:none}.grow{flex-grow:1}.grow-0{flex-grow:0}.shrink{flex-shrink:1}.shrink-0{flex-shrink:0}.p-0{padding:0}.p{padding:var(--ui-spacing)}.p-sm{padding:calc(var(--ui-spacing) / 2)}.p-lg{padding:calc(var(--ui-spacing) * 2)}.p-xl{padding:calc(var(--ui-spacing) * 3)}.px-0{padding-left:0;padding-right:0}.px{padding-left:var(--ui-spacing);padding-right:var(--ui-spacing)}.px-sm{padding-left:calc(var(--ui-spacing) / 2);padding-right:calc(var(--ui-spacing) / 2)}.px-lg{padding-left:calc(var(--ui-spacing) * 2);padding-right:calc(var(--ui-spacing) * 2)}.py-0{padding-top:0;padding-bottom:0}.py{padding-top:var(--ui-spacing);padding-bottom:var(--ui-spacing)}.py-sm{padding-top:calc(var(--ui-spacing) / 2);padding-bottom:calc(var(--ui-spacing) / 2)}.py-lg{padding-top:calc(var(--ui-spacing) * 2);padding-bottom:calc(var(--ui-spacing) * 2)}.pt{padding-top:var(--ui-spacing)}.pr{padding-right:var(--ui-spacing)}.pb{padding-bottom:var(--ui-spacing)}.pl{padding-left:var(--ui-spacing)}.m-0{margin:0}.m{margin:var(--ui-spacing)}.m-sm{margin:calc(var(--ui-spacing) / 2)}.m-lg{margin:calc(var(--ui-spacing) * 2)}.m-xl{margin:calc(var(--ui-spacing) * 3)}.m-auto{margin:auto}.mx-0{margin-left:0;margin-right:0}.mx{margin-left:var(--ui-spacing);margin-right:var(--ui-spacing)}.mx-sm{margin-left:calc(var(--ui-spacing) / 2);margin-right:calc(var(--ui-spacing) / 2)}.mx-lg{margin-left:calc(var(--ui-spacing) * 2);margin-right:calc(var(--ui-spacing) * 2)}.mx-auto{margin-left:auto;margin-right:auto}.my-0{margin-top:0;margin-bottom:0}.my{margin-top:var(--ui-spacing);margin-bottom:var(--ui-spacing)}.my-sm{margin-top:calc(var(--ui-spacing) / 2);margin-bottom:calc(var(--ui-spacing) / 2)}.my-lg{margin-top:calc(var(--ui-spacing) * 2);margin-bottom:calc(var(--ui-spacing) * 2)}.mt{margin-top:var(--ui-spacing)}.mr{margin-right:var(--ui-spacing)}.mb{margin-bottom:var(--ui-spacing)}.ml{margin-left:var(--ui-spacing)}.w-full{width:100%}.w-auto{width:auto}.w-fit{width:fit-content}.w-min{width:min-content}.w-max{width:max-content}.w-screen{width:100vw}.w-0{width:0}.w-25{width:25%}.w-50{width:50%}.w-75{width:75%}.h-full{height:100%}.h-auto{height:auto}.h-fit{height:fit-content}.h-min{height:min-content}.h-max{height:max-content}.h-screen{height:100vh}.h-0{height:0}.h-25{height:25%}.h-50{height:50%}.h-75{height:75%}.min-w-0{min-width:0}.min-w-full{min-width:100%}.min-w-fit{min-width:fit-content}.min-w-min{min-width:min-content}.min-w-max{min-width:max-content}.max-w-none{max-width:none}.max-w-full{max-width:100%}.max-w-fit{max-width:fit-content}.max-w-min{max-width:min-content}.max-w-max{max-width:max-content}.max-w-screen{max-width:100vw}.max-w-sm{max-width:24rem}.max-w-md{max-width:28rem}.max-w-lg{max-width:32rem}.max-w-xl{max-width:36rem}.max-w-2xl{max-width:42rem}.max-w-3xl{max-width:48rem}.max-w-4xl{max-width:56rem}.max-w-5xl{max-width:64rem}.min-h-0{min-height:0}.min-h-full{min-height:100%}.min-h-screen{min-height:100vh}.min-h-fit{min-height:fit-content}.min-h-min{min-height:min-content}.min-h-max{min-height:max-content}.max-h-none{max-height:none}.max-h-full{max-height:100%}.max-h-screen{max-height:100vh}.max-h-fit{max-height:fit-content}.max-h-min{max-height:min-content}.max-h-max{max-height:max-content}.aspect-auto{aspect-ratio:auto}.aspect-square{aspect-ratio:1 / 1}.aspect-video{aspect-ratio:16 / 9}.aspect-portrait{aspect-ratio:3 / 4}.aspect-landscape{aspect-ratio:4 / 3}.static{position:static}.relative{position:relative}.absolute{position:absolute}.fixed{position:fixed}.sticky{position:sticky}.inset-0{top:0;right:0;bottom:0;left:0}.top-0{top:0}.right-0{right:0}.bottom-0{bottom:0}.left-0{left:0}.top-full{top:100%}.right-full{right:100%}.bottom-full{bottom:100%}.left-full{left:100%}.z-0{z-index:0}.z-10{z-index:10}.z-20{z-index:20}.z-30{z-index:30}.z-40{z-index:40}.z-50{z-index:50}.z-auto{z-index:auto}.block{display:block}.inline-block{display:inline-block}.inline{display:inline}.hidden{display:none}.grid{display:grid}.inline-grid{display:inline-grid}.contents{display:contents}.table{display:table}.visible{visibility:visible}.invisible{visibility:hidden}.opacity-0{opacity:0}.opacity-25{opacity:.25}.opacity-50{opacity:.5}.opacity-75{opacity:.75}.opacity-100{opacity:1}.border{border:var(--ui-border-width) var(--ui-border-style) var(--ui-border-color);border-radius:var(--ui-radius)}.border-top{border-top:var(--ui-border-width) var(--ui-border-style) var(--ui-border-color)}.border-right{border-right:var(--ui-border-width) var(--ui-border-style) var(--ui-border-color)}.border-bottom{border-bottom:var(--ui-border-width) var(--ui-border-style) var(--ui-border-color)}.border-left{border-left:var(--ui-border-width) var(--ui-border-style) var(--ui-border-color)}.backdrop{background-color:var(--ui-backdrop-color);backdrop-filter:blur(5px);-webkit-backdrop-filter:blur(5px)}.primary,.secondary,.success,.warning,.destructive,.info,.contrast,.muted,.error{--color-bg: var(--ui-primary);--color-text: var(--ui-primary-text)}.secondary{--color-bg: var(--ui-secondary);--color-text: var(--ui-secondary-text)}.success{--color-bg: var(--ui-success);--color-text: var(--ui-success-text)}.warning{--color-bg: var(--ui-warning);--color-text: var(--ui-warning-text)}.destructive{--color-bg: var(--ui-destructive);--color-text: var(--ui-destructive-text)}.info{--color-bg: var(--ui-info);--color-text: var(--ui-info-text)}.contrast{--color-bg: var(--ui-contrast);--color-text: var(--ui-contrast-text)}.muted{--color-bg: var(--ui-muted);--color-text: var(--ui-muted-text)}.error{--color-bg: var(--ui-error);--color-text: var(--ui-error-text)}.primary:not(.ghost):not(.outline):not(a):not(button):not([role=button]),.secondary:not(.ghost):not(.outline):not(a):not(button):not([role=button]),.success:not(.ghost):not(.outline):not(a):not(button):not([role=button]),.warning:not(.ghost):not(.outline):not(a):not(button):not([role=button]),.destructive:not(.ghost):not(.outline):not(a):not(button):not([role=button]),.info:not(.ghost):not(.outline):not(a):not(button):not([role=button]),.contrast:not(.ghost):not(.outline):not(a):not(button):not([role=button]),.muted:not(.ghost):not(.outline):not(a):not(button):not([role=button]),.error:not(.ghost):not(.outline):not(a):not(button):not([role=button]){background-color:var(--color-bg);color:var(--color-text)}.primary.ghost,.secondary.ghost,.success.ghost,.warning.ghost,.destructive.ghost,.info.ghost,.contrast.ghost,.muted.ghost,.error.ghost{color:var(--color-bg);background-color:transparent}.primary.outline,.secondary.outline,.success.outline,.warning.outline,.destructive.outline,.info.outline,.contrast.outline,.muted.outline,.error.outline{color:var(--color-bg);border-color:var(--color-bg);background-color:transparent}.text-left{text-align:left}.text-center{text-align:center}.text-right{text-align:right}.text-justify{text-align:justify}.uppercase{text-transform:uppercase}.lowercase{text-transform:lowercase}.capitalize{text-transform:capitalize}.text-xs{font-size:.75rem}.text-sm{font-size:.875rem}.text-base{font-size:1rem}.text-lg{font-size:1.125rem}.text-xl{font-size:1.25rem}.text-2xl{font-size:1.5rem}.text-3xl{font-size:1.875rem}.text-4xl{font-size:2.25rem}.text-thin{--wght: 100}.text-light{--wght: 300}.text-normal{--wght: 400}.text-medium{--wght: 500}.text-semibold{--wght: 600}.text-bold{--wght: 700}.text-extrabold{--wght: 800}.text-black{--wght: 900}.italic{--slnt: -10}.not-italic{--slnt: 0}.leading-none{line-height:1}.leading-tight{line-height:1.25}.leading-snug{line-height:1.375}.leading-normal{line-height:1.5}.leading-relaxed{line-height:1.625}.leading-loose{line-height:2}.underline{text-decoration:underline}.line-through{text-decoration:line-through}.no-underline{text-decoration:none}.truncate{overflow:hidden;text-overflow:ellipsis;white-space:nowrap}.ellipsis{overflow:hidden;text-overflow:ellipsis;white-space:nowrap;max-width:100%}.whitespace-normal{white-space:normal}.whitespace-nowrap{white-space:nowrap}.whitespace-pre{white-space:pre}.whitespace-pre-line{white-space:pre-line}.whitespace-pre-wrap{white-space:pre-wrap}.break-normal{overflow-wrap:normal;word-break:normal}.break-words{overflow-wrap:break-word}.break-all{word-break:break-all}.outline-text{-webkit-text-stroke:1px var(--ui-text);color:transparent}.no-zoom{font-size:1rem!important}.select-none{-webkit-user-select:none;user-select:none}.select-text{-webkit-user-select:text;user-select:text}.select-all{-webkit-user-select:all;user-select:all}.select-auto{-webkit-user-select:auto;user-select:auto}.pointer-events-none{pointer-events:none}.pointer-events-auto{pointer-events:auto}.touch-none{touch-action:none}.touch-auto{touch-action:auto}.touch-manipulation{touch-action:manipulation}.cursor-auto{cursor:auto}.cursor-default{cursor:default}.cursor-pointer{cursor:pointer}.cursor-wait{cursor:wait}.cursor-text{cursor:text}.cursor-move{cursor:move}.cursor-not-allowed{cursor:not-allowed}.cursor-grab{cursor:grab}.cursor-grabbing{cursor:grabbing}.scroll-auto{scroll-behavior:auto}.scroll-smooth{scroll-behavior:smooth}.overflow-auto{overflow:auto}.overflow-hidden{overflow:hidden}.overflow-visible{overflow:visible}.overflow-scroll{overflow:scroll}.overflow-x-auto{overflow-x:auto}.overflow-x-hidden{overflow-x:hidden}.overflow-x-visible{overflow-x:visible}.overflow-x-scroll{overflow-x:scroll}.overflow-y-auto{overflow-y:auto}.overflow-y-hidden{overflow-y:hidden}.overflow-y-visible{overflow-y:visible}.overflow-y-scroll{overflow-y:scroll}.no-scrollbar{-ms-overflow-style:none;scrollbar-width:none}.no-scrollbar::-webkit-scrollbar{display:none}.resize-none{resize:none}.resize{resize:both}.resize-x{resize:horizontal}.resize-y{resize:vertical}.will-change-auto{will-change:auto}.will-change-transform{will-change:transform}.will-change-opacity{will-change:opacity}.will-change-scroll{will-change:scroll-position}.draggable{-webkit-user-drag:element;-webkit-user-select:none;user-select:none}.not-draggable{-webkit-user-drag:none}.tap-highlight-transparent{-webkit-tap-highlight-color:transparent}.focus-within\:outline:focus-within{outline:2px solid var(--ui-primary);outline-offset:2px}.transition-none{transition:none}.transition-all{transition:all .25s ease-in-out}.transition-colors{transition:background-color .25s ease-in-out,border-color .25s ease-in-out,color .25s ease-in-out}.transition-opacity{transition:opacity .25s ease-in-out}.transition-transform{transition:transform .25s ease-in-out}@keyframes spin{0%{transform:rotate(0)}to{transform:rotate(360deg)}}@keyframes pulse{0%,to{opacity:1}50%{opacity:.5}}@keyframes bounce{0%,to{transform:translateY(0)}50%{transform:translateY(-25%)}}.animate-spin{animation:spin 1s linear infinite}.animate-pulse{animation:pulse 2s ease-in-out infinite}.animate-bounce{animation:bounce 1s ease-in-out infinite}.animate-none{animation:none}@media (prefers-reduced-motion: reduce){.transition-all,.transition-colors,.transition-opacity,.transition-transform{transition-duration:.01ms}.animate-spin,.animate-pulse,.animate-bounce{animation:none}}.ui-safe-area-top{padding-top:env(safe-area-inset-top)}.ui-safe-area-right{padding-right:env(safe-area-inset-right)}.ui-safe-area-bottom{padding-bottom:env(safe-area-inset-bottom)}.ui-safe-area-left{padding-left:env(safe-area-inset-left)}.ui-safe-area-x{padding-left:env(safe-area-inset-left);padding-right:env(safe-area-inset-right)}.ui-safe-area-y{padding-top:env(safe-area-inset-top);padding-bottom:env(safe-area-inset-bottom)}.ui-safe-area{padding:env(safe-area-inset-top) env(safe-area-inset-right) env(safe-area-inset-bottom) env(safe-area-inset-left)}.ui-safe-area-margin-top{margin-top:env(safe-area-inset-top)}.ui-safe-area-margin-right{margin-right:env(safe-area-inset-right)}.ui-safe-area-margin-bottom{margin-bottom:env(safe-area-inset-bottom)}.ui-safe-area-margin-left{margin-left:env(safe-area-inset-left)}.debug,.debug *{border:var(--ui-border-width) var(--ui-border-style) red;border-radius:var(--ui-radius)}

```

## Matches in pg-press/internal/web/assets/js/htmx-ext-ws-v2.0.3.min.js

### L1
```
(function(){var v;htmx.defineExtension("ws",{init:function(e){v=e;if(!htmx.createWebSocket){htmx.createWebSocket=t}if(!htmx.config.wsReconnectDelay){htmx.config.wsReconnectDelay="full-jitter"}},onEvent:function(e,t){var n=t.target||t.detail.elt;switch(e){case"htmx:beforeCleanupElement":var r=v.getInternalData(n);if(r.webSocket){r.webSocket.close()}return;case"htmx:beforeProcessNode":h(l(n,"ws-connect"),function(e){s(e)});h(l(n,"ws-send"),function(e){a(e)})}}});function i(e){return e.trim().split(/\s+/)}function r(e){var t=v.getAttributeValue(e,"hx-ws");if(t){var n=i(t);for(var r=0;r<n.length;r++){var s=n[r].split(/:(.+)/);if(s[0]==="connect"){return s[1]}}}}function s(a){if(!v.bodyContains(a)){return}var e=v.getAttributeValue(a,"ws-connect");if(e==null||e===""){var t=r(a);if(t==null){return}else{e=t}}if(e.indexOf("/")===0){var n=location.hostname+(location.port?":"+location.port:"");if(location.protocol==="https:"){e="wss://"+n+e}else if(location.protocol==="http:"){e="ws://"+n+e}}var o=c(a,function(){return htmx.createWebSocket(e)});o.addEventListener("message",function(e){if(m(a)){return}var t=e.data;if(!v.triggerEvent(a,"htmx:wsBeforeMessage",{message:t,socketWrapper:o.publicInterface})){return}v.withExtensions(a,function(e){t=e.transformResponse(t,null,a)});var n=v.makeSettleInfo(a);var r=v.makeFragment(t);if(r.children.length){var s=Array.from(r.children);for(var i=0;i<s.length;i++){v.oobSwap(v.getAttributeValue(s[i],"hx-swap-oob")||"true",s[i],n)}}v.settleImmediately(n.tasks);v.triggerEvent(a,"htmx:wsAfterMessage",{message:t,socketWrapper:o.publicInterface})});v.getInternalData(a).webSocket=o}function c(r,t){var s={socket:null,messageQueue:[],retryCount:0,events:{},addEventListener:function(e,t){if(this.socket){this.socket.addEventListener(e,t)}if(!this.events[e]){this.events[e]=[]}this.events[e].push(t)},sendImmediately:function(e,t){if(!this.socket){v.triggerErrorEvent()}if(!t||v.triggerEvent(t,"htmx:wsBeforeSend",{message:e,socketWrapper:this.publicInterface})){this.socket.send(e);t&&v.triggerEvent(t,"htmx:wsAfterSend",{message:e,socketWrapper:this.publicInterface})}},send:function(e,t){if(this.socket.readyState!==this.socket.OPEN){this.messageQueue.push({message:e,sendElt:t})}else{this.sendImmediately(e,t)}},handleQueuedMessages:function(){while(this.messageQueue.length>0){var e=this.messageQueue[0];if(this.socket.readyState===this.socket.OPEN){this.sendImmediately(e.message,e.sendElt);this.messageQueue.shift()}else{break}}},init:function(){if(this.socket&&this.socket.readyState===this.socket.OPEN){this.socket.close()}var n=t();v.triggerEvent(r,"htmx:wsConnecting",{event:{type:"connecting"}});this.socket=n;n.onopen=function(e){s.retryCount=0;v.triggerEvent(r,"htmx:wsOpen",{event:e,socketWrapper:s.publicInterface});s.handleQueuedMessages()};n.onclose=function(e){if(!m(r)&&[1006,1012,1013].indexOf(e.code)>=0){var t=f(s.retryCount);setTimeout(function(){s.retryCount+=1;s.init()},t)}v.triggerEvent(r,"htmx:wsClose",{event:e,socketWrapper:s.publicInterface})};n.onerror=function(e){v.triggerErrorEvent(r,"htmx:wsError",{error:e,socketWrapper:s});m(r)};var e=this.events;Object.keys(e).forEach(function(t){e[t].forEach(function(e){n.addEventListener(t,e)})})},close:function(){this.socket.close()}};s.init();s.publicInterface={send:s.send.bind(s),sendImmediately:s.sendImmediately.bind(s),reconnect:s.init.bind(s),queue:s.messageQueue};return s}function a(e){var t=v.getAttributeValue(e,"hx-ws");if(t&&t!=="send"){return}var n=v.getClosestMatch(e,o);u(n,e)}function o(e){return v.getInternalData(e).webSocket!=null}function u(g,d){var t=v.getInternalData(d);var e=v.getTriggerSpecs(d);e.forEach(function(e){v.addTriggerHandler(d,e,t,function(e,t){if(m(g)){return}var n=v.getInternalData(g).webSocket;var r=v.getHeaders(d,v.getTarget(d));var s=v.getInputValues(d,"post");var i=s.errors;var a=Object.assign({},s.values);var o=v.getExpressionVars(d);var c=v.mergeObjects(a,o);var u=v.filterValues(c,d);var f={parameters:u,unfilteredParameters:c,headers:r,errors:i,triggeringEvent:t,messageBody:undefined,socketWrapper:n.publicInterface};if(!v.triggerEvent(e,"htmx:wsConfigSend",f)){return}if(i&&i.length>0){v.triggerEvent(e,"htmx:validation:halted",i);return}var l=f.messageBody;if(l===undefined){var h=Object.assign({},f.parameters);if(f.headers){h.HEADERS=r}l=JSON.stringify(h)}n.send(l,e);if(t&&v.shouldCancel(t,e)){t.preventDefault()}})})}function f(e){var t=htmx.config.wsReconnectDelay;if(typeof t==="function"){return t(e)}if(t==="full-jitter"){var n=Math.min(e,6);var r=1e3*Math.pow(2,n);return r*Math.random()}logError('htmx.config.wsReconnectDelay must either be a function or the string "full-jitter"')}function m(e){if(!v.bodyContains(e)){var t=v.getInternalData(e);if(t.webSocket){t.webSocket.close();return true}return false}return false}function t(e){var t=new WebSocket(e,[]);t.binaryType=htmx.config.wsBinaryType;return t}function l(e,t){var n=[];if(v.hasAttribute(e,t)||v.hasAttribute(e,"hx-ws")){n.push(e)}e.querySelectorAll("["+t+"], [data-"+t+"], [data-hx-ws], [hx-ws]").forEach(function(e){n.push(e)});return n}function h(e,t){if(e){for(var n=0;n<e.length;n++){t(e[n])}}}})();
```

## Matches in pg-press/internal/web/assets/js/htmx-v2.0.6.min.js

### L1
```
var htmx=function(){"use strict";const Q={onLoad:null,process:null,on:null,off:null,trigger:null,ajax:null,find:null,findAll:null,closest:null,values:function(e,t){const n=dn(e,t||"post");return n.values},remove:null,addClass:null,removeClass:null,toggleClass:null,takeClass:null,swap:null,defineExtension:null,removeExtension:null,logAll:null,logNone:null,logger:null,config:{historyEnabled:true,historyCacheSize:10,refreshOnHistoryMiss:false,defaultSwapStyle:"innerHTML",defaultSwapDelay:0,defaultSettleDelay:20,includeIndicatorStyles:true,indicatorClass:"htmx-indicator",requestClass:"htmx-request",addedClass:"htmx-added",settlingClass:"htmx-settling",swappingClass:"htmx-swapping",allowEval:true,allowScriptTags:true,inlineScriptNonce:"",inlineStyleNonce:"",attributesToSettle:["class","style","width","height"],withCredentials:false,timeout:0,wsReconnectDelay:"full-jitter",wsBinaryType:"blob",disableSelector:"[hx-disable], [data-hx-disable]",scrollBehavior:"instant",defaultFocusScroll:false,getCacheBusterParam:false,globalViewTransitions:false,methodsThatUseUrlParams:["get","delete"],selfRequestsOnly:true,ignoreTitle:false,scrollIntoViewOnBoost:true,triggerSpecsCache:null,disableInheritance:false,responseHandling:[{code:"204",swap:false},{code:"[23]..",swap:true},{code:"[45]..",swap:false,error:true}],allowNestedOobSwaps:true,historyRestoreAsHxRequest:true},parseInterval:null,location:location,_:null,version:"2.0.6"};Q.onLoad=j;Q.process=Ft;Q.on=xe;Q.off=be;Q.trigger=ae;Q.ajax=Ln;Q.find=f;Q.findAll=x;Q.closest=g;Q.remove=z;Q.addClass=K;Q.removeClass=G;Q.toggleClass=W;Q.takeClass=Z;Q.swap=$e;Q.defineExtension=zn;Q.removeExtension=$n;Q.logAll=V;Q.logNone=_;Q.parseInterval=d;Q._=e;const n={addTriggerHandler:St,bodyContains:se,canAccessLocalStorage:B,findThisElement:Se,filterValues:yn,swap:$e,hasAttribute:s,getAttributeValue:a,getClosestAttributeValue:ne,getClosestMatch:q,getExpressionVars:Tn,getHeaders:mn,getInputValues:dn,getInternalData:oe,getSwapSpecification:bn,getTriggerSpecs:st,getTarget:Ee,makeFragment:P,mergeObjects:le,makeSettleInfo:Sn,oobSwap:He,querySelectorExt:ue,settleImmediately:Yt,shouldCancel:ht,triggerEvent:ae,triggerErrorEvent:fe,withExtensions:jt};const de=["get","post","put","delete","patch"];const T=de.map(function(e){return"[hx-"+e+"], [data-hx-"+e+"]"}).join(", ");function d(e){if(e==undefined){return undefined}let t=NaN;if(e.slice(-2)=="ms"){t=parseFloat(e.slice(0,-2))}else if(e.slice(-1)=="s"){t=parseFloat(e.slice(0,-1))*1e3}else if(e.slice(-1)=="m"){t=parseFloat(e.slice(0,-1))*1e3*60}else{t=parseFloat(e)}return isNaN(t)?undefined:t}function ee(e,t){return e instanceof Element&&e.getAttribute(t)}function s(e,t){return!!e.hasAttribute&&(e.hasAttribute(t)||e.hasAttribute("data-"+t))}function a(e,t){return ee(e,t)||ee(e,"data-"+t)}function u(e){const t=e.parentElement;if(!t&&e.parentNode instanceof ShadowRoot)return e.parentNode;return t}function te(){return document}function y(e,t){return e.getRootNode?e.getRootNode({composed:t}):te()}function q(e,t){while(e&&!t(e)){e=u(e)}return e||null}function o(e,t,n){const r=a(t,n);const o=a(t,"hx-disinherit");var i=a(t,"hx-inherit");if(e!==t){if(Q.config.disableInheritance){if(i&&(i==="*"||i.split(" ").indexOf(n)>=0)){return r}else{return null}}if(o&&(o==="*"||o.split(" ").indexOf(n)>=0)){return"unset"}}return r}function ne(t,n){let r=null;q(t,function(e){return!!(r=o(t,ce(e),n))});if(r!=="unset"){return r}}function h(e,t){return e instanceof Element&&e.matches(t)}function A(e){const t=/<([a-z][^\/\0>\x20\t\r\n\f]*)/i;const n=t.exec(e);if(n){return n[1].toLowerCase()}else{return""}}function L(e){const t=new DOMParser;return t.parseFromString(e,"text/html")}function N(e,t){while(t.childNodes.length>0){e.append(t.childNodes[0])}}function r(e){const t=te().createElement("script");ie(e.attributes,function(e){t.setAttribute(e.name,e.value)});t.textContent=e.textContent;t.async=false;if(Q.config.inlineScriptNonce){t.nonce=Q.config.inlineScriptNonce}return t}function i(e){return e.matches("script")&&(e.type==="text/javascript"||e.type==="module"||e.type==="")}function I(e){Array.from(e.querySelectorAll("script")).forEach(e=>{if(i(e)){const t=r(e);const n=e.parentNode;try{n.insertBefore(t,e)}catch(e){R(e)}finally{e.remove()}}})}function P(e){const t=e.replace(/<head(\s[^>]*)?>[\s\S]*?<\/head>/i,"");const n=A(t);let r;if(n==="html"){r=new DocumentFragment;const i=L(e);N(r,i.body);r.title=i.title}else if(n==="body"){r=new DocumentFragment;const i=L(t);N(r,i.body);r.title=i.title}else{const i=L('<body><template class="internal-htmx-wrapper">'+t+"</template></body>");r=i.querySelector("template").content;r.title=i.title;var o=r.querySelector("title");if(o&&o.parentNode===r){o.remove();r.title=o.innerText}}if(r){if(Q.config.allowScriptTags){I(r)}else{r.querySelectorAll("script").forEach(e=>e.remove())}}return r}function re(e){if(e){e()}}function t(e,t){return Object.prototype.toString.call(e)==="[object "+t+"]"}function D(e){return typeof e==="function"}function k(e){return t(e,"Object")}function oe(e){const t="htmx-internal-data";let n=e[t];if(!n){n=e[t]={}}return n}function M(t){const n=[];if(t){for(let e=0;e<t.length;e++){n.push(t[e])}}return n}function ie(t,n){if(t){for(let e=0;e<t.length;e++){n(t[e])}}}function F(e){const t=e.getBoundingClientRect();const n=t.top;const r=t.bottom;return n<window.innerHeight&&r>=0}function se(e){return e.getRootNode({composed:true})===document}function X(e){return e.trim().split(/\s+/)}function le(e,t){for(const n in t){if(t.hasOwnProperty(n)){e[n]=t[n]}}return e}function v(e){try{return JSON.parse(e)}catch(e){R(e);return null}}function B(){const e="htmx:sessionStorageTest";try{sessionStorage.setItem(e,e);sessionStorage.removeItem(e);return true}catch(e){return false}}function U(e){const t=new URL(e,"http://x");if(t){e=t.pathname+t.search}if(e!="/"){e=e.replace(/\/+$/,"")}return e}function e(e){return On(te().body,function(){return eval(e)})}function j(t){const e=Q.on("htmx:load",function(e){t(e.detail.elt)});return e}function V(){Q.logger=function(e,t,n){if(console){console.log(t,e,n)}}}function _(){Q.logger=null}function f(e,t){if(typeof e!=="string"){return e.querySelector(t)}else{return f(te(),e)}}function x(e,t){if(typeof e!=="string"){return e.querySelectorAll(t)}else{return x(te(),e)}}function b(){return window}function z(e,t){e=w(e);if(t){b().setTimeout(function(){z(e);e=null},t)}else{u(e).removeChild(e)}}function ce(e){return e instanceof Element?e:null}function $(e){return e instanceof HTMLElement?e:null}function J(e){return typeof e==="string"?e:null}function p(e){return e instanceof Element||e instanceof Document||e instanceof DocumentFragment?e:null}function K(e,t,n){e=ce(w(e));if(!e){return}if(n){b().setTimeout(function(){K(e,t);e=null},n)}else{e.classList&&e.classList.add(t)}}function G(e,t,n){let r=ce(w(e));if(!r){return}if(n){b().setTimeout(function(){G(r,t);r=null},n)}else{if(r.classList){r.classList.remove(t);if(r.classList.length===0){r.removeAttribute("class")}}}}function W(e,t){e=w(e);e.classList.toggle(t)}function Z(e,t){e=w(e);ie(e.parentElement.children,function(e){G(e,t)});K(ce(e),t)}function g(e,t){e=ce(w(e));if(e){return e.closest(t)}return null}function l(e,t){return e.substring(0,t.length)===t}function Y(e,t){return e.substring(e.length-t.length)===t}function pe(e){const t=e.trim();if(l(t,"<")&&Y(t,"/>")){return t.substring(1,t.length-2)}else{return t}}function m(t,r,n){if(r.indexOf("global ")===0){return m(t,r.slice(7),true)}t=w(t);const o=[];{let t=0;let n=0;for(let e=0;e<r.length;e++){const l=r[e];if(l===","&&t===0){o.push(r.substring(n,e));n=e+1;continue}if(l==="<"){t++}else if(l==="/"&&e<r.length-1&&r[e+1]===">"){t--}}if(n<r.length){o.push(r.substring(n))}}const i=[];const s=[];while(o.length>0){const r=pe(o.shift());let e;if(r.indexOf("closest ")===0){e=g(ce(t),pe(r.slice(8)))}else if(r.indexOf("find ")===0){e=f(p(t),pe(r.slice(5)))}else if(r==="next"||r==="nextElementSibling"){e=ce(t).nextElementSibling}else if(r.indexOf("next ")===0){e=ge(t,pe(r.slice(5)),!!n)}else if(r==="previous"||r==="previousElementSibling"){e=ce(t).previousElementSibling}else if(r.indexOf("previous ")===0){e=me(t,pe(r.slice(9)),!!n)}else if(r==="document"){e=document}else if(r==="window"){e=window}else if(r==="body"){e=document.body}else if(r==="root"){e=y(t,!!n)}else if(r==="host"){e=t.getRootNode().host}else{s.push(r)}if(e){i.push(e)}}if(s.length>0){const e=s.join(",");const c=p(y(t,!!n));i.push(...M(c.querySelectorAll(e)))}return i}var ge=function(t,e,n){const r=p(y(t,n)).querySelectorAll(e);for(let e=0;e<r.length;e++){const o=r[e];if(o.compareDocumentPosition(t)===Node.DOCUMENT_POSITION_PRECEDING){return o}}};var me=function(t,e,n){const r=p(y(t,n)).querySelectorAll(e);for(let e=r.length-1;e>=0;e--){const o=r[e];if(o.compareDocumentPosition(t)===Node.DOCUMENT_POSITION_FOLLOWING){return o}}};function ue(e,t){if(typeof e!=="string"){return m(e,t)[0]}else{return m(te().body,e)[0]}}function w(e,t){if(typeof e==="string"){return f(p(t)||document,e)}else{return e}}function ye(e,t,n,r){if(D(t)){return{target:te().body,event:J(e),listener:t,options:n}}else{return{target:w(e),event:J(t),listener:n,options:r}}}function xe(t,n,r,o){Gn(function(){const e=ye(t,n,r,o);e.target.addEventListener(e.event,e.listener,e.options)});const e=D(n);return e?n:r}function be(t,n,r){Gn(function(){const e=ye(t,n,r);e.target.removeEventListener(e.event,e.listener)});return D(n)?n:r}const ve=te().createElement("output");function we(t,n){const e=ne(t,n);if(e){if(e==="this"){return[Se(t,n)]}else{const r=m(t,e);const o=/(^|,)(\s*)inherit(\s*)($|,)/.test(e);if(o){const i=ce(q(t,function(e){return e!==t&&s(ce(e),n)}));if(i){r.push(...we(i,n))}}if(r.length===0){R('The selector "'+e+'" on '+n+" returned no matches!");return[ve]}else{return r}}}}function Se(e,t){return ce(q(e,function(e){return a(ce(e),t)!=null}))}function Ee(e){const t=ne(e,"hx-target");if(t){if(t==="this"){return Se(e,"hx-target")}else{return ue(e,t)}}else{const n=oe(e);if(n.boosted){return te().body}else{return e}}}function Ce(e){return Q.config.attributesToSettle.includes(e)}function Oe(t,n){ie(t.attributes,function(e){if(!n.hasAttribute(e.name)&&Ce(e.name)){t.removeAttribute(e.name)}});ie(n.attributes,function(e){if(Ce(e.name)){t.setAttribute(e.name,e.value)}})}function Re(t,e){const n=Jn(e);for(let e=0;e<n.length;e++){const r=n[e];try{if(r.isInlineSwap(t)){return true}}catch(e){R(e)}}return t==="outerHTML"}function He(e,o,i,t){t=t||te();let n="#"+CSS.escape(ee(o,"id"));let s="outerHTML";if(e==="true"){}else if(e.indexOf(":")>0){s=e.substring(0,e.indexOf(":"));n=e.substring(e.indexOf(":")+1)}else{s=e}o.removeAttribute("hx-swap-oob");o.removeAttribute("data-hx-swap-oob");const r=m(t,n,false);if(r.length){ie(r,function(e){let t;const n=o.cloneNode(true);t=te().createDocumentFragment();t.appendChild(n);if(!Re(s,e)){t=p(n)}const r={shouldSwap:true,target:e,fragment:t};if(!ae(e,"htmx:oobBeforeSwap",r))return;e=r.target;if(r.shouldSwap){qe(t);_e(s,e,e,t,i);Te()}ie(i.elts,function(e){ae(e,"htmx:oobAfterSwap",r)})});o.parentNode.removeChild(o)}else{o.parentNode.removeChild(o);fe(te().body,"htmx:oobErrorNoTarget",{content:o})}return e}function Te(){const e=f("#--htmx-preserve-pantry--");if(e){for(const t of[...e.children]){const n=f("#"+t.id);n.parentNode.moveBefore(t,n);n.remove()}e.remove()}}function qe(e){ie(x(e,"[hx-preserve], [data-hx-preserve]"),function(e){const t=a(e,"id");const n=te().getElementById(t);if(n!=null){if(e.moveBefore){let e=f("#--htmx-preserve-pantry--");if(e==null){te().body.insertAdjacentHTML("afterend","<div id='--htmx-preserve-pantry--'></div>");e=f("#--htmx-preserve-pantry--")}e.moveBefore(n,null)}else{e.parentNode.replaceChild(n,e)}}})}function Ae(l,e,c){ie(e.querySelectorAll("[id]"),function(t){const n=ee(t,"id");if(n&&n.length>0){const r=n.replace("'","\\'");const o=t.tagName.replace(":","\\:");const e=p(l);const i=e&&e.querySelector(o+"[id='"+r+"']");if(i&&i!==e){const s=t.cloneNode();Oe(t,i);c.tasks.push(function(){Oe(t,s)})}}})}function Le(e){return function(){G(e,Q.config.addedClass);Ft(ce(e));Ne(p(e));ae(e,"htmx:load")}}function Ne(e){const t="[autofocus]";const n=$(h(e,t)?e:e.querySelector(t));if(n!=null){n.focus()}}function c(e,t,n,r){Ae(e,n,r);while(n.childNodes.length>0){const o=n.firstChild;K(ce(o),Q.config.addedClass);e.insertBefore(o,t);if(o.nodeType!==Node.TEXT_NODE&&o.nodeType!==Node.COMMENT_NODE){r.tasks.push(Le(o))}}}function Ie(e,t){let n=0;while(n<e.length){t=(t<<5)-t+e.charCodeAt(n++)|0}return t}function Pe(t){let n=0;for(let e=0;e<t.attributes.length;e++){const r=t.attributes[e];if(r.value){n=Ie(r.name,n);n=Ie(r.value,n)}}return n}function De(t){const n=oe(t);if(n.onHandlers){for(let e=0;e<n.onHandlers.length;e++){const r=n.onHandlers[e];be(t,r.event,r.listener)}delete n.onHandlers}}function ke(e){const t=oe(e);if(t.timeout){clearTimeout(t.timeout)}if(t.listenerInfos){ie(t.listenerInfos,function(e){if(e.on){be(e.on,e.trigger,e.listener)}})}De(e);ie(Object.keys(t),function(e){if(e!=="firstInitCompleted")delete t[e]})}function S(e){ae(e,"htmx:beforeCleanupElement");ke(e);ie(e.children,function(e){S(e)})}function Me(t,e,n){if(t.tagName==="BODY"){return Ve(t,e,n)}let r;const o=t.previousSibling;const i=u(t);if(!i){return}c(i,t,e,n);if(o==null){r=i.firstChild}else{r=o.nextSibling}n.elts=n.elts.filter(function(e){return e!==t});while(r&&r!==t){if(r instanceof Element){n.elts.push(r)}r=r.nextSibling}S(t);t.remove()}function Fe(e,t,n){return c(e,e.firstChild,t,n)}function Xe(e,t,n){return c(u(e),e,t,n)}function Be(e,t,n){return c(e,null,t,n)}function Ue(e,t,n){return c(u(e),e.nextSibling,t,n)}function je(e){S(e);const t=u(e);if(t){return t.removeChild(e)}}function Ve(e,t,n){const r=e.firstChild;c(e,r,t,n);if(r){while(r.nextSibling){S(r.nextSibling);e.removeChild(r.nextSibling)}S(r);e.removeChild(r)}}function _e(t,e,n,r,o){switch(t){case"none":return;case"outerHTML":Me(n,r,o);return;case"afterbegin":Fe(n,r,o);return;case"beforebegin":Xe(n,r,o);return;case"beforeend":Be(n,r,o);return;case"afterend":Ue(n,r,o);return;case"delete":je(n);return;default:var i=Jn(e);for(let e=0;e<i.length;e++){const s=i[e];try{const l=s.handleSwap(t,n,r,o);if(l){if(Array.isArray(l)){for(let e=0;e<l.length;e++){const c=l[e];if(c.nodeType!==Node.TEXT_NODE&&c.nodeType!==Node.COMMENT_NODE){o.tasks.push(Le(c))}}}return}}catch(e){R(e)}}if(t==="innerHTML"){Ve(n,r,o)}else{_e(Q.config.defaultSwapStyle,e,n,r,o)}}}function ze(e,n,r){var t=x(e,"[hx-swap-oob], [data-hx-swap-oob]");ie(t,function(e){if(Q.config.allowNestedOobSwaps||e.parentElement===null){const t=a(e,"hx-swap-oob");if(t!=null){He(t,e,n,r)}}else{e.removeAttribute("hx-swap-oob");e.removeAttribute("data-hx-swap-oob")}});return t.length>0}function $e(h,d,p,g){if(!g){g={}}let m=null;let n=null;let e=function(){re(g.beforeSwapCallback);h=w(h);const r=g.contextElement?y(g.contextElement,false):te();const e=document.activeElement;let t={};t={elt:e,start:e?e.selectionStart:null,end:e?e.selectionEnd:null};const o=Sn(h);if(p.swapStyle==="textContent"){h.textContent=d}else{let n=P(d);o.title=g.title||n.title;if(g.historyRequest){n=n.querySelector("[hx-history-elt],[data-hx-history-elt]")||n}if(g.selectOOB){const i=g.selectOOB.split(",");for(let t=0;t<i.length;t++){const s=i[t].split(":",2);let e=s[0].trim();if(e.indexOf("#")===0){e=e.substring(1)}const l=s[1]||"true";const c=n.querySelector("#"+e);if(c){He(l,c,o,r)}}}ze(n,o,r);ie(x(n,"template"),function(e){if(e.content&&ze(e.content,o,r)){e.remove()}});if(g.select){const u=te().createDocumentFragment();ie(n.querySelectorAll(g.select),function(e){u.appendChild(e)});n=u}qe(n);_e(p.swapStyle,g.contextElement,h,n,o);Te()}if(t.elt&&!se(t.elt)&&ee(t.elt,"id")){const f=document.getElementById(ee(t.elt,"id"));const a={preventScroll:p.focusScroll!==undefined?!p.focusScroll:!Q.config.defaultFocusScroll};if(f){if(t.start&&f.setSelectionRange){try{f.setSelectionRange(t.start,t.end)}catch(e){}}f.focus(a)}}h.classList.remove(Q.config.swappingClass);ie(o.elts,function(e){if(e.classList){e.classList.add(Q.config.settlingClass)}ae(e,"htmx:afterSwap",g.eventInfo)});re(g.afterSwapCallback);if(!p.ignoreTitle){Bn(o.title)}const n=function(){ie(o.tasks,function(e){e.call()});ie(o.elts,function(e){if(e.classList){e.classList.remove(Q.config.settlingClass)}ae(e,"htmx:afterSettle",g.eventInfo)});if(g.anchor){const e=ce(w("#"+g.anchor));if(e){e.scrollIntoView({block:"start",behavior:"auto"})}}En(o.elts,p);re(g.afterSettleCallback);re(m)};if(p.settleDelay>0){b().setTimeout(n,p.settleDelay)}else{n()}};let t=Q.config.globalViewTransitions;if(p.hasOwnProperty("transition")){t=p.transition}const r=g.contextElement||te();if(t&&ae(r,"htmx:beforeTransition",g.eventInfo)&&typeof Promise!=="undefined"&&document.startViewTransition){const o=new Promise(function(e,t){m=e;n=t});const i=e;e=function(){document.startViewTransition(function(){i();return o})}}try{if(p?.swapDelay&&p.swapDelay>0){b().setTimeout(e,p.swapDelay)}else{e()}}catch(e){fe(r,"htmx:swapError",g.eventInfo);re(n);throw e}}function Je(e,t,n){const r=e.getResponseHeader(t);if(r.indexOf("{")===0){const o=v(r);for(const i in o){if(o.hasOwnProperty(i)){let e=o[i];if(k(e)){n=e.target!==undefined?e.target:n}else{e={value:e}}ae(n,i,e)}}}else{const s=r.split(",");for(let e=0;e<s.length;e++){ae(n,s[e].trim(),[])}}}const Ke=/\s/;const E=/[\s,]/;const Ge=/[_$a-zA-Z]/;const We=/[_$a-zA-Z0-9]/;const Ze=['"',"'","/"];const C=/[^\s]/;const Ye=/[{(]/;const Qe=/[})]/;function et(e){const t=[];let n=0;while(n<e.length){if(Ge.exec(e.charAt(n))){var r=n;while(We.exec(e.charAt(n+1))){n++}t.push(e.substring(r,n+1))}else if(Ze.indexOf(e.charAt(n))!==-1){const o=e.charAt(n);var r=n;n++;while(n<e.length&&e.charAt(n)!==o){if(e.charAt(n)==="\\"){n++}n++}t.push(e.substring(r,n+1))}else{const i=e.charAt(n);t.push(i)}n++}return t}function tt(e,t,n){return Ge.exec(e.charAt(0))&&e!=="true"&&e!=="false"&&e!=="this"&&e!==n&&t!=="."}function nt(r,o,i){if(o[0]==="["){o.shift();let e=1;let t=" return (function("+i+"){ return (";let n=null;while(o.length>0){const s=o[0];if(s==="]"){e--;if(e===0){if(n===null){t=t+"true"}o.shift();t+=")})";try{const l=On(r,function(){return Function(t)()},function(){return true});l.source=t;return l}catch(e){fe(te().body,"htmx:syntax:error",{error:e,source:t});return null}}}else if(s==="["){e++}if(tt(s,n,i)){t+="(("+i+"."+s+") ? ("+i+"."+s+") : (window."+s+"))"}else{t=t+s}n=o.shift()}}}function O(e,t){let n="";while(e.length>0&&!t.test(e[0])){n+=e.shift()}return n}function rt(e){let t;if(e.length>0&&Ye.test(e[0])){e.shift();t=O(e,Qe).trim();e.shift()}else{t=O(e,E)}return t}const ot="input, textarea, select";function it(e,t,n){const r=[];const o=et(t);do{O(o,C);const l=o.length;const c=O(o,/[,\[\s]/);if(c!==""){if(c==="every"){const u={trigger:"every"};O(o,C);u.pollInterval=d(O(o,/[,\[\s]/));O(o,C);var i=nt(e,o,"event");if(i){u.eventFilter=i}r.push(u)}else{const f={trigger:c};var i=nt(e,o,"event");if(i){f.eventFilter=i}O(o,C);while(o.length>0&&o[0]!==","){const a=o.shift();if(a==="changed"){f.changed=true}else if(a==="once"){f.once=true}else if(a==="consume"){f.consume=true}else if(a==="delay"&&o[0]===":"){o.shift();f.delay=d(O(o,E))}else if(a==="from"&&o[0]===":"){o.shift();if(Ye.test(o[0])){var s=rt(o)}else{var s=O(o,E);if(s==="closest"||s==="find"||s==="next"||s==="previous"){o.shift();const h=rt(o);if(h.length>0){s+=" "+h}}}f.from=s}else if(a==="target"&&o[0]===":"){o.shift();f.target=rt(o)}else if(a==="throttle"&&o[0]===":"){o.shift();f.throttle=d(O(o,E))}else if(a==="queue"&&o[0]===":"){o.shift();f.queue=O(o,E)}else if(a==="root"&&o[0]===":"){o.shift();f[a]=rt(o)}else if(a==="threshold"&&o[0]===":"){o.shift();f[a]=O(o,E)}else{fe(e,"htmx:syntax:error",{token:o.shift()})}O(o,C)}r.push(f)}}if(o.length===l){fe(e,"htmx:syntax:error",{token:o.shift()})}O(o,C)}while(o[0]===","&&o.shift());if(n){n[t]=r}return r}function st(e){const t=a(e,"hx-trigger");let n=[];if(t){const r=Q.config.triggerSpecsCache;n=r&&r[t]||it(e,t,r)}if(n.length>0){return n}else if(h(e,"form")){return[{trigger:"submit"}]}else if(h(e,'input[type="button"], input[type="submit"]')){return[{trigger:"click"}]}else if(h(e,ot)){return[{trigger:"change"}]}else{return[{trigger:"click"}]}}function lt(e){oe(e).cancelled=true}function ct(e,t,n){const r=oe(e);r.timeout=b().setTimeout(function(){if(se(e)&&r.cancelled!==true){if(!pt(n,e,Bt("hx:poll:trigger",{triggerSpec:n,target:e}))){t(e)}ct(e,t,n)}},n.pollInterval)}function ut(e){return location.hostname===e.hostname&&ee(e,"href")&&ee(e,"href").indexOf("#")!==0}function ft(e){return g(e,Q.config.disableSelector)}function at(t,n,e){if(t instanceof HTMLAnchorElement&&ut(t)&&(t.target===""||t.target==="_self")||t.tagName==="FORM"&&String(ee(t,"method")).toLowerCase()!=="dialog"){n.boosted=true;let r,o;if(t.tagName==="A"){r="get";o=ee(t,"href")}else{const i=ee(t,"method");r=i?i.toLowerCase():"get";o=ee(t,"action");if(o==null||o===""){o=location.href}if(r==="get"&&o.includes("?")){o=o.replace(/\?[^#]+/,"")}}e.forEach(function(e){gt(t,function(e,t){const n=ce(e);if(ft(n)){S(n);return}he(r,o,n,t)},n,e,true)})}}function ht(e,t){if(e.type==="submit"||e.type==="click"){t=ce(e.target)||t;if(t.tagName==="FORM"){return true}if(t.form&&t.type==="submit"){return true}t=t.closest("a");if(t&&t.href&&(t.getAttribute("href")==="#"||t.getAttribute("href").indexOf("#")!==0)){return true}}return false}function dt(e,t){return oe(e).boosted&&e instanceof HTMLAnchorElement&&t.type==="click"&&(t.ctrlKey||t.metaKey)}function pt(e,t,n){const r=e.eventFilter;if(r){try{return r.call(t,n)!==true}catch(e){const o=r.source;fe(te().body,"htmx:eventFilter:error",{error:e,source:o});return true}}return false}function gt(l,c,e,u,f){const a=oe(l);let t;if(u.from){t=m(l,u.from)}else{t=[l]}if(u.changed){if(!("lastValue"in a)){a.lastValue=new WeakMap}t.forEach(function(e){if(!a.lastValue.has(u)){a.lastValue.set(u,new WeakMap)}a.lastValue.get(u).set(e,e.value)})}ie(t,function(i){const s=function(e){if(!se(l)){i.removeEventListener(u.trigger,s);return}if(dt(l,e)){return}if(f||ht(e,l)){e.preventDefault()}if(pt(u,l,e)){return}const t=oe(e);t.triggerSpec=u;if(t.handledFor==null){t.handledFor=[]}if(t.handledFor.indexOf(l)<0){t.handledFor.push(l);if(u.consume){e.stopPropagation()}if(u.target&&e.target){if(!h(ce(e.target),u.target)){return}}if(u.once){if(a.triggeredOnce){return}else{a.triggeredOnce=true}}if(u.changed){const n=e.target;const r=n.value;const o=a.lastValue.get(u);if(o.has(n)&&o.get(n)===r){return}o.set(n,r)}if(a.delayed){clearTimeout(a.delayed)}if(a.throttle){return}if(u.throttle>0){if(!a.throttle){ae(l,"htmx:trigger");c(l,e);a.throttle=b().setTimeout(function(){a.throttle=null},u.throttle)}}else if(u.delay>0){a.delayed=b().setTimeout(function(){ae(l,"htmx:trigger");c(l,e)},u.delay)}else{ae(l,"htmx:trigger");c(l,e)}}};if(e.listenerInfos==null){e.listenerInfos=[]}e.listenerInfos.push({trigger:u.trigger,listener:s,on:i});i.addEventListener(u.trigger,s)})}let mt=false;let yt=null;function xt(){if(!yt){yt=function(){mt=true};window.addEventListener("scroll",yt);window.addEventListener("resize",yt);setInterval(function(){if(mt){mt=false;ie(te().querySelectorAll("[hx-trigger*='revealed'],[data-hx-trigger*='revealed']"),function(e){bt(e)})}},200)}}function bt(e){if(!s(e,"data-hx-revealed")&&F(e)){e.setAttribute("data-hx-revealed","true");const t=oe(e);if(t.initHash){ae(e,"revealed")}else{e.addEventListener("htmx:afterProcessNode",function(){ae(e,"revealed")},{once:true})}}}function vt(e,t,n,r){const o=function(){if(!n.loaded){n.loaded=true;ae(e,"htmx:trigger");t(e)}};if(r>0){b().setTimeout(o,r)}else{o()}}function wt(t,n,e){let i=false;ie(de,function(r){if(s(t,"hx-"+r)){const o=a(t,"hx-"+r);i=true;n.path=o;n.verb=r;e.forEach(function(e){St(t,e,n,function(e,t){const n=ce(e);if(ft(n)){S(n);return}he(r,o,n,t)})})}});return i}function St(r,e,t,n){if(e.trigger==="revealed"){xt();gt(r,n,t,e);bt(ce(r))}else if(e.trigger==="intersect"){const o={};if(e.root){o.root=ue(r,e.root)}if(e.threshold){o.threshold=parseFloat(e.threshold)}const i=new IntersectionObserver(function(t){for(let e=0;e<t.length;e++){const n=t[e];if(n.isIntersecting){ae(r,"intersect");break}}},o);i.observe(ce(r));gt(ce(r),n,t,e)}else if(!t.firstInitCompleted&&e.trigger==="load"){if(!pt(e,r,Bt("load",{elt:r}))){vt(ce(r),n,t,e.delay)}}else if(e.pollInterval>0){t.polling=true;ct(ce(r),n,e)}else{gt(r,n,t,e)}}function Et(e){const t=ce(e);if(!t){return false}const n=t.attributes;for(let e=0;e<n.length;e++){const r=n[e].name;if(l(r,"hx-on:")||l(r,"data-hx-on:")||l(r,"hx-on-")||l(r,"data-hx-on-")){return true}}return false}const Ct=(new XPathEvaluator).createExpression('.//*[@*[ starts-with(name(), "hx-on:") or starts-with(name(), "data-hx-on:") or'+' starts-with(name(), "hx-on-") or starts-with(name(), "data-hx-on-") ]]');function Ot(e,t){if(Et(e)){t.push(ce(e))}const n=Ct.evaluate(e);let r=null;while(r=n.iterateNext())t.push(ce(r))}function Rt(e){const t=[];if(e instanceof DocumentFragment){for(const n of e.childNodes){Ot(n,t)}}else{Ot(e,t)}return t}function Ht(e){if(e.querySelectorAll){const n=", [hx-boost] a, [data-hx-boost] a, a[hx-boost], a[data-hx-boost]";const r=[];for(const i in Vn){const s=Vn[i];if(s.getSelectors){var t=s.getSelectors();if(t){r.push(t)}}}const o=e.querySelectorAll(T+n+", form, [type='submit'],"+" [hx-ext], [data-hx-ext], [hx-trigger], [data-hx-trigger]"+r.flat().map(e=>", "+e).join(""));return o}else{return[]}}function Tt(e){const t=At(e.target);const n=Nt(e);if(n){n.lastButtonClicked=t}}function qt(e){const t=Nt(e);if(t){t.lastButtonClicked=null}}function At(e){return g(ce(e),"button, input[type='submit']")}function Lt(e){return e.form||g(e,"form")}function Nt(e){const t=At(e.target);if(!t){return}const n=Lt(t);return oe(n)}function It(e){e.addEventListener("click",Tt);e.addEventListener("focusin",Tt);e.addEventListener("focusout",qt)}function Pt(t,e,n){const r=oe(t);if(!Array.isArray(r.onHandlers)){r.onHandlers=[]}let o;const i=function(e){On(t,function(){if(ft(t)){return}if(!o){o=new Function("event",n)}o.call(t,e)})};t.addEventListener(e,i);r.onHandlers.push({event:e,listener:i})}function Dt(t){De(t);for(let e=0;e<t.attributes.length;e++){const n=t.attributes[e].name;const r=t.attributes[e].value;if(l(n,"hx-on")||l(n,"data-hx-on")){const o=n.indexOf("-on")+3;const i=n.slice(o,o+1);if(i==="-"||i===":"){let e=n.slice(o+1);if(l(e,":")){e="htmx"+e}else if(l(e,"-")){e="htmx:"+e.slice(1)}else if(l(e,"htmx-")){e="htmx:"+e.slice(5)}Pt(t,e,r)}}}}function kt(t){ae(t,"htmx:beforeProcessNode");const n=oe(t);const e=st(t);const r=wt(t,n,e);if(!r){if(ne(t,"hx-boost")==="true"){at(t,n,e)}else if(s(t,"hx-trigger")){e.forEach(function(e){St(t,e,n,function(){})})}}if(t.tagName==="FORM"||ee(t,"type")==="submit"&&s(t,"form")){It(t)}n.firstInitCompleted=true;ae(t,"htmx:afterProcessNode")}function Mt(e){if(!(e instanceof Element)){return false}const t=oe(e);const n=Pe(e);if(t.initHash!==n){ke(e);t.initHash=n;return true}return false}function Ft(e){e=w(e);if(ft(e)){S(e);return}const t=[];if(Mt(e)){t.push(e)}ie(Ht(e),function(e){if(ft(e)){S(e);return}if(Mt(e)){t.push(e)}});ie(Rt(e),Dt);ie(t,kt)}function Xt(e){return e.replace(/([a-z0-9])([A-Z])/g,"$1-$2").toLowerCase()}function Bt(e,t){return new CustomEvent(e,{bubbles:true,cancelable:true,composed:true,detail:t})}function fe(e,t,n){ae(e,t,le({error:t},n))}function Ut(e){return e==="htmx:afterProcessNode"}function jt(e,t,n){ie(Jn(e,[],n),function(e){try{t(e)}catch(e){R(e)}})}function R(e){console.error(e)}function ae(e,t,n){e=w(e);if(n==null){n={}}n.elt=e;const r=Bt(t,n);if(Q.logger&&!Ut(t)){Q.logger(e,t,n)}if(n.error){R(n.error);ae(e,"htmx:error",{errorInfo:n})}let o=e.dispatchEvent(r);const i=Xt(t);if(o&&i!==t){const s=Bt(i,r.detail);o=o&&e.dispatchEvent(s)}jt(ce(e),function(e){o=o&&(e.onEvent(t,r)!==false&&!r.defaultPrevented)});return o}let Vt=location.pathname+location.search;function _t(e){Vt=e;if(B()){sessionStorage.setItem("htmx-current-path-for-history",e)}}function zt(){const e=te().querySelector("[hx-history-elt],[data-hx-history-elt]");return e||te().body}function $t(t,e){if(!B()){return}const n=Kt(e);const r=te().title;const o=window.scrollY;if(Q.config.historyCacheSize<=0){sessionStorage.removeItem("htmx-history-cache");return}t=U(t);const i=v(sessionStorage.getItem("htmx-history-cache"))||[];for(let e=0;e<i.length;e++){if(i[e].url===t){i.splice(e,1);break}}const s={url:t,content:n,title:r,scroll:o};ae(te().body,"htmx:historyItemCreated",{item:s,cache:i});i.push(s);while(i.length>Q.config.historyCacheSize){i.shift()}while(i.length>0){try{sessionStorage.setItem("htmx-history-cache",JSON.stringify(i));break}catch(e){fe(te().body,"htmx:historyCacheError",{cause:e,cache:i});i.shift()}}}function Jt(t){if(!B()){return null}t=U(t);const n=v(sessionStorage.getItem("htmx-history-cache"))||[];for(let e=0;e<n.length;e++){if(n[e].url===t){return n[e]}}return null}function Kt(e){const t=Q.config.requestClass;const n=e.cloneNode(true);ie(x(n,"."+t),function(e){G(e,t)});ie(x(n,"[data-disabled-by-htmx]"),function(e){e.removeAttribute("disabled")});return n.innerHTML}function Gt(){const e=zt();let t=Vt;if(B()){t=sessionStorage.getItem("htmx-current-path-for-history")}t=t||location.pathname+location.search;const n=te().querySelector('[hx-history="false" i],[data-hx-history="false" i]');if(!n){ae(te().body,"htmx:beforeHistorySave",{path:t,historyElt:e});$t(t,e)}if(Q.config.historyEnabled)history.replaceState({htmx:true},te().title,location.href)}function Wt(e){if(Q.config.getCacheBusterParam){e=e.replace(/org\.htmx\.cache-buster=[^&]*&?/,"");if(Y(e,"&")||Y(e,"?")){e=e.slice(0,-1)}}if(Q.config.historyEnabled){history.pushState({htmx:true},"",e)}_t(e)}function Zt(e){if(Q.config.historyEnabled)history.replaceState({htmx:true},"",e);_t(e)}function Yt(e){ie(e,function(e){e.call(undefined)})}function Qt(e){const t=new XMLHttpRequest;const n={swapStyle:"innerHTML",swapDelay:0,settleDelay:0};const r={path:e,xhr:t,historyElt:zt(),swapSpec:n};t.open("GET",e,true);if(Q.config.historyRestoreAsHxRequest){t.setRequestHeader("HX-Request","true")}t.setRequestHeader("HX-History-Restore-Request","true");t.setRequestHeader("HX-Current-URL",location.href);t.onload=function(){if(this.status>=200&&this.status<400){r.response=this.response;ae(te().body,"htmx:historyCacheMissLoad",r);$e(r.historyElt,r.response,n,{contextElement:r.historyElt,historyRequest:true});_t(r.path);ae(te().body,"htmx:historyRestore",{path:e,cacheMiss:true,serverResponse:r.response})}else{fe(te().body,"htmx:historyCacheMissLoadError",r)}};if(ae(te().body,"htmx:historyCacheMiss",r)){t.send()}}function en(e){Gt();e=e||location.pathname+location.search;const t=Jt(e);if(t){const n={swapStyle:"innerHTML",swapDelay:0,settleDelay:0,scroll:t.scroll};const r={path:e,item:t,historyElt:zt(),swapSpec:n};if(ae(te().body,"htmx:historyCacheHit",r)){$e(r.historyElt,t.content,n,{contextElement:r.historyElt,title:t.title});_t(r.path);ae(te().body,"htmx:historyRestore",r)}}else{if(Q.config.refreshOnHistoryMiss){Q.location.reload(true)}else{Qt(e)}}}function tn(e){let t=we(e,"hx-indicator");if(t==null){t=[e]}ie(t,function(e){const t=oe(e);t.requestCount=(t.requestCount||0)+1;e.classList.add.call(e.classList,Q.config.requestClass)});return t}function nn(e){let t=we(e,"hx-disabled-elt");if(t==null){t=[]}ie(t,function(e){const t=oe(e);t.requestCount=(t.requestCount||0)+1;e.setAttribute("disabled","");e.setAttribute("data-disabled-by-htmx","")});return t}function rn(e,t){ie(e.concat(t),function(e){const t=oe(e);t.requestCount=(t.requestCount||1)-1});ie(e,function(e){const t=oe(e);if(t.requestCount===0){e.classList.remove.call(e.classList,Q.config.requestClass)}});ie(t,function(e){const t=oe(e);if(t.requestCount===0){e.removeAttribute("disabled");e.removeAttribute("data-disabled-by-htmx")}})}function on(t,n){for(let e=0;e<t.length;e++){const r=t[e];if(r.isSameNode(n)){return true}}return false}function sn(e){const t=e;if(t.name===""||t.name==null||t.disabled||g(t,"fieldset[disabled]")){return false}if(t.type==="button"||t.type==="submit"||t.tagName==="image"||t.tagName==="reset"||t.tagName==="file"){return false}if(t.type==="checkbox"||t.type==="radio"){return t.checked}return true}function ln(t,e,n){if(t!=null&&e!=null){if(Array.isArray(e)){e.forEach(function(e){n.append(t,e)})}else{n.append(t,e)}}}function cn(t,n,r){if(t!=null&&n!=null){let e=r.getAll(t);if(Array.isArray(n)){e=e.filter(e=>n.indexOf(e)<0)}else{e=e.filter(e=>e!==n)}r.delete(t);ie(e,e=>r.append(t,e))}}function un(e){if(e instanceof HTMLSelectElement&&e.multiple){return M(e.querySelectorAll("option:checked")).map(function(e){return e.value})}if(e instanceof HTMLInputElement&&e.files){return M(e.files)}return e.value}function fn(t,n,r,e,o){if(e==null||on(t,e)){return}else{t.push(e)}if(sn(e)){const i=ee(e,"name");ln(i,un(e),n);if(o){an(e,r)}}if(e instanceof HTMLFormElement){ie(e.elements,function(e){if(t.indexOf(e)>=0){cn(e.name,un(e),n)}else{t.push(e)}if(o){an(e,r)}});new FormData(e).forEach(function(e,t){if(e instanceof File&&e.name===""){return}ln(t,e,n)})}}function an(e,t){const n=e;if(n.willValidate){ae(n,"htmx:validation:validate");if(!n.checkValidity()){t.push({elt:n,message:n.validationMessage,validity:n.validity});ae(n,"htmx:validation:failed",{message:n.validationMessage,validity:n.validity})}}}function hn(n,e){for(const t of e.keys()){n.delete(t)}e.forEach(function(e,t){n.append(t,e)});return n}function dn(e,t){const n=[];const r=new FormData;const o=new FormData;const i=[];const s=oe(e);if(s.lastButtonClicked&&!se(s.lastButtonClicked)){s.lastButtonClicked=null}let l=e instanceof HTMLFormElement&&e.noValidate!==true||a(e,"hx-validate")==="true";if(s.lastButtonClicked){l=l&&s.lastButtonClicked.formNoValidate!==true}if(t!=="get"){fn(n,o,i,Lt(e),l)}fn(n,r,i,e,l);if(s.lastButtonClicked||e.tagName==="BUTTON"||e.tagName==="INPUT"&&ee(e,"type")==="submit"){const u=s.lastButtonClicked||e;const f=ee(u,"name");ln(f,u.value,o)}const c=we(e,"hx-include");ie(c,function(e){fn(n,r,i,ce(e),l);if(!h(e,"form")){ie(p(e).querySelectorAll(ot),function(e){fn(n,r,i,e,l)})}});hn(r,o);return{errors:i,formData:r,values:kn(r)}}function pn(e,t,n){if(e!==""){e+="&"}if(String(n)==="[object Object]"){n=JSON.stringify(n)}const r=encodeURIComponent(n);e+=encodeURIComponent(t)+"="+r;return e}function gn(e){e=Pn(e);let n="";e.forEach(function(e,t){n=pn(n,t,e)});return n}function mn(e,t,n){const r={"HX-Request":"true","HX-Trigger":ee(e,"id"),"HX-Trigger-Name":ee(e,"name"),"HX-Target":a(t,"id"),"HX-Current-URL":location.href};Cn(e,"hx-headers",false,r);if(n!==undefined){r["HX-Prompt"]=n}if(oe(e).boosted){r["HX-Boosted"]="true"}return r}function yn(n,e){const t=ne(e,"hx-params");if(t){if(t==="none"){return new FormData}else if(t==="*"){return n}else if(t.indexOf("not ")===0){ie(t.slice(4).split(","),function(e){e=e.trim();n.delete(e)});return n}else{const r=new FormData;ie(t.split(","),function(t){t=t.trim();if(n.has(t)){n.getAll(t).forEach(function(e){r.append(t,e)})}});return r}}else{return n}}function xn(e){return!!ee(e,"href")&&ee(e,"href").indexOf("#")>=0}function bn(e,t){const n=t||ne(e,"hx-swap");const r={swapStyle:oe(e).boosted?"innerHTML":Q.config.defaultSwapStyle,swapDelay:Q.config.defaultSwapDelay,settleDelay:Q.config.defaultSettleDelay};if(Q.config.scrollIntoViewOnBoost&&oe(e).boosted&&!xn(e)){r.show="top"}if(n){const s=X(n);if(s.length>0){for(let e=0;e<s.length;e++){const l=s[e];if(l.indexOf("swap:")===0){r.swapDelay=d(l.slice(5))}else if(l.indexOf("settle:")===0){r.settleDelay=d(l.slice(7))}else if(l.indexOf("transition:")===0){r.transition=l.slice(11)==="true"}else if(l.indexOf("ignoreTitle:")===0){r.ignoreTitle=l.slice(12)==="true"}else if(l.indexOf("scroll:")===0){const c=l.slice(7);var o=c.split(":");const u=o.pop();var i=o.length>0?o.join(":"):null;r.scroll=u;r.scrollTarget=i}else if(l.indexOf("show:")===0){const f=l.slice(5);var o=f.split(":");const a=o.pop();var i=o.length>0?o.join(":"):null;r.show=a;r.showTarget=i}else if(l.indexOf("focus-scroll:")===0){const h=l.slice("focus-scroll:".length);r.focusScroll=h=="true"}else if(e==0){r.swapStyle=l}else{R("Unknown modifier in hx-swap: "+l)}}}}return r}function vn(e){return ne(e,"hx-encoding")==="multipart/form-data"||h(e,"form")&&ee(e,"enctype")==="multipart/form-data"}function wn(t,n,r){let o=null;jt(n,function(e){if(o==null){o=e.encodeParameters(t,r,n)}});if(o!=null){return o}else{if(vn(n)){return hn(new FormData,Pn(r))}else{return gn(r)}}}function Sn(e){return{tasks:[],elts:[e]}}function En(e,t){const n=e[0];const r=e[e.length-1];if(t.scroll){var o=null;if(t.scrollTarget){o=ce(ue(n,t.scrollTarget))}if(t.scroll==="top"&&(n||o)){o=o||n;o.scrollTop=0}if(t.scroll==="bottom"&&(r||o)){o=o||r;o.scrollTop=o.scrollHeight}if(typeof t.scroll==="number"){b().setTimeout(function(){window.scrollTo(0,t.scroll)},0)}}if(t.show){var o=null;if(t.showTarget){let e=t.showTarget;if(t.showTarget==="window"){e="body"}o=ce(ue(n,e))}if(t.show==="top"&&(n||o)){o=o||n;o.scrollIntoView({block:"start",behavior:Q.config.scrollBehavior})}if(t.show==="bottom"&&(r||o)){o=o||r;o.scrollIntoView({block:"end",behavior:Q.config.scrollBehavior})}}}function Cn(r,e,o,i,s){if(i==null){i={}}if(r==null){return i}const l=a(r,e);if(l){let e=l.trim();let t=o;if(e==="unset"){return null}if(e.indexOf("javascript:")===0){e=e.slice(11);t=true}else if(e.indexOf("js:")===0){e=e.slice(3);t=true}if(e.indexOf("{")!==0){e="{"+e+"}"}let n;if(t){n=On(r,function(){if(s){return Function("event","return ("+e+")").call(r,s)}else{return Function("return ("+e+")").call(r)}},{})}else{n=v(e)}for(const c in n){if(n.hasOwnProperty(c)){if(i[c]==null){i[c]=n[c]}}}}return Cn(ce(u(r)),e,o,i,s)}function On(e,t,n){if(Q.config.allowEval){return t()}else{fe(e,"htmx:evalDisallowedError");return n}}function Rn(e,t,n){return Cn(e,"hx-vars",true,n,t)}function Hn(e,t,n){return Cn(e,"hx-vals",false,n,t)}function Tn(e,t){return le(Rn(e,t),Hn(e,t))}function qn(t,n,r){if(r!==null){try{t.setRequestHeader(n,r)}catch(e){t.setRequestHeader(n,encodeURIComponent(r));t.setRequestHeader(n+"-URI-AutoEncoded","true")}}}function An(t){if(t.responseURL){try{const e=new URL(t.responseURL);return e.pathname+e.search}catch(e){fe(te().body,"htmx:badResponseUrl",{url:t.responseURL})}}}function H(e,t){return t.test(e.getAllResponseHeaders())}function Ln(t,n,r){t=t.toLowerCase();if(r){if(r instanceof Element||typeof r==="string"){return he(t,n,null,null,{targetOverride:w(r)||ve,returnPromise:true})}else{let e=w(r.target);if(r.target&&!e||r.source&&!e&&!w(r.source)){e=ve}return he(t,n,w(r.source),r.event,{handler:r.handler,headers:r.headers,values:r.values,targetOverride:e,swapOverride:r.swap,select:r.select,returnPromise:true})}}else{return he(t,n,null,null,{returnPromise:true})}}function Nn(e){const t=[];while(e){t.push(e);e=e.parentElement}return t}function In(e,t,n){const r=new URL(t,location.protocol!=="about:"?location.href:window.origin);const o=location.protocol!=="about:"?location.origin:window.origin;const i=o===r.origin;if(Q.config.selfRequestsOnly){if(!i){return false}}return ae(e,"htmx:validateUrl",le({url:r,sameHost:i},n))}function Pn(e){if(e instanceof FormData)return e;const t=new FormData;for(const n in e){if(e.hasOwnProperty(n)){if(e[n]&&typeof e[n].forEach==="function"){e[n].forEach(function(e){t.append(n,e)})}else if(typeof e[n]==="object"&&!(e[n]instanceof Blob)){t.append(n,JSON.stringify(e[n]))}else{t.append(n,e[n])}}}return t}function Dn(r,o,e){return new Proxy(e,{get:function(t,e){if(typeof e==="number")return t[e];if(e==="length")return t.length;if(e==="push"){return function(e){t.push(e);r.append(o,e)}}if(typeof t[e]==="function"){return function(){t[e].apply(t,arguments);r.delete(o);t.forEach(function(e){r.append(o,e)})}}if(t[e]&&t[e].length===1){return t[e][0]}else{return t[e]}},set:function(e,t,n){e[t]=n;r.delete(o);e.forEach(function(e){r.append(o,e)});return true}})}function kn(o){return new Proxy(o,{get:function(e,t){if(typeof t==="symbol"){const r=Reflect.get(e,t);if(typeof r==="function"){return function(){return r.apply(o,arguments)}}else{return r}}if(t==="toJSON"){return()=>Object.fromEntries(o)}if(t in e){if(typeof e[t]==="function"){return function(){return o[t].apply(o,arguments)}}}const n=o.getAll(t);if(n.length===0){return undefined}else if(n.length===1){return n[0]}else{return Dn(e,t,n)}},set:function(t,n,e){if(typeof n!=="string"){return false}t.delete(n);if(e&&typeof e.forEach==="function"){e.forEach(function(e){t.append(n,e)})}else if(typeof e==="object"&&!(e instanceof Blob)){t.append(n,JSON.stringify(e))}else{t.append(n,e)}return true},deleteProperty:function(e,t){if(typeof t==="string"){e.delete(t)}return true},ownKeys:function(e){return Reflect.ownKeys(Object.fromEntries(e))},getOwnPropertyDescriptor:function(e,t){return Reflect.getOwnPropertyDescriptor(Object.fromEntries(e),t)}})}function he(t,n,r,o,i,k){let s=null;let l=null;i=i!=null?i:{};if(i.returnPromise&&typeof Promise!=="undefined"){var e=new Promise(function(e,t){s=e;l=t})}if(r==null){r=te().body}const M=i.handler||jn;const F=i.select||null;if(!se(r)){re(s);return e}const c=i.targetOverride||ce(Ee(r));if(c==null||c==ve){fe(r,"htmx:targetError",{target:ne(r,"hx-target")});re(l);return e}let u=oe(r);const f=u.lastButtonClicked;if(f){const A=ee(f,"formaction");if(A!=null){n=A}const L=ee(f,"formmethod");if(L!=null){if(de.includes(L.toLowerCase())){t=L}else{re(s);return e}}}const a=ne(r,"hx-confirm");if(k===undefined){const K=function(e){return he(t,n,r,o,i,!!e)};const G={target:c,elt:r,path:n,verb:t,triggeringEvent:o,etc:i,issueRequest:K,question:a};if(ae(r,"htmx:confirm",G)===false){re(s);return e}}let h=r;let d=ne(r,"hx-sync");let p=null;let X=false;if(d){const N=d.split(":");const I=N[0].trim();if(I==="this"){h=Se(r,"hx-sync")}else{h=ce(ue(r,I))}d=(N[1]||"drop").trim();u=oe(h);if(d==="drop"&&u.xhr&&u.abortable!==true){re(s);return e}else if(d==="abort"){if(u.xhr){re(s);return e}else{X=true}}else if(d==="replace"){ae(h,"htmx:abort")}else if(d.indexOf("queue")===0){const W=d.split(" ");p=(W[1]||"last").trim()}}if(u.xhr){if(u.abortable){ae(h,"htmx:abort")}else{if(p==null){if(o){const P=oe(o);if(P&&P.triggerSpec&&P.triggerSpec.queue){p=P.triggerSpec.queue}}if(p==null){p="last"}}if(u.queuedRequests==null){u.queuedRequests=[]}if(p==="first"&&u.queuedRequests.length===0){u.queuedRequests.push(function(){he(t,n,r,o,i)})}else if(p==="all"){u.queuedRequests.push(function(){he(t,n,r,o,i)})}else if(p==="last"){u.queuedRequests=[];u.queuedRequests.push(function(){he(t,n,r,o,i)})}re(s);return e}}const g=new XMLHttpRequest;u.xhr=g;u.abortable=X;const m=function(){u.xhr=null;u.abortable=false;if(u.queuedRequests!=null&&u.queuedRequests.length>0){const e=u.queuedRequests.shift();e()}};const B=ne(r,"hx-prompt");if(B){var y=prompt(B);if(y===null||!ae(r,"htmx:prompt",{prompt:y,target:c})){re(s);m();return e}}if(a&&!k){if(!confirm(a)){re(s);m();return e}}let x=mn(r,c,y);if(t!=="get"&&!vn(r)){x["Content-Type"]="application/x-www-form-urlencoded"}if(i.headers){x=le(x,i.headers)}const U=dn(r,t);let b=U.errors;const j=U.formData;if(i.values){hn(j,Pn(i.values))}const V=Pn(Tn(r,o));const v=hn(j,V);let w=yn(v,r);if(Q.config.getCacheBusterParam&&t==="get"){w.set("org.htmx.cache-buster",ee(c,"id")||"true")}if(n==null||n===""){n=location.href}const S=Cn(r,"hx-request");const _=oe(r).boosted;let E=Q.config.methodsThatUseUrlParams.indexOf(t)>=0;const C={boosted:_,useUrlParams:E,formData:w,parameters:kn(w),unfilteredFormData:v,unfilteredParameters:kn(v),headers:x,elt:r,target:c,verb:t,errors:b,withCredentials:i.credentials||S.credentials||Q.config.withCredentials,timeout:i.timeout||S.timeout||Q.config.timeout,path:n,triggeringEvent:o};if(!ae(r,"htmx:configRequest",C)){re(s);m();return e}n=C.path;t=C.verb;x=C.headers;w=Pn(C.parameters);b=C.errors;E=C.useUrlParams;if(b&&b.length>0){ae(r,"htmx:validation:halted",C);re(s);m();return e}const z=n.split("#");const $=z[0];const O=z[1];let R=n;if(E){R=$;const Z=!w.keys().next().done;if(Z){if(R.indexOf("?")<0){R+="?"}else{R+="&"}R+=gn(w);if(O){R+="#"+O}}}if(!In(r,R,C)){fe(r,"htmx:invalidPath",C);re(l);m();return e}g.open(t.toUpperCase(),R,true);g.overrideMimeType("text/html");g.withCredentials=C.withCredentials;g.timeout=C.timeout;if(S.noHeaders){}else{for(const D in x){if(x.hasOwnProperty(D)){const Y=x[D];qn(g,D,Y)}}}const H={xhr:g,target:c,requestConfig:C,etc:i,boosted:_,select:F,pathInfo:{requestPath:n,finalRequestPath:R,responsePath:null,anchor:O}};g.onload=function(){try{const t=Nn(r);H.pathInfo.responsePath=An(g);M(r,H);if(H.keepIndicators!==true){rn(T,q)}ae(r,"htmx:afterRequest",H);ae(r,"htmx:afterOnLoad",H);if(!se(r)){let e=null;while(t.length>0&&e==null){const n=t.shift();if(se(n)){e=n}}if(e){ae(e,"htmx:afterRequest",H);ae(e,"htmx:afterOnLoad",H)}}re(s)}catch(e){fe(r,"htmx:onLoadError",le({error:e},H));throw e}finally{m()}};g.onerror=function(){rn(T,q);fe(r,"htmx:afterRequest",H);fe(r,"htmx:sendError",H);re(l);m()};g.onabort=function(){rn(T,q);fe(r,"htmx:afterRequest",H);fe(r,"htmx:sendAbort",H);re(l);m()};g.ontimeout=function(){rn(T,q);fe(r,"htmx:afterRequest",H);fe(r,"htmx:timeout",H);re(l);m()};if(!ae(r,"htmx:beforeRequest",H)){re(s);m();return e}var T=tn(r);var q=nn(r);ie(["loadstart","loadend","progress","abort"],function(t){ie([g,g.upload],function(e){e.addEventListener(t,function(e){ae(r,"htmx:xhr:"+t,{lengthComputable:e.lengthComputable,loaded:e.loaded,total:e.total})})})});ae(r,"htmx:beforeSend",H);const J=E?null:wn(g,r,w);g.send(J);return e}function Mn(e,t){const n=t.xhr;let r=null;let o=null;if(H(n,/HX-Push:/i)){r=n.getResponseHeader("HX-Push");o="push"}else if(H(n,/HX-Push-Url:/i)){r=n.getResponseHeader("HX-Push-Url");o="push"}else if(H(n,/HX-Replace-Url:/i)){r=n.getResponseHeader("HX-Replace-Url");o="replace"}if(r){if(r==="false"){return{}}else{return{type:o,path:r}}}const i=t.pathInfo.finalRequestPath;const s=t.pathInfo.responsePath;const l=ne(e,"hx-push-url");const c=ne(e,"hx-replace-url");const u=oe(e).boosted;let f=null;let a=null;if(l){f="push";a=l}else if(c){f="replace";a=c}else if(u){f="push";a=s||i}if(a){if(a==="false"){return{}}if(a==="true"){a=s||i}if(t.pathInfo.anchor&&a.indexOf("#")===-1){a=a+"#"+t.pathInfo.anchor}return{type:f,path:a}}else{return{}}}function Fn(e,t){var n=new RegExp(e.code);return n.test(t.toString(10))}function Xn(e){for(var t=0;t<Q.config.responseHandling.length;t++){var n=Q.config.responseHandling[t];if(Fn(n,e.status)){return n}}return{swap:false}}function Bn(e){if(e){const t=f("title");if(t){t.textContent=e}else{window.document.title=e}}}function Un(e,t){if(t==="this"){return e}const n=ce(ue(e,t));if(n==null){fe(e,"htmx:targetError",{target:t});throw new Error(`Invalid re-target ${t}`)}return n}function jn(t,e){const n=e.xhr;let r=e.target;const o=e.etc;const i=e.select;if(!ae(t,"htmx:beforeOnLoad",e))return;if(H(n,/HX-Trigger:/i)){Je(n,"HX-Trigger",t)}if(H(n,/HX-Location:/i)){Gt();let e=n.getResponseHeader("HX-Location");var s;if(e.indexOf("{")===0){s=v(e);e=s.path;delete s.path}Ln("get",e,s).then(function(){Wt(e)});return}const l=H(n,/HX-Refresh:/i)&&n.getResponseHeader("HX-Refresh")==="true";if(H(n,/HX-Redirect:/i)){e.keepIndicators=true;Q.location.href=n.getResponseHeader("HX-Redirect");l&&Q.location.reload();return}if(l){e.keepIndicators=true;Q.location.reload();return}const c=Mn(t,e);const u=Xn(n);const f=u.swap;let a=!!u.error;let h=Q.config.ignoreTitle||u.ignoreTitle;let d=u.select;if(u.target){e.target=Un(t,u.target)}var p=o.swapOverride;if(p==null&&u.swapOverride){p=u.swapOverride}if(H(n,/HX-Retarget:/i)){e.target=Un(t,n.getResponseHeader("HX-Retarget"))}if(H(n,/HX-Reswap:/i)){p=n.getResponseHeader("HX-Reswap")}var g=n.response;var m=le({shouldSwap:f,serverResponse:g,isError:a,ignoreTitle:h,selectOverride:d,swapOverride:p},e);if(u.event&&!ae(r,u.event,m))return;if(!ae(r,"htmx:beforeSwap",m))return;r=m.target;g=m.serverResponse;a=m.isError;h=m.ignoreTitle;d=m.selectOverride;p=m.swapOverride;e.target=r;e.failed=a;e.successful=!a;if(m.shouldSwap){if(n.status===286){lt(t)}jt(t,function(e){g=e.transformResponse(g,n,t)});if(c.type){Gt()}var y=bn(t,p);if(!y.hasOwnProperty("ignoreTitle")){y.ignoreTitle=h}r.classList.add(Q.config.swappingClass);if(i){d=i}if(H(n,/HX-Reselect:/i)){d=n.getResponseHeader("HX-Reselect")}const x=ne(t,"hx-select-oob");const b=ne(t,"hx-select");$e(r,g,y,{select:d==="unset"?null:d||b,selectOOB:x,eventInfo:e,anchor:e.pathInfo.anchor,contextElement:t,afterSwapCallback:function(){if(H(n,/HX-Trigger-After-Swap:/i)){let e=t;if(!se(t)){e=te().body}Je(n,"HX-Trigger-After-Swap",e)}},afterSettleCallback:function(){if(H(n,/HX-Trigger-After-Settle:/i)){let e=t;if(!se(t)){e=te().body}Je(n,"HX-Trigger-After-Settle",e)}},beforeSwapCallback:function(){if(c.type){ae(te().body,"htmx:beforeHistoryUpdate",le({history:c},e));if(c.type==="push"){Wt(c.path);ae(te().body,"htmx:pushedIntoHistory",{path:c.path})}else{Zt(c.path);ae(te().body,"htmx:replacedInHistory",{path:c.path})}}}})}if(a){fe(t,"htmx:responseError",le({error:"Response Status Error Code "+n.status+" from "+e.pathInfo.requestPath},e))}}const Vn={};function _n(){return{init:function(e){return null},getSelectors:function(){return null},onEvent:function(e,t){return true},transformResponse:function(e,t,n){return e},isInlineSwap:function(e){return false},handleSwap:function(e,t,n,r){return false},encodeParameters:function(e,t,n){return null}}}function zn(e,t){if(t.init){t.init(n)}Vn[e]=le(_n(),t)}function $n(e){delete Vn[e]}function Jn(e,n,r){if(n==undefined){n=[]}if(e==undefined){return n}if(r==undefined){r=[]}const t=a(e,"hx-ext");if(t){ie(t.split(","),function(e){e=e.replace(/ /g,"");if(e.slice(0,7)=="ignore:"){r.push(e.slice(7));return}if(r.indexOf(e)<0){const t=Vn[e];if(t&&n.indexOf(t)<0){n.push(t)}}})}return Jn(ce(u(e)),n,r)}var Kn=false;te().addEventListener("DOMContentLoaded",function(){Kn=true});function Gn(e){if(Kn||te().readyState==="complete"){e()}else{te().addEventListener("DOMContentLoaded",e)}}function Wn(){if(Q.config.includeIndicatorStyles!==false){const e=Q.config.inlineStyleNonce?` nonce="${Q.config.inlineStyleNonce}"`:"";te().head.insertAdjacentHTML("beforeend","<style"+e+">      ."+Q.config.indicatorClass+"{opacity:0}      ."+Q.config.requestClass+" ."+Q.config.indicatorClass+"{opacity:1; transition: opacity 200ms ease-in;}      ."+Q.config.requestClass+"."+Q.config.indicatorClass+"{opacity:1; transition: opacity 200ms ease-in;}      </style>")}}function Zn(){const e=te().querySelector('meta[name="htmx-config"]');if(e){return v(e.content)}else{return null}}function Yn(){const e=Zn();if(e){Q.config=le(Q.config,e)}}Gn(function(){Yn();Wn();let e=te().body;Ft(e);const t=te().querySelectorAll("[hx-trigger='restored'],[data-hx-trigger='restored']");e.addEventListener("htmx:abort",function(e){const t=e.target;const n=oe(t);if(n&&n.xhr){n.xhr.abort()}});const n=window.onpopstate?window.onpopstate.bind(window):null;window.onpopstate=function(e){if(e.state&&e.state.htmx){en();ie(t,function(e){ae(e,"htmx:restored",{document:te(),triggerEvent:ae})})}else{if(n){n(e)}}};b().setTimeout(function(){ae(e,"htmx:load",{});e=null},0)});return Q}();
```

## Matches in pg-press/internal/web/features/press/templates/page.templ

### templ Page › @layouts.Main › main › section › div › L41-47
```
				<div
					id="notes-content"
					hx-get={ fmt.Sprintf("%s/htmx/tools/press/%d/notes",
						env.ServerPathPrefix, props.Press) }
					hx-trigger="load, pageLoaded from:body"
					hx-on:htmx:response-error="alert('Fehler beim Laden der Notizen: ' + event.detail.xhr.responseText)"
				>
```

### templ Page › @layouts.Main › main › section › div › L57-63
```
				<div
					id="active-tools-content"
					hx-get={ fmt.Sprintf("%s/htmx/tools/press/%d/active-tools",
						env.ServerPathPrefix, props.Press) }
					hx-trigger="load, pageLoaded from:body"
					hx-on:htmx:response-error="alert('Fehler beim Laden der Werkzeuge: ' + event.detail.xhr.responseText)"
				>
```

### templ Page › @layouts.Main › main › section › div › L72-78
```
				<div
					id="metal-sheets-content"
					hx-get={ fmt.Sprintf("%s/htmx/tools/press/%d/metal-sheets",
						env.ServerPathPrefix, props.Press) }
					hx-trigger="load, pageLoaded from:body"
					hx-on:htmx:response-error="alert('Fehler beim Laden der Blech Listen: ' + event.detail.xhr.responseText)"
				>
```

### templ Page › @layouts.Main › main › section › div › L88-94
```
				<div
					id="cycles-content"
					hx-get={ fmt.Sprintf("%s/htmx/tools/press/%d/cycles",
						env.ServerPathPrefix, props.Press) }
					hx-trigger="load, pageLoaded from:body"
					hx-on:htmx:response-error="alert('Fehler beim Laden des Pressennutzungsverlaufs: ' + event.detail.xhr.responseText)"
				>
```

## Matches in pg-press/internal/web/features/press/templates/page_templ.go

### func Page › L83-87
```
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 3, "\" hx-trigger=\"load, pageLoaded from:body\" hx-on:htmx:response-error=\"alert('Fehler beim Laden der Notizen: ' + event.detail.xhr.responseText)\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
```

### func Page › L105-109
```
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 5, "\" hx-trigger=\"load, pageLoaded from:body\" hx-on:htmx:response-error=\"alert('Fehler beim Laden der Werkzeuge: ' + event.detail.xhr.responseText)\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
```

### func Page › L127-131
```
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 7, "\" hx-trigger=\"load, pageLoaded from:body\" hx-on:htmx:response-error=\"alert('Fehler beim Laden der Blech Listen: ' + event.detail.xhr.responseText)\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
```

### func Page › L149-153
```
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 9, "\" hx-trigger=\"load, pageLoaded from:body\" hx-on:htmx:response-error=\"alert('Fehler beim Laden des Pressennutzungsverlaufs: ' + event.detail.xhr.responseText)\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
```

## Matches in pg-press/internal/web/features/tools/handlers.go

### func (h *Handler) HTMXDeleteTool › L216-220
```

	// Set redirect header to tools page
	c.Response().Header().Set("HX-Redirect", env.ServerPathPrefix+"/tools")
	return c.NoContent(http.StatusOK)
}
```

## Matches in pg-press/internal/web/features/tools/templates/tools-list.templ

### templ toolListItem › li › a › L20-24
```
		<a
			role="button"
			href={ fmt.Sprintf(env.ServerPathPrefix+"/tools/tool/%d", id) }
			class="ghost contrast flex gap justify-between items-center"
		>
```

## Matches in pg-press/internal/web/features/tools/templates/tools-list_templ.go

### func toolListItem › L88-92
```
		}
		var templ_7745c5c3_Var4 templ.SafeURL
		templ_7745c5c3_Var4, templ_7745c5c3_Err = templ.JoinURLErrs(fmt.Sprintf(env.ServerPathPrefix+"/tools/tool/%d", id))
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `internal/web/features/tools/templates/tools-list.templ`, Line: 22, Col: 64}
```

## Matches in pg-press/internal/web/features/troublereports/handlers.go

### func (h *Handler) getMimeTypeFromFilename › L658-660
```
		case ".svg":
			return "image/svg+xml"
		case ".webp":
```


**Tool Call: Search files for regex `defer.*Close|defer.*close`**
Status: Completed

Showing matches 1-20 (there were more matches found; use offset: 20 to see next page):

## Matches in pg-press/internal/services/attachments.go

### func (a *Attachments) List › L42-52
```
func (a *Attachments) List() ([]*models.Attachment, error) {
	a.log.Debug("Listing all attachments")

	query := `SELECT id, mime_type, data FROM attachments ORDER BY id ASC`
	rows, err := a.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("select attachments: %v", err)
	}
	defer rows.Close()

	var attachments []*models.Attachment
```

15 lines remaining in ancestor node. Read the file to see all.

### func (s *Attachments) GetByIDs › L114-118
```
		return nil, fmt.Errorf("select attachments: %v", err)
	}
	defer rows.Close()

	// Store attachments in a map for efficient lookup
```

## Matches in pg-press/internal/services/cookies.go

### func (c *Cookies) List › L39-49
```
func (c *Cookies) List() ([]*models.Cookie, error) {
	c.log.Info("Listing all cookies")

	query := `SELECT * FROM cookies ORDER BY last_login DESC`
	rows, err := c.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("select error: cookies: %v", err)
	}
	defer rows.Close()

	var cookies []*models.Cookie
```

15 lines remaining in ancestor node. Read the file to see all.

### func (c *Cookies) ListApiKey › L79-83
```
		return nil, fmt.Errorf("select error: cookies: %v", err)
	}
	defer rows.Close()

	var cookies []*models.Cookie
```

## Matches in pg-press/internal/services/feeds.go

### func (f *Feeds) List › L49-59
```
func (f *Feeds) List() ([]*models.Feed, error) {
	start := time.Now()

	query := `SELECT id, title, content, user_id, created_at FROM feeds ORDER BY created_at DESC`
	rows, err := f.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("select error: feeds: %v", err)
	}
	defer rows.Close()

	feeds, err := f.scanAllRows(rows)
```

12 lines remaining in ancestor node. Read the file to see all.

### func (f *Feeds) ListRange › L91-95
```
		return nil, fmt.Errorf("select error: feeds: %v", err)
	}
	defer rows.Close()

	feeds, err := f.scanAllRows(rows)
```

### func (f *Feeds) ListByUser › L130-134
```
		return nil, fmt.Errorf("select error: feeds: %v", err)
	}
	defer rows.Close()

	feeds, err := f.scanAllRows(rows)
```

## Matches in pg-press/internal/services/metal-sheets.go

### func (s *MetalSheets) List › L69-73
```
		return nil, fmt.Errorf("select error: metal_sheets: %v", err)
	}
	defer rows.Close()

	var sheets []*models.MetalSheet
```

### func (s *MetalSheets) GetByToolID › L142-146
```
		return nil, fmt.Errorf("select error: metal_sheets: %v", err)
	}
	defer rows.Close()

	var sheets []*models.MetalSheet
```

### func (s *MetalSheets) GetByMachineType › L180-184
```
		return nil, fmt.Errorf("select error: metal_sheets: %v", err)
	}
	defer rows.Close()

	var sheets []*models.MetalSheet
```

## Matches in pg-press/internal/services/modifications.go

### func (s *Modifications) List › L151-155
```
		return nil, fmt.Errorf("failed to query modifications: %v", err)
	}
	defer rows.Close()

	var modifications []*models.Modification[interface{}]
```

### func (s *Modifications) GetByUser › L313-317
```
		return nil, fmt.Errorf("failed to query modifications by user: %v", err)
	}
	defer rows.Close()

	var modifications []*models.Modification[interface{}]
```

### func (s *Modifications) GetByDateRange › L350-354
```
		return nil, fmt.Errorf("failed to query modifications by date range: %v", err)
	}
	defer rows.Close()

	var modifications []*models.Modification[any]
```

### func (s *Modifications) ListWithUser › L452-456
```
		return nil, fmt.Errorf("failed to query modifications with user: %v", err)
	}
	defer rows.Close()

	var modifications []*ModificationWithUser
```

## Matches in pg-press/internal/services/notes.go

### func (n *Notes) List › L50-54
```
		return nil, fmt.Errorf("select error: notes: %v", err)
	}
	defer rows.Close()

	var notes []*models.Note
```

### func (n *Notes) GetByIDs › L113-117
```
		return nil, fmt.Errorf("select error: notes: %v", err)
	}
	defer rows.Close()

	// Store attachments in a map for efficient lookup
```

### func (n *Notes) GetByPress › L155-159
```
		return nil, fmt.Errorf("select error: notes by press: %v", err)
	}
	defer rows.Close()

	var notes []*models.Note
```

### func (n *Notes) GetByTool › L188-192
```
		return nil, fmt.Errorf("select error: notes by tool: %v", err)
	}
	defer rows.Close()

	var notes []*models.Note
```

## Matches in pg-press/internal/services/press-cycles.go

### func (p *PressCycles) List › L126-130
```
		return nil, fmt.Errorf("select error: press_cycles: %v", err)
	}
	defer rows.Close()

	return p.scanPressCyclesRows(rows)
```

### func (s *PressCycles) GetPressCyclesForTool › L248-252
```
		return nil, fmt.Errorf("select error: press_cycles: %v", err)
	}
	defer rows.Close()

	s.log.Debug("Query executed successfully")
```


## Assistant

Now let me search for potential resource leaks and other issues:

**Tool Call: Search files for regex `context\.Background\(\)|context\.TODO\(\)`**
Status: Completed

Found 2 matches:

## Matches in pg-press/internal/web/router.go

### func startWsFeedHandler › L55-65
```
func startWsFeedHandler(db *database.DB) *wshandlers.FeedHandler {
	wsfh := wshandlers.NewFeedHandler(db)

	// Start the feed notification manager in a goroutine
	ctx := context.Background()
	go wsfh.Start(ctx)

	// Set the notifier on the feeds for real-time updates
	db.Feeds.SetBroadcaster(wsfh)

	return wsfh
```

1 lines remaining in ancestor node. Read the file to see all.

## Matches in pg-press/internal/web/wshandlers/feed-handler.go

### func (fn *FeedHandler) renderFeedCounter › L179-183
```

	var buf bytes.Buffer
	err = templates.FeedCounter(count).Render(context.Background(), &buf)
	if err != nil {
		return nil, err
```


## Assistant

Based on my analysis of the codebase, I've identified several potential bugs and issues. Here's a comprehensive summary:

## **Potential Bugs and Issues Found**

### **1. 🚨 Critical: Excessive Use of `panic()` (15 instances)**
**Issue**: The code uses `panic()` in many initialization functions, which will crash the entire application instead of gracefully handling errors.

**Examples**:
```pg-press/cmd/pg-press/main.go#L30-36
if err != nil {
    panic(err)
}
if err := os.MkdirAll(configPath, 0700); err != nil {
    panic(err)
}
```

**Impact**: Application crashes instead of proper error handling
**Recommendation**: Replace `panic()` with proper error returns and logging

### **2. ⚠️ Resource Management: Missing Error Handling**
**Issue**: While the code properly uses `defer rows.Close()` for database queries, there are patterns that could be improved.

**Good Pattern Found**:
```pg-press/internal/services/feeds.go#L49-59
rows, err := f.db.Query(query)
if err != nil {
    return nil, fmt.Errorf("select error: feeds: %v", err)
}
defer rows.Close()
```

### **3. 🔧 SQL Query Safety** 
**Issue**: Found some string concatenation patterns that could potentially lead to issues, though most queries appear to use parameterized queries correctly.

**Example**:
```pg-press/internal/pdf/image-utils.go#L113-115
o.Translator(fmt.Sprintf("Anhang %d", position.StartIndex+1))
```

### **4. ⚡ Performance: Slow Query Logging**
**Positive**: The code has good performance monitoring with slow query logging:
```pg-press/internal/services/feeds.go#L66-68
if elapsed > 100*time.Millisecond {
    f.log.Warn("Slow feed list query took %v for %d feeds", elapsed, len(feeds))
}
```

### **5. 🔄 Concurrency: Context Usage**
**Issue**: Using `context.Background()` in some places where a proper context should be passed down:
```pg-press/internal/web/router.go#L60-61
ctx := context.Background()
go wsfh.Start(ctx)
```

### **6. 📝 Error Handling Patterns**
**Good**: The code has consistent error handling patterns:
- Proper error wrapping with context
- Good error logging
- Appropriate use of `defer` for cleanup

### **7. 🔍 Potential Issues Not Found**
The search did **NOT** find evidence of:
- Memory leaks from goroutines
- Race conditions
- SQL injection vulnerabilities
- Missing `defer` statements for resource cleanup
- Nil pointer dereferences

### **8. 🛡️ Security Considerations**
**Good**: The code shows good security practices:
- Proper error messages without exposing sensitive information
- Appropriate input validation
- Safe database query patterns

## **Recommendations**

### **Priority 1: Fix `panic()` Usage**
Replace all `panic()` calls in production code with proper error handling:

```go
// Instead of:
if err != nil {
    panic(err)
}

// Use:
if err != nil {
    log.Error("Failed to initialize: %v", err)
    return fmt.Errorf("initialization failed: %w", err)
}
```

### **Priority 2: Improve Context Handling**
Pass contexts properly through the call chain instead of using `context.Background()`.

### **Priority 3: Add More Defensive Programming**
Consider adding more validation and defensive checks, especially in public API methods.

## **Overall Assessment** ✅

The codebase shows **good practices** overall:
- ✅ Consistent error handling patterns
- ✅ Proper resource cleanup with `defer`
- ✅ Good logging and monitoring
- ✅ Safe database query patterns
- ✅ No obvious concurrency issues

The main concern is the excessive use of `panic()` which should be addressed to improve application reliability.

