# WebSocket Suspension Fix - Complete Guide

## 🎯 Problem Overview

Previously, the PG-VIS application had aggressive WebSocket timeout handling that caused connections to disconnect when browser tabs were suspended (backgrounded) or when users switched between tabs. This resulted in:

- **Timeout errors** when tabs were suspended for more than 60 seconds
- **Connection drops** during normal mobile browsing behavior
- **Poor user experience** with feed updates stopping unexpectedly
- **Unnecessary server load** from aggressive reconnection attempts

## ✅ Solution Implemented

### 1. Server-Side Improvements (`routes/internal/notifications/feed_notifier.go`)

**Timeout Adjustments:**

- ❌ **Removed aggressive read deadlines** - no more 60-second disconnections
- ⏰ **Increased write timeouts** from 10s to 30s - handles slow connections
- 📡 **Extended ping intervals** from 54s to 5 minutes - reduces traffic
- 🕒 **Increased feed update timeouts** from 5s to 30s - accommodates suspension

**Smart Error Handling:**

```go
// New intelligent error detection
func isTemporaryError(err error) bool {
    errStr := err.Error()
    return strings.Contains(errStr, "timeout") ||
           strings.Contains(errStr, "deadline") ||
           strings.Contains(errStr, "broken pipe") ||
           strings.Contains(errStr, "connection reset")
}
```

### 2. Client-Side WebSocket Manager (`routes/assets/js/websocket-manager.js`)

**Suspension Detection:**

- 👁️ **Page visibility monitoring** - detects tab suspension/restoration
- 🌐 **Network state awareness** - coordinates with online/offline events
- 🔄 **Smart reconnection logic** - only reconnects when appropriate
- 📈 **Exponential backoff** - prevents server overload

**Key Features:**

```javascript
// Handles browser suspension gracefully
handleVisibilityChange() {
    if (!wasVisible && this.isPageVisible) {
        // Page became visible - check and reconnect websockets
        this.checkAndReconnectAll();
    }
}
```

### 3. Enhanced PWA Integration (`routes/assets/js/pwa-manager.js`)

**Coordinated Recovery:**

- 🔗 **WebSocket health checks** after visibility changes
- 🔄 **Service worker coordination** for offline/online transitions
- ⚡ **Automatic connection restoration** when tabs become active

## 🧪 Testing Instructions

### Quick Test Setup

1. **Start the server:**

```bash
cd pg-vis
go run ./cmd/pg-vis
```

2. **Open the main application:**

```bash
# Open in browser
open http://localhost:8080/
```

### Test Scenarios

#### ✋ Manual Suspension Test

1. Navigate to any page with the feed counter (bell icon in navigation)
2. Verify WebSocket connection is active (feed counter should be functional)
3. **Switch to another tab** for 60+ seconds (old system would disconnect here)
4. **Return to the PG-VIS tab**
5. ✅ **Expected:** Connection automatically restores, feed counter updates resume

#### 🔄 Browser Console Test

1. Open browser developer tools console
2. Use `wsManagerDebug.stats()` to check connection status
3. Switch tabs and return
4. ✅ **Expected:** Connection statistics show successful reconnection

#### 🌐 Network Resilience Test

1. Navigate to main application
2. Simulate network disconnect/reconnect
3. Check feed counter functionality
4. ✅ **Expected:** Connection recovers after network restoration

#### 📱 Mobile Testing

1. Open main PG-VIS application on mobile device
2. Verify feed counter is working
3. **Switch between apps** for 2+ minutes
4. Return to browser
5. ✅ **Expected:** Connection restored automatically

### Real-World Testing

#### Feed Counter Integration Test

1. **Navigate to main PG-VIS application:**

```bash
open http://localhost:8080/
```

2. **Create test feeds:**

```bash
curl -X POST "http://localhost:8080/test/feed/create?type=user_add&count=3"
```

3. **Test suspension scenario:**
    - Notice feed counter in navigation
    - Switch tabs for 2+ minutes
    - Return to PG-VIS tab
    - Create more test feeds
    - ✅ **Expected:** Counter updates immediately, no reconnection delay

#### Multi-Tab Testing

1. Open **multiple PG-VIS tabs**
2. Connect WebSocket in each tab
3. **Suspend some tabs** by switching away
4. Create test feeds in one active tab
5. **Restore suspended tabs**
6. ✅ **Expected:** All tabs receive updates when restored

## 📊 Monitoring and Verification

### Server-Side Monitoring

**Log Messages to Watch For:**

```bash
# Good - connection maintained during suspension
[FeedConnection] Connection closed normally for user X

# Good - smart timeout handling
[FeedConnection] Temporary error for user X (possibly suspended): timeout

# Bad - would indicate fix isn't working
[FeedConnection] Error reading message from user X: deadline exceeded
```

**Monitor Connection Metrics:**

```bash
# Check active connections
curl -s http://localhost:8080/debug/websocket-stats

# Expected: Stable connection count even during tab switching
```

### Client-Side Monitoring

**Browser Console Logs:**

```javascript
// Good indicators
[WebSocketManager] Page became visible, checking websocket connections
[WebSocketManager] WebSocket opened: ./nav/feed-counter
[WebSocketManager] Reconnecting stale connection: ./nav/feed-counter

// Use debug commands
wsManagerDebug.stats()        // Show connection statistics
wsManagerDebug.reconnectAll() // Force reconnection test
```

**Browser Console Metrics:**

```javascript
// Check connection statistics
wsManagerDebug.stats();
// Expected output should show:
// - totalConnections: reasonable number
// - isPageVisible: true when tab is active
// - isOnline: true when network is available
// - reconnectAttempts: minimal failed attempts
```

## 🔧 Technical Implementation Details

### Architecture Changes

```
Before:
┌─────────────┐    60s timeout    ┌──────────────┐
│ Browser Tab │ ────────────────► │ Server       │
│ (suspended) │ ◄──── disconnect ─┤ (aggressive) │
└─────────────┘                   └──────────────┘

After:
┌─────────────┐   no timeout     ┌──────────────┐
│ Browser Tab │ ────────────────► │ Server       │
│ (suspended) │ ◄──── maintain ──┤ (patient)    │
└─────────────┘                   └──────────────┘
        │                                │
        ▼         on restore             ▼
┌─────────────┐ ────────────────► ┌──────────────┐
│ WebSocket   │ ◄──── reconnect ──┤ Smart Logic  │
│ Manager     │                   │ (client)     │
└─────────────┘                   └──────────────┘
```

### Key Configuration Changes

**Server Configuration:**

```go
// Old aggressive timeouts
conn.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
ticker := time.NewTicker(54 * time.Second)

// New patient timeouts
// conn.Conn.SetReadDeadline() // Removed entirely
ticker := time.NewTicker(5 * time.Minute)
```

**Client Configuration:**

```javascript
// Smart reconnection logic
reconnectDelay: 1000,        // Start with 1 second
maxReconnectDelay: 30000,    // Max 30 seconds
maxReconnectAttempts: 5,     // Reasonable retry limit
```

## 🐛 Troubleshooting

### Issue: WebSocket Still Disconnects During Suspension

**Diagnosis:**

```bash
# Check if websocket-manager.js is loaded
curl -s http://localhost:8080/js/websocket-manager.js | head -5

# Verify browser console for errors
# Expected: "[WebSocketManager] WebSocket suspension handling initialized"
```

**Solution:**

1. Ensure `websocket-manager.js` is included in main layout
2. Verify HTMX WebSocket extension is loaded first
3. Check browser compatibility (requires modern features)

### Issue: Reconnection Not Working

**Diagnosis:**

```javascript
// In browser console
console.log(window.webSocketManager); // Should be defined
console.log(window.wsManagerDebug.stats()); // Should show stats
```

**Solution:**

1. Check page visibility API support: `document.hidden`
2. Verify WebSocket Manager initialization
3. Test with manual reconnection: `wsManagerDebug.reconnectAll()`

### Issue: High Server CPU During Reconnections

**Symptoms:**

```bash
# High connection churn in logs
grep "Connection registered" server.log | wc -l  # Should be reasonable
grep "Connection closed" server.log | wc -l     # Should match connects
```

**Solution:**

1. Verify exponential backoff is working
2. Check for JavaScript errors preventing proper cleanup
3. Monitor connection metrics in test page

### Issue: Feed Updates Missing After Suspension

**Test:**

```bash
# Create test feed while tab suspended
curl -X POST "http://localhost:8080/test/feed/create?type=user_add&count=1"

# Return to tab - should see update within 5 seconds
```

**Solution:**

1. Check feed counter template rendering
2. Verify HTMX target elements exist
3. Test with browser developer tools Network tab

## 🔍 Performance Impact

### Positive Changes

✅ **Reduced server load** - Fewer unnecessary disconnections
✅ **Lower bandwidth usage** - Less frequent ping messages (5min vs 54s)
✅ **Better mobile experience** - Handles app switching gracefully
✅ **Improved reliability** - Smart error handling vs aggressive timeouts

### Metrics to Monitor

**Server Metrics:**

- WebSocket connection count stability
- CPU usage during peak tab switching
- Memory usage per connection
- Error rates in WebSocket handling

**Client Metrics:**

- Time to reconnection after suspension
- Success rate of automatic recovery
- User-perceived downtime
- Battery impact on mobile devices

## 📝 Configuration Options

### Server Tuning

```go
// In feed_notifier.go - adjust these values based on usage patterns
ticker := time.NewTicker(5 * time.Minute)     // Ping interval
conn.Conn.SetWriteDeadline(30 * time.Second)  // Write timeout
case <-time.After(30 * time.Second):          // Feed update timeout
```

### Client Tuning

```javascript
// In websocket-manager.js - customize reconnection behavior
maxReconnectAttempts: 5,      // How many times to retry
reconnectDelay: 1000,         // Initial delay between attempts
maxReconnectDelay: 30000,     // Maximum delay (exponential backoff)
```

## 🚀 Deployment Checklist

- [ ] **Server changes deployed** - Updated `feed_notifier.go`
- [ ] **Client scripts deployed** - New `websocket-manager.js`
- [ ] **Templates updated** - Include WebSocket manager in layout
- [ ] **Main application tested** - Verify feed counter WebSocket functionality
- [ ] **Monitoring enabled** - Track connection metrics
- [ ] **Load testing completed** - Verify performance under load
- [ ] **Mobile testing done** - Test on actual mobile devices
- [ ] **Documentation updated** - Team trained on new behavior

## 🎉 Success Criteria

### ✅ Technical Success

- [ ] No timeout disconnections during normal tab switching
- [ ] Automatic reconnection within 5 seconds of tab restoration
- [ ] Stable connection count during suspension periods
- [ ] Reduced error rates in server logs
- [ ] Feed updates resume immediately after suspension

### ✅ User Experience Success

- [ ] Feed counter updates consistently across tab switches
- [ ] No user intervention required for connection recovery
- [ ] Mobile app switching works seamlessly
- [ ] Improved perceived reliability of real-time features
- [ ] Reduced user complaints about missing notifications

## 📞 Support

**For issues with this implementation:**

1. **Check browser console first:** Use `wsManagerDebug.stats()` for diagnostics
2. **Review server logs:** Look for WebSocket connection patterns
3. **Browser developer tools:** Check console for JavaScript errors
4. **Network monitoring:** Verify WebSocket upgrade requests succeed

**Debug Commands:**

```javascript
// Browser console debugging
wsManagerDebug.stats(); // Connection statistics
wsManagerDebug.reconnectAll(); // Force reconnection
window.webSocketManager.getStats(); // Detailed manager state
```

This fix provides a robust, production-ready solution for WebSocket suspension handling that improves both user experience and system reliability.
