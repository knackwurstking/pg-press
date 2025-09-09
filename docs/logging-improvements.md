# Logging Improvements Documentation

## Overview

This document outlines the comprehensive logging improvements made across the pg-press application to enhance debugging, monitoring, and troubleshooting capabilities. The improvements focus on meaningful, actionable logging while avoiding excessive debug noise.

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
[INFO ] 2024/01/15 10:30:15 [Middleware] API key authentication successful for user TestUser from 192.168.1.1
[WARN ] 2024/01/15 10:30:16 [Middleware] Slow request: GET /tools took 1.2s (user-agent: Mozilla/5.0...)
[WARN ] 2024/01/15 10:30:17 [Middleware] Authentication failed from 192.168.1.1: invalid credentials
```

### 2. Server Command Enhancements

**File: `cmd/pg-press/commands-server.go`**

#### Improvements Made:

- **Startup Sequence**: Key server initialization milestones
- **Error Context**: Enhanced error messages with common causes and troubleshooting hints
- **Log File Redirection**: Status of log output redirection
- **Critical Failures**: Server startup failures with diagnostic information

#### Example Log Entries:

```
[INFO ] 2024/01/15 10:25:00 [Server] Redirected logs to file: /var/log/pgpress.log
[INFO ] 2024/01/15 10:25:01 [Server] Starting HTTP server on localhost:8080
[ERROR] 2024/01/15 10:25:02 [Server] Server startup failed on localhost:8080: address already in use
[ERROR] 2024/01/15 10:25:02 [Server] Common causes: port already in use, permission denied, invalid address format
```

### 3. Database Service Logging

#### User Service (`internal/database/services/user/service.go`)

**Performance Monitoring:**

- Query execution timing with warnings for slow operations (>100ms)
- Focus on operations that impact user experience
- Bulk operation performance tracking

**Enhanced Error Context:**

- User identification in critical operations
- Clear validation failure messages
- Transaction state information when relevant

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
[INFO ] 2024/01/15 11:00:00 [HTMX Handler Tools] User TestUser creating new tool
[INFO ] 2024/01/15 11:00:01 [HTMX Handler Tools] Created tool ID 456 (Type=FC, Code=G01) by user TestUser
[INFO ] 2024/01/15 11:00:02 [HTMX Handler Tools] User TestUser deleting tool 789
[WARN ] 2024/01/15 11:00:03 [HTMX Handler Tools] Slow tools query took 250ms for 150 tools
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
[INFO ] 2024/01/15 09:00:00 [Handler Auth] Successful login for user from 192.168.1.1
[INFO ] 2024/01/15 09:00:01 [Handler Auth] User TestUser logging out
[WARN ] 2024/01/15 09:00:02 [Handler Auth] Failed login attempt from 192.168.1.1
[INFO ] 2024/01/15 09:00:03 [Handler Auth] Administrator AdminUser logged in from 192.168.1.1
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
[INFO ] 2024/01/15 12:00:00 [HTMX Handler Nav] WebSocket connection established for user TestUser from 192.168.1.1
[INFO ] 2024/01/15 12:00:01 [HTMX Handler Nav] WebSocket connection closed for user TestUser from 192.168.1.1
[ERROR] 2024/01/15 12:00:02 [HTMX Handler Nav] WebSocket authentication failed from 192.168.1.100: invalid user
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
[INFO ] 2024/01/15 10:30:18 [Handler Auth] Successful login for user from 192.168.1.1
```

## Error Handling Improvements

### Context-Rich Error Messages

Error logs now include:

- **Operation Context**: What was being attempted
- **User Information**: Who initiated the operation (when relevant)
- **System State**: Key information for troubleshooting
- **Recovery Actions**: Cleanup or rollback status when applicable

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

### Log Level Guidelines

- **DEBUG**: Reserved for development troubleshooting (minimal use in production)
- **INFO**: Successful operations, user actions, and significant events
- **WARN**: Performance issues, recoverable errors, and potential problems
- **ERROR**: Operation failures, system errors, and critical issues

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
2. **Focus on Value**: Only log information that aids troubleshooting or monitoring
3. **User Context**: Include user information for important operations and security events
4. **Performance Awareness**: Log timing for operations that could impact user experience
5. **Maintain Consistency**: Use consistent log formats across components
6. **Avoid Noise**: Minimize debug logging in production environments

### Security Considerations

1. **Sensitive Data**: Never log passwords, API keys, or personal information in full
2. **Data Masking**: Use masking for sensitive data (e.g., `api_key: ****1234`)
3. **IP Logging**: Log IP addresses for security monitoring
4. **Rate Limiting**: Avoid excessive logging that could impact performance

### Performance Impact

1. **Early Exit**: Use early exit patterns for filtered log levels
2. **Selective Logging**: Focus on meaningful events rather than verbose debugging
3. **Performance Thresholds**: Only log timing when operations exceed expected durations
4. **Log Rotation**: Implement log rotation for file-based logging
5. **Resource Awareness**: Avoid excessive logging that could impact application performance

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
