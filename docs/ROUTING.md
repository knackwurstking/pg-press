# Routing Documentation

This document provides a comprehensive overview of all HTTP routes available in the PG Press application. Routes are organized by feature and include HTTP methods, paths, and descriptions.

## Table of Contents

- [Route Structure](#route-structure)
- [Static Files](#static-files)
- [Authentication Routes](#authentication-routes)
- [Home & Dashboard](#home--dashboard)
- [Tool Management](#tool-management)
  - [Tools Overview](#tools-overview)
  - [HTMX Tool Operations](#htmx-tool-operations)
  - [HTMX Tool Sections](#htmx-tool-sections)
  - [HTMX Admin Operations](#htmx-admin-operations)
- [Individual Tool Pages](#individual-tool-pages)
  - [Tool Details](#tool-details)
  - [HTMX Tool Regenerations](#htmx-tool-regenerations)
  - [HTMX Tool Status Management](#htmx-tool-status-management)
  - [HTMX Tool Sections & Data](#htmx-tool-sections--data)
  - [HTMX Tool Cycle Management](#htmx-tool-cycle-management)
  - [HTMX Tool Binding Operations](#htmx-tool-binding-operations)
- [Press Management](#press-management)
  - [Press Pages](#press-pages)
  - [HTMX Press Sections](#htmx-press-sections)
  - [Press PDF Reports](#press-pdf-reports)
- [Umbau (Tool Change) Operations](#umbau-tool-change-operations)
- [Notes Management](#notes-management)
  - [Notes Pages](#notes-pages)
  - [HTMX Notes Operations](#htmx-notes-operations)
- [Trouble Reports](#trouble-reports)
  - [Trouble Report Pages](#trouble-report-pages)
  - [HTMX Trouble Report Operations](#htmx-trouble-report-operations)
- [Metal Sheets Management](#metal-sheets-management)
  - [HTMX Metal Sheet Operations](#htmx-metal-sheet-operations)
- [Activity Feed](#activity-feed)
  - [Feed Pages](#feed-pages)
  - [HTMX Feed Operations](#htmx-feed-operations)
- [Navigation & WebSocket](#navigation--websocket)
  - [HTMX Navigation](#htmx-navigation)
- [User Profile](#user-profile)
  - [Profile Pages](#profile-pages)
  - [HTMX Profile Operations](#htmx-profile-operations)
- [Editor](#editor)
  - [Editor Pages](#editor-pages)
  - [Editor Operations](#editor-operations)
- [Help & Documentation](#help--documentation)
  - [Help Pages](#help-pages)
- [WebSocket Endpoints](#websocket-endpoints)
  - [Real-time Updates](#real-time-updates)
- [Route Patterns](#route-patterns)
  - [Parameter Patterns](#parameter-patterns)
  - [Common Query Parameters](#common-query-parameters)
- [Authentication & Authorization](#authentication--authorization)
- [Content Types](#content-types)
- [Error Handling](#error-handling)

## Route Structure

All routes are prefixed with the `SERVER_PATH_PREFIX` environment variable (default: `/pg-press`).

Routes ending without a trailing slash automatically support both variants (with and without trailing slash).

## Static Files

| Method | Path        | Description                                       |
| ------ | ----------- | ------------------------------------------------- |
| GET    | `/assets/*` | Static assets (CSS, JS, images, icons, manifests) |

## Authentication Routes

| Method | Path      | Description                 |
| ------ | --------- | --------------------------- |
| GET    | `/login`  | Login page                  |
| GET    | `/logout` | Logout handler and redirect |

## Home & Dashboard

| Method | Path | Description              |
| ------ | ---- | ------------------------ |
| GET    | `/`  | Main dashboard/home page |

## Tool Management

### Tools Overview

| Method | Path     | Description         |
| ------ | -------- | ------------------- |
| GET    | `/tools` | Tools overview page |

### HTMX Tool Operations

| Method | Path                    | Description          |
| ------ | ----------------------- | -------------------- |
| GET    | `/htmx/tools/edit`      | Get tool edit dialog |
| POST   | `/htmx/tools/edit`      | Create new tool      |
| PUT    | `/htmx/tools/edit`      | Update existing tool |
| DELETE | `/htmx/tools/delete`    | Delete tool          |
| PATCH  | `/htmx/tools/mark-dead` | Mark tool as dead    |

### HTMX Tool Sections

| Method | Path                        | Description               |
| ------ | --------------------------- | ------------------------- |
| GET    | `/htmx/tools/section/press` | Get press section content |
| GET    | `/htmx/tools/section/tools` | Get tools section content |

### HTMX Admin Operations

| Method | Path                                  | Description                      |
| ------ | ------------------------------------- | -------------------------------- |
| GET    | `/htmx/tools/admin/overlapping-tools` | Get overlapping tools admin view |

## Individual Tool Pages

### Tool Details

| Method | Path              | Description                 |
| ------ | ----------------- | --------------------------- |
| GET    | `/tools/tool/:id` | Individual tool detail page |

### HTMX Tool Regenerations

| Method | Path                                       | Description                  |
| ------ | ------------------------------------------ | ---------------------------- |
| GET    | `/htmx/tools/tool/:id/edit-regeneration`   | Get regeneration edit dialog |
| PUT    | `/htmx/tools/tool/:id/edit-regeneration`   | Update tool regeneration     |
| DELETE | `/htmx/tools/tool/:id/delete-regeneration` | Delete tool regeneration     |

### HTMX Tool Status Management

| Method | Path                         | Description          |
| ------ | ---------------------------- | -------------------- |
| GET    | `/htmx/tools/status-edit`    | Get status edit form |
| GET    | `/htmx/tools/status-display` | Get status display   |
| PUT    | `/htmx/tools/status`         | Update tool status   |

### HTMX Tool Sections & Data

| Method | Path                       | Description                   |
| ------ | -------------------------- | ----------------------------- |
| GET    | `/htmx/tools/notes`        | Get tool notes section        |
| GET    | `/htmx/tools/metal-sheets` | Get tool metal sheets section |
| GET    | `/htmx/tools/cycles`       | Get tool cycles table         |
| GET    | `/htmx/tools/total-cycles` | Get tool total cycles count   |

### HTMX Tool Cycle Management

| Method | Path                       | Description            |
| ------ | -------------------------- | ---------------------- |
| GET    | `/htmx/tools/cycle/edit`   | Get cycle edit dialog  |
| POST   | `/htmx/tools/cycle/edit`   | Create new cycle entry |
| PUT    | `/htmx/tools/cycle/edit`   | Update cycle entry     |
| DELETE | `/htmx/tools/cycle/delete` | Delete cycle entry     |

### HTMX Tool Binding Operations

| Method | Path                          | Description            |
| ------ | ----------------------------- | ---------------------- |
| PATCH  | `/htmx/tools/tool/:id/bind`   | Bind tool to press     |
| PATCH  | `/htmx/tools/tool/:id/unbind` | Unbind tool from press |

## Press Management

### Press Pages

| Method | Path                  | Description                    |
| ------ | --------------------- | ------------------------------ |
| GET    | `/tools/press/:press` | Individual press overview page |

### HTMX Press Sections

| Method | Path                                    | Description                |
| ------ | --------------------------------------- | -------------------------- |
| GET    | `/htmx/tools/press/:press/active-tools` | Get active tools for press |
| GET    | `/htmx/tools/press/:press/metal-sheets` | Get metal sheets for press |
| GET    | `/htmx/tools/press/:press/cycles`       | Get cycles for press       |
| GET    | `/htmx/tools/press/:press/notes`        | Get notes for press        |

### Press PDF Reports

| Method | Path                                         | Description                |
| ------ | -------------------------------------------- | -------------------------- |
| GET    | `/htmx/tools/press/:press/cycle-summary-pdf` | Generate cycle summary PDF |

## Umbau (Tool Change) Operations

| Method | Path                        | Description                |
| ------ | --------------------------- | -------------------------- |
| GET    | `/tools/press/:press/umbau` | Tool change page for press |
| POST   | `/tools/press/:press/umbau` | Process tool change        |

## Notes Management

### Notes Pages

| Method | Path     | Description         |
| ------ | -------- | ------------------- |
| GET    | `/notes` | Notes overview page |

### HTMX Notes Operations

| Method | Path                 | Description          |
| ------ | -------------------- | -------------------- |
| GET    | `/htmx/notes/edit`   | Get note edit dialog |
| POST   | `/htmx/notes/edit`   | Create new note      |
| PUT    | `/htmx/notes/edit`   | Update existing note |
| DELETE | `/htmx/notes/delete` | Delete note          |

## Trouble Reports

### Trouble Report Pages

| Method | Path                                 | Description                         |
| ------ | ------------------------------------ | ----------------------------------- |
| GET    | `/trouble-reports`                   | Trouble reports overview page       |
| GET    | `/trouble-reports/share-pdf`         | Share trouble report as PDF         |
| GET    | `/trouble-reports/attachment`        | Get trouble report attachment       |
| GET    | `/trouble-reports/modifications/:id` | Get modification history for report |

### HTMX Trouble Report Operations

| Method | Path                                        | Description                     |
| ------ | ------------------------------------------- | ------------------------------- |
| GET    | `/htmx/trouble-reports/data`                | Get trouble reports data        |
| DELETE | `/htmx/trouble-reports/data`                | Delete trouble report           |
| GET    | `/htmx/trouble-reports/attachments-preview` | Get attachments preview         |
| POST   | `/htmx/trouble-reports/rollback`            | Rollback trouble report changes |

## Metal Sheets Management

### HTMX Metal Sheet Operations

| Method | Path                        | Description                 |
| ------ | --------------------------- | --------------------------- |
| GET    | `/htmx/metal-sheets/edit`   | Get metal sheet edit dialog |
| POST   | `/htmx/metal-sheets/edit`   | Create new metal sheet      |
| PUT    | `/htmx/metal-sheets/edit`   | Update existing metal sheet |
| DELETE | `/htmx/metal-sheets/delete` | Delete metal sheet          |

## Activity Feed

### Feed Pages

| Method | Path    | Description        |
| ------ | ------- | ------------------ |
| GET    | `/feed` | Activity feed page |

### HTMX Feed Operations

| Method | Path              | Description            |
| ------ | ----------------- | ---------------------- |
| GET    | `/htmx/feed/list` | Get activity feed list |

## Navigation & WebSocket

### HTMX Navigation

| Method | Path                     | Description                                           |
| ------ | ------------------------ | ----------------------------------------------------- |
| GET    | `/htmx/nav/feed-counter` | WebSocket endpoint for real-time feed counter updates |

## User Profile

### Profile Pages

| Method | Path       | Description       |
| ------ | ---------- | ----------------- |
| GET    | `/profile` | User profile page |

### HTMX Profile Operations

| Method | Path                    | Description           |
| ------ | ----------------------- | --------------------- |
| GET    | `/htmx/profile/cookies` | Get user cookies data |
| DELETE | `/htmx/profile/cookies` | Delete user cookies   |

## Editor

### Editor Pages

| Method | Path      | Description         |
| ------ | --------- | ------------------- |
| GET    | `/editor` | Content editor page |

### Editor Operations

| Method | Path           | Description         |
| ------ | -------------- | ------------------- |
| POST   | `/editor/save` | Save editor content |

## Help & Documentation

### Help Pages

| Method | Path             | Description        |
| ------ | ---------------- | ------------------ |
| GET    | `/help/markdown` | Markdown help page |

## WebSocket Endpoints

### Real-time Updates

| Endpoint                 | Protocol  | Description                             |
| ------------------------ | --------- | --------------------------------------- |
| `/htmx/nav/feed-counter` | WebSocket | Real-time activity feed counter updates |

## Route Patterns

### Parameter Patterns

- `:id` - Numeric tool/entity ID
- `:press` - Press number (0-5)

### Common Query Parameters

- Various filters and pagination parameters depending on the endpoint
- Form data for POST/PUT operations
- File uploads for attachment handling

## Authentication & Authorization

Most routes require authentication via cookie-based sessions or API keys. Static assets and the login page are publicly accessible.

## Content Types

- **HTML Pages**: Return full HTML pages with navigation
- **HTMX Endpoints**: Return HTML fragments for dynamic page updates
- **PDF Endpoints**: Return PDF documents with appropriate headers
- **WebSocket**: Real-time bidirectional communication
- **File Downloads**: Return file attachments with appropriate MIME types

## Error Handling

All routes implement comprehensive error handling with:

- Proper HTTP status codes
- User-friendly error messages
- Logging for debugging
- Graceful fallbacks where appropriate
