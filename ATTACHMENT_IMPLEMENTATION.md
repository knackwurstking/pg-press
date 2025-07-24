# Attachment Implementation for Trouble Reports

## Overview

This document describes the implementation of file attachment functionality for trouble reports in the PG-VIS application. The system allows users to upload, manage, and download attachments associated with trouble reports with comprehensive validation and security features.

## Features Implemented

### Core Functionality

- **File Upload**: Support for multiple file uploads with drag-and-drop interface
- **File Management**: View, download, reorder, and delete attachments
- **Validation**: Comprehensive file validation including size, type, and count limits
- **Security**: MIME type validation and secure file handling
- **User Interface**: Intuitive drag-and-drop interface with file previews

### Technical Specifications

- **Maximum File Size**: 10MB per attachment
- **Maximum Attachments**: 10 attachments per trouble report
- **Supported File Types**:
    - Images: JPG, JPEG, PNG, GIF, BMP, SVG, WebP
    - Documents: PDF, DOC, DOCX, TXT, RTF, ODT
    - Archives: ZIP, RAR, 7Z, TAR, GZ, BZ2

## API Endpoints

### Main Trouble Report Endpoints (Updated)

- `GET /trouble-reports/dialog-edit` - Display edit dialog with attachment support
- `POST /trouble-reports/dialog-edit` - Create new trouble report with attachments
- `PUT /trouble-reports/dialog-edit?id={id}` - Update existing trouble report with attachments

### New Attachment Management Endpoints

- `GET /trouble-reports/attachments?id={report_id}&attachment_id={attachment_id}` - Download/view attachment
- `DELETE /trouble-reports/attachments?id={report_id}&attachment_id={attachment_id}` - Delete specific attachment
- `POST /trouble-reports/attachments/reorder` - Reorder attachments

## Database Schema Changes

The existing `trouble_reports` table already supports attachments through the `linked_attachments` BLOB field, which stores JSON-serialized attachment data. No schema changes were required.

### Attachment Data Structure

```json
{
    "id": "unique_attachment_identifier",
    "mime_type": "application/pdf",
    "data": "base64_encoded_file_data"
}
```

## Backend Implementation

### Updated Files

#### `dialog-edit.go`

- Enhanced `EditDialogTemplateData` with attachment error handling
- Updated form validation to process file uploads
- Added `processAttachments()` function for handling file uploads
- Added `processFileUpload()` for individual file processing
- Added attachment management handlers:
    - `handlePostAttachmentReorder()`
    - `handleDeleteAttachment()`
    - `handleGetAttachment()`

#### `routes/constants/routes.go`

Added new form field constants:

```go
AttachmentsFormField     = "attachments"
AttachmentOrderField     = "attachment_order"
ExistingAttachmentPrefix = "existing_attachment_"
```

#### `troublereports.go`

- Added new route registrations for attachment management
- Added error message constants for attachment operations

#### `routes/internal/utils/utils.go`

- Added `GetCurrentTimeMillis()` utility function for unique ID generation

### Key Implementation Details

#### File Upload Processing

```go
func (h *Handler) processFileUpload(fileHeader *multipart.FileHeader, index int) (*pgvis.Attachment, error) {
    // Size validation
    if fileHeader.Size > pgvis.MaxAttachmentDataSize {
        return nil, fmt.Errorf("file %s is too large (max 10MB)", fileHeader.Filename)
    }

    // Read file data
    file, err := fileHeader.Open()
    if err != nil {
        return nil, fmt.Errorf("failed to open file: %w", err)
    }
    defer file.Close()

    data, err := io.ReadAll(file)
    if err != nil {
        return nil, fmt.Errorf("failed to read file: %w", err)
    }

    // Generate unique ID and detect MIME type
    attachmentID := h.generateUniqueAttachmentID(fileHeader.Filename, index)
    mimeType := h.detectMimeType(fileHeader, data)

    // Create and validate attachment
    attachment := &pgvis.Attachment{
        ID:       attachmentID,
        MimeType: mimeType,
        Data:     data,
    }

    return attachment, attachment.Validate()
}
```

#### MIME Type Detection

The system uses a multi-layered approach for MIME type detection:

1. HTTP header Content-Type
2. Go's built-in `http.DetectContentType()`
3. File extension-based fallback mapping

#### Unique ID Generation

Attachment IDs are generated using:

- Sanitized filename (without extension)
- Current timestamp in milliseconds
- Index counter
- Format: `{sanitized_filename}_{timestamp}_{index}`

## Frontend Implementation

### Updated Template Features

#### Enhanced Dialog (`dialog-edit.html`)

- **Drag-and-Drop Zone**: Visual file upload area with hover effects
- **File Preview**: Shows selected files before upload with size validation
- **Existing Attachments**: Displays current attachments with management options
- **Sortable Interface**: Drag-to-reorder existing attachments using SortableJS
- **Validation Feedback**: Real-time file size and count validation

#### JavaScript Functions

```javascript
// File selection and drag-drop handling
function handleFileSelect(event)
function handleFileDrop(event)
function handleDragOver(event)
function handleDragLeave(event)

// Attachment management
function viewAttachment(reportId, attachmentId)
function deleteAttachment(reportId, attachmentId)
function removeFileFromPreview(index)

// Utility functions
function formatFileSize(bytes)
function updateAttachmentOrderInput()
```

### User Interface Features

#### Visual Indicators

- **File Type Icons**: Different Bootstrap icons for images, documents, archives
- **File Size Display**: Human-readable file sizes with error highlighting
- **Drag Handles**: Visual grip indicators for reordering
- **Progress Feedback**: Loading states during upload/delete operations

#### Validation Messages

- File size exceeded warnings
- File type restriction notifications
- Maximum attachment count limits
- Real-time validation during file selection

## Security Considerations

### File Validation

- **Size Limits**: Strict 10MB per file limit
- **Type Validation**: MIME type checking with whitelist approach
- **Content Validation**: Uses Go's content detection for additional security
- **Count Limits**: Maximum 10 attachments per report

### Data Handling

- **Binary Storage**: Files stored as binary data in database
- **Sanitized Filenames**: Special characters removed from attachment IDs
- **Secure Downloads**: Proper Content-Type and Content-Disposition headers

### Access Control

- **Authentication Required**: All attachment operations require valid user session
- **Report Ownership**: Users can only manage attachments on accessible reports
- **Admin Privileges**: Deletion operations may require admin rights (configurable)

## Usage Examples

### Creating a Trouble Report with Attachments

1. Open trouble report creation dialog
2. Fill in title and content
3. Drag files to upload area or click to select
4. Review file preview and remove unwanted files
5. Submit form - attachments are processed and stored

### Managing Existing Attachments

1. Open existing trouble report for editing
2. View current attachments in sortable list
3. Drag to reorder attachments
4. Click "View" to download/preview attachments
5. Click "Delete" to remove unwanted attachments
6. Add new files if needed
7. Submit to save changes

### Downloading Attachments

Access attachments via:

```
GET /trouble-reports/attachments?id={report_id}&attachment_id={attachment_id}
```

## Limitations and Considerations

### Current Limitations

- **Storage**: Files stored in database as BLOB data (consider file system storage for large deployments)
- **Processing**: No virus scanning or advanced content validation
- **Thumbnails**: No automatic thumbnail generation for images
- **Compression**: No automatic file compression

### Performance Considerations

- **Database Size**: Large attachments increase database size significantly
- **Memory Usage**: File processing loads entire files into memory
- **Transfer Speed**: Large files may cause timeout issues on slow connections

### Future Enhancements

- **File System Storage**: Move to file system with database references
- **Image Processing**: Automatic thumbnail generation and resizing
- **Virus Scanning**: Integration with antivirus scanning
- **Cloud Storage**: Support for cloud storage providers (AWS S3, etc.)
- **Preview Generation**: PDF and document preview generation
- **Compression**: Automatic file compression for supported formats

## Error Handling

### Client-Side Validation

- File size validation before upload
- File type checking using extension and MIME type
- Attachment count limits enforcement
- Real-time feedback for validation errors

### Server-Side Validation

- Comprehensive file validation in `processFileUpload()`
- Database transaction rollback on validation failures
- Proper HTTP error codes and messages
- Detailed error logging for debugging

### Error Messages

- `attachmentTooLargeMessage`: "attachment exceeds maximum size limit (10MB)"
- `tooManyAttachmentsMessage`: "too many attachments (maximum 10 allowed)"
- `attachmentNotFoundMessage`: "attachment not found"
- `invalidAttachmentMessage`: "invalid attachment data"

## Testing Recommendations

### Manual Testing Scenarios

1. **Upload Various File Types**: Test all supported file formats
2. **Size Limit Testing**: Attempt to upload files exceeding 10MB
3. **Count Limit Testing**: Try uploading more than 10 attachments
4. **Drag-and-Drop**: Test drag-and-drop functionality across browsers
5. **Reordering**: Test attachment reordering with multiple attachments
6. **Download/View**: Verify all attachments download correctly
7. **Delete Operations**: Test attachment deletion and persistence

### Automated Testing Areas

- File upload validation logic
- MIME type detection accuracy
- Unique ID generation collision testing
- Database transaction integrity
- Error handling and recovery

## Deployment Notes

### Requirements

- **SortableJS**: CDN dependency for drag-and-drop reordering
- **Bootstrap Icons**: Required for file type and action icons
- **Browser Support**: Modern browsers with File API support

### Configuration

- Attachment limits configurable via constants in `pgvis/attachment.go`
- MIME type mappings can be extended in the mime type detection function
- Error messages are configurable in handler constants

This implementation provides a robust, user-friendly attachment system for trouble reports while maintaining security and performance considerations.
