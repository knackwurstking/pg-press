# HTMX API Documentation

## Overview

PG Press is built on an HTMX-based architecture that provides dynamic, server-rendered web interactions without traditional REST APIs. Instead of JSON APIs, the application uses HTMX endpoints that return HTML fragments for seamless page updates.

## Architecture

### Technology Stack

- **Backend**: Go with Echo web framework
- **Frontend**: HTMX for dynamic interactions + vanilla JavaScript
- **Templates**: Templ for type-safe HTML generation
- **Real-time**: WebSocket for live updates
- **Authentication**: Cookie-based sessions with API key support

### Request/Response Pattern

Unlike REST APIs, HTMX endpoints:

- Accept form data or URL parameters
- Return HTML fragments (not JSON)
- Update specific page sections dynamically
- Support standard HTTP methods (GET, POST, PUT, DELETE)

## Authentication

### Session-Based Authentication

All routes (except `/login`) require user authentication via session cookies.

**Login Process:**

```
POST /login
Content-Type: application/x-www-form-urlencoded

api-key=your-api-key-here
```

**Response:** Redirects to profile page on success, stays on login page with error on failure.

**Logout:**

```
GET /logout
```

**Response:** Redirects to login page and clears session.

## Page Routes

### Authentication & User Management

#### Login Page

```
GET /login
```

Renders the login form. Accepts `api-key` form parameter for authentication.

#### Logout

```
GET /logout
```

Clears user session and redirects to login page.

### Main Application Pages

#### Home Dashboard

```
GET /
GET /home
```

Main dashboard showing system overview and recent activity.

#### User Profile

```
GET /profile
```

User profile page with session management and user information.

#### Activity Feed

```
GET /feed
```

Chronological activity feed showing system events and user actions.

### Tools & Press Management

#### Tools Overview

```
GET /tools
```

Main tools management page with press sections and tool listings.

#### Individual Tool Page

```
GET /tools/tool/{id}
```

Detailed view for a specific tool with cycles, notes, and management options.

#### Press-Specific View

```
GET /tools/press/{press_number}
```

View all tools and activity for a specific press (0-5).

### Trouble Reports

#### Trouble Reports List

```
GET /trouble-reports
```

List all trouble reports with search and filtering capabilities.

#### PDF Export

```
GET /trouble-reports/share-pdf?id={report_id}
```

Generate and download PDF version of trouble report.

#### Attachment Download

```
GET /trouble-reports/attachment?id={attachment_id}
```

Download file attachment from trouble report.

#### Modifications History

```
GET /trouble-reports/modifications/{id}
```

View modification history for a specific trouble report.

### Content Editor

#### Editor Page

```
GET /editor?type={content_type}&id={id}&return_url={url}
```

**Parameters:**

- `type` (required): Content type (`troublereport`)
- `id` (optional): ID of existing item to edit
- `return_url` (optional): URL to redirect after save

**Save Content:**

```
POST /editor/save
Content-Type: multipart/form-data

type=troublereport&content=...&title=...&attachments=...
```

### Notes Management

#### Notes Page

```
GET /notes
```

Comprehensive notes management with filtering and search.

### Metal Sheets

#### Metal Sheets Management

```
GET /metal-sheets
```

Metal sheet inventory and tool assignment management.

## HTMX Endpoints

HTMX endpoints return HTML fragments for dynamic page updates.

### Feed Management

#### Feed List

```
GET /htmx/feed/list?limit={limit}&offset={offset}
```

Returns HTML fragment with feed entries for infinite scroll or pagination.

### Navigation

#### Feed Counter WebSocket

```
WebSocket: /ws/feed-counter
```

Real-time feed counter updates.

**Message Format:**

```json
{
  "type": "counter_update",
  "count": 5,
  "timestamp": 1699123456789
}
```

### Profile Management

#### Get User Cookies

```
GET /htmx/profile/cookies
```

Returns HTML table of user's active sessions.

#### Delete User Cookie

```
DELETE /htmx/profile/cookies?value={cookie_value}
```

Deletes a specific user session.

### Tools Management

#### Tools List by Press

```
GET /htmx/tools/section/tools?press={press_number}
```

Returns HTML fragment with tools for specified press.

#### Press Section

```
GET /htmx/tools/section/press?press={press_number}
```

Returns HTML fragment for press section with tools and controls.

#### Tool Edit Dialog

```
GET /htmx/tools/edit?id={id}
POST /htmx/tools/edit
PUT /htmx/tools/edit?id={id}
```

CRUD operations for tools, returns HTML dialogs and updated content.

#### Delete Tool

```
DELETE /htmx/tools/delete?id={id}
```

Deletes tool and returns updated tools section.

### Press Cycles Management

#### Cycles Section

```
GET /htmx/tools/tool/{id}/cycles
```

Returns HTML fragment with cycle data for specific tool.

#### Total Cycles Calculator

```
GET /htmx/tools/tool/{id}/total-cycles-calc?press={press}&total={total}
```

Returns HTML fragment with cycle calculation validation.

#### Cycle CRUD Operations

```
GET /htmx/tools/tool/{id}/cycle/edit?cycle_id={id}
POST /htmx/tools/tool/{id}/cycle/edit
PUT /htmx/tools/tool/{id}/cycle/edit?cycle_id={id}
DELETE /htmx/tools/tool/{id}/cycle/delete?cycle_id={id}
```

Full CRUD operations for press cycles.

### Trouble Reports Management

#### Reports Data

```
GET /htmx/trouble-reports/data?search={query}&limit={limit}&offset={offset}
```

Returns HTML fragment with filtered/paginated trouble reports.

#### Delete Trouble Report

```
DELETE /htmx/trouble-reports/data?id={id}
```

Deletes trouble report and returns updated list.

#### Attachments Preview

```
GET /htmx/trouble-reports/attachments-preview?id={id}
```

Returns HTML preview of report attachments.

#### Restore Modification

```
POST /htmx/trouble-reports/rollback
Content-Type: application/x-www-form-urlencoded

id={report_id}&mod_index={index}
```

Restores trouble report to previous version.

### Notes Management

#### Note CRUD Operations

```
GET /htmx/notes/edit?id={id}&link_to_tables={entity}
POST /htmx/notes/edit?link_to_tables={entity}
PUT /htmx/notes/edit?id={id}&link_to_tables={entity}
DELETE /htmx/notes/delete?id={id}
```

**Link Formats:**

- `tool_{id}` - Links to specific tool
- `press_{number}` - Links to specific press
- `{type}_{id}` - Links to any entity type

#### Notes by Tool

```
GET /htmx/tools/notes?tool_id={id}
```

Returns HTML fragment with notes for specific tool.

## WebSocket Endpoints

### Feed Counter Updates

**Endpoint:** `/ws/feed-counter`

**Connection:** Standard WebSocket connection

**Message Types:**

**Counter Update:**

```json
{
  "type": "counter_update",
  "count": 12,
  "timestamp": 1699123456789
}
```

**Error Message:**

```json
{
  "type": "error",
  "message": "Connection error description"
}
```

## Error Handling

### Standard Error Responses

HTMX endpoints return appropriate HTTP status codes with HTML error content:

#### 400 Bad Request

```html
<div class="error">Invalid request parameters</div>
```

#### 401 Unauthorized

```html
<div class="error">Authentication required</div>
```

#### 404 Not Found

```html
<div class="error">Resource not found</div>
```

#### 500 Internal Server Error

```html
<div class="error">Internal server error</div>
```

## Request Patterns

### Form Submissions

Most HTMX endpoints expect form data:

```html
<form hx-post="/htmx/tools/edit" hx-target="#tools-section">
  <input name="position" value="top" />
  <input name="code" value="T001" />
  <input name="type" value="punch" />
  <button type="submit">Save</button>
</form>
```

### HTMX Triggers

Common HTMX response headers for page updates:

```http
HX-Trigger: toolCreated, pageLoaded
HX-Trigger: toolUpdated, pageLoaded
HX-Trigger: toolDeleted, pageLoaded
```

### File Uploads

File uploads use multipart forms:

```html
<form hx-post="/editor/save" hx-encoding="multipart/form-data" hx-target="body">
  <input type="file" name="attachments" multiple />
  <textarea name="content"></textarea>
  <button type="submit">Save</button>
</form>
```

## Security

### Authentication Requirements

- All non-login routes require valid session
- Session established via API key authentication
- Session cookies are HTTP-only and secure

### Input Validation

- All form inputs are validated server-side
- File uploads restricted by type and size
- SQL injection protection through parameterized queries

### CSRF Protection

- Form-based endpoints include CSRF protection
- Session-based validation for all state-changing operations

## Performance Considerations

### Caching Strategy

- Static assets cached for 1 year (CSS, JS, fonts)
- Images cached for 30 days
- HTML responses not cached for dynamic content

### HTMX Optimizations

- Partial page updates reduce bandwidth
- Client-side state management minimal
- Server-side rendering for all dynamic content

## Development and Testing

### Local Development

```bash
make dev  # Starts server with hot reloading
```

**Default URL:** `http://localhost:8080`

### Testing HTMX Endpoints

Use curl with appropriate headers:

```bash
# Test HTMX endpoint
curl -H "HX-Request: true" \
     -H "Cookie: session=your-session-cookie" \
     http://localhost:8080/htmx/tools/section/tools?press=1

# Test form submission
curl -X POST \
     -H "Content-Type: application/x-www-form-urlencoded" \
     -H "Cookie: session=your-session-cookie" \
     -d "position=top&code=T001&type=punch" \
     http://localhost:8080/htmx/tools/edit
```

## Migration from Traditional APIs

If integrating with external systems that expect JSON APIs:

1. **Consider adding API endpoints** alongside HTMX endpoints
2. **Reuse business logic** from existing handlers
3. **Maintain authentication** system compatibility
4. **Use content negotiation** to serve JSON vs HTML based on Accept header

## Best Practices

### HTMX Development

1. **Target specific elements** for updates rather than full page reloads
2. **Use appropriate HTTP methods** (GET for reads, POST for creates, etc.)
3. **Include CSRF tokens** in forms
4. **Handle loading states** with HTMX indicators
5. **Validate inputs** both client and server-side

### Error Handling

1. **Return appropriate HTTP status codes**
2. **Provide user-friendly error messages**
3. **Log errors** for debugging
4. **Handle network failures** gracefully

### Performance

1. **Minimize HTML fragment size** for HTMX responses
2. **Cache static assets** aggressively
3. **Use WebSockets** for real-time updates
4. **Optimize database queries** for frequently accessed data

This documentation reflects the actual HTMX-based architecture of PG Press and should be used as the primary reference for understanding the application's API patterns and endpoints.
