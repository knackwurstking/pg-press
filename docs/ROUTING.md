# Routing Table

## Page Routes

| Method | Path                        | Handler                                    | Description                          |
| ------ | --------------------------- | ------------------------------------------ | ------------------------------------ |
| GET    | /                           | `html.Home.handleHome`                     | Renders the home page.               |
| GET    | /login                      | `html.Auth.handleLogin`                    | Renders the login page.              |
| GET    | /logout                     | `html.Auth.handleLogout`                   | Logs the user out.                   |
| GET    | /feed                       | `html.Feed.handleFeed`                     | Renders the feed page.               |
| GET    | /profile                    | `html.Profile.handleProfile`               | Renders the profile page.            |
| GET    | /tools                      | `html.Tools.handleTools`                   | Renders the tools page.              |
| GET    | /tools/press/:press         | `html.Tools.handlePressPage`               | Renders a specific press page.       |
| GET    | /tools/tool/:id             | `html.Tools.handleToolPage`                | Renders a specific tool page.        |
| GET    | /trouble-reports            | `html.TroubleReports.handleTroubleReports` | Renders the trouble reports page.    |
| GET    | /trouble-reports/share-pdf  | `html.TroubleReports.handleGetSharePdf`    | Generates a PDF of a trouble report. |
| GET    | /trouble-reports/attachment | `html.TroubleReports.handleGetAttachment`  | Serves a trouble report attachment.  |

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
