# Editor Feature Implementation

This document describes the new reusable editor feature that was created to replace the trouble reports edit dialog and provide a consistent markdown editing experience across the application.

## Overview

The editor feature provides a standalone, full-page editor with markdown support that can be used by different content types throughout the application. It replaces the previous dialog-based editing approach with a more user-friendly and feature-rich editing experience.

## Features

- **Full-page editing experience** with proper space for content creation
- **Live markdown preview** with real-time rendering
- **Markdown toolbar** with common formatting tools (bold, italic, headers, quotes, code, lists)
- **File attachments support** (for supported content types)
- **Drag & drop file uploads**
- **Responsive design** that works on mobile and desktop
- **Type-agnostic architecture** for easy extension to other content types

## Architecture

### Directory Structure

```
pg-press/internal/web/features/editor/
├── handlers.go           # HTTP handlers for editor routes
├── routes.go            # Route registration
└── templates/
    ├── editor.templ     # Main editor template
    └── editor_templ.go  # Generated template code
```

### Key Components

#### Handler (`handlers.go`)

- `GetEditorPage()` - Renders the editor page with existing content (for edits) or blank (for new content)
- `PostSaveContent()` - Processes form submissions and saves content based on type
- `loadExistingContent()` - Loads existing content for editing
- `saveContent()` - Saves content with type-specific logic
- `processAttachments()` - Handles file uploads and attachments

#### Template (`templates/editor.templ`)

- Full-page editor layout with markdown support
- Live preview panel with split-screen view
- Markdown toolbar with formatting tools
- Attachment management for supported types
- Responsive design with mobile support

#### Routes (`routes.go`)

- `GET /editor` - Editor page with query parameters for configuration
- `POST /editor/save` - Save content endpoint

## Usage

### URL Parameters

The editor is accessed via `/editor` with the following query parameters:

- `type` (required) - Content type (currently supports: `troublereport`)
- `id` (optional) - ID of existing item to edit (omit for new items)
- `return_url` (optional) - URL to redirect to after successful save

### Examples

**Create new trouble report:**

```
/editor?type=troublereport&return_url=/trouble-reports
```

**Edit existing trouble report:**

```
/editor?type=troublereport&id=123&return_url=/trouble-reports
```

## Changes Made

### Trouble Reports Refactoring

#### Removed Files

- `dialog-edit-trouble-report.templ` - Old dialog template
- `dialog-edit-trouble-report_templ.go` - Generated dialog code

#### Updated Files

- `page.templ` - Create button now links to editor instead of opening dialog
- `list.templ` - Edit buttons now link to editor instead of opening dialog
- `handlers.go` - Removed dialog-related handler methods:
  - `HTMXGetEditTroubleReportDialog()`
  - `HTMXPostEditTroubleReportDialog()`
  - `HTMXPutEditTroubleReportDialog()`
  - `validateDialogEditFormData()`
- `routes.go` - Removed dialog edit routes

#### Router Integration

- Added editor routes registration in `internal/web/router.go`
- Removed TODO comment about creating markdown editor

## Content Type Support

Currently supported content types:

### Trouble Reports (`troublereport`)

- **Fields**: Title, Content, UseMarkdown flag
- **Attachments**: Supported (images up to 10MB)
- **Database operations**: Full CRUD with attachment management
- **Feed integration**: Creates feed entries for create/update operations

### Future Extensions

The editor is designed to be easily extensible for other content types. To add support for a new type:

1. Add the type case in `loadExistingContent()`
2. Add the type case in `saveContent()`
3. Update `getTypeName()` in the template
4. Update `supportsAttachments()` if attachments are needed
5. Add appropriate redirect URL in `PostSaveContent()`

## Markdown Features

The editor supports the following markdown features:

- **Headers**: `#`, `##`, `###`
- **Bold**: `**text**`
- **Italic**: `*text*`
- **Code**: `` `code` ``
- **Quotes**: `> quote`
- **Lists**:
  - Unordered: `- item`
  - Ordered: `1. item`
- **Line breaks**: Automatic conversion of newlines
- **Paragraphs**: Double newlines create new paragraphs

## File Attachments

For content types that support attachments:

- **Supported formats**: Images (JPEG, PNG, GIF, WebP, SVG)
- **File size limit**: 10MB per file
- **Upload methods**:
  - Click to browse files
  - Drag & drop
- **Preview**: Shows file name and size before upload
- **Management**: Can remove files before saving
- **Existing attachments**: Can view and delete existing attachments when editing

## User Experience Improvements

Compared to the previous dialog approach:

1. **More space**: Full-page editor provides ample room for content creation
2. **Better preview**: Side-by-side markdown preview with live updates
3. **Tool assistance**: Markdown toolbar for easy formatting
4. **File management**: Better drag & drop and file preview experience
5. **Mobile friendly**: Responsive design that works on all devices
6. **Navigation**: Clear back/cancel options with return URL support
7. **Persistent state**: No risk of losing content when accidentally clicking outside

## Technical Notes

- Uses Templ for server-side rendering
- JavaScript handles client-side markdown preview and file management
- Attachment processing reuses existing models and database operations
- Maintains backward compatibility with existing markdown rendering
- Follows existing application patterns for handlers and routes
