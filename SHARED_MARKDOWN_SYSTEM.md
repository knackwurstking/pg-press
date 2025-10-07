# Shared Markdown System

This document describes the shared markdown rendering system implemented in the pg-press application. The system provides centralized markdown functionality that can be reused across different features.

## Overview

The shared markdown system consolidates markdown rendering logic, styles, and functionality into reusable components located in the `internal/web/shared/components/` directory. This approach eliminates code duplication and ensures consistent markdown behavior throughout the application.

## Architecture

### Components

The shared markdown system consists of three main components in `internal/web/shared/components/markdown.templ`:

#### 1. `MarkdownScript()`

Provides a hybrid approach combining templ script functions and window object functions:

**Templ Script Functions** (scoped, efficient):

- **`renderMarkdownToHTML(content)`** - Core markdown-to-HTML conversion function

**Window Object Functions** (global, cross-template access):

- **`window.processMarkdownContent()`** - Automatically processes elements with `data-markdown-content` attributes
- **`window.renderMarkdownInContainer(containerId, content)`** - Renders markdown in a specific container
- **`window.updateMarkdownPreview(textareaId, previewId)`** - Updates live preview for editors

**Auto-initialization**: Automatically processes markdown content on DOM load and HTMX events

#### 2. `MarkdownStyles()`

Provides CSS styles for rendered markdown content:

- Typography for headers (h1-h6)
- Paragraph and list styling
- Code block and inline code formatting
- Emphasis and strong text styling
- Consistent theming using CSS variables

#### 3. `MarkdownContent(content, useMarkdown)`

Renders markdown content conditionally:

- If `useMarkdown` is true: Creates a div with `markdown-content` class and `data-markdown-content` attribute
- If `useMarkdown` is false: Displays content in a `<pre>` tag

## Supported Markdown Features

The system supports the following markdown syntax:

### Headers

```markdown
# Header 1

## Header 2

### Header 3
```

### Text Formatting

```markdown
**bold text**
_italic text_
`inline code`
```

### Lists

```markdown
- Unordered list item
- Another item

1. Ordered list item
2. Another numbered item
```

### Blockquotes

```markdown
> This is a blockquote
> Continued line
```

### Paragraphs and Line Breaks

- Double newlines create new paragraphs
- Single newlines become `<br>` tags within paragraphs

## Usage

### Basic Implementation

To add markdown support to a feature:

1. **Import the components** in your template:

```go
import "github.com/knackwurstking/pgpress/internal/web/shared/components"
```

2. **Include the script and styles** in your template:

```go
@components.MarkdownScript()  // Includes both script functions and auto-initialization
@components.MarkdownStyles()
```

3. **Render content using the component**:

```go
@components.MarkdownContent(item.Content, item.UseMarkdown)
```

**Note**: No additional initialization is needed - `MarkdownScript()` automatically handles DOM processing and HTMX integration.

### Live Preview (Editor)

For editor functionality with live preview:

1. **Include the shared components**:

```go
@components.MarkdownScript()  // Includes all functions and auto-initialization
@components.MarkdownStyles()
```

2. **Use the preview update function**:

```javascript
function updatePreview() {
  window.updateMarkdownPreview("content", "preview-content");
}
```

3. **Set up event listeners**:

```javascript
textarea.addEventListener("input", updatePreview);
```

**Benefits of the hybrid approach**:

- Core rendering function (`renderMarkdownToHTML`) is scoped and efficient
- DOM manipulation functions are globally accessible where needed
- Automatic initialization eliminates boilerplate code

## Current Implementations

### Editor Feature

**Location**: `internal/web/features/editor/templates/editor.templ`

**Usage**:

- Includes `MarkdownScript()` (hybrid functions + auto-init) and `MarkdownStyles()`
- Uses `window.updateMarkdownPreview()` for live preview functionality
- Provides full-featured markdown editing experience with automatic processing

### Trouble Reports Feature

**Location**: `internal/web/features/troublereports/templates/list.templ`

**Usage**:

- Includes `MarkdownScript()` (hybrid functions + auto-init) and `MarkdownStyles()`
- Uses `MarkdownContent()` component for content display
- Automatically processes markdown with no additional setup required

## Benefits

### Code Reusability

- Single source of truth for markdown functionality
- Hybrid approach optimizes both performance and accessibility
- No duplication of rendering logic or styles
- Easy to maintain and update

### Consistency

- Uniform markdown rendering across all features
- Consistent styling and behavior
- Same supported markdown features everywhere

### Performance

- Shared JavaScript functions loaded once per page
- Consistent CSS that can be cached
- Automatic processing with optimized event handling

### Maintainability

- Changes to markdown functionality only need to be made in one place
- Easy to add new markdown features globally
- Centralized debugging and testing

## Technical Details

### Hybrid Architecture

The system uses a hybrid approach combining:

**Templ Script Functions**:

- `renderMarkdownToHTML()` - Core processing, scoped to avoid namespace pollution
- Compiled efficiently, no global object pollution

**Window Object Functions**:

- DOM manipulation functions that need cross-template access
- `window.processMarkdownContent()`, `window.updateMarkdownPreview()`, etc.

### Automatic Processing

The system automatically processes markdown content on:

- DOM content loaded
- HTMX afterSwap events
- Manual triggers via JavaScript functions
- No manual initialization required

### CSS Integration

Styles use CSS variables from the ui.min.css framework:

- `--ui-text` for text color
- Theme-aware styling that works in light and dark modes
- Consistent with application design system

### Error Handling

- Graceful fallback to plain text if markdown rendering fails
- Safe HTML escaping for non-markdown content
- Null/empty content handling

## Future Enhancements

Potential improvements to consider:

### Additional Markdown Features

- Tables
- Images with captions
- Strikethrough text
- Links with title attributes
- Nested blockquotes

### Performance Optimizations

- Lazy loading of markdown processor
- Caching of rendered content
- Virtual scrolling for large documents

### Editor Enhancements

- Syntax highlighting
- Real-time error detection
- Keyboard shortcuts for formatting

## Migration Guide

### From Feature-Specific Implementation

1. **Remove local markdown functions** from feature templates
2. **Remove duplicate CSS styles** for markdown content
3. **Import shared components** package
4. **Replace local implementations** with shared component calls
5. **Update event listeners** to use shared functions
6. **Test functionality** to ensure consistent behavior

### Example Migration

**Before** (feature-specific):

```go
script renderMarkdownToHTML() {
    window.renderMarkdownToHTML = function(content) {
        // Local implementation...
    };
}
```

**After** (shared hybrid):

```go
@components.MarkdownScript()  // Includes both templ script functions and window functions
```

## Troubleshooting

### Common Issues

1. **Markdown not rendering**: Ensure `MarkdownScript()` is included in the template
2. **Styling issues**: Verify `MarkdownStyles()` is included and CSS variables are defined
3. **HTMX integration**: Auto-processing should work automatically - check console for errors
4. **Preview not updating**: Ensure `window.updateMarkdownPreview()` is called correctly

### Debugging

- Use browser console to check if `renderMarkdownToHTML` function is defined (not on window)
- Verify window functions like `window.processMarkdownContent` are available
- Check that `data-markdown-content` attributes are set properly
- Confirm CSS variables are available in the current theme
- Look for JavaScript errors that might prevent script function compilation

## Conclusion

The shared markdown system provides a robust, maintainable, and consistent approach to markdown handling across the pg-press application. The hybrid architecture combines the efficiency of templ script functions with the accessibility of global window functions, providing optimal performance while maintaining cross-template compatibility. By centralizing functionality and styles, it reduces complexity, improves consistency, and makes future enhancements easier to implement.
