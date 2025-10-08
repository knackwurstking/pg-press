# Quick Server Path Prefix Test Guide

This guide provides a fast way to verify that the help system correctly handles the `SERVER_PATH_PREFIX` environment variable.

## Quick Tests

### Test 1: No Path Prefix (Default)

```bash
# Terminal 1: Start server without prefix
unset SERVER_PATH_PREFIX
go run ./cmd/pg-press server --addr localhost:8080

# Terminal 2: Test help page
curl -I http://localhost:8080/help/markdown
# Expected: HTTP/200 OK
```

**Browser Test:**

- Visit: `http://localhost:8080/help/markdown`
- Click "Startseite" → Should go to `http://localhost:8080/`

### Test 2: With Path Prefix

```bash
# Terminal 1: Start server with prefix
export SERVER_PATH_PREFIX="/pg-press"
go run ./cmd/pg-press server --addr localhost:8080

# Terminal 2: Test help page
curl -I http://localhost:8080/pg-press/help/markdown
# Expected: HTTP/200 OK

curl -I http://localhost:8080/help/markdown
# Expected: HTTP/404 Not Found
```

**Browser Test:**

- Visit: `http://localhost:8080/pg-press/help/markdown`
- Click "Startseite" → Should go to `http://localhost:8080/pg-press/`

### Test 3: Editor Integration

```bash
# With server running (either with or without prefix)
# Navigate to editor page and verify:
```

**Manual Steps:**

1. Find markdown checkbox: "Markdown-Formatierung verwenden"
2. Look for info icon (ⓘ) next to the description
3. Click info icon → Should open help page in new tab with correct URL
4. Enable markdown checkbox to show toolbar
5. Look for help icon (❓) in "Markdown-Werkzeuge" section
6. Click help icon → Should open help page in new tab with correct URL

## Expected Results Summary

| Scenario         | Help URL                           | Home URL     | Status |
| ---------------- | ---------------------------------- | ------------ | ------ |
| No prefix        | `/help/markdown`                   | `/`          | ✅     |
| With `/pg-press` | `/pg-press/help/markdown`          | `/pg-press/` | ✅     |
| Wrong URL        | `/help/markdown` (when prefix set) | N/A          | ❌ 404 |

## Quick Troubleshooting

**404 Error:**

- Check `SERVER_PATH_PREFIX` is set correctly
- Verify you're using the prefixed URL
- Ensure server restarted after setting prefix

**Links Don't Work:**

- Run `templ generate` after template changes
- Check browser dev tools for actual href values
- Verify `go build ./...` succeeds

**CSS/Styling Issues:**

- Check browser network tab for failed asset loads
- Verify static files are served with prefix

## One-Liner Tests

```bash
# Test no prefix
unset SERVER_PATH_PREFIX && go run ./cmd/pg-press server --addr :8080 &
sleep 3 && curl -s -o /dev/null -w "No prefix: %{http_code}\n" http://localhost:8080/help/markdown
kill %1

# Test with prefix
export SERVER_PATH_PREFIX="/test" && go run ./cmd/pg-press server --addr :8081 &
sleep 3 && curl -s -o /dev/null -w "With prefix: %{http_code}\n" http://localhost:8081/test/help/markdown
kill %1
```

## Success Criteria

✅ **All tests pass if:**

- Help page loads with correct prefix
- Navigation links include prefix
- Editor help links work with prefix
- Wrong URLs return 404
- Interactive demo works regardless of prefix

This quick test should take less than 5 minutes and verify core path prefix functionality.
