# Attachment Usage Guide

This guide provides practical examples for using the attachment functionality in PG-VIS trouble reports.

## Basic Usage Examples

### Creating a Trouble Report with Attachments

1. **Open the Create Dialog**

    ```
    GET /trouble-reports/dialog-edit
    ```

2. **Upload Files via Web Interface**
    - Click the upload area or drag files directly
    - Select multiple files (up to 10 files, 10MB each)
    - Preview shows selected files with validation
    - Submit form to create report with attachments

3. **Example Form Data**
    ```html
    <form enctype="multipart/form-data" method="POST">
        <input name="title" value="Server Hardware Issue" />
        <input
            name="content"
            value="Server experiencing intermittent failures"
        />
        <input type="file" name="attachments" multiple />
        <input type="hidden" name="attachment_order" value="" />
    </form>
    ```

### Managing Existing Attachments

1. **Edit Existing Report**

    ```
    GET /trouble-reports/dialog-edit?id=123
    ```

2. **Reorder Attachments**
    - Drag attachments using the grip handle
    - Order is automatically saved
    - Uses hidden field `attachment_order` with comma-separated IDs

3. **Delete Specific Attachment**
    ```
    DELETE /trouble-reports/attachments?id=123&attachment_id=screenshot_1234567890_0
    ```

## API Usage Examples

### Download Attachment

```bash
# Download a specific attachment
curl -H "Cookie: pgvis-api-key=your-api-key" \
     "https://your-server/trouble-reports/attachments?id=123&attachment_id=error_log_1234567890_0" \
     -o downloaded_file.txt
```

### Reorder Attachments

```bash
# Reorder attachments via API
curl -X POST \
     -H "Cookie: pgvis-api-key=your-api-key" \
     -d "new_order=file1_id,file3_id,file2_id" \
     "https://your-server/trouble-reports/attachments/reorder?id=123"
```

### Upload via API

```bash
# Create trouble report with attachments
curl -X POST \
     -H "Cookie: pgvis-api-key=your-api-key" \
     -F "title=Network Issue" \
     -F "content=Detailed description of the network problem" \
     -F "attachments=@screenshot.png" \
     -F "attachments=@error_log.txt" \
     "https://your-server/trouble-reports/dialog-edit"
```

## Supported File Types

### Images

- **JPEG/JPG**: `.jpg`, `.jpeg` - Photos, screenshots
- **PNG**: `.png` - Screenshots, diagrams
- **GIF**: `.gif` - Animated images, simple graphics
- **SVG**: `.svg` - Vector graphics, diagrams
- **WebP**: `.webp` - Modern image format
- **BMP**: `.bmp` - Bitmap images

### Documents

- **PDF**: `.pdf` - Reports, documentation
- **Word**: `.doc`, `.docx` - Text documents
- **Text**: `.txt` - Plain text files, logs
- **RTF**: `.rtf` - Rich text format
- **OpenDocument**: `.odt` - Open office documents

### Archives

- **ZIP**: `.zip` - Compressed archives
- **RAR**: `.rar` - WinRAR archives
- **7-Zip**: `.7z` - 7-Zip archives
- **TAR**: `.tar` - Unix archives
- **GZIP**: `.gz` - Compressed files
- **BZIP2**: `.bz2` - Compressed files

## Configuration Examples

### Customizing File Limits

To modify attachment limits, update constants in `pgvis/attachment.go`:

```go
const (
    MaxAttachmentDataSize = 20 * 1024 * 1024 // Change to 20MB
    MaxAttachmentIDLength = 500              // Longer IDs
)
```

### Adding New MIME Types

Extend the `getMimeTypeFromFilename` function in `dialog-edit.go`:

```go
switch ext {
case ".mp4":
    return "video/mp4"
case ".xlsx":
    return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
// Add more types as needed
}
```

## Common Use Cases

### 1. Bug Report with Screenshots

```html
<!-- Typical bug report scenario -->
Title: "Login button not working" Content: "Users cannot log in on mobile
devices" Attachments: - mobile_screenshot.png (showing the issue) -
browser_console.txt (error logs) - network_trace.har (network analysis)
```

### 2. Hardware Issue Documentation

```html
Title: "Server temperature warning" Content: "Server room temperature exceeded
normal range" Attachments: - temperature_graph.pdf - sensor_data.csv -
server_photo.jpg
```

### 3. Software Deployment Issues

```html
Title: "Deployment failed on production" Content: "Application deployment script
failed" Attachments: - deployment_log.txt - configuration.zip -
error_screenshot.png
```

## JavaScript Integration Examples

### Custom File Validation

```javascript
// Add custom validation before upload
document.getElementById("attachments").addEventListener("change", function (e) {
    const files = Array.from(e.target.files);

    // Custom validation
    const invalidFiles = files.filter((file) => {
        return file.name.includes("temp") || file.name.startsWith("~");
    });

    if (invalidFiles.length > 0) {
        alert("Temporary files are not allowed");
        e.target.value = "";
        return;
    }

    // Proceed with normal handling
    handleFileSelect(e);
});
```

### Progress Tracking

```javascript
// Add upload progress tracking
function trackUploadProgress() {
    const form = document.querySelector("form");

    form.addEventListener("submit", function (e) {
        // Show progress indicator
        const progressDiv = document.createElement("div");
        progressDiv.innerHTML = "Uploading attachments...";
        progressDiv.className = "upload-progress";
        form.appendChild(progressDiv);
    });
}
```

## Troubleshooting

### Common Issues

1. **File Too Large**

    ```
    Error: "attachment exceeds maximum size limit (10MB)"
    Solution: Compress files or split large files
    ```

2. **Too Many Files**

    ```
    Error: "too many attachments (maximum 10 allowed)"
    Solution: Remove unnecessary files or combine related files
    ```

3. **Unsupported File Type**

    ```
    Error: Browser rejects file selection
    Solution: Check file extension against supported types
    ```

4. **Upload Timeout**
    ```
    Issue: Large files timeout during upload
    Solution: Increase server timeout or use smaller files
    ```

### Debugging Steps

1. **Check Browser Console**

    ```javascript
    // Enable debug logging
    console.log(
        "Selected files:",
        document.getElementById("attachments").files,
    );
    ```

2. **Verify MIME Types**

    ```bash
    # Check file MIME type
    file --mime-type filename.ext
    ```

3. **Test File Size**
    ```javascript
    // Check file sizes in browser
    Array.from(files).forEach((file) => {
        console.log(`${file.name}: ${file.size} bytes`);
    });
    ```

## Security Considerations

### File Upload Security

1. **MIME Type Validation**
    - Server validates both extension and content
    - Multiple validation layers prevent bypass

2. **Size Limits**
    - Hard limits prevent DoS attacks
    - Client and server-side validation

3. **Content Scanning**
    - Files stored as binary data in database
    - No direct file system access

### Best Practices

1. **User Training**
    - Educate users on acceptable file types
    - Provide clear guidelines for file naming

2. **Regular Cleanup**
    - Monitor database size growth
    - Archive old reports with large attachments

3. **Access Control**
    - Verify user permissions before file access
    - Log all file operations for audit

## Performance Optimization

### Database Considerations

```sql
-- Monitor attachment storage usage
SELECT
    COUNT(*) as report_count,
    SUM(LENGTH(linked_attachments)) as total_attachment_size
FROM trouble_reports;
```

### Client-Side Optimization

```javascript
// Compress images before upload (example)
function compressImage(file, maxWidth = 1920) {
    return new Promise((resolve) => {
        const canvas = document.createElement("canvas");
        const ctx = canvas.getContext("2d");
        const img = new Image();

        img.onload = () => {
            const ratio = Math.min(maxWidth / img.width, maxWidth / img.height);
            canvas.width = img.width * ratio;
            canvas.height = img.height * ratio;

            ctx.drawImage(img, 0, 0, canvas.width, canvas.height);
            canvas.toBlob(resolve, "image/jpeg", 0.8);
        };

        img.src = URL.createObjectURL(file);
    });
}
```

This guide provides comprehensive examples for implementing and using the attachment functionality effectively in your PG-VIS deployment.
