# REST API Documentation

This document provides comprehensive documentation for the PG Press REST API, designed for programmatic access to the manufacturing management system.

## Overview

The PG Press REST API provides programmatic access to all core functionality:

- Tool and press management
- Cycle tracking and reporting
- Trouble report creation and management
- Notes and documentation system
- User management and authentication
- File attachment handling

### API Characteristics

- **Protocol**: HTTP/HTTPS
- **Data Format**: JSON
- **Authentication**: API Key or Session Cookie
- **Rate Limiting**: 1000 requests/hour per API key
- **Versioning**: v1 (current)

### Base URL

```
Production: https://your-domain.com/api/v1
Development: http://localhost:8080/api/v1
```

## Authentication

### API Key Authentication

Primary method for programmatic access:

```http
GET /api/v1/tools
Authorization: Bearer your-api-key-here
```

### Session Cookie Authentication

For web application integration:

```http
GET /api/v1/tools
Cookie: session=your-session-cookie
```

### Getting an API Key

API keys are generated through the web interface or CLI:

```bash
# CLI method
./pg-press user add-key --name "My API Client"

# Returns: api_key_abc123def456
```

## Common Patterns

### Request Headers

```http
Content-Type: application/json
Authorization: Bearer your-api-key
Accept: application/json
```

### Response Format

All responses follow a consistent structure:

**Success Response:**

```json
{
  "success": true,
  "data": {
    "id": 123,
    "name": "Tool ABC"
  },
  "meta": {
    "timestamp": "2023-12-01T10:30:00Z",
    "version": "v1"
  }
}
```

**Error Response:**

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Tool code is required",
    "details": {
      "field": "code",
      "value": null
    }
  },
  "meta": {
    "timestamp": "2023-12-01T10:30:00Z",
    "request_id": "req_abc123"
  }
}
```

### Pagination

List endpoints support pagination:

```http
GET /api/v1/tools?page=2&limit=50&sort=code&order=asc
```

**Pagination Response:**

```json
{
  "success": true,
  "data": [...],
  "pagination": {
    "page": 2,
    "limit": 50,
    "total": 150,
    "pages": 3,
    "has_next": true,
    "has_prev": true
  }
}
```

## Tools API

### List Tools

```http
GET /api/v1/tools
```

**Query Parameters:**

- `page` (integer): Page number (default: 1)
- `limit` (integer): Items per page (default: 20, max: 100)
- `press` (integer): Filter by press number (0-5)
- `position` (string): Filter by position (top, bottom)
- `type` (string): Filter by tool type
- `search` (string): Search in code and type
- `sort` (string): Sort field (code, type, press, created_at)
- `order` (string): Sort order (asc, desc)

**Response:**

```json
{
  "success": true,
  "data": [
    {
      "id": 123,
      "position": "top",
      "type": "cutting",
      "code": "ABC-123",
      "press": 5,
      "regenerating": false,
      "format": {
        "height": 100,
        "width": 200,
        "material": "steel"
      },
      "total_cycles": 15420,
      "created_at": "2023-11-01T09:00:00Z",
      "updated_at": "2023-12-01T10:30:00Z"
    }
  ],
  "pagination": {...}
}
```

### Get Tool

```http
GET /api/v1/tools/{id}
```

**Path Parameters:**

- `id` (integer): Tool ID

**Response:**

```json
{
  "success": true,
  "data": {
    "id": 123,
    "position": "top",
    "type": "cutting",
    "code": "ABC-123",
    "press": 5,
    "regenerating": false,
    "format": {
      "height": 100,
      "width": 200,
      "material": "steel"
    },
    "cycles": {
      "total": 15420,
      "recent": [
        {
          "id": 456,
          "date": "2023-12-01T08:00:00Z",
          "cycles": 150,
          "press_number": 5,
          "performed_by": {
            "id": 789,
            "name": "John Doe"
          }
        }
      ]
    },
    "notes": [
      {
        "id": 321,
        "level": 1,
        "content": "Requires calibration",
        "created_at": "2023-11-30T15:00:00Z"
      }
    ],
    "created_at": "2023-11-01T09:00:00Z",
    "updated_at": "2023-12-01T10:30:00Z"
  }
}
```

### Create Tool

```http
POST /api/v1/tools
```

**Request Body:**

```json
{
  "position": "top",
  "type": "cutting",
  "code": "ABC-124",
  "press": 3,
  "format": {
    "height": 100,
    "width": 200,
    "material": "steel",
    "specifications": {
      "tolerance": "±0.1mm",
      "surface_finish": "Ra 0.8"
    }
  }
}
```

**Response:**

```json
{
  "success": true,
  "data": {
    "id": 124,
    "position": "top",
    "type": "cutting",
    "code": "ABC-124",
    "press": 3,
    "regenerating": false,
    "format": {...},
    "created_at": "2023-12-01T10:30:00Z",
    "updated_at": "2023-12-01T10:30:00Z"
  }
}
```

### Update Tool

```http
PUT /api/v1/tools/{id}
```

**Path Parameters:**

- `id` (integer): Tool ID

**Request Body:** Same as create (partial updates supported)

### Delete Tool

```http
DELETE /api/v1/tools/{id}
```

**Path Parameters:**

- `id` (integer): Tool ID

**Response:**

```json
{
  "success": true,
  "data": {
    "message": "Tool deleted successfully",
    "deleted_id": 123
  }
}
```

## Press Cycles API

### List Cycles

```http
GET /api/v1/cycles
```

**Query Parameters:**

- `tool_id` (integer): Filter by tool ID
- `press_number` (integer): Filter by press number (0-5)
- `date_from` (date): Start date filter (YYYY-MM-DD)
- `date_to` (date): End date filter (YYYY-MM-DD)
- `performed_by` (integer): Filter by user ID

**Response:**

```json
{
  "success": true,
  "data": [
    {
      "id": 456,
      "tool_id": 123,
      "press_number": 5,
      "date": "2023-12-01T08:00:00Z",
      "total_cycles": 150,
      "performed_by": {
        "id": 789,
        "name": "John Doe"
      },
      "tool": {
        "id": 123,
        "code": "ABC-123",
        "type": "cutting"
      }
    }
  ]
}
```

### Create Cycle Record

```http
POST /api/v1/cycles
```

**Request Body:**

```json
{
  "tool_id": 123,
  "press_number": 5,
  "total_cycles": 150,
  "date": "2023-12-01T08:00:00Z",
  "notes": "Regular maintenance cycle"
}
```

### Get Cycle Statistics

```http
GET /api/v1/cycles/stats
```

**Query Parameters:**

- `tool_id` (integer, optional): Tool-specific stats
- `press_number` (integer, optional): Press-specific stats
- `period` (string): Time period (day, week, month, year)

**Response:**

```json
{
  "success": true,
  "data": {
    "total_cycles": 150000,
    "average_per_day": 500,
    "by_press": {
      "0": 25000,
      "1": 30000,
      "2": 20000,
      "3": 35000,
      "4": 25000,
      "5": 15000
    },
    "by_period": [
      {
        "date": "2023-12-01",
        "cycles": 500
      }
    ]
  }
}
```

## Trouble Reports API

### List Reports

```http
GET /api/v1/trouble-reports
```

**Query Parameters:**

- `status` (string): Filter by status (open, in_progress, resolved)
- `priority` (string): Filter by priority (low, medium, high, critical)
- `search` (string): Search in title and content
- `created_from` (date): Filter reports created after date
- `created_to` (date): Filter reports created before date

**Response:**

```json
{
  "success": true,
  "data": [
    {
      "id": 789,
      "title": "Press 3 Hydraulic Issue",
      "content": "Hydraulic system showing pressure drops...",
      "status": "open",
      "priority": "high",
      "attachments": [
        {
          "id": 101,
          "filename": "hydraulic_reading.jpg",
          "mime_type": "image/jpeg",
          "size": 245760,
          "url": "/api/v1/attachments/101"
        }
      ],
      "created_at": "2023-12-01T09:00:00Z",
      "updated_at": "2023-12-01T10:30:00Z"
    }
  ]
}
```

### Create Report

```http
POST /api/v1/trouble-reports
Content-Type: multipart/form-data
```

**Form Fields:**

- `title` (string): Report title
- `content` (string): Detailed description
- `priority` (string): Priority level (low, medium, high, critical)
- `attachments[]` (file): Multiple file uploads supported

**Response:**

```json
{
  "success": true,
  "data": {
    "id": 790,
    "title": "Press 3 Hydraulic Issue",
    "content": "Hydraulic system showing pressure drops...",
    "status": "open",
    "priority": "high",
    "attachments": [...],
    "created_at": "2023-12-01T10:30:00Z"
  }
}
```

### Export Report PDF

```http
GET /api/v1/trouble-reports/{id}/pdf
```

**Path Parameters:**

- `id` (integer): Report ID

**Response:** PDF file download with appropriate headers

### Update Report

```http
PUT /api/v1/trouble-reports/{id}
```

**Request Body:**

```json
{
  "title": "Updated title",
  "content": "Updated content",
  "status": "in_progress",
  "priority": "medium"
}
```

## Notes API

### List Notes

```http
GET /api/v1/notes
```

**Query Parameters:**

- `linked` (string): Filter by linked entity (e.g., "tool_123", "press_5")
- `level` (integer): Filter by priority level (0=INFO, 1=ATTENTION, 2=BROKEN)
- `search` (string): Search in note content

**Response:**

```json
{
  "success": true,
  "data": [
    {
      "id": 321,
      "level": 1,
      "content": "Tool requires calibration every 100 cycles",
      "linked": "tool_123",
      "linked_entity": {
        "type": "tool",
        "id": 123,
        "code": "ABC-123"
      },
      "created_at": "2023-11-30T15:00:00Z"
    }
  ]
}
```

### Create Note

```http
POST /api/v1/notes
```

**Request Body:**

```json
{
  "level": 1,
  "content": "Important maintenance note",
  "linked": "tool_123"
}
```

### Update Note

```http
PUT /api/v1/notes/{id}
```

### Delete Note

```http
DELETE /api/v1/notes/{id}
```

## Users API

### List Users

```http
GET /api/v1/users
```

**Response:**

```json
{
  "success": true,
  "data": [
    {
      "telegram_id": 123456789,
      "user_name": "John Doe",
      "last_feed": "feed_456",
      "created_at": "2023-10-01T09:00:00Z",
      "last_login": "2023-12-01T08:30:00Z"
    }
  ]
}
```

### Get Current User

```http
GET /api/v1/users/me
```

**Response:**

```json
{
  "success": true,
  "data": {
    "telegram_id": 123456789,
    "user_name": "John Doe",
    "api_key": "api_key_abc123def456",
    "permissions": ["read", "write", "admin"],
    "last_feed": "feed_456",
    "active_sessions": 3,
    "last_login": "2023-12-01T08:30:00Z"
  }
}
```

### Update User Profile

```http
PUT /api/v1/users/me
```

**Request Body:**

```json
{
  "user_name": "Updated Name"
}
```

## Activity Feed API

### Get Feed

```http
GET /api/v1/feed
```

**Query Parameters:**

- `limit` (integer): Items per page (max: 100)
- `since` (string): Get entries after this feed ID
- `type` (string): Filter by event type

**Response:**

```json
{
  "success": true,
  "data": [
    {
      "id": "feed_789",
      "time": "2023-12-01T10:30:00Z",
      "type": "tool_created",
      "data": {
        "user": "John Doe",
        "entity_type": "tool",
        "entity_id": 123,
        "action": "created",
        "details": {
          "tool_code": "ABC-123",
          "press": 5
        }
      }
    }
  ],
  "meta": {
    "latest_id": "feed_789",
    "has_more": false
  }
}
```

### Mark Feed as Read

```http
POST /api/v1/feed/mark-read
```

**Request Body:**

```json
{
  "feed_id": "feed_789"
}
```

## Attachments API

### Upload Attachment

```http
POST /api/v1/attachments
Content-Type: multipart/form-data
```

**Form Fields:**

- `file` (file): File to upload
- `description` (string, optional): File description

**Response:**

```json
{
  "success": true,
  "data": {
    "id": 101,
    "filename": "hydraulic_reading.jpg",
    "mime_type": "image/jpeg",
    "size": 245760,
    "url": "/api/v1/attachments/101",
    "created_at": "2023-12-01T10:30:00Z"
  }
}
```

### Download Attachment

```http
GET /api/v1/attachments/{id}
```

**Path Parameters:**

- `id` (integer): Attachment ID

**Response:** File download with appropriate headers

### Delete Attachment

```http
DELETE /api/v1/attachments/{id}
```

## Error Codes

### HTTP Status Codes

- `200 OK`: Successful request
- `201 Created`: Resource created successfully
- `204 No Content`: Successful request with no response body
- `400 Bad Request`: Invalid request data
- `401 Unauthorized`: Authentication required or invalid
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Resource not found
- `409 Conflict`: Resource conflict (e.g., duplicate code)
- `422 Unprocessable Entity`: Validation errors
- `429 Too Many Requests`: Rate limit exceeded
- `500 Internal Server Error`: Server error

### Application Error Codes

| Code                  | Description                | HTTP Status |
| --------------------- | -------------------------- | ----------- |
| `AUTH_REQUIRED`       | Authentication required    | 401         |
| `INVALID_API_KEY`     | API key invalid or expired | 401         |
| `FORBIDDEN`           | Insufficient permissions   | 403         |
| `NOT_FOUND`           | Resource not found         | 404         |
| `VALIDATION_ERROR`    | Request validation failed  | 422         |
| `DUPLICATE_CODE`      | Tool code already exists   | 409         |
| `INVALID_PRESS`       | Press number must be 0-5   | 422         |
| `FILE_TOO_LARGE`      | File exceeds size limit    | 413         |
| `INVALID_FILE_TYPE`   | Unsupported file type      | 422         |
| `RATE_LIMIT_EXCEEDED` | Too many requests          | 429         |

## Rate Limiting

### Limits

- **Free tier**: 1,000 requests/hour
- **Premium tier**: 10,000 requests/hour
- **Enterprise**: Custom limits

### Headers

Rate limit information is included in response headers:

```http
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 995
X-RateLimit-Reset: 1701435600
X-RateLimit-Window: 3600
```

### Rate Limit Exceeded

```json
{
  "success": false,
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Rate limit exceeded. Try again in 45 minutes.",
    "retry_after": 2700
  }
}
```

## SDKs and Client Libraries

### Official SDKs

- **Go**: `github.com/your-org/pg-press-go`
- **Python**: `pip install pg-press-client`
- **JavaScript/Node.js**: `npm install pg-press-client`

### Usage Examples

**Go:**

```go
import "github.com/your-org/pg-press-go"

client := pgpress.NewClient("your-api-key")
tools, err := client.Tools.List(context.Background(), &pgpress.ToolsListOptions{
    Press: 5,
    Limit: 50,
})
```

**Python:**

```python
from pg_press import Client

client = Client(api_key="your-api-key")
tools = client.tools.list(press=5, limit=50)
```

**JavaScript:**

```javascript
import PGPress from "pg-press-client";

const client = new PGPress({ apiKey: "your-api-key" });
const tools = await client.tools.list({ press: 5, limit: 50 });
```

## Webhooks

### Configuration

Configure webhooks to receive real-time notifications:

```http
POST /api/v1/webhooks
```

**Request Body:**

```json
{
  "url": "https://your-app.com/webhooks/pg-press",
  "events": ["tool.created", "cycle.recorded", "report.created"],
  "secret": "your-webhook-secret"
}
```

### Event Types

- `tool.created` - New tool added
- `tool.updated` - Tool modified
- `tool.deleted` - Tool removed
- `cycle.recorded` - New cycle data
- `report.created` - New trouble report
- `report.updated` - Report modified
- `note.created` - New note added

### Payload Format

```json
{
  "event": "tool.created",
  "timestamp": "2023-12-01T10:30:00Z",
  "data": {
    "tool": {
      "id": 123,
      "code": "ABC-123",
      "press": 5
    },
    "user": {
      "id": 789,
      "name": "John Doe"
    }
  },
  "signature": "sha256=abc123def456..."
}
```

## Best Practices

### Authentication

- Store API keys securely
- Rotate keys regularly
- Use HTTPS in production
- Implement proper error handling

### Performance

- Use pagination for large datasets
- Cache responses when appropriate
- Implement exponential backoff for retries
- Monitor rate limits

### Error Handling

- Check HTTP status codes
- Parse error response details
- Implement retry logic for transient errors
- Log errors for debugging

### Data Validation

- Validate input data before sending
- Handle validation errors gracefully
- Use appropriate data types
- Follow naming conventions

## Changelog

### v1.0.0 (Current)

- Initial API release
- Full CRUD operations for all resources
- Authentication and rate limiting
- File upload support
- Webhook notifications

### Future Versions

- GraphQL endpoint
- Bulk operations
- Advanced filtering and search
- Real-time subscriptions
- Enhanced analytics endpoints

## Support

For API support and questions:

- Documentation: https://docs.pg-press.com/api
- GitHub Issues: https://github.com/your-org/pg-press/issues
- Email: api-support@your-domain.com

## Terms of Service

By using the PG Press API, you agree to our Terms of Service and Privacy Policy. Please review these documents before integrating with our API.
