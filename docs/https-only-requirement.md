# HTTPS-Only Requirement for PDF Sharing

## Overview

The PDF sharing feature in PG-VIS now enforces a strict **HTTPS-only** policy for Web Share API usage. This ensures reliable, secure sharing behavior and prevents confusing user experiences with broken share dialogs.

## Policy Details

### Web Share API Usage

- **HTTPS connections**: Web Share API is attempted for native sharing
- **HTTP connections**: Web Share API is **completely skipped**, direct download only
- **localhost**: Treated as HTTP (download behavior, no Web Share API)

### Security Rationale

1. **Browser Security**: Web Share API requires secure context (HTTPS)
2. **File Sharing**: Browsers block file sharing on non-HTTPS connections
3. **User Experience**: Prevents empty share dialogs with missing PDF files
4. **Consistency**: Clear, predictable behavior based on connection type

## Behavior Matrix

| Connection Type | Share Button     | User Action | Result                       |
| --------------- | ---------------- | ----------- | ---------------------------- |
| **HTTPS**       | 📤 Share icon    | Click share | Native share dialog with PDF |
| **HTTP**        | 📥 Download icon | Click share | PDF downloads immediately    |
| **localhost**   | 📥 Download icon | Click share | PDF downloads immediately    |

## Technical Implementation

### Detection Logic

```javascript
const isHTTPS = location.protocol === "https:";
const canUseWebShare = isHTTPS; // No localhost exception

if (canUseWebShare && navigator.share && navigator.canShare) {
    // Attempt Web Share API
} else {
    // Direct download
}
```

### Visual Indicators

- **HTTPS**: Blue share button with share icon
- **HTTP/localhost**: Gray download button with download icon, dashed border
- **Tooltip changes**: "Als PDF teilen" vs "PDF herunterladen (HTTPS erforderlich für Teilen)"

### User Communication

### Console Logging

- **HTTPS**: `✅ Decision: HTTPS detected - will attempt Web Share API`
- **HTTP**: `🚫 Decision: FORCE DIRECT DOWNLOAD - HTTPS required for Web Share API`

## Development Considerations

### Testing Environments

- **localhost**: Uses download behavior (same as HTTP)
- **HTTPS development**: Requires proper SSL certificates
- **Debug tools**: Available on localhost for testing, but Web Share API still disabled

### Recommended Setup

1. Use HTTPS proxy (ngrok, cloudflare tunnel)
2. Set up local SSL certificates
3. Test on actual HTTPS domain
4. Use mobile devices with HTTPS for full feature testing

## Migration Impact

### No Breaking Changes

- Existing HTTPS deployments: No change in behavior
- HTTP deployments: Better UX (download vs broken share)
- Mobile users: Optimal experience on HTTPS
- Desktop users: Consistent download behavior

### Benefits

1. **Eliminates confusion**: No more empty share dialogs
2. **Predictable behavior**: Users know what to expect
3. **Security compliant**: Follows browser security policies
4. **Better UX**: Immediate download better than broken sharing

## Production Deployment

### Requirements

- **HTTPS must be configured** for optimal user experience
- SSL certificates properly installed
- Redirect HTTP to HTTPS (recommended)
- Test sharing functionality on mobile devices

### Monitoring

- Track usage patterns between HTTP and HTTPS users
- Monitor download success rates
- Collect feedback on sharing experience

## Debug Tools

### Available in Development

- **Orange bug button**: Test Web Share API capabilities (development mode only)
- **Console logging**: Detailed decision path

### Debug Commands

```javascript
// Check current environment
console.log("Protocol:", location.protocol);
console.log("Can use Web Share:", location.protocol === "https:");

// Test Web Share API capabilities
window.debugWebShareAPI();
```

## Common Questions

**Q: Why doesn't localhost get Web Share API?**
A: For consistency and security. Even in development, the behavior should match production HTTP deployments.

**Q: Can I override this for testing?**
A: Debug tools are available, but actual Web Share API requires HTTPS for security reasons.

**Q: What about development workflow?**
A: Use HTTPS development setup or test the download functionality (which works everywhere).

**Q: Will this affect existing users?**
A: HTTPS users see no change. HTTP users get better experience (reliable download vs broken sharing).

## Summary

This change ensures:

- ✅ Reliable sharing experience on HTTPS
- ✅ No broken share dialogs on HTTP
- ✅ Clear user expectations
- ✅ Security compliance
- ✅ Consistent behavior across environments

The HTTPS-only policy provides a better, more predictable user experience while maintaining security best practices.
