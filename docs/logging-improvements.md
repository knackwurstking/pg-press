# Logging Improvements Documentation

## Overview

This document outlines the comprehensive logging improvements made across the pg-press application to enhance debugging, monitoring, and troubleshooting capabilities.

## Key Improvements

### 1. Enhanced Middleware Logging

**File: `cmd/pg-press/middleware.go`**

#### Improvements Made:

- **Security Logging**: Added logging for authentication failures, admin actions, and unauthenticated requests to sensitive endpoints
- **Performance Monitoring**: Log slow requests (>1s) with detailed timing information
- **Request Context**: Include user agent, content length, and remote IP in logs
- **User Agent Validation**: Log mismatches between cookie and request user agents as potential security concerns
- **Session Management**: Detailed logging of cookie updates and timestamp changes

#### Example Log Entries:

```
[INFO ] 2024/01/15 10:30:15 [Middleware] Admin action: POST /htmx/tools/edit by AdminUser
[WARN ] 2024/01/15 10:30:16 [Middleware] Slow request: GET /tools took 1.2s (user-agent: Mozilla/5.0...)
[WARN ] 2024/01/15 10:30:17 [Middleware] User agent mismatch for user TestUser from 192.168.1.1
```

### 2. Server Command Enhancements

**File: `cmd/pg-press/commands-server.go`**

#### Improvements Made:

- **Startup Sequence**: Detailed logging of server initialization steps
- **Database Connection**: Log database path and connection status
- **Error Context**: Enhanced error messages with common causes and troubleshooting hints
- **Configuration Logging**: Log file redirection and middleware setup progress

#### Example Log Entries:

```
[INFO ] 2024/01/15 10:25:00 [Server] Opening database at: /path/to/database.db
[INFO ] 2024/01/15 10:25:01 [Server] Database opened successfully
[INFO ] 2024/01/15 10:25:02 [Server] Starting HTTP server on localhost:8080
[ERROR] 2024/01/15 10:25:03 [Server] Server startup failed on localhost:8080: address already in use
```

### 3. Database Service Logging

#### User Service (`internal/database/services/user/service.go`)

**Performance Monitoring:**

- Query execution timing with warnings for slow operations (>100ms)
- Detailed operation tracking with elapsed time measurements
- Row count and data size logging

**Enhanced Error Context:**

- User identification in all operations
- Validation failure details with field-specific information
- Transaction rollback logging with cleanup status

**Security and Audit Trail:**

- Admin user action tracking
- Authentication attempt logging with source IP context
- API key validation with length and format checks

#### Feed Service (`internal/database/services/feed/service.go`)

**Query Performance:**

- Individual query timing with performance warnings
- Pagination parameter validation and logging
- Bulk operation progress tracking

**Error Handling:**

- Database connection error details
- Row scanning failure context
- Broadcasting notification status

#### Trouble Report Service (`internal/database/services/troublereport/service.go`)

**Comprehensive Operation Logging:**

- Attachment processing with size and type information
- Multi-step operation timing (marshal → database → feed creation)
- Cleanup operation status on failures

**Enhanced Error Recovery:**

- Failed attachment cleanup logging
- Transaction rollback status
- Orphaned data identification

### 4. HTMX Handler Improvements

**File: `internal/web/htmx/tools.go`**

#### Improvements Made:

- **Request Tracking**: Log remote IP, user agent, and request context
- **Form Validation**: Detailed validation error logging with field-specific messages
- **Performance Metrics**: Database query timing and render performance tracking
- **User Activity**: Comprehensive audit trail for create, update, and delete operations

#### Example Log Entries:

```
[INFO ] 2024/01/15 11:00:00 [HTMX Handler Tools] User TestUser (ID: 123) creating new tool from 192.168.1.1
[INFO ] 2024/01/15 11:00:01 [HTMX Handler Tools] Creating tool: Type=FC, Code=G01, Position=top, Format=100x50 by user TestUser
[INFO ] 2024/01/15 11:00:02 [HTMX Handler Tools] Successfully created tool ID 456 in 250ms (db: 200ms)
[WARN ] 2024/01/15 11:00:03 [HTMX Handler Tools] Invalid width value from 192.168.1.1: abc
```

### 5. Authentication Handler Improvements

**File: `internal/web/html/auth.go`**

#### Improvements Made:

- **Login Process Tracking**: Complete login flow timing and status
- **Session Management**: Cookie creation and cleanup logging
- **Security Monitoring**: Failed authentication attempts with context
- **User Experience**: Redirect timing and error handling

#### Example Log Entries:

```
[INFO ] 2024/01/15 09:00:00 [Handler Auth] Login page request from 192.168.1.1
[INFO ] 2024/01/15 09:00:01 [Handler Auth] Processing login attempt from 192.168.1.1 with API key (length: 32)
[INFO ] 2024/01/15 09:00:02 [Handler Auth] User TestUser (ID: 123) authenticated from 192.168.1.1 (db lookup: 50ms)
[INFO ] 2024/01/15 09:00:03 [Handler Auth] Successfully created session for user TestUser in 100ms
```

### 6. WebSocket Connection Logging

**File: `internal/web/htmx/nav.go`**

#### Improvements Made:

- **Connection Lifecycle**: Complete WebSocket connection tracking
- **User Association**: Link WebSocket connections to authenticated users
- **Performance Monitoring**: Connection setup and registration timing
- **Error Handling**: Connection failure diagnosis and cleanup logging

#### Example Log Entries:

```
[INFO ] 2024/01/15 12:00:00 [HTMX Handler Nav] WebSocket upgrade request from 192.168.1.1
[INFO ] 2024/01/15 12:00:01 [HTMX Handler Nav] WebSocket user authenticated: TestUser (ID: 123, LastFeed: 456) from 192.168.1.1
[INFO ] 2024/01/15 12:00:02 [HTMX Handler Nav] WebSocket connection registered for user TestUser in 50ms
```

## Performance Monitoring

### Timing Thresholds

The logging system now includes performance warnings for operations exceeding these thresholds:

| Operation Type       | Warning Threshold | Critical Threshold |
| -------------------- | ----------------- | ------------------ |
| Database Queries     | 100ms             | 500ms              |
| HTTP Requests        | 1000ms            | 5000ms             |
| Template Rendering   | 50ms              | 200ms              |
| WebSocket Operations | 100ms             | 500ms              |
| File I/O Operations  | 200ms             | 1000ms             |

### Performance Log Format

```
[WARN ] 2024/01/15 10:30:15 [Component] Slow operation took 1.5s for context (expected: <200ms)
```

## Security Logging

### Authentication Events

- **Login Attempts**: Success/failure with IP address and user agent
- **Session Management**: Cookie creation, updates, and cleanup
- **Authorization**: Admin actions and permission checks
- **API Key Usage**: Invalid keys and suspicious patterns

### Security Log Examples

```
[WARN ] 2024/01/15 10:30:15 [Middleware] Authentication failed from 192.168.1.100: invalid credentials
[INFO ] 2024/01/15 10:30:16 [Handler Auth] Administrator TestAdmin logged in from 192.168.1.1
[WARN ] 2024/01/15 10:30:17 [Middleware] Unauthenticated POST request to /htmx/tools/edit from 192.168.1.200
```

## Error Handling Improvements

### Context-Rich Error Messages

All error logs now include:

- **Operation Context**: What was being attempted
- **User Information**: Who initiated the operation
- **Timing Information**: How long the operation took before failing
- **System State**: Relevant system information at time of error
- **Recovery Actions**: What cleanup or rollback occurred

### Error Log Format

```
[ERROR] 2024/01/15 10:30:15 [Component] Failed to perform operation for user TestUser in 250ms: detailed error message
```

## Debugging Enhancements

### Request Correlation

- **HTTP Requests**: Include remote IP, user agent, and session information
- **Database Operations**: Link to initiating user and operation context
- **WebSocket Connections**: Track complete connection lifecycle
- **File Operations**: Include user context and operation timing

### Debug Log Levels

- **DEBUG**: Detailed operation flow and state changes
- **INFO**: Successful operations and significant events
- **WARN**: Performance issues and recoverable errors
- **ERROR**: Operation failures and system errors

## Configuration

### Environment Variables

The logging system respects these environment variables:

- `LOG_LEVEL`: Set minimum log level (DEBUG, INFO, WARN, ERROR)
- `NO_COLOR`: Disable colored output
- `FORCE_COLOR`: Force colored output even when not in terminal

### Log Output

Logs can be directed to:

- **Standard Error**: Default output for console applications
- **File Output**: Using the `--log-file` command line option
- **Structured Format**: JSON format for log aggregation systems

## Best Practices

### Log Message Guidelines

1. **Be Specific**: Include relevant IDs, names, and context
2. **Include Timing**: Log operation duration for performance monitoring
3. **Add User Context**: Include user information for audit trails
4. **Provide Actionability**: Include enough information for troubleshooting
5. **Maintain Consistency**: Use consistent log formats across components

### Security Considerations

1. **Sensitive Data**: Never log passwords, API keys, or personal information in full
2. **Data Masking**: Use masking for sensitive data (e.g., `api_key: ****1234`)
3. **IP Logging**: Log IP addresses for security monitoring
4. **Rate Limiting**: Avoid excessive logging that could impact performance

### Performance Impact

1. **Early Exit**: Use early exit patterns for filtered log levels
2. **Lazy Evaluation**: Use format strings instead of string concatenation
3. **Structured Logging**: Consider structured formats for high-volume logging
4. **Log Rotation**: Implement log rotation for file-based logging

## Monitoring and Alerting

### Key Metrics to Monitor

1. **Error Rate**: Monitor ERROR level log frequency
2. **Performance Degradation**: Track WARN level performance messages
3. **Authentication Failures**: Monitor failed login attempts
4. **Database Performance**: Track slow query warnings
5. **WebSocket Connections**: Monitor connection establishment/failure rates

### Alert Thresholds

- **High Error Rate**: >10 errors/minute
- **Performance Degradation**: >5 slow operations/minute
- **Authentication Issues**: >20 failed logins/hour
- **Database Issues**: >50% queries above warning threshold

This comprehensive logging enhancement provides full visibility into application behavior, performance characteristics, and security events while maintaining good performance and readability.
