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

| Method | Path            | Handler                          | Description                |
| ------ | --------------- | -------------------------------- | -------------------------- |
| GET    | /htmx/feed/list | `htmxhandler.Feed.handleListGET` | Fetches the list of feeds. |

### Navigation

| Method | Path                   | Handler                                          | Description                 |
| ------ | ---------------------- | ------------------------------------------------ | --------------------------- |
| GET    | /htmx/nav/feed-counter | `htmxhandler.Nav.handleFeedCounterWebSocketEcho` | WebSocket for feed counter. |

### Profile

| Method | Path                  | Handler                                   | Description                 |
| ------ | --------------------- | ----------------------------------------- | --------------------------- |
| GET    | /htmx/profile/cookies | `htmxhandler.Profile.handleGetCookies`    | Fetches the user's cookies. |
| DELETE | /htmx/profile/cookies | `htmxhandler.Profile.handleDeleteCookies` | Deletes a user's cookie.    |

### Tools

| Method | Path                     | Handler                                 | Description                    |
| ------ | ------------------------ | --------------------------------------- | ------------------------------ |
| GET    | /htmx/tools/list-all     | `htmxhandler.Tools.handleListAll`       | Fetches all tools.             |
| GET    | /htmx/tools/edit         | `htmxhandler.Tools.handleEdit`          | Renders the tool edit dialog.  |
| POST   | /htmx/tools/edit         | `htmxhandler.Tools.handleEditPOST`      | Creates a new tool.            |
| PUT    | /htmx/tools/edit         | `htmxhandler.Tools.handleEditPUT`       | Updates a tool.                |
| DELETE | /htmx/tools/delete       | `htmxhandler.Tools.handleDelete`        | Deletes a tool.                |
| GET    | /htmx/tools/cycles       | `htmxhandler.Tools.handleCyclesSection` | Fetches the cycles section.    |
| GET    | /htmx/tools/cycle/edit   | `htmxhandler.Tools.handleCycleEditGET`  | Renders the cycle edit dialog. |
| POST   | /htmx/tools/cycle/edit   | `htmxhandler.Tools.handleCycleEditPOST` | Creates a new cycle.           |
| PUT    | /htmx/tools/cycle/edit   | `htmxhandler.Tools.handleCycleEditPUT`  | Updates a cycle.               |
| DELETE | /htmx/tools/cycle/delete | `htmxhandler.Tools.handleCycleDELETE`   | Deletes a cycle.               |

### Trouble Reports

| Method | Path                                      | Handler                                                  | Description                   |
| ------ | ----------------------------------------- | -------------------------------------------------------- | ----------------------------- |
| GET    | /htmx/trouble-reports/dialog-edit         | `htmxhandler.TroubleReports.handleGetDialogEdit`         | Renders the edit dialog.      |
| POST   | /htmx/trouble-reports/dialog-edit         | `htmxhandler.TroubleReports.handlePostDialogEdit`        | Creates a new trouble report. |
| PUT    | /htmx/trouble-reports/dialog-edit         | `htmxhandler.TroubleReports.handlePutDialogEdit`         | Updates a trouble report.     |
| GET    | /htmx/trouble-reports/data                | `htmxhandler.TroubleReports.handleGetData`               | Fetches trouble report data.  |
| DELETE | /htmx/trouble-reports/data                | `htmxhandler.TroubleReports.handleDeleteData`            | Deletes a trouble report.     |
| GET    | /htmx/trouble-reports/attachments-preview | `htmxhandler.TroubleReports.handleGetAttachmentsPreview` | Fetches attachment previews.  |
| GET    | /htmx/trouble-reports/modifications/:id   | `htmxhandler.TroubleReports.handleGetModifications`      | Fetches modifications.        |
| POST   | /htmx/trouble-reports/modifications/:id   | `htmxhandler.TroubleReports.handlePostModifications`     | Restores a modification.      |
