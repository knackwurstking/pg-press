#!/bin/bash

# WebSocket Suspension Fix Validation Script
# This script verifies that all components of the WebSocket suspension fix are properly implemented

set -e

echo "🔍 Validating WebSocket Suspension Fix Implementation..."
echo "=================================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Counters
PASS=0
FAIL=0
WARN=0

# Helper functions
pass() {
    echo -e "✅ ${GREEN}PASS${NC}: $1"
    PASS=$((PASS + 1))
}

fail() {
    echo -e "❌ ${RED}FAIL${NC}: $1"
    FAIL=$((FAIL + 1))
}

warn() {
    echo -e "⚠️  ${YELLOW}WARN${NC}: $1"
    WARN=$((WARN + 1))
}

info() {
    echo -e "ℹ️  ${BLUE}INFO${NC}: $1"
}

# Check if we're in the right directory
if [ ! -f "go.mod" ] || [ ! -d "routes" ]; then
    fail "Not in pg-vis root directory. Please run from project root."
    exit 1
fi

echo
echo "🔧 Checking Server-Side Implementation..."
echo "----------------------------------------"

# Check feed_notifier.go exists
if [ -f "routes/internal/notifications/feed_notifier.go" ]; then
    pass "feed_notifier.go exists"

    # Check for timeout modifications
    if grep -q "5 \* time.Minute" routes/internal/notifications/feed_notifier.go; then
        pass "Ping interval increased to 5 minutes"
    else
        fail "Ping interval not updated (should be 5 * time.Minute)"
    fi

    if grep -q "30 \* time.Second" routes/internal/notifications/feed_notifier.go; then
        pass "Write deadline increased to 30 seconds"
    else
        fail "Write deadline not updated (should be 30 * time.Second)"
    fi

    # Check for read deadline removal
    if grep -q "// conn.Conn.SetReadDeadline" routes/internal/notifications/feed_notifier.go; then
        pass "Read deadline properly commented out"
    else
        warn "Read deadline may still be active (should be commented out)"
    fi

    # Check for isTemporaryError function
    if grep -q "isTemporaryError" routes/internal/notifications/feed_notifier.go; then
        pass "isTemporaryError function implemented"
    else
        fail "isTemporaryError function missing"
    fi

    # Check for improved timeout values
    if grep -q "30 \* time.Second" routes/internal/notifications/feed_notifier.go && grep -c "30 \* time.Second" routes/internal/notifications/feed_notifier.go | grep -q "[23456789]"; then
        pass "Feed update timeouts increased to 30 seconds"
    else
        fail "Feed update timeouts not properly increased"
    fi

else
    fail "feed_notifier.go not found"
fi

echo
echo "🌐 Checking Client-Side Implementation..."
echo "----------------------------------------"

# Check websocket-manager.js exists
if [ -f "routes/assets/js/websocket-manager.js" ]; then
    pass "websocket-manager.js exists"

    # Check for key functions
    if grep -q "class WebSocketManager" routes/assets/js/websocket-manager.js; then
        pass "WebSocketManager class defined"
    else
        fail "WebSocketManager class not found"
    fi

    if grep -q "handleVisibilityChange" routes/assets/js/websocket-manager.js; then
        pass "handleVisibilityChange method implemented"
    else
        fail "handleVisibilityChange method missing"
    fi

    if grep -q "exponential backoff" routes/assets/js/websocket-manager.js; then
        pass "Exponential backoff documented"
    else
        warn "Exponential backoff logic may be missing"
    fi

    if grep -q "checkAndReconnectAll" routes/assets/js/websocket-manager.js; then
        pass "checkAndReconnectAll method implemented"
    else
        fail "checkAndReconnectAll method missing"
    fi

    # Check for HTMX integration
    if grep -q "htmx:wsOpen" routes/assets/js/websocket-manager.js; then
        pass "HTMX WebSocket event handling implemented"
    else
        fail "HTMX WebSocket event integration missing"
    fi

else
    fail "websocket-manager.js not found"
fi

# Check main layout integration
if [ -f "routes/templates/layouts/main.html" ]; then
    pass "main.html layout found"

    if grep -q "websocket-manager.js" routes/templates/layouts/main.html; then
        pass "websocket-manager.js included in main layout"
    else
        fail "websocket-manager.js not included in main layout"
    fi

    # Check script loading order
    if grep -n "htmx-ext-ws" routes/templates/layouts/main.html | head -1 | cut -d: -f1 | \
       awk -v ws_line="$(grep -n "websocket-manager.js" routes/templates/layouts/main.html | cut -d: -f1)" \
       'BEGIN{getline; if($1 < ws_line) print "correct"; else print "incorrect"}' | grep -q "correct"; then
        pass "Script loading order is correct (HTMX before WebSocket Manager)"
    else
        warn "Check script loading order (HTMX extensions should load before WebSocket Manager)"
    fi

else
    fail "main.html layout not found"
fi

# Check PWA manager updates
if [ -f "routes/assets/js/pwa-manager.js" ]; then
    pass "pwa-manager.js exists"

    if grep -q "webSocketManager" routes/assets/js/pwa-manager.js; then
        pass "PWA manager integrates with WebSocket manager"
    else
        warn "PWA manager may not be coordinating with WebSocket manager"
    fi

    if grep -q "visibilitychange" routes/assets/js/pwa-manager.js && \
       ! grep -q "//document.addEventListener.*visibilitychange" routes/assets/js/pwa-manager.js; then
        pass "Visibility change handling enabled in PWA manager"
    else
        fail "Visibility change handling not properly enabled"
    fi

else
    warn "pwa-manager.js not found (optional but recommended)"
fi

echo
echo "🧪 Checking Test Infrastructure..."
echo "---------------------------------"

# Check documentation
if [ -f "WEBSOCKET_SUSPENSION_FIX.md" ]; then
    pass "Documentation exists (WEBSOCKET_SUSPENSION_FIX.md)"
else
    warn "Documentation file not found (recommended)"
fi

echo
echo "🔨 Checking Build and Dependencies..."
echo "------------------------------------"

# Check Go build
info "Testing Go build..."
if go build -o /tmp/pg-vis-validate ./cmd/pg-vis 2>/dev/null; then
    pass "Go application builds successfully"
    rm -f /tmp/pg-vis-validate
else
    fail "Go build failed - check for compilation errors"
fi

# Check Go modules
if go mod tidy && [ $? -eq 0 ]; then
    pass "Go modules are clean"
else
    warn "Go modules may need attention"
fi

# Check for golang.org/x/net/websocket dependency
if grep -q "golang.org/x/net" go.mod; then
    pass "golang.org/x/net dependency present"
else
    fail "golang.org/x/net dependency missing"
fi

# Check that gorilla/websocket is not present
if ! grep -q "gorilla/websocket" go.mod; then
    pass "gorilla/websocket dependency removed"
else
    warn "gorilla/websocket dependency still present (should be removed)"
fi

echo
echo "🔍 Running Static Analysis..."
echo "-----------------------------"

# Check for potential issues with staticcheck if available
if command -v staticcheck >/dev/null 2>&1; then
    info "Running staticcheck analysis..."
    if staticcheck ./... 2>/dev/null; then
        pass "No static analysis issues found"
    else
        warn "Static analysis found potential issues"
    fi
else
    info "staticcheck not available (install with: go install honnef.co/go/tools/cmd/staticcheck@latest)"
fi

# Check with go vet
if go vet ./... 2>/dev/null; then
    pass "go vet found no issues"
else
    warn "go vet found potential issues"
fi

echo
echo "📋 Checking File Permissions and Structure..."
echo "---------------------------------------------"

# Check that JavaScript files are readable
if [ -r "routes/assets/js/websocket-manager.js" ]; then
    pass "websocket-manager.js is readable"
else
    fail "websocket-manager.js permission issues"
fi

# Check for any leftover test files or temporary files
if find . -name "*.tmp" -o -name "*.bak" -o -name "*~" | grep -q .; then
    warn "Temporary files found - consider cleaning up"
else
    pass "No temporary files found"
fi

# Check directory structure
expected_dirs=("routes/assets/js" "routes/templates/layouts" "routes/internal/notifications")
for dir in "${expected_dirs[@]}"; do
    if [ -d "$dir" ]; then
        pass "Directory $dir exists"
    else
        fail "Directory $dir missing"
    fi
done

echo
echo "🎯 Functional Validation Suggestions..."
echo "--------------------------------------"

info "Manual testing steps:"
echo "  1. Start server: go run ./cmd/pg-vis"
echo "  2. Open main application: http://localhost:8080/"
echo "  3. Navigate to a page with feed counter and switch tabs for 60+ seconds"
echo "  4. Return to tab - connection should restore automatically"
echo "  5. Create test feeds: curl -X POST 'http://localhost:8080/test/feed/create?type=user_add&count=1'"

info "Browser console commands for debugging:"
echo "  - wsManagerDebug.stats() - Show connection statistics"
echo "  - wsManagerDebug.reconnectAll() - Force reconnection"
echo "  - window.webSocketManager.getStats() - Detailed state"

echo
echo "📊 Validation Summary"
echo "===================="
echo -e "✅ ${GREEN}Passed${NC}: $PASS"
echo -e "❌ ${RED}Failed${NC}: $FAIL"
echo -e "⚠️  ${YELLOW}Warnings${NC}: $WARN"

if [ $FAIL -eq 0 ]; then
    echo
    echo -e "🎉 ${GREEN}All critical checks passed!${NC}"
    echo "The WebSocket suspension fix appears to be properly implemented."
    echo "Proceed with manual testing using the main application."
    exit 0
else
    echo
    echo -e "⚠️  ${RED}Some checks failed.${NC}"
    echo "Please review the failed items above before proceeding."
    exit 1
fi
