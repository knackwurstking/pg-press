# API Routing Documentation

This document provides comprehensive documentation for all API endpoints and routes in the PG Press application, including request/response formats, authentication requirements, and usage examples.

## Overview

PG Press uses a hybrid routing approach combining:

- **Page Routes**: Traditional HTTP routes serving HTML pages
- **HTMX Routes**: Dynamic content endpoints for interactive UI updates
- **WebSocket Routes**: Real-time communication endpoints
- **Asset Routes**: Static file serving with caching

All routes support the `SERVER_PATH_PREFIX` environment variable for deployment under a subpath.

## Authentication

### Authentication Methods

1. **Session Cookies**: Primary method for web interface
2. **API Keys**: For programmatic access via `X-API-Key` header
3. **WebSocket Authentication**: Via cookie or query parameter

### Authentication Headers

```http
# API Key Authentication
X-API-Key: your-api-key-here

# Session Cookie (automatically handled by browser)
Cookie: session=cookie-value-here
```

### Error Responses

```http
HTTP/1.1 401 Unauthorized
Content-Type: application/json

{
  "error": "Authentication required",
  "code": "AUTH_REQUIRED"
}
```

## Page Routes

Traditional HTTP routes that serve complete HTML pages.

### Authentication & User Management

#### Login Page

```http
GET /login
```

**Description**: Renders the user login page.

**Response**: HTML login form

**Status Codes**:

- `200 OK`: Login page rendered successfully
- `302 Found`: Redirect if already authenticated

---

#### Logout

```http
GET /logout
```

**Description**: Logs the user out and clears session.

**Response**: Redirect to login page

**Status Codes**:

- `302 Found`: Redirect to login after logout
- `401 Unauthorized`: Not authenticated

---

### Main Application Pages

#### Home Dashboard

```http
GET /
```

**Description**: Main dashboard with system overview and quick actions.

**Authentication**: Required

**Response**: HTML dashboard page with:

- System status overview
- Recent activity feed
- Quick action buttons
- Navigation menu

**Status Codes**:

- `200 OK`: Dashboard rendered successfully
- `401 Unauthorized`: Not authenticated
- `302 Found`: Redirect to login if not authenticated

---

#### Activity Feed

```http
GET /feed
```

**Description**: Displays comprehensive activity feed with real-time updates.

**Authentication**: Required

**Response**: HTML page with:

- Paginated activity feed
- WebSocket connection for live updates
- Filter controls
- Feed item details

**Status Codes**:

- `200 OK`: Feed page rendered successfully
- `401 Unauthorized`: Not authenticated

---

#### User Profile

```http
GET /profile
```

**Description**: User profile management page.

**Authentication**: Required

**Response**: HTML page with:

- User information display
- Session management
- Cookie management interface
- API key information

**Status Codes**:

- `200 OK`: Profile page rendered successfully
- `401 Unauthorized`: Not authenticated

---

### Tools & Press Management

#### Tools Overview

```http
GET /tools
```

**Description**: Main tools management page with comprehensive tool listing.

**Authentication**: Required

**Response**: HTML page with:

- Tools list with filtering and sorting
- Press assignment overview
- Quick action buttons
- Tool creation form

**Status Codes**:

- `200 OK`: Tools page rendered successfully
- `401 Unauthorized`: Not authenticated

---

#### Individual Press Page

```http
GET /tools/press/:press
```

**Path Parameters**:

- `press` (integer): Press number (0-5)

**Description**: Detailed view of a specific press with assigned tools.

**Authentication**: Required

**Response**: HTML page with:

- Press-specific tool assignments
- Press performance metrics
- Press cycle history
- Press-specific notes

**Status Codes**:

- `200 OK`: Press page rendered successfully
- `404 Not Found`: Invalid press number
- `401 Unauthorized`: Not authenticated

**Example**:

```http
GET /tools/press/5
```

---

#### Individual Tool Page

```http
GET /tools/tool/:id
```

**Path Parameters**:

- `id` (integer): Tool ID

**Description**: Detailed view of a specific tool with complete information.

**Authentication**: Required

**Response**: HTML page with:

- Complete tool specifications
- Cycle history and performance
- Linked notes and documentation
- Regeneration history
- Metal sheet assignments

**Status Codes**:

- `200 OK`: Tool page rendered successfully
- `404 Not Found`: Tool not found
- `401 Unauthorized`: Not authenticated

**Example**:

```http
GET /tools/tool/123
```

---

### Trouble Reports

#### Trouble Reports List

```http
GET /trouble-reports
```

**Description**: Main trouble reports management page.

**Authentication**: Required

**Response**: HTML page with:

- Paginated reports list
- Search and filter controls
- Report creation form
- Status indicators

**Status Codes**:

- `200 OK`: Reports page rendered successfully
- `401 Unauthorized`: Not authenticated

---

#### PDF Report Export

```http
GET /trouble-reports/share-pdf
```

**Query Parameters**:

- `id` (integer, required): Trouble report ID

**Description**: Generates and downloads PDF version of trouble report.

**Authentication**: Required

**Response**: PDF file download

**Status Codes**:

- `200 OK`: PDF generated and served
- `404 Not Found`: Report not found
- `401 Unauthorized`: Not authenticated
- `500 Internal Server Error`: PDF generation failed

**Example**:

```http
GET /trouble-reports/share-pdf?id=456
```

---

#### Attachment Download

```http
GET /trouble-reports/attachment
```

**Query Parameters**:

- `attachment_id` (integer, required): Attachment ID

**Description**: Serves trouble report attachment file.

**Authentication**: Required

**Response**: Binary file data with appropriate Content-Type

**Status Codes**:

- `200 OK`: File served successfully
- `404 Not Found`: Attachment not found
- `401 Unauthorized`: Not authenticated

**Example**:

```http
GET /trouble-reports/attachment?attachment_id=789
```

## HTMX Routes

Dynamic content endpoints for interactive UI updates using HTMX.

### Feed Management

#### Feed List

```http
GET /htmx/feed/list
```

**Description**: Fetches paginated activity feed entries.

**Authentication**: Required

**Query Parameters**:

- `page` (integer, optional): Page number (default: 1)
- `limit` (integer, optional): Items per page (default: 20)
- `filter` (string, optional): Filter by event type

**Response**: HTML fragment with feed entries

**Status Codes**:

- `200 OK`: Feed entries returned
- `401 Unauthorized`: Not authenticated

**Response Example**:

```html
<div class="feed-entries">
  <div class="feed-entry" data-id="123">
    <span class="timestamp">2023-12-01 10:30:00</span>
    <span class="event">Tool Created</span>
    <span class="details">Tool ABC-123 added to Press 3</span>
  </div>
  <!-- More entries... -->
</div>
```

---

### Navigation

#### Feed Counter WebSocket

```http
GET /htmx/nav/feed-counter
```

**Description**: WebSocket endpoint for real-time feed counter updates.

**Authentication**: Required

**Protocol**: WebSocket upgrade

**Messages**:

- Server sends counter updates as JSON
- Client sends heartbeat messages

**Message Format**:

```json
{
  "type": "counter_update",
  "count": 15,
  "timestamp": "2023-12-01T10:30:00Z"
}
```

---

### Profile Management

#### Get User Cookies

```http
GET /htmx/profile/cookies
```

**Description**: Fetches current user's active session cookies.

**Authentication**: Required

**Response**: HTML table fragment with cookie information

**Status Codes**:

- `200 OK`: Cookies list returned
- `401 Unauthorized`: Not authenticated

**Response Example**:

```html
<table class="cookies-table">
  <tr>
    <td>Chrome on Windows</td>
    <td>2023-12-01 10:30:00</td>
    <td>
      <button hx-delete="/htmx/profile/cookies?value=abc123">Delete</button>
    </td>
  </tr>
</table>
```

---

#### Delete User Cookie

```http
DELETE /htmx/profile/cookies
```

**Description**: Deletes a specific user session cookie.

**Authentication**: Required

**Query Parameters**:

- `value` (string, required): Cookie value to delete

**Response**: Success message or updated cookies list

**Status Codes**:

- `200 OK`: Cookie deleted successfully
- `404 Not Found`: Cookie not found
- `401 Unauthorized`: Not authenticated

---

### Tools Management

#### Tools List

```http
GET /htmx/tools/list
```

**Description**: Fetches comprehensive tools list with filtering and sorting.

**Authentication**: Required

**Query Parameters**:

- `press` (integer, optional): Filter by press number
- `status` (string, optional): Filter by status (active, regenerating)
- `sort` (string, optional): Sort field (code, press, type)
- `order` (string, optional): Sort order (asc, desc)

**Response**: HTML fragment with tools list

**Status Codes**:

- `200 OK`: Tools list returned
- `401 Unauthorized`: Not authenticated

---

#### Tool Edit Dialog

```http
GET /htmx/tools/edit
```

**Description**: Renders tool creation or editing dialog.

**Authentication**: Required

**Query Parameters**:

- `id` (integer, optional): Tool ID for editing
- `close` (boolean, optional): Close dialog flag

**Response**: HTML dialog fragment

**Status Codes**:

- `200 OK`: Dialog rendered
- `404 Not Found`: Tool not found (when editing)
- `401 Unauthorized`: Not authenticated

---

#### Create Tool

```http
POST /htmx/tools/edit
```

**Description**: Creates a new tool.

**Authentication**: Required

**Request Body** (form-encoded):

```
position=top
type=cutting
code=ABC-123
press=5
format[height]=100
format[width]=200
```

**Response**: Success message and updated tools list

**Status Codes**:

- `201 Created`: Tool created successfully
- `400 Bad Request`: Validation errors
- `409 Conflict`: Tool code already exists
- `401 Unauthorized`: Not authenticated

---

#### Update Tool

```http
PUT /htmx/tools/edit
```

**Description**: Updates an existing tool.

**Authentication**: Required

**Query Parameters**:

- `id` (integer, required): Tool ID

**Request Body** (form-encoded): Same as create

**Response**: Success message and updated tool information

**Status Codes**:

- `200 OK`: Tool updated successfully
- `400 Bad Request`: Validation errors
- `404 Not Found`: Tool not found
- `409 Conflict`: Tool code conflict
- `401 Unauthorized`: Not authenticated

---

#### Delete Tool

```http
DELETE /htmx/tools/delete
```

**Description**: Deletes a tool and all related data.

**Authentication**: Required

**Query Parameters**:

- `id` (integer, required): Tool ID

**Response**: Success message and updated tools list

**Status Codes**:

- `200 OK`: Tool deleted successfully
- `404 Not Found`: Tool not found
- `409 Conflict`: Tool cannot be deleted (has dependencies)
- `401 Unauthorized`: Not authenticated

---

### Press Cycles Management

#### Cycles Section

```http
GET /htmx/tools/cycles
```

**Description**: Fetches cycles section for a specific tool.

**Authentication**: Required

**Query Parameters**:

- `tool_id` (integer, required): Tool ID

**Response**: HTML fragment with cycles information

**Status Codes**:

- `200 OK`: Cycles section returned
- `404 Not Found`: Tool not found
- `401 Unauthorized`: Not authenticated

---

#### Total Cycles Calculator

```http
GET /htmx/tools/total-cycles
```

**Description**: Calculates total cycles for a tool with optional input validation.

**Authentication**: Required

**Query Parameters**:

- `tool_id` (integer, required): Tool ID
- `input` (integer, optional): New cycle count for validation

**Response**: JSON with total cycles information

**Status Codes**:

- `200 OK`: Calculation successful
- `404 Not Found`: Tool not found
- `401 Unauthorized`: Not authenticated

**Response Example**:

```json
{
  "total_cycles": 15420,
  "last_update": "2023-12-01T10:30:00Z",
  "validation": {
    "input": 150,
    "new_total": 15570,
    "valid": true
  }
}
```

---

#### Cycle Edit Dialog

```http
GET /htmx/tools/cycle/edit
```

**Description**: Renders cycle creation or editing dialog.

**Authentication**: Required

**Query Parameters**:

- `tool_id` (integer, required): Tool ID
- `cycle_id` (integer, optional): Cycle ID for editing
- `close` (boolean, optional): Close dialog flag

**Response**: HTML dialog fragment

**Status Codes**:

- `200 OK`: Dialog rendered
- `404 Not Found`: Tool or cycle not found
- `401 Unauthorized`: Not authenticated

---

#### Create Cycle Record

```http
POST /htmx/tools/cycle/edit
```

**Description**: Creates a new cycle record for a tool.

**Authentication**: Required

**Query Parameters**:

- `tool_id` (integer, required): Tool ID

**Request Body** (form-encoded):

```
press_number=5
total_cycles=150
date=2023-12-01T10:30:00Z
notes=Regular maintenance cycle
```

**Response**: Success message and updated cycles list

**Status Codes**:

- `201 Created`: Cycle created successfully
- `400 Bad Request`: Validation errors
- `404 Not Found`: Tool not found
- `401 Unauthorized`: Not authenticated

---

#### Update Cycle Record

```http
PUT /htmx/tools/cycle/edit
```

**Description**: Updates an existing cycle record.

**Authentication**: Required

**Query Parameters**:

- `tool_id` (integer, required): Tool ID
- `cycle_id` (integer, required): Cycle ID

**Request Body** (form-encoded): Same as create

**Response**: Success message and updated cycle information

**Status Codes**:

- `200 OK`: Cycle updated successfully
- `400 Bad Request`: Validation errors
- `404 Not Found`: Tool or cycle not found
- `401 Unauthorized`: Not authenticated

---

#### Delete Cycle Record

```http
DELETE /htmx/tools/cycle/delete
```

**Description**: Deletes a cycle record.

**Authentication**: Required

**Query Parameters**:

- `tool_id` (integer, required): Tool ID
- `cycle_id` (integer, required): Cycle ID

**Response**: Success message and updated cycles list

**Status Codes**:

- `200 OK`: Cycle deleted successfully
- `404 Not Found`: Tool or cycle not found
- `401 Unauthorized`: Not authenticated

---

### Metal Sheets Management

#### Metal Sheet Edit Dialog

```http
GET /htmx/metal-sheets/edit
```

**Description**: Renders metal sheet creation or editing dialog.

**Authentication**: Required

**Query Parameters**:

- `id` (integer, optional): Metal sheet ID for editing
- `tool_id` (integer, optional): Pre-assign to tool
- `close` (boolean, optional): Close dialog flag

**Response**: HTML dialog fragment

**Status Codes**:

- `200 OK`: Dialog rendered
- `404 Not Found`: Metal sheet not found (when editing)
- `401 Unauthorized`: Not authenticated

---

### Trouble Reports Management

#### Edit Dialog

```http
GET /htmx/trouble-reports/dialog-edit
```

**Description**: Renders trouble report creation or editing dialog.

**Authentication**: Required

**Query Parameters**:

- `id` (integer, optional): Report ID for editing
- `close` (boolean, optional): Close dialog flag

**Response**: HTML dialog fragment with form

**Status Codes**:

- `200 OK`: Dialog rendered
- `404 Not Found`: Report not found (when editing)
- `401 Unauthorized`: Not authenticated

---

#### Create Trouble Report

```http
POST /htmx/trouble-reports/dialog-edit
```

**Description**: Creates a new trouble report.

**Authentication**: Required

**Request Body** (multipart/form-data):

```
title=Press 3 Hydraulic Issue
content=Detailed description of the problem...
attachments[]=@file1.jpg
attachments[]=@file2.pdf
priority=high
```

**Response**: Success message and updated reports list

**Status Codes**:

- `201 Created`: Report created successfully
- `400 Bad Request`: Validation errors
- `413 Payload Too Large`: File too large
- `401 Unauthorized`: Not authenticated

---

#### Update Trouble Report

```http
PUT /htmx/trouble-reports/dialog-edit
```

**Description**: Updates an existing trouble report.

**Authentication**: Required

**Query Parameters**:

- `id` (integer, required): Report ID

**Request Body** (multipart/form-data): Same as create

**Response**: Success message and updated report information

**Status Codes**:

- `200 OK`: Report updated successfully
- `400 Bad Request`: Validation errors
- `404 Not Found`: Report not found
- `401 Unauthorized`: Not authenticated

---

#### Reports Data

```http
GET /htmx/trouble-reports/data
```

**Description**: Fetches paginated trouble reports data.

**Authentication**: Required

**Query Parameters**:

- `page` (integer, optional): Page number
- `search` (string, optional): Search query
- `status` (string, optional): Filter by status

**Response**: HTML fragment with reports table

**Status Codes**:

- `200 OK`: Reports data returned
- `401 Unauthorized`: Not authenticated

---

#### Delete Trouble Report

```http
DELETE /htmx/trouble-reports/data
```

**Description**: Deletes a trouble report and all attachments.

**Authentication**: Required

**Query Parameters**:

- `id` (integer, required): Report ID

**Response**: Success message and updated reports list

**Status Codes**:

- `200 OK`: Report deleted successfully
- `404 Not Found`: Report not found
- `401 Unauthorized`: Not authenticated

---

#### Attachments Preview

```http
GET /htmx/trouble-reports/attachments-preview
```

**Description**: Generates preview of report attachments.

**Authentication**: Required

**Query Parameters**:

- `id` (integer, required): Report ID
- `time` (string, optional): Modification time filter

**Response**: HTML fragment with attachment previews

**Status Codes**:

- `200 OK`: Previews generated
- `404 Not Found`: Report not found
- `401 Unauthorized`: Not authenticated

---

#### Modifications History

```http
GET /htmx/trouble-reports/modifications/:id
```

**Description**: Fetches modification history for a trouble report.

**Authentication**: Required

**Path Parameters**:

- `id` (integer): Report ID

**Response**: HTML fragment with modification history

**Status Codes**:

- `200 OK`: History returned
- `404 Not Found`: Report not found
- `401 Unauthorized`: Not authenticated

---

#### Restore Modification

```http
POST /htmx/trouble-reports/modifications/:id
```

**Description**: Restores a trouble report to a previous state.

**Authentication**: Required

**Path Parameters**:

- `id` (integer): Report ID

**Query Parameters**:

- `time` (string, required): Timestamp to restore to

**Response**: Success message and updated report

**Status Codes**:

- `200 OK`: Report restored successfully
- `404 Not Found`: Report or modification not found
- `400 Bad Request`: Invalid timestamp
- `401 Unauthorized`: Not authenticated

## WebSocket Endpoints

Real-time communication endpoints using WebSocket protocol.

### Feed Counter Updates

```
WS /htmx/nav/feed-counter
```

**Description**: Real-time feed counter updates for navigation.

**Authentication**: Required (via cookie or query parameter)

**Connection**: WebSocket upgrade

**Message Types**:

**Client → Server**:

```json
{
  "type": "ping",
  "timestamp": "2023-12-01T10:30:00Z"
}
```

**Server → Client**:

```json
{
  "type": "counter_update",
  "count": 15,
  "last_update": "2023-12-01T10:30:00Z"
}
```

**Connection Lifecycle**:

1. Client initiates WebSocket connection
2. Server validates authentication
3. Server sends initial counter value
4. Server sends updates when counter changes
5. Client sends periodic ping for keep-alive

## Error Handling

### Standard Error Responses

All endpoints follow consistent error response formats:

#### 400 Bad Request

```json
{
  "error": "Validation failed",
  "code": "VALIDATION_ERROR",
  "details": {
    "field": "tool_code",
    "message": "Tool code is required"
  }
}
```

#### 401 Unauthorized

```json
{
  "error": "Authentication required",
  "code": "AUTH_REQUIRED"
}
```

#### 403 Forbidden

```json
{
  "error": "Insufficient permissions",
  "code": "FORBIDDEN",
  "required_role": "admin"
}
```

#### 404 Not Found

```json
{
  "error": "Resource not found",
  "code": "NOT_FOUND",
  "resource": "tool",
  "id": 123
}
```

#### 409 Conflict

```json
{
  "error": "Resource conflict",
  "code": "CONFLICT",
  "details": "Tool code already exists"
}
```

#### 500 Internal Server Error

```json
{
  "error": "Internal server error",
  "code": "INTERNAL_ERROR",
  "request_id": "req_abc123"
}
```

## Rate Limiting

### Limits

- **API Endpoints**: 100 requests per minute per API key
- **HTMX Endpoints**: 200 requests per minute per session
- **File Uploads**: 10 MB per request, 100 MB per hour per user
- **WebSocket Connections**: 5 concurrent connections per user

### Rate Limit Headers

```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1638360000
```

## Caching

### Cache Headers

Static assets include appropriate cache headers:

```http
Cache-Control: public, max-age=31536000, immutable
ETag: "abc123def456"
Last-Modified: Wed, 01 Dec 2023 10:30:00 GMT
```

### Asset Versioning

All assets include version parameters:

```
/css/ui.min.css?v=1701425400
/js/app.min.js?v=1701425400
```

## Development and Testing

### Development Server

```bash
# Start development server with hot reload
make dev

# Server runs on http://localhost:8080
```

### API Testing

```bash
# Test with curl
curl -H "X-API-Key: your-key" \
     -X GET \
     http://localhost:8080/htmx/tools/list

# Test WebSocket with wscat
wscat -c "ws://localhost:8080/htmx/nav/feed-counter" \
      -H "Cookie: session=your-session"
```

### Response Time Targets

- **Page Routes**: < 200ms
- **HTMX Routes**: < 100ms
- **WebSocket Messages**: < 50ms
- **File Downloads**: Based on file size

## Security

### HTTPS Requirements

Production deployments should use HTTPS for:

- Secure cookie transmission
- WebSocket connections
- File uploads/downloads
- API key protection

### Content Security Policy

```
Content-Security-Policy: default-src 'self';
                        script-src 'self' 'unsafe-inline';
                        style-src 'self' 'unsafe-inline';
                        img-src 'self' data:;
                        connect-src 'self' ws: wss:;
```

### CORS Configuration

CORS is configured for development environments:

```
Access-Control-Allow-Origin: http://localhost:3000
Access-Control-Allow-Methods: GET, POST, PUT, DELETE
Access-Control-Allow-Headers: X-API-Key, Content-Type
```

This routing documentation provides comprehensive coverage of all endpoints, their usage patterns, and integration requirements. Keep it updated as new endpoints are added or existing ones are modified.
