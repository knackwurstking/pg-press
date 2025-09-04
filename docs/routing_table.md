# Routing Table

This document outlines the routing structure of the web application, detailing the available endpoints, their purposes, and the corresponding handlers.

## General Routes

| Method | Path                        | Handler               | Description                                     |
| ------ | --------------------------- | --------------------- | ----------------------------------------------- |
| GET    | /                           | `html.Home`           | Serves the home page.                           |
| GET    | /login                      | `html.Auth`           | Serves the login page.                          |
| GET    | /logout                     | `html.Auth`           | Logs the user out.                              |
| GET    | /profile                    | `html.Profile`        | Serves the user profile page.                   |
| GET    | /feed                       | `html.Feed`           | Serves the feed page.                           |
| GET    | /tools                      | `html.Tools`          | Serves the tools page.                          |
| GET    | /tools/press/:press         | `html.Tools`          | Serves the tools page for a specific press.     |
| GET    | /tools/tool/:id             | `html.Tools`          | Serves the page for a specific tool.            |
| GET    | /trouble-reports            | `html.TroubleReports` | Serves the trouble reports page.                |
| GET    | /trouble-reports/share-pdf  | `html.TroubleReports` | Generates and serves a PDF of a trouble report. |
| GET    | /trouble-reports/attachment | `html.TroubleReports` | Serves a trouble report attachment.             |

---

## HTMX Routes

### Feed

| Method | Path            | Handler     | Description                                 |
| ------ | --------------- | ----------- | ------------------------------------------- |
| GET    | /htmx/feed/list | `htmx.Feed` | Fetches and renders the list of feed items. |

### Nav

| Method | Path                   | Handler    | Description                                              |
| ------ | ---------------------- | ---------- | -------------------------------------------------------- |
| GET    | /htmx/nav/feed-counter | `htmx.Nav` | Establishes a WebSocket connection for the feed counter. |

### Profile

| Method | Path                  | Handler        | Description                             |
| ------ | --------------------- | -------------- | --------------------------------------- |
| GET    | /htmx/profile/cookies | `htmx.Profile` | Fetches and renders the user's cookies. |
| DELETE | /htmx/profile/cookies | `htmx.Profile` | Deletes a user's cookie.                |

### Tools

| Method | Path                     | Handler      | Description                                      |
| ------ | ------------------------ | ------------ | ------------------------------------------------ |
| GET    | /htmx/tools/list         | `htmx.Tools` | Fetches and renders the list of all tools.       |
| GET    | /htmx/tools/edit         | `htmx.Tools` | Renders the tool edit dialog.                    |
| POST   | /htmx/tools/edit         | `htmx.Tools` | Creates a new tool.                              |
| PUT    | /htmx/tools/edit         | `htmx.Tools` | Updates an existing tool.                        |
| DELETE | /htmx/tools/delete       | `htmx.Tools` | Deletes a tool.                                  |
| GET    | /htmx/tools/cycles       | `htmx.Tools` | Fetches and renders the cycles for a tool.       |
| GET    | /htmx/tools/total-cycles | `htmx.Tools` | Fetches and renders the total cycles for a tool. |
| GET    | /htmx/tools/cycle/edit   | `htmx.Tools` | Renders the cycle edit dialog.                   |
| POST   | /htmx/tools/cycle/edit   | `htmx.Tools` | Creates a new cycle.                             |
| PUT    | /htmx/tools/cycle/edit   | `htmx.Tools` | Updates an existing cycle.                       |
| DELETE | /htmx/tools/cycle/delete | `htmx.Tools` | Deletes a cycle.                                 |

### Trouble Reports

| Method | Path                                      | Handler               | Description                                                 |
| ------ | ----------------------------------------- | --------------------- | ----------------------------------------------------------- |
| GET    | /htmx/trouble-reports/dialog-edit         | `htmx.TroubleReports` | Renders the trouble report edit dialog.                     |
| POST   | /htmx/trouble-reports/dialog-edit         | `htmx.TroubleReports` | Creates a new trouble report.                               |
| PUT    | /htmx/trouble-reports/dialog-edit         | `htmx.TroubleReports` | Updates an existing trouble report.                         |
| GET    | /htmx/trouble-reports/data                | `htmx.TroubleReports` | Fetches and renders the list of trouble reports.            |
| DELETE | /htmx/trouble-reports/data                | `htmx.TroubleReports` | Deletes a trouble report.                                   |
| GET    | /htmx/trouble-reports/attachments-preview | `htmx.TroubleReports` | Renders the attachments preview for a trouble report.       |
| GET    | /htmx/trouble-reports/modifications/:id   | `htmx.TroubleReports` | Fetches and renders the modifications for a trouble report. |
| POST   | /htmx/trouble-reports/modifications/:id   | `htmx.TroubleReports` | Restores a modification for a trouble report.               |
