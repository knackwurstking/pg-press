# Real-Time Feed Notification System

## Overview

The Real-Time Feed Notification System provides instant updates to users when new feed entries are created in the PG-VIS application. Using WebSocket connections with the HTMX WebSocket extension, users receive live feed counter updates without needing to refresh the page.

## Architecture

### Components

1. **Feed Notifier Manager** (`routes/internal/notifications/feed_notifier.go`)
    - Manages WebSocket connections
    - Broadcasts updates to all connected clients
    - Handles connection registration/unregistration

2. **Enhanced Feeds Model** (`pgvis/feeds.go`)
    - Integrated with notification system
    - Triggers notifications when new feeds are added

3. **WebSocket Handler** (`routes/handlers/nav/nav.go`)
    - Handles WebSocket connections for feed counter updates
    - Manages user authentication and connection lifecycle

4. **Frontend Integration** (HTMX WebSocket Extension)
    - Establishes WebSocket connections
    - Updates DOM in real-time

## How It Works

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │   WebSocket     │    │   Feed          │
│   (HTMX)        │    │   Handler       │    │   Notifier      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │ 1. Connect             │                       │
         ├──────────────────────→│                       │
         │                       │ 2. Register           │
         │                       ├──────────────────────→│
         │                       │                       │
         │ 3. Initial Counter     │                       │
         │←──────────────────────┤                       │
         │                       │                       │
         │                       │ 4. New Feed Added     │
         │                       │←──────────────────────┤
         │                       │                       │
         │ 5. Updated Counter     │                       │
         │←──────────────────────┤                       │
```

### Flow Description

1. **Connection Establishment**: Client connects via HTMX WebSocket extension
2. **Authentication**: User authentication is verified before connection registration
3. **Registration**: Connection is registered with the Feed Notifier Manager
4. **Initial Data**: Client receives initial feed counter data
5. **Real-time Updates**: When new feeds are added, all connected clients receive updates

## Setup and Configuration

### 1. Server Initialization

The notification system is automatically initialized in `routes/router.go`:

```go
// Initialize feed notification system
feedNotifier := notifications.NewFeedNotifier(o.DB, templates)

// Start the feed notification manager
ctx := context.Background()
go feedNotifier.Start(ctx)

// Set the notifier on the feeds for real-time updates
o.DB.Feeds.SetNotifier(feedNotifier)
```

### 2. Frontend Integration

Add the WebSocket connection to your navigation component:

```html
<a
    hx-ws="connect:./nav/feed-counter"
    hx-ext="ws"
    hx-target="#feedCounter"
    hx-swap="outerHTML"
>
    Feed Notifications
    <span id="feedCounter"></span>
</a>
```

### 3. Required Scripts

Ensure HTMX and the WebSocket extension are loaded:

```html
<script src="./js/htmx-v2.0.6.min.js"></script>
<script src="./js/htmx-ext-ws-v2.0.3.min.js"></script>
```

## API Endpoints

### WebSocket Endpoint

**GET** `/nav/feed-counter`

Establishes a WebSocket connection for real-time feed counter updates.

**Requirements:**

- User must be authenticated
- Connection is upgraded to WebSocket protocol

**Events:**

- `htmx:wsConnecting` - Connection attempt started
- `htmx:wsOpen` - Connection established
- `htmx:wsClose` - Connection closed
- `htmx:wsError` - Connection error
- `htmx:wsBeforeMessage` - Before receiving message
- `htmx:wsAfterMessage` - After processing message

### Test Endpoints (Development)

**POST** `/test/feed/create`

Creates test feeds for development and testing.

**Query Parameters:**

- `type` - Feed type (user_add, user_remove, trouble_report_add, etc.)
- `message` - Custom message for the feed
- `count` - Number of feeds to create (1-10)

**Example:**

```bash
curl -X POST "http://localhost:8080/test/feed/create?type=user_add&count=3"
```

## Configuration Options

### WebSocket Settings

```go
upgrader := websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true // Configure based on your security requirements
    },
}
```

### Connection Limits

The system includes built-in protections:

- Read limit: 512 bytes per message
- Write timeout: 10 seconds
- Read timeout: 60 seconds with ping/pong keepalive
- Ping interval: 54 seconds

### Buffer Sizes

- Send channel buffer: 256 messages
- Broadcast channel buffer: 100 notifications

## Testing

### Automated Testing

Use the comprehensive test page (`real-time-feed-test.html`):

1. **Connection Testing**: Verify WebSocket establishment
2. **Single Feed Creation**: Test individual feed types
3. **Batch Creation**: Test multiple feeds at once
4. **Stress Testing**: Create many feeds rapidly
5. **Connection Monitoring**: Real-time event logging

### Manual Testing

1. Start the PG-VIS server
2. Open the test page in a browser
3. Authenticate with the application
4. Click "Feed Notifications" to connect
5. Use test buttons to create feeds
6. Observe real-time counter updates

### Test Scenarios

```javascript
// Create single feed
await fetch("./test/feed/create?type=user_add", { method: "POST" });

// Create multiple feeds
await fetch("./test/feed/create?type=user_add&count=5", { method: "POST" });

// Stress test
for (let i = 0; i < 50; i++) {
    fetch("./test/feed/create?type=user_add", { method: "POST" });
}
```

## Performance Considerations

### Memory Management

- Connections are automatically cleaned up when clients disconnect
- Goroutines are properly terminated with context cancellation
- Buffer sizes prevent memory leaks

### Scalability

- Each connection runs in its own goroutine
- Broadcast operations are non-blocking
- Failed connections are automatically removed

### Network Efficiency

- Only HTML fragments are sent, not full pages
- Ping/pong keeps connections alive efficiently
- Graceful degradation on network issues

## Troubleshooting

### Common Issues

**WebSocket connection fails**

- Check user authentication
- Verify WebSocket endpoint is accessible
- Ensure HTMX WebSocket extension is loaded

**No real-time updates**

- Verify Feed Notifier is running
- Check if feeds are being created successfully
- Monitor browser developer console for errors

**High memory usage**

- Check for connection leaks
- Monitor number of active connections
- Verify goroutines are being cleaned up

### Debug Logging

Enable debug logging to monitor the system:

```go
log.Printf("[FeedNotifier] Broadcasting to %d connections", count)
log.Printf("[WebSocket] User authenticated: %s", user.UserName)
```

### Connection Monitoring

The test page includes real-time monitoring:

- Connection count
- Message count
- Feed creation statistics
- Detailed event logging

## Security Considerations

### Authentication

- All WebSocket connections require user authentication
- User context is validated before connection registration
- Connections are tied to specific user sessions

### Origin Validation

Configure `CheckOrigin` based on your security requirements:

```go
CheckOrigin: func(r *http.Request) bool {
    // Add your origin validation logic
    return true
}
```

### Rate Limiting

Consider implementing rate limiting for:

- WebSocket connection attempts
- Feed creation endpoints
- Message sending frequency

## Future Enhancements

### Planned Features

1. **Selective Notifications**: Filter updates based on user preferences
2. **Connection Persistence**: Reconnection with state recovery
3. **Message Queuing**: Offline message delivery
4. **Analytics**: Connection and usage metrics

### Extension Points

1. **Custom Feed Types**: Add new feed types with specific handling
2. **Push Notifications**: Integration with browser push API
3. **Multi-Channel**: Support for different notification channels
4. **User Preferences**: Customizable notification settings

## API Reference

### FeedNotifier Interface

```go
type FeedNotifier interface {
    NotifyNewFeed()
}
```

### FeedConnection Structure

```go
type FeedConnection struct {
    UserID   int64
    LastFeed int64
    Conn     *websocket.Conn
    Send     chan []byte
    Done     chan struct{}
}
```

### Key Methods

- `RegisterConnection(userID, lastFeed int64, conn *websocket.Conn)`
- `UnregisterConnection(conn *FeedConnection)`
- `NotifyNewFeed()`
- `Start(ctx context.Context)`

## License

This system is part of the PG-VIS project and follows the same licensing terms.
