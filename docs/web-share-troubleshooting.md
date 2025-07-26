# Web Share API Troubleshooting Guide

## Overview

This guide helps resolve common issues with the PDF sharing feature in PG-VIS that uses the Web Share API.

## Quick Diagnosis

### Step 1: Check Your Environment

- **Protocol**: Are you using HTTPS? (Required for Web Share API)
- **Browser**: Which browser and version are you using?
- **Device**: Mobile or desktop?

### Step 2: Use the Debug Button

1. Add `?debug=true` to your URL or access via localhost
2. Click the orange debug button (🐛) next to any share button
3. Check the console output and alert messages

## Common Error Messages

### "The request is not allowed by the user agent or the platform in the current context"

**What it means**: Web Share API security restriction

**Root causes**:

1. **HTTP instead of HTTPS** - Web Share API requires secure context
2. **No user gesture** - Share must be triggered by direct user click
3. **Browser restrictions** - Some contexts don't allow sharing

**Solutions**:

- ✅ Access site via HTTPS (not HTTP)
- ✅ Click the share button directly (don't use automation)
- ✅ Try a different browser or device
- ✅ Check browser permissions for the site

### "Web Share API not supported"

**What it means**: Your browser doesn't support Web Share API

**Expected behavior**: PDF will download automatically

**Affected browsers**:

- Firefox (all versions)
- Internet Explorer
- Older browsers

**No action needed** - the download fallback works fine.

### "Cannot share this file type"

**What it means**: Browser supports Web Share API but not file sharing

**Common scenarios**:

- Desktop Chrome/Edge (file sharing only on mobile)
- Older Safari versions

**Expected behavior**: PDF downloads, you can share manually

## Browser Support Matrix

| Platform               | Text Sharing | File Sharing | Experience                  |
| ---------------------- | ------------ | ------------ | --------------------------- |
| **Mobile Safari 14+**  | ✅           | ✅           | Native share sheet with PDF |
| **Mobile Chrome 89+**  | ✅           | ✅           | Android share menu with PDF |
| **Desktop Safari 14+** | ✅           | ❌           | Download PDF                |
| **Desktop Chrome 89+** | ✅           | ❌           | Download PDF                |
| **Firefox (all)**      | ❌           | ❌           | Download PDF                |

## HTTPS Requirements

### Why HTTPS is Required

- Web Share API is a secure context feature
- Browsers block it on HTTP for security reasons
- File sharing especially requires HTTPS

### Check Your Connection

```javascript
// In browser console:
console.log("Protocol:", location.protocol);
console.log("HTTPS:", location.protocol === "https:");
```

### Development Setup

- Use HTTPS proxy for testing
- Use `ngrok` or similar for HTTPS testing
- Set up local HTTPS certificates for development

## Troubleshooting Steps

### Step 1: Environment Check

```bash
# Check if you're on HTTPS
echo $URL | grep https

# Or in browser console:
console.log(location.protocol);
```

### Step 2: Browser Capability Test

1. Open browser developer tools (F12)
2. Go to Console tab
3. Run debug test:

```javascript
// Test Web Share API
console.log("Web Share API:", "share" in navigator);
console.log("Can Share method:", "canShare" in navigator);

// Test file sharing
if (navigator.canShare) {
    const testFile = new File(["test"], "test.pdf", {
        type: "application/pdf",
    });
    console.log("Can share files:", navigator.canShare({ files: [testFile] }));
}
```

### Step 3: Test Actual Sharing

1. Click the blue share button
2. Check browser console for error messages
3. Note the behavior (native share dialog vs download)

### Step 4: Fallback Verification

- Ensure PDF downloads work if sharing fails
- Check downloaded file opens correctly
- Verify file has correct content

## Expected Behaviors

### ✅ Working File Sharing

- Click share button
- Native share dialog opens
- PDF file appears in share options
- Can share to apps (WhatsApp, Email, etc.)

### ✅ Working Text Sharing + Download

- Click share button
- PDF downloads automatically
- Native share dialog opens for text
- Can share link/info to apps

### ✅ Download Fallback

- Click share button
- PDF downloads automatically
- Green download icon appears
- Can manually share downloaded file

## Development Tools

### Debug Button Features

- **Environment detection**: Shows browser capabilities
- **File sharing test**: Tests with sample PDF
- **Console logging**: Detailed diagnostic information
- **Error handling**: Shows specific error types

### Enabling Debug Mode

Add any of these to enable debug button:

- Access via `localhost`
- Add `?debug=true` to URL
- Access from hostname containing `dev`

### Console Debug Commands

```javascript
// Manual capability check
window.debugWebShareAPI();

// Test specific file
const testPDF = new File(["%PDF-1.4..."], "test.pdf", {
    type: "application/pdf",
});
navigator.canShare({ files: [testPDF] });

// Check security context
console.log("Secure context:", window.isSecureContext);
```

## Production Considerations

### User Communication

- Show HTTPS notice on HTTP connections
- Provide clear feedback for different behaviors
- Explain why download happens instead of sharing

### Graceful Degradation

- Always provide download fallback
- Don't rely solely on Web Share API
- Test on multiple devices/browsers

### Performance

- Cache capability detection results
- Minimize file size for sharing
- Provide progress indicators

## Common Questions

**Q: Why does it download instead of share?**
A: Your browser/device doesn't support file sharing. This is normal and expected.

**Q: Why do I need HTTPS?**
A: Web Share API is a security-sensitive feature that requires secure connections.

**Q: Can I force file sharing to work?**
A: No, it depends on browser/device capabilities. Download fallback is the reliable solution.

**Q: Does this work on all mobile devices?**
A: Most modern mobile browsers support file sharing, but there are exceptions.

## Support Contact

If you continue experiencing issues:

1. Include browser/device information
2. Include console error messages
3. Specify exact steps taken
4. Note whether debug button is available
