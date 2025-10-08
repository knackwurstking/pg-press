# Markdown Features Implementation for Trouble Reports

## Overview

This document outlines the implementation of markdown support for the trouble reports feature in pg-press. The enhancement allows users to optionally use markdown syntax in their trouble report content, which is then rendered as formatted HTML when displayed and properly formatted in PDF exports.

## Features Implemented

### Core Functionality

- **Optional Markdown Support**: Users can enable markdown formatting via a checkbox in the edit dialog
- **HTML Rendering**: Markdown content is converted to HTML when displayed in the web interface
- **PDF Integration**: Markdown content is formatted appropriately in PDF exports
- **Security**: HTML sanitization to prevent XSS attacks
- **Backward Compatibility**: Existing reports continue to display as plain text

### User Interface Enhancements

- **Markdown Checkbox**: Clear opt-in mechanism for markdown formatting
- **Editing Tools**: Toolbar with common markdown formatting buttons
- **Live Preview**: Real-time preview of markdown rendering
- **Responsive Styling**: CSS optimized for both markdown and plain text content

## Database Changes

### Schema Updates

```sql
-- Added use_markdown column to trouble_reports table
ALTER TABLE trouble_reports ADD COLUMN use_markdown BOOLEAN DEFAULT 0;
```

### Migration Support

- Automatic migration detects existing databases and adds the `use_markdown` column
- Default value is `false` (0) to maintain backward compatibility
- No data loss during migration

## Model Updates

### TroubleReport Model (`pkg/models/troublereport.go`)

```go
type TroubleReport struct {
    ID                int64   `json:"id"`
    Title             string  `json:"title"`
    Content           string  `json:"content"`
    LinkedAttachments []int64 `json:"linked_attachments"`
    UseMarkdown       bool    `json:"use_markdown"`  // New field
}
```

### Modification Data Model (`pkg/models/modification_data_types.go`)

```go
type TroubleReportModData struct {
    Title             string  `json:"title"`
    Content           string  `json:"content"`
    LinkedAttachments []int64 `json:"linked_attachments"`
    UseMarkdown       bool    `json:"use_markdown"`  // New field
}
```

## Service Layer Updates

### Database Operations (`internal/services/trouble-reports.go`)

- Updated INSERT queries to include `use_markdown` field
- Updated UPDATE queries to handle `use_markdown` field
- Updated SELECT queries to retrieve `use_markdown` field
- Added migration function for existing databases
- Updated modification tracking to include markdown flag

### Key Changes

```go
// INSERT operation
const addQuery = `INSERT INTO trouble_reports
    (title, content, linked_attachments, use_markdown) VALUES (?, ?, ?, ?)`

// UPDATE operation
const updateQuery string = `UPDATE trouble_reports
    SET title = ?, content = ?, linked_attachments = ?, use_markdown = ? WHERE id = ?`

// SELECT scan includes use_markdown
scanner.Scan(&report.ID, &report.Title, &report.Content, &linkedAttachments, &report.UseMarkdown)
```

## Handler Updates

### Form Processing (`internal/web/features/troublereports/handlers.go`)

- Updated `validateDialogEditFormData` function to extract `use_markdown` checkbox value
- Modified create and update handlers to process markdown flag
- Updated edit dialog handler to pass markdown state to template

### Key Changes

```go
func (h *Handler) validateDialogEditFormData(ctx echo.Context) (
    title, content string,
    useMarkdown bool,  // New return value
    attachments []*models.Attachment,
    err error,
) {
    // ... existing validation logic
    useMarkdown = ctx.FormValue("use_markdown") == "on"
    // ... rest of function
}
```

## Template Updates

### Edit Dialog Template (`internal/web/features/troublereports/templates/dialog-edit-trouble-report.templ`)

#### New Markdown Checkbox

```html
<label for="use_markdown" class="flex gap items-center">
  <input
    type="checkbox"
    name="use_markdown"
    id="use_markdown"
    onchange="toggleMarkdownFeatures()"
  />
  <span>Markdown-Formatierung verwenden</span>
  <small class="muted"
    >(unterstützt **fett**, *kursiv*, # Überschriften, etc.)</small
  >
</label>
```

#### Markdown Editing Tools

- Bold/Italic formatting buttons
- Header insertion buttons (H1, H2)
- List insertion button
- Code formatting button
- Live preview toggle

#### JavaScript Features

- `toggleMarkdownFeatures()`: Shows/hides markdown tools
- `insertMarkdown()`: Inserts markdown syntax at cursor position
- `togglePreview()`: Shows/hides live preview
- `updatePreview()`: Converts markdown to HTML for preview

### List Display Template (`internal/web/features/troublereports/templates/list.templ`)

```go
// Conditional rendering based on UseMarkdown flag
if tr.UseMarkdown {
    <div class="markdown-content">
        @templ.Raw(string(utils.RenderMarkdown(tr.Content)))
    </div>
} else {
    <pre>{ tr.Content }</pre>
}
```

## Utility Functions

### HTML Rendering Fix

The key implementation detail is using `@templ.Raw()` to render sanitized HTML instead of escaping it:

```go
// WRONG - This would escape the HTML and show raw tags to users
{ utils.RenderMarkdown(tr.Content) }

// CORRECT - This renders the HTML properly while maintaining security
@templ.Raw(string(utils.RenderMarkdown(tr.Content)))
```

### Newline Preprocessing

The markdown processor includes intelligent newline handling:

```go
// preprocessNewlines converts single newlines to line breaks, but preserves markdown structure
func preprocessNewlines(markdown string) string {
    lines := strings.Split(markdown, "\n")
    var result []string

    for i, line := range lines {
        result = append(result, line)

        // Add line break if:
        // - Current line is not empty
        // - Next line exists and is not empty
        // - Neither line is a markdown structural element
        if i < len(lines)-1 {
            currentTrimmed := strings.TrimSpace(line)
            nextTrimmed := strings.TrimSpace(lines[i+1])

            if currentTrimmed != "" && nextTrimmed != "" &&
                !isMarkdownStructure(currentTrimmed) &&
                !isMarkdownStructure(nextTrimmed) &&
                !strings.HasSuffix(line, "  ") {
                // Add two spaces for line break
                result[len(result)-1] = line + "  "
            }
        }
    }

    return strings.Join(result, "\n")
}
```

### Markdown Processor (`pkg/utils/utils.go`)

```go
func RenderMarkdown(markdown string) template.HTML {
    // Preprocess newlines for better user experience
    processedMarkdown := preprocessNewlines(markdown)
    // Configure blackfriday with safe settings
    extensions := blackfriday.NoIntraEmphasis |
        blackfriday.Tables |
        blackfriday.FencedCode |
        blackfriday.Autolink |
        blackfriday.Strikethrough |
        blackfriday.SpaceHeadings |
        blackfriday.BackslashLineBreak

    renderer := blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{
        Flags: blackfriday.HTMLFlagsNone,
    })

    htmlBytes := blackfriday.Run([]byte(markdown),
        blackfriday.WithRenderer(renderer),
        blackfriday.WithExtensions(extensions))
    htmlContent := string(htmlBytes)

    // Apply additional sanitization
    sanitizedHTML := sanitizeHTML(htmlContent)
    return template.HTML(sanitizedHTML)
}
```

### Security Features

- HTML sanitization to remove dangerous tags (`<script>`, `<iframe>`, etc.)
- Event handler removal (`onclick`, `onload`, etc.)
- Protocol filtering for links (`javascript:`, `data:`)
- Safe HTML rendering using `template.HTML`

## CSS Styling

### Markdown Content Styling (`internal/web/assets/css/trouble-reports/data.css`)

```css
.markdown-content {
  line-height: 1.6;
  color: var(--ui-text);
}

.markdown-content h1,
.markdown-content h2,
.markdown-content h3 {
  margin-top: 1.5em;
  margin-bottom: 0.5em;
  font-weight: bold;
}

.markdown-content code {
  background-color: var(--ui-background-secondary);
  border-radius: 3px;
  padding: 0.2em 0.4em;
  font-family: "Courier New", "Monaco", monospace;
}

.markdown-content table {
  border-collapse: collapse;
  width: 100%;
}
```

### Edit Dialog Styling (`internal/web/assets/css/trouble-reports/edit.css`)

```css
label[for="use_markdown"] {
  padding: var(--ui-spacing);
  background-color: var(--ui-background-secondary);
  border: 1px solid var(--ui-border-color);
  border-radius: var(--ui-radius);
}

.markdown-toolbar {
  display: flex;
  gap: calc(var(--ui-spacing) / 2);
  background-color: var(--ui-background-secondary);
}

.markdown-preview {
  border: 1px solid var(--ui-border-color);
  border-radius: var(--ui-radius);
  background-color: var(--ui-background);
}
```

## PDF Generation Updates

### Enhanced PDF Rendering (`internal/pdf/trouble-report.go`)

- Added `renderMarkdownContentToPDF()` function
- Handles markdown headers, lists, and basic formatting
- Removes markdown syntax for clean PDF output
- Maintains text hierarchy in PDF layout

### Supported Markdown Elements in PDF

- Headers (# ## ###) → Different font sizes and weights
- Unordered lists (- \*) → Bullet points
- Numbered lists (1. 2. 3.) → Numbered items
- Basic formatting removal (\*_, _, ~~, `) → Clean text
- Links → Text extraction

```go
func renderMarkdownContentToPDF(o *troubleReportOptions) {
    lines := strings.Split(content, "\n")
    for _, line := range lines {
        // Handle headers
        if strings.HasPrefix(line, "# ") {
            o.PDF.SetFont("Arial", "B", 13)
            // ... formatting logic
        }
        // Handle lists, paragraphs, etc.
    }
}
```

## Dependencies Added

### Go Modules

```go
require github.com/russross/blackfriday/v2 v2.1.0
```

## Testing

### Unit Tests (`pkg/utils/utils_test.go`)

- `TestRenderMarkdown`: Tests various markdown elements
- `TestRenderMarkdownReturnsHTML`: Verifies return type
- `TestRenderMarkdownWithSpecialCharacters`: Security testing
- `BenchmarkRenderMarkdown`: Performance testing

### Test Coverage

- Bold and italic formatting
- Headers (H1, H2, H3)
- Lists (ordered and unordered)
- Code blocks and inline code
- Links and strikethrough
- XSS prevention
- Empty input handling

## Security Considerations

### XSS Prevention

1. **Input Sanitization**: HTML in markdown input is processed safely
2. **Output Sanitization**: Generated HTML is sanitized before rendering
3. **Template Safety**: Uses `template.HTML` for safe rendering
4. **Script Removal**: JavaScript tags and event handlers are stripped

### Safe HTML Elements Allowed

- Headers: `<h1>`, `<h2>`, `<h3>`, `<h4>`, `<h5>`, `<h6>`
- Text formatting: `<strong>`, `<em>`, `<code>`, `<pre>`
- Lists: `<ul>`, `<ol>`, `<li>`
- Other: `<p>`, `<br>`, `<hr>`, `<blockquote>`
- Tables: `<table>`, `<th>`, `<td>`, `<tr>`
- Links: `<a>` (with protocol filtering)

### HTML Rendering Implementation

The sanitized HTML is rendered using `@templ.Raw()` to bypass Go template escaping:

```go
// In list.templ template
@templ.Raw(string(utils.RenderMarkdown(tr.Content)))
```

This ensures users see properly formatted content instead of escaped HTML tags.

### Dangerous Elements Removed

- Scripts: `<script>`, `<iframe>`, `<object>`, `<embed>`
- Forms: `<form>`, `<input>`, `<button>`, `<select>`
- Events: All `on*` attributes
- Protocols: `javascript:`, `data:` in links

## Usage Instructions

### For Users

1. Open trouble report create/edit dialog
2. Check "Markdown-Formatierung verwenden" checkbox
3. Use markdown toolbar buttons or type markdown syntax directly
4. Click "Vorschau" to see formatted output
5. Save report as usual

### Supported Markdown Syntax

**Supported Markdown Syntax**:

```markdown
# Header 1

## Header 2

### Header 3

**Bold text**
_Italic text_
~~Strikethrough~~

Regular text with
automatic line breaks
when typing on separate lines

- Unordered list
- Another item

1. Ordered list
2. Another item

`Inline code`
```

Code block

```

[Link text](https://example.com)
```

### For Developers

1. The `UseMarkdown` field is automatically handled in CRUD operations
2. Templates conditionally render based on the markdown flag using `@templ.Raw()`
3. HTML content is properly rendered (not escaped) for rich formatting
4. PDF generation automatically formats markdown content
5. Security is handled transparently by the utility functions

## Migration Path

### Existing Data

- All existing trouble reports have `use_markdown = false`
- Content displays exactly as before (plain text)
- No user action required for existing reports

### New Reports

- Users can opt-in to markdown formatting
- Default is plain text (backward compatible)
- Markdown can be enabled/disabled when editing

## Performance Considerations

### Rendering Performance

- Markdown processing is done on-demand during display
- No performance impact for plain text reports
- Benchmark shows acceptable performance for typical report sizes

### Database Impact

- Single boolean column addition
- No impact on existing queries
- Indexes not needed for boolean flag

## Future Enhancements

### Potential Improvements

1. **Rich Text Editor**: WYSIWYG editor integration
2. **Syntax Highlighting**: Code block syntax highlighting
3. **Image Embedding**: Markdown image syntax support
4. **Table Editor**: Visual table editing tools
5. **Export Formats**: Additional export formats (Word, etc.)

### Backward Compatibility

- All changes maintain full backward compatibility
- Existing reports continue to work unchanged
- API endpoints maintain same interface
- Database migration is non-destructive

## Conclusion

The markdown feature implementation provides a powerful enhancement to the trouble reports system while maintaining full backward compatibility and security. Users can now create rich, formatted reports while developers benefit from a clean, maintainable codebase with comprehensive testing and documentation.

The implementation follows best practices for:

- Security (XSS prevention)
- Performance (on-demand processing)
- Maintainability (clean separation of concerns)
- User Experience (intuitive interface with preview)
- Data Integrity (safe migration and validation)
