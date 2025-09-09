# Routing Table

## Page Routes

| Method | Path                        | Handler                                    | Query Parameters           | Description                          |
| ------ | --------------------------- | ------------------------------------------ | -------------------------- | ------------------------------------ |
| GET    | /                           | `html.Home.handleHome`                     | -                          | Renders the home page.               |
| GET    | /login                      | `html.Auth.handleLogin`                    | -                          | Renders the login page.              |
| GET    | /logout                     | `html.Auth.handleLogout`                   | -                          | Logs the user out.                   |
| GET    | /feed                       | `html.Feed.handleFeed`                     | -                          | Renders the feed page.               |
| GET    | /profile                    | `html.Profile.handleProfile`               | -                          | Renders the profile page.            |
| GET    | /tools                      | `html.Tools.handleTools`                   | -                          | Renders the tools page.              |
| GET    | /tools/press/:press         | `html.Tools.handlePressPage`               | -                          | Renders a specific press page.       |
| GET    | /tools/tool/:id             | `html.Tools.handleToolPage`                | -                          | Renders a specific tool page.        |
| GET    | /trouble-reports            | `html.TroubleReports.handleTroubleReports` | -                          | Renders the trouble reports page.    |
| GET    | /trouble-reports/share-pdf  | `html.TroubleReports.handleGetSharePdf`    | `id` (required)            | Generates a PDF of a trouble report. |
| GET    | /trouble-reports/attachment | `html.TroubleReports.handleGetAttachment`  | `attachment_id` (required) | Serves a trouble report attachment.  |

## HTMX Routes

### Feed

| Method | Path            | Handler                   | Query Parameters | Description                |
| ------ | --------------- | ------------------------- | ---------------- | -------------------------- |
| GET    | /htmx/feed/list | `htmx.Feed.handleListGET` | -                | Fetches the list of feeds. |

### Navigation

| Method | Path                   | Handler                                   | Query Parameters | Description                 |
| ------ | ---------------------- | ----------------------------------------- | ---------------- | --------------------------- |
| GET    | /htmx/nav/feed-counter | `htmx.Nav.handleFeedCounterWebSocketEcho` | -                | WebSocket for feed counter. |

### Profile

| Method | Path                  | Handler                            | Query Parameters   | Description                 |
| ------ | --------------------- | ---------------------------------- | ------------------ | --------------------------- |
| GET    | /htmx/profile/cookies | `htmx.Profile.handleGetCookies`    | -                  | Fetches the user's cookies. |
| DELETE | /htmx/profile/cookies | `htmx.Profile.handleDeleteCookies` | `value` (required) | Deletes a user's cookie.    |

### Tools

| Method | Path                     | Handler                          | Query Parameters                                                | Description                          |
| ------ | ------------------------ | -------------------------------- | --------------------------------------------------------------- | ------------------------------------ |
| GET    | /htmx/tools/list         | `htmx.Tools.handleList`          | -                                                               | Fetches all tools.                   |
| GET    | /htmx/tools/edit         | `htmx.Tools.handleEdit`          | `id` (optional), `close` (optional)                             | Renders the tool edit dialog.        |
| POST   | /htmx/tools/edit         | `htmx.Tools.handleEditPOST`      | -                                                               | Creates a new tool.                  |
| PUT    | /htmx/tools/edit         | `htmx.Tools.handleEditPUT`       | `id` (required)                                                 | Updates a tool.                      |
| DELETE | /htmx/tools/delete       | `htmx.Tools.handleDelete`        | `id` (required)                                                 | Deletes a tool.                      |
| GET    | /htmx/tools/cycles       | `htmx.Cycles.handleSection`      | `tool_id` (required)                                            | Fetches the cycles section.          |
| GET    | /htmx/tools/total-cycles | `htmx.Cycles.handleTotalCycles`  | `tool_id` (required), `input` (optional)                        | Fetches the total cycles for a tool. |
| GET    | /htmx/tools/cycle/edit   | `htmx.Cycles.handleEditGET`      | `tool_id` (required), `cycle_id` (optional), `close` (optional) | Renders the cycle edit dialog.       |
| POST   | /htmx/tools/cycle/edit   | `htmx.Cycles.handleEditPOST`     | `tool_id` (required)                                            | Creates a new cycle.                 |
| PUT    | /htmx/tools/cycle/edit   | `htmx.Cycles.handleEditPUT`      | `tool_id` (required), `cycle_id` (required)                     | Updates a cycle.                     |
| DELETE | /htmx/tools/cycle/delete | `htmx.Cycles.handleDELETE`       | `tool_id` (required), `cycle_id` (required)                     | Deletes a cycle.                     |
| GET    | /htmx/metal-sheets/edit  | `htmx.MetalSheets.handleEditGET` | -                                                               | Renders the metal sheet edit dialog. |

### Trouble Reports

| Method | Path                                      | Handler                                           | Query Parameters                    | Description                   |
| ------ | ----------------------------------------- | ------------------------------------------------- | ----------------------------------- | ----------------------------- |
| GET    | /htmx/trouble-reports/dialog-edit         | `htmx.TroubleReports.handleGetDialogEdit`         | `id` (optional), `close` (optional) | Renders the edit dialog.      |
| POST   | /htmx/trouble-reports/dialog-edit         | `htmx.TroubleReports.handlePostDialogEdit`        | -                                   | Creates a new trouble report. |
| PUT    | /htmx/trouble-reports/dialog-edit         | `htmx.TroubleReports.handlePutDialogEdit`         | `id` (required)                     | Updates a trouble report.     |
| GET    | /htmx/trouble-reports/data                | `htmx.TroubleReports.handleGetData`               | -                                   | Fetches trouble report data.  |
| DELETE | /htmx/trouble-reports/data                | `htmx.TroubleReports.handleDeleteData`            | `id` (required)                     | Deletes a trouble report.     |
| GET    | /htmx/trouble-reports/attachments-preview | `htmx.TroubleReports.handleGetAttachmentsPreview` | `id` (required), `time` (optional)  | Fetches attachment previews.  |
| GET    | /htmx/trouble-reports/modifications/:id   | `htmx.TroubleReports.handleGetModifications`      | -                                   | Fetches modifications.        |
| POST   | /htmx/trouble-reports/modifications/:id   | `htmx.TroubleReports.handlePostModifications`     | `time` (required)                   | Restores a modification.      |
