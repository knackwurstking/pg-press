# Trouble Reports PDF Share Feature

## Overview

The PDF Share feature allows users to generate and download a comprehensive PDF document containing all the details of a trouble report. This feature is designed to facilitate sharing trouble reports with stakeholders who may not have direct access to the PG-VIS system.

## How to Use

1. Navigate to the trouble reports page
2. Locate the trouble report you want to share
3. In the actions section (right side of each trouble report), click the blue share button (📤 icon)
4. The system will generate a PDF and either:
    - **HTTPS + Modern browsers with file sharing support**: Open the native share dialog to share via installed apps (email, messaging, cloud storage, etc.)
    - **HTTP connections or unsupported browsers**: Automatically download the PDF to your device (no sharing dialog)

## Troubleshooting

### Common Issues and Solutions

#### 1. PDF Downloads Instead of Sharing Dialog

**Most common cause**: Not using HTTPS connection
**What happens**: System skips Web Share API entirely and downloads PDF directly
**Solution**: Access the site via HTTPS for native sharing experience
**Note**: This is intentional behavior - HTTP connections cannot securely share files

#### 2. "The request is not allowed by the user agent or the platform in the current context"

**Cause**: Web Share API security restrictions (only occurs on HTTPS)
**Solutions**:

- **User Gesture**: Make sure you're clicking the button directly (not automated)
- **Browser Context**: Some browsers restrict sharing in certain contexts
- **Browser Permissions**: Check site permissions in browser settings

#### 3. Share button only shares text without PDF file

**Cause**: Browser supports Web Share API but not file sharing
**Solutions**:

- **Check browser support**: Click the orange debug button (🐛) to test capabilities
- **Use mobile browser**: Mobile browsers have better file sharing support
- **Expected behavior**: PDF downloads first, then text sharing dialog opens

#### 4. Debug Button Shows "Permission Denied"

**Cause**: Browser security policies or user settings
**Solutions**:

- Check browser permissions for the site
- Try a different browser or device
- Ensure you're on HTTPS
- Clear browser cache and cookies

#### 5. PDF Downloads Instead of Sharing

**Expected behavior in**:

- Desktop Chrome/Firefox (file sharing not supported)
- HTTP connections (security limitation)
- Browsers without Web Share API support

### Using the Debug Button

1. **Click the orange debug button (🐛)** next to any share button
2. **Check the console output** for detailed capability information
3. **Read the alert messages** for your browser's specific capabilities
4. **Test results**:
    - ✅ "Web Share API supports PDF files!" = File sharing should work
    - ❌ "Does not support PDF files" = Will download instead
    - ❌ "Web Share API not supported" = Download-only fallback

### Browser-Specific Notes

- **Mobile Safari/Chrome**: Best support for file sharing
- **Desktop browsers**: Limited file sharing, automatic download fallback
- **Firefox**: No Web Share API support, always downloads
- **Older browsers**: Automatic download fallback

## PDF Content

The generated PDF includes the following information:

### Header Section

- Document title: "Fehlerbericht" (Trouble Report)
- Report ID (e.g., "Report-ID: #123")

### Main Content

- **Title**: The trouble report title in a bordered section
- **Content**: The full content/description of the trouble report in a bordered section

### Metadata Section

- **Creation Date**: When the trouble report was first created
- **Creator**: Username of the person who created the report
- **Last Modified**: Date and time of the most recent modification (if different from creation)
- **Last Modifier**: Username of the person who made the last modification
- **Number of Modifications**: Total count of changes made to the report

### Attachments Section (if applicable)

- **Attachment Count**: Total number of attachments
- **Image Attachments**: Actual images embedded in PDF with:
    - Full-size images scaled to fit page width (max 100mm width)
    - Image captions showing attachment number and MIME type
    - Support for JPG, PNG, and GIF formats
    - Automatic page breaks for large images
    - Error handling for corrupted/unsupported image formats
- **Other Attachment Details**: For non-image attachments:
    - Attachment number
    - MIME type
    - Category (Document, Archive, or Other)

### Footer

- Document generation timestamp
- Note indicating the document was automatically generated from PG-VIS

### Technical Implementation

### Dependencies

- **gofpdf/v2**: PDF generation library (`github.com/jung-kurt/gofpdf/v2`)
- **Web Share API**: Browser native sharing functionality
- **Fetch API**: For asynchronous PDF retrieval

### Character Encoding

- **German Umlauts**: Automatic conversion for PDF compatibility
    - `ä` → `ae`, `ö` → `oe`, `ü` → `ue`
    - `Ä` → `Ae`, `Ö` → `Oe`, `Ü` → `Ue`
    - `ß` → `ss`
- **Rationale**: Standard PDF fonts require ASCII-compatible characters
- **Scope**: Applies to all text content including titles, content, usernames, and error messages

### Route

- **Endpoint**: `GET /trouble-reports/share-pdf?id={trouble_report_id}`
- **Handler**: `DataHandler.handleGetSharePdf`

### Method

JavaScript fetch request with Web Share API integration

- **Debug endpoint**: Orange debug button for testing browser capabilities

### File Naming Convention

Generated PDF files follow this naming pattern:

```
fehlerbericht_{report_id}_{date}.pdf
```

Example: `fehlerbericht_123_2024-01-15.pdf`

### Security Considerations

- Users must be authenticated to access this feature
- The same access controls that apply to viewing trouble reports also apply to PDF generation
- All PDF generation activities are logged for audit purposes

### Styling and Layout

- **Page Format**: A4 Portrait
- **Margins**: 20mm on all sides
- **Fonts**: Arial family with varying sizes and weights
- **Colors**:
    - Header text: Dark blue (#003366)
    - Section headers: Blue background (#F0F8FF)
    - Metadata: Gray text (#808080)
    - Image captions: Italic gray text
- **Sections**: Bordered layout with clear visual separation
- **Images**:
    - Maximum width: 100mm (scaled to maintain aspect ratio)
    - Centered horizontally on page
    - Automatic page breaks for oversized content
    - Captions below each image

### Performance Considerations

- **Image Processing**: Large images are automatically scaled to optimize PDF size
- **Memory Usage**: Multiple high-resolution images may increase memory consumption during PDF generation
- **File Size Impact**: PDFs with embedded images will be significantly larger than text-only reports
- **Processing Time**: Generation time increases with number and size of image attachments
- **Recommended Limits**:
    - Keep image attachments under 5MB each for optimal performance
    - Maximum 10 image attachments per trouble report
    - Consider image compression before uploading for faster PDF generation

### Error Handling

- Invalid or non-existent trouble report IDs return appropriate error responses
- PDF generation failures are logged and return user-friendly error messages
- Network timeouts and database errors are handled gracefully
- Web Share API failures automatically fall back to download
- JavaScript errors display user-friendly alert messages
- Button loading states prevent multiple simultaneous requests
- Image processing errors are handled gracefully with fallback text descriptions
- Corrupted or unsupported image formats display error messages instead of breaking PDF generation
- German character encoding issues are prevented by automatic ASCII conversion

### Debug Features

- **Debug Button**: Orange bug icon (🐛) next to each share button (development mode only)
- **Console Logging**: Detailed logs for troubleshooting sharing issues
- **Capability Testing**: Tests browser's Web Share API and file sharing support
- **Test File Sharing**: Allows testing with a small sample PDF file

## Logging

The feature includes comprehensive logging:

- PDF generation requests (with report ID)
- Successful PDF creation (with file size)
- Error conditions during PDF generation or data retrieval

## Browser Compatibility

### Web Share API File Sharing Support

| Connection + Browser              | Text Sharing | File Sharing | Experience                     |
| --------------------------------- | ------------ | ------------ | ------------------------------ |
| **HTTPS + Mobile Chrome/Safari**  | ✅           | ✅           | Native share with PDF file     |
| **HTTPS + Desktop Chrome/Safari** | ✅           | ❌           | Downloads PDF, shares text     |
| **HTTPS + Firefox**               | ❌           | ❌           | Downloads PDF only             |
| **HTTP + Any Browser**            | ❌           | ❌           | Downloads PDF only (by design) |
| **Other browsers**                | Varies       | ❌           | Downloads PDF only             |

### Detection and Behavior

- **HTTPS Check**: System first checks if connection is secure (HTTPS only)
- **HTTP Connections**: Skip Web Share API entirely, download PDF directly
- **HTTPS Connections**: Check `navigator.canShare()` for file support, fallback to download if needed
- **Debug mode**: Orange debug button (🐛) tests browser capabilities and explains expected behavior

### User Experience Patterns

1. **Best case** (HTTPS + Mobile Chrome/Safari): Native share dialog with PDF file attached
2. **Partial support** (HTTPS + Desktop Chrome/Safari): PDF downloads + text sharing dialog
3. **Direct download** (HTTP connections): PDF downloads immediately, no sharing dialog
4. **Fallback** (HTTPS + Firefox/older browsers): PDF downloads automatically
5. **Error handling**: User-friendly messages with retry options

## Future Enhancements

Potential improvements for future versions:

- ✅ Include actual images in PDF (implemented)
- Add company branding/logo to the PDF header
- Support for custom PDF templates
- Batch PDF generation for multiple reports
- Enhanced Web Share API integration with custom share targets
- PDF password protection options
- Progress indicators for large PDF generation
- Offline PDF generation capabilities
- Custom share messages and metadata
- Remove debug button in production builds
- Add browser-specific optimization hints
- Image compression options for smaller PDF file sizes
- Thumbnail generation for faster PDF processing
- Progress indicators for large PDF generation with many images
- UTF-8 font support for native German character display
- Configurable character encoding options
