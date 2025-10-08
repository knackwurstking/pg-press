# Markdown Improvements

This document outlines the recent improvements made to the markdown rendering system in pg-press.

## Summary of Changes

Two key improvements have been implemented to enhance the markdown functionality:

1. **Added underline support** - Users can now create underlined text using the `__text__` syntax
2. **Fixed blockquote newline detection** - Blockquotes now properly preserve line breaks between consecutive quote lines

## Detailed Changes

### 1. Underline Support

**Files Modified:**

- `internal/web/shared/components/markdown.templ`
- `internal/web/features/editor/templates/editor.templ`

**Changes Made:**

- Added regex pattern to convert `__text__` to `<u>text</u>` in the markdown renderer
- Added CSS styling for `<u>` elements with `text-decoration: underline`
- Added underline button to the editor toolbar with Bootstrap icon `bi-type-underline`
- Button calls `insertMarkdown('__', '__')` to wrap selected text

**Usage:**

```markdown
This text has **underlined words** in the sentence.
You can combine **bold**, _italic_, and **underlined** formatting.
```

**Rendered as:**

- This text has <u>underlined words</u> in the sentence.
- You can combine **bold**, _italic_, and <u>underlined</u> formatting.

### 2. Blockquote Newline Fix

**Files Modified:**

- `internal/web/shared/components/markdown.templ`

**Problem Fixed:**
Previously, consecutive blockquote lines were joined with `<br>` tags, which didn't properly separate the lines in the final rendered output.

**Solution:**
Changed the replacement pattern from `<br>` to `\n` (newline) to ensure proper line separation within blockquote blocks.

**Before:**

```javascript
.replace(/<\/bq-line>/g, '<br>');
```

**After:**

```javascript
.replace(/<\/bq-line>/g, '\n');
```

**Usage:**

```markdown
> This is the first line of a blockquote
> This is the second line
> This is the third line
```

**Rendered as:**

> This is the first line of a blockquote
> This is the second line
> This is the third line

### 3. Editor Toolbar Enhancement

**New Button Added:**

- Icon: `bi-type-underline` (Bootstrap Icons)
- Title: "Unterstrichen" (German for "Underlined")
- Function: `insertMarkdown('__', '__')`
- Position: Between italic and heading buttons

**Help Integration:**

- Added help links in markdown checkbox section and toolbar
- Links use `env.ServerPathPrefix` for proper URL construction
- Help opens in new tabs to preserve editor content
- All links automatically handle server path prefix configuration

## Technical Implementation

### Regex Processing Order

The markdown processor handles formatting in this order:

1. Headers (`#`, `##`, `###`)
2. **Underline** (`__text__`) - **NEW**
3. Bold (`**text**`)
4. Italic (`*text*`)
5. Inline code (`` `code` ``)
6. Lists and blockquotes

### CSS Styling

Added specific styling for underlined text:

```css
.markdown-content u {
  text-decoration: underline;
}
```

## Testing

A comprehensive test file (`test_markdown.html`) has been created to verify both improvements work correctly. The test includes:

- Underline syntax testing
- Blockquote newline preservation
- Combined formatting scenarios
- Interactive preview functionality

## Backward Compatibility

These changes are fully backward compatible:

- Existing markdown content continues to render correctly
- No breaking changes to the API
- All existing functionality remains unchanged

## Browser Support

The underline feature uses standard HTML `<u>` tags with CSS `text-decoration: underline`, which is supported in all modern browsers.

## Files Modified

1. `internal/web/shared/components/markdown.templ` - Core markdown rendering logic
2. `internal/web/features/editor/templates/editor.templ` - Editor toolbar with help links
3. `internal/web/features/help/` - Complete new help feature module
4. `internal/web/router.go` - Added help routes registration
5. `internal/web/shared/components/markdown_templ.go` - Auto-generated (via `templ generate`)
6. `internal/web/features/editor/templates/editor_templ.go` - Auto-generated (via `templ generate`)
7. `internal/web/features/help/templates/markdown_templ.go` - Auto-generated (via `templ generate`)

## How to Use

### For Users

1. **Underline text**: Wrap text with double underscores: `__your text__`
2. **Blockquotes**: Use `> ` at the start of lines - newlines are now properly preserved
3. **Editor toolbar**: Click the underline button (U icon) to format selected text
4. **Get help**: Click the help icons (ⓘ or ❓) in the editor to access comprehensive markdown documentation
5. **Access help directly**: Visit `{SERVER_PATH_PREFIX}/help/markdown` where prefix is your configured path

### For Developers

- Run `templ generate` after modifying `.templ` files
- The markdown processor automatically handles the new syntax
- No additional configuration required

## Future Enhancements

Potential future improvements could include:

- Strikethrough support (`~~text~~`)
- Better code block handling with syntax highlighting
- Table support
- Link rendering improvements

## Server Path Prefix Support

All help system components fully support the `SERVER_PATH_PREFIX` environment variable:

**Implementation:**

- Help page accessible at `{prefix}/help/markdown` where prefix is configurable
- Editor help links use `env.ServerPathPrefix` for proper URL construction
- Navigation links automatically handle prefix for seamless experience
- Zero additional configuration required

**Benefits:**

- Works with reverse proxies and custom deployment paths
- Maintains consistent user experience regardless of URL structure
- Automatic prefix handling requires no manual URL updates
- Compatible with existing server path prefix configuration
