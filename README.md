# pg-vis

A web application for press visualization and management with efficient asset caching.

## Features

- Press management and visualization
- Trouble report generation and PDF export
- Real-time feed updates via WebSockets
- Efficient asset caching for optimal performance
- User authentication and authorization

## Asset Caching

This application implements a comprehensive asset caching strategy to improve performance:

### Cache Headers

Static assets are cached with appropriate headers based on file type:

- **CSS/JS files**: 1 year cache with `immutable` flag
- **Font files**: 1 year cache with `immutable` flag
- **Images**: 30 days cache
- **Icons/Favicons**: 1 week cache
- **JSON files**: 1 day cache

### Asset Versioning

Assets include version parameters for cache invalidation:

- URLs like `/css/ui.min.css?v=1705405800` ensure fresh assets after updates
- Version based on server startup timestamp
- Automatic cache invalidation on server restarts/deployments

### Testing Cache Implementation

Run the caching test script to verify headers:

```bash
# Start the development server
make dev

# In another terminal, test caching
./scripts/test-caching.sh
```

See [docs/CACHING.md](docs/CACHING.md) for detailed implementation details.

# Routing Table

## Page Routes

| Method | Path                        | Handler                                       | Description                          |
| ------ | --------------------------- | --------------------------------------------- | ------------------------------------ |
| GET    | /                           | `handler.Home.handleHome`                     | Renders the home page.               |
| GET    | /login                      | `handler.Auth.handleLogin`                    | Renders the login page.              |
| GET    | /logout                     | `handler.Auth.handleLogout`                   | Logs the user out.                   |
| GET    | /feed                       | `handler.Feed.handleFeed`                     | Renders the feed page.               |
| GET    | /profile                    | `handler.Profile.handleProfile`               | Renders the profile page.            |
| GET    | /tools                      | `handler.Tools.handleTools`                   | Renders the tools page.              |
| GET    | /tools/press/:press         | `handler.Tools.handlePressPage`               | Renders a specific press page.       |
| GET    | /tools/tool/:id             | `handler.Tools.handleToolPage`                | Renders a specific tool page.        |
| GET    | /trouble-reports            | `handler.TroubleReports.handleTroubleReports` | Renders the trouble reports page.    |
| GET    | /trouble-reports/share-pdf  | `handler.TroubleReports.handleGetSharePdf`    | Generates a PDF of a trouble report. |
| GET    | /trouble-reports/attachment | `handler.TroubleReports.handleGetAttachment`  | Serves a trouble report attachment.  |

## HTMX Routes

### Feed

| Method | Path            | Handler                   | Description                |
| ------ | --------------- | ------------------------- | -------------------------- |
| GET    | /htmx/feed/list | `htmx.Feed.handleListGET` | Fetches the list of feeds. |

### Navigation

| Method | Path                   | Handler                                   | Description                 |
| ------ | ---------------------- | ----------------------------------------- | --------------------------- |
| GET    | /htmx/nav/feed-counter | `htmx.Nav.handleFeedCounterWebSocketEcho` | WebSocket for feed counter. |

### Profile

| Method | Path                  | Handler                            | Description                 |
| ------ | --------------------- | ---------------------------------- | --------------------------- |
| GET    | /htmx/profile/cookies | `htmx.Profile.handleGetCookies`    | Fetches the user's cookies. |
| DELETE | /htmx/profile/cookies | `htmx.Profile.handleDeleteCookies` | Deletes a user's cookie.    |

### Tools

| Method | Path                     | Handler                          | Description                          |
| ------ | ------------------------ | -------------------------------- | ------------------------------------ |
| GET    | /htmx/tools/list         | `htmx.Tools.handleList`          | Fetches all tools.                   |
| GET    | /htmx/tools/edit         | `htmx.Tools.handleEdit`          | Renders the tool edit dialog.        |
| POST   | /htmx/tools/edit         | `htmx.Tools.handleEditPOST`      | Creates a new tool.                  |
| PUT    | /htmx/tools/edit         | `htmx.Tools.handleEditPUT`       | Updates a tool.                      |
| DELETE | /htmx/tools/delete       | `htmx.Tools.handleDelete`        | Deletes a tool.                      |
| GET    | /htmx/tools/cycles       | `htmx.Tools.handleCyclesSection` | Fetches the cycles section.          |
| GET    | /htmx/tools/total-cycles | `htmx.Tools.handleTotalCycles`   | Fetches the total cycles for a tool. |
| GET    | /htmx/tools/cycle/edit   | `htmx.Tools.handleCycleEditGET`  | Renders the cycle edit dialog.       |
| POST   | /htmx/tools/cycle/edit   | `htmx.Tools.handleCycleEditPOST` | Creates a new cycle.                 |
| PUT    | /htmx/tools/cycle/edit   | `htmx.Tools.handleCycleEditPUT`  | Updates a cycle.                     |
| DELETE | /htmx/tools/cycle/delete | `htmx.Tools.handleCycleDELETE`   | Deletes a cycle.                     |

### Trouble Reports

| Method | Path                                      | Handler                                           | Description                   |
| ------ | ----------------------------------------- | ------------------------------------------------- | ----------------------------- |
| GET    | /htmx/trouble-reports/dialog-edit         | `htmx.TroubleReports.handleGetDialogEdit`         | Renders the edit dialog.      |
| POST   | /htmx/trouble-reports/dialog-edit         | `htmx.TroubleReports.handlePostDialogEdit`        | Creates a new trouble report. |
| PUT    | /htmx/trouble-reports/dialog-edit         | `htmx.TroubleReports.handlePutDialogEdit`         | Updates a trouble report.     |
| GET    | /htmx/trouble-reports/data                | `htmx.TroubleReports.handleGetData`               | Fetches trouble report data.  |
| DELETE | /htmx/trouble-reports/data                | `htmx.TroubleReports.handleDeleteData`            | Deletes a trouble report.     |
| GET    | /htmx/trouble-reports/attachments-preview | `htmx.TroubleReports.handleGetAttachmentsPreview` | Fetches attachment previews.  |
| GET    | /htmx/trouble-reports/modifications/:id   | `htmx.TroubleReports.handleGetModifications`      | Fetches modifications.        |
| POST   | /htmx/trouble-reports/modifications/:id   | `htmx.TroubleReports.handlePostModifications`     | Restores a modification.      |

## Database

The database is a SQLite database, and the schema is defined in the `internal/database` directory.

### Tables

#### `attachments`

Stores file attachments.

| Column    | Type    | Description                      |
| --------- | ------- | -------------------------------- |
| id        | INTEGER | The attachment ID.               |
| mime_type | TEXT    | The MIME type of the attachment. |
| data      | BLOB    | The attachment data.             |

#### `cookies`

Stores user session cookies.

| Column     | Type    | Description                             |
| ---------- | ------- | --------------------------------------- |
| user_agent | TEXT    | The user agent of the client.           |
| value      | TEXT    | The cookie value.                       |
| api_key    | TEXT    | The API key associated with the cookie. |
| last_login | INTEGER | The timestamp of the last login.        |

#### `feeds`

Stores activity feed entries.

| Column    | Type    | Description                 |
| --------- | ------- | --------------------------- |
| id        | INTEGER | The feed entry ID.          |
| time      | INTEGER | The timestamp of the event. |
| data_type | TEXT    | The type of the event.      |
| data      | BLOB    | The event data.             |

#### `metal_sheets`

Stores information about metal sheets.

| Column       | Type     | Description                                  |
| ------------ | -------- | -------------------------------------------- |
| id           | INTEGER  | The metal sheet ID.                          |
| tile_height  | REAL     | The tile height.                             |
| value        | REAL     | The value.                                   |
| marke_height | INTEGER  | The marke height.                            |
| stf          | REAL     | The STF value.                               |
| stf_max      | REAL     | The maximum STF value.                       |
| tool_id      | INTEGER  | The ID of the tool the sheet is assigned to. |
| notes        | BLOB     | Linked notes.                                |
| mods         | BLOB     | Modification history.                        |
| created_at   | DATETIME | The timestamp of creation.                   |
| updated_at   | DATETIME | The timestamp of the last update.            |

#### `notes`

Stores notes that can be linked to other items.

| Column     | Type     | Description                                     |
| ---------- | -------- | ----------------------------------------------- |
| id         | INTEGER  | The note ID.                                    |
| level      | INTEGER  | The note level (e.g., INFO, ATTENTION, BROKEN). |
| content    | TEXT     | The note content.                               |
| created_at | DATETIME | The timestamp of creation.                      |

#### `press_cycles`

Stores press cycle information for tools.

| Column       | Type     | Description                                  |
| ------------ | -------- | -------------------------------------------- |
| id           | INTEGER  | The press cycle ID.                          |
| press_number | INTEGER  | The press number.                            |
| tool_id      | INTEGER  | The tool ID.                                 |
| date         | DATETIME | The date of the cycle.                       |
| total_cycles | INTEGER  | The total number of cycles.                  |
| performed_by | INTEGER  | The ID of the user who performed the action. |

#### `tool_regenerations`

Stores tool regeneration history.

| Column                 | Type     | Description                                       |
| ---------------------- | -------- | ------------------------------------------------- |
| id                     | INTEGER  | The regeneration ID.                              |
| tool_id                | INTEGER  | The tool ID.                                      |
| regenerated_at         | DATETIME | The timestamp of regeneration.                    |
| cycles_at_regeneration | INTEGER  | The number of cycles at the time of regeneration. |
| reason                 | TEXT     | The reason for regeneration.                      |
| performed_by           | INTEGER  | The ID of the user who performed the action.      |

#### `tools`

Stores tool information.

| Column   | Type    | Description                                |
| -------- | ------- | ------------------------------------------ |
| id       | INTEGER | The tool ID.                               |
| position | TEXT    | The tool position (e.g., 'top', 'bottom'). |
| format   | BLOB    | The tool format.                           |
| type     | TEXT    | The tool type.                             |
| code     | TEXT    | The tool code.                             |
| status   | TEXT    | The tool status.                           |
| press    | INTEGER | The press number the tool is on.           |
| notes    | BLOB    | Linked notes.                              |
| mods     | BLOB    | Modification history.                      |

#### `trouble_reports`

Stores trouble reports.

| Column             | Type    | Description                |
| ------------------ | ------- | -------------------------- |
| id                 | INTEGER | The trouble report ID.     |
| title              | TEXT    | The title of the report.   |
| content            | TEXT    | The content of the report. |
| linked_attachments | TEXT    | Linked attachments.        |
| mods               | BLOB    | Modification history.      |

#### `users`

Stores user information.

| Column      | Type    | Description                                      |
| ----------- | ------- | ------------------------------------------------ |
| telegram_id | INTEGER | The user's Telegram ID.                          |
| user_name   | TEXT    | The user's name.                                 |
| api_key     | TEXT    | The user's API key.                              |
| last_feed   | TEXT    | The ID of the last feed entry the user has seen. |
