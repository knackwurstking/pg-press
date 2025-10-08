# Help Feature

This directory contains the help system for pg-press, providing comprehensive documentation and interactive examples for users.

## Overview

The help feature provides in-depth documentation and tutorials for various aspects of the application, with a focus on markdown formatting capabilities.

## Features

### Markdown Help Page (`/help/markdown`)

A comprehensive interactive guide that includes:

- **Quick Reference Card**: Visual examples of all supported markdown syntax
- **Interactive Demo**: Live preview where users can test markdown syntax
- **Detailed Documentation**: Step-by-step examples for each feature
- **Tips & Best Practices**: Guidelines for effective markdown usage
- **Common Mistakes**: Examples of what to avoid and correct alternatives

### Supported Markdown Features

The help page documents all supported markdown features:

#### Text Formatting

- **Bold text**: `**text**` → **text**
- **Italic text**: `*text*` → _text_
- **Underlined text**: `__text__` → <u>text</u>
- **Inline code**: `` `code` `` → `code`

#### Structure Elements

- **Headers**: `#`, `##`, `###`
- **Lists**: Unordered (`-`) and ordered (`1.`)
- **Blockquotes**: `> quote text`

#### Interactive Features

- Live preview as you type
- Copy-paste examples
- Visual syntax highlighting

## Integration

### Editor Integration

The help system is integrated with the editor through:

1. **Checkbox Help Link**: Info icon next to "Markdown-Formatierung verwenden"
2. **Toolbar Help Link**: Question mark icon in the markdown tools section
3. **Target Blank Links**: Help opens in new tab to preserve editor content

### Access Points

- Direct URL: `{ServerPathPrefix}/help/markdown` (properly handles server path prefix)
- Editor checkbox section: Info icon link (uses `env.ServerPathPrefix`)
- Markdown toolbar: Help icon link (uses `env.ServerPathPrefix`)

## File Structure

```
internal/web/features/help/
├── README.md           # This file
├── handlers.go         # HTTP handlers for help routes
├── routes.go          # Route definitions
└── templates/
    ├── markdown.templ     # Main help page template
    └── markdown_templ.go  # Generated template file
```

## Implementation Details

### Template Structure

The markdown help page uses:

- `layouts.Main` for consistent page structure
- `components.MarkdownScript` for rendering functionality
- Custom CSS for help-specific styling
- Interactive JavaScript for live demo

### Route Registration

Routes are registered in `internal/web/router.go`:

```go
help.RegisterRoutes(e)
```

No database dependency is required for help pages. All URLs properly handle the `SERVER_PATH_PREFIX` environment variable through `env.ServerPathPrefix`.

### Styling

The help page includes:

- Responsive grid layouts
- Interactive demo containers
- Example code highlighting
- Accessibility-friendly color schemes
- Mobile-responsive design

## Usage for Developers

### Adding New Help Pages

1. Create new handler in `handlers.go`
2. Add template in `templates/`
3. Register route in `routes.go`
4. Run `templ generate` to create Go template

### Customizing Content

The markdown help content is defined directly in the template and can be modified by editing `templates/markdown.templ`.

### Styling Customization

Help-specific styles are in the `helpStyles()` template function and can be customized without affecting global styles.

## User Experience

### Interactive Demo

- Real-time preview as users type
- Pre-populated with examples
- Copy-paste friendly syntax examples

### Navigation

- Clean, scannable layout
- Logical progression from basic to advanced
- Quick reference for power users
- Server path prefix aware URLs for deployment flexibility

### Accessibility

- Semantic HTML structure
- Keyboard navigation support
- Screen reader friendly
- High contrast colors

## Performance

- Static content with minimal JavaScript
- CSS optimized for fast rendering
- No external dependencies
- Minimal bundle size impact
- Server path prefix handling adds no performance overhead

## Maintenance

### Updating Documentation

- Edit `templates/markdown.templ` for content changes
- Run `templ generate` after template modifications
- Test interactive examples after updates

### Adding Features

- Update quick reference section
- Add new interactive examples
- Update tips and best practices
- Test with real markdown processor

## Future Enhancements

Potential future improvements could include:

- Search functionality within help pages
- Contextual help tooltips
- Video tutorials integration
- Multi-language support
- Advanced markdown features documentation
- Keyboard shortcuts reference

## Related Documentation

For comprehensive information about the help system implementation:

- **[Help System Overview](../../../docs/HELP_SYSTEM.md)** - Complete implementation summary and technical details
- **[Markdown Enhancements](../../../docs/MARKDOWN_ENHANCEMENTS.md)** - Details on underline support and blockquote fixes
- **[Path Prefix Testing](../../../docs/PATH_PREFIX_TESTING.md)** - Quick testing guide for server path prefix functionality
- **[Markdown Implementation](../../../docs/MARKDOWN_IMPLEMENTATION.md)** - Core markdown system documentation
- **[Shared Markdown System](../../../docs/SHARED_MARKDOWN_SYSTEM.md)** - Shared components and utilities

## Server Path Prefix Support

The help system fully supports the `SERVER_PATH_PREFIX` environment variable:

- All internal links use `env.ServerPathPrefix` for proper URL construction
- Help page accessible at `{prefix}/help/markdown` where prefix is configurable
- Editor integration links automatically include the correct prefix
- Navigation links properly handle prefix for seamless user experience
- No additional configuration required - works automatically with existing setup
