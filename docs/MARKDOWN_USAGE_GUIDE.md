# Markdown Usage Guide for pg-press

## Overview

This comprehensive guide covers all markdown features supported in pg-press, including syntax examples, best practices, and common use cases. The markdown system in pg-press provides a powerful way to create rich, formatted content while maintaining readability in plain text.

## Table of Contents

- [Basic Text Formatting](#basic-text-formatting)
- [Headers](#headers)
- [Lists](#lists)
- [Code](#code)
- [Links](#links)
- [Blockquotes](#blockquotes)
- [Line Breaks and Paragraphs](#line-breaks-and-paragraphs)
- [Best Practices](#best-practices)
- [Common Mistakes](#common-mistakes)
- [Advanced Usage](#advanced-usage)
- [Technical Implementation](#technical-implementation)

## Basic Text Formatting

### Bold Text

Use double asterisks (`**`) to make text bold.

**Syntax:**

```markdown
**This text will be bold**
```

**Result:**
**This text will be bold**

**Use Cases:**

- Emphasizing important information
- Highlighting key terms or concepts
- Creating visual hierarchy in text

### Italic Text

Use single asterisks (`*`) to make text italic.

**Syntax:**

```markdown
_This text will be italic_
```

**Result:**
_This text will be italic_

**Use Cases:**

- Adding emphasis to words or phrases
- Indicating technical terms or foreign words
- Providing subtle emphasis without strong weight

### Underlined Text

Use double underscores (`__`) to underline text.

**Syntax:**

```markdown
**This text will be underlined**
```

**Result:**
<u>This text will be underlined</u>

**Use Cases:**

- Highlighting important information
- Indicating clickable elements (though links are preferred)
- Creating visual distinction in formatted text

### Combining Formats

You can combine different formatting styles within the same text.

**Syntax:**

```markdown
This is **bold**, this is _italic_, and this is **underlined**.
You can even combine them: **_bold and italic_** or **_underlined and italic_**.
```

**Result:**
This is **bold**, this is _italic_, and this is <u>underlined</u>.
You can even combine them: **_bold and italic_** or <u>_underlined and italic_</u>.

## Headers

Headers create a hierarchy of content and improve document structure.

### Header Levels

**Syntax:**

```markdown
# Header 1 (Largest)

## Header 2 (Large)

### Header 3 (Medium)

#### Header 4 (Small)

##### Header 5 (Smaller)

###### Header 6 (Smallest)
```

**Best Practices for Headers:**

- Always include a space after the `#` symbols
- Use headers to create logical document structure
- Don't skip header levels (don't go from H1 to H3)
- Use Header 1 sparingly (typically for document title)
- Headers 2 and 3 are most commonly used for sections

**Example Document Structure:**

```markdown
# Project Documentation

## Getting Started

### Installation

#### Prerequisites

### Configuration

## User Guide

### Basic Usage

### Advanced Features
```

## Lists

### Unordered Lists

Use hyphens (`-`) to create bullet points.

**Syntax:**

```markdown
- First item
- Second item
- Third item
  - Nested item (use 2 spaces for indentation)
  - Another nested item
- Fourth item
```

**Result:**

- First item
- Second item
- Third item
  - Nested item
  - Another nested item
- Fourth item

### Ordered Lists

Use numbers followed by periods (`1.`) to create numbered lists.

**Syntax:**

```markdown
1. First item
2. Second item
3. Third item
   1. Nested numbered item
   2. Another nested item
4. Fourth item
```

**Result:**

1. First item
2. Second item
3. Third item
   1. Nested numbered item
   2. Another nested item
4. Fourth item

**List Best Practices:**

- Use consistent indentation (2 spaces for nesting)
- Leave blank lines before and after lists for better readability
- Choose ordered lists for step-by-step instructions
- Choose unordered lists for simple collections of items

## Code

### Inline Code

Use backticks (`` ` ``) to mark inline code.

**Syntax:**

```markdown
Use the `console.log()` function to print output.
The `npm install` command installs dependencies.
```

**Result:**
Use the `console.log()` function to print output.
The `npm install` command installs dependencies.

### Code Blocks

Use triple backticks (` ``` `) to create code blocks.

**Syntax:**

````markdown
```javascript
function greet(name) {
  return `Hello, ${name}!`;
}

console.log(greet("World"));
```
````

**Result:**

```javascript
function greet(name) {
  return `Hello, ${name}!`;
}

console.log(greet("World"));
```

**Code Best Practices:**

- Always specify the language for syntax highlighting
- Use inline code for short code snippets, variable names, or commands
- Use code blocks for multi-line code examples
- Include comments in code blocks to explain complex logic

## Links

Create clickable links to external resources or internal pages.

### Basic Links

**Syntax:**

```markdown
[Link text](https://example.com)
[pg-press Documentation](https://github.com/user/pg-press)
```

**Result:**
[Link text](https://example.com)
[pg-press Documentation](https://github.com/user/pg-press)

### Link Best Practices

- Use descriptive link text instead of "click here"
- Verify links are working before publishing
- Consider using relative paths for internal links
- Include protocol (https://) for external links

**Good Examples:**

```markdown
[View the installation guide](docs/installation.md)
[Download the latest release](https://github.com/user/repo/releases/latest)
```

**Avoid:**

```markdown
Click [here](https://example.com) to see more.
[https://example.com](https://example.com)
```

## Blockquotes

Use the greater-than symbol (`>`) to create blockquotes for citations or highlighted text.

### Basic Blockquotes

**Syntax:**

```markdown
> This is a blockquote.
> It can span multiple lines.
> Each line should start with >.
```

**Result:**

> This is a blockquote.
> It can span multiple lines.
> Each line should start with >.

### Multi-paragraph Blockquotes

**Syntax:**

```markdown
> This is the first paragraph of a blockquote.
>
> This is the second paragraph in the same blockquote.
```

### Nested Blockquotes

**Syntax:**

```markdown
> This is a blockquote.
>
> > This is a nested blockquote.
>
> Back to the first level.
```

**Blockquote Best Practices:**

- Use for quotations, important notes, or callouts
- Include attribution when quoting sources
- Don't overuse - they should stand out as special content
- Consider using blockquotes for warnings or important information

## Line Breaks and Paragraphs

### Paragraphs

Separate paragraphs with blank lines.

**Syntax:**

```markdown
This is the first paragraph. It contains some text that explains a concept.

This is the second paragraph. It's separated from the first by a blank line.

This is the third paragraph.
```

### Line Breaks

For line breaks within a paragraph, end a line with two spaces or use manual line break.

**Syntax:**

```markdown
First line of paragraph
Second line of same paragraph (note the two spaces above)

Or use a blank line to start a new paragraph.
```

## Best Practices

### Document Structure

1. **Use Consistent Formatting**
   - Stick to one style for headers, lists, and emphasis
   - Be consistent with spacing and indentation
   - Use the same formatting for similar content types

2. **Create Logical Hierarchy**
   - Start with H1 for the document title
   - Use H2 for major sections
   - Use H3 for subsections
   - Don't skip header levels

3. **Improve Readability**
   - Use blank lines to separate content blocks
   - Keep paragraphs reasonably short
   - Use lists to break up long text blocks
   - Add spaces around formatting characters

### Content Guidelines

1. **Write Clear Link Text**

   ```markdown
   Good: [Download the user manual](docs/manual.pdf)
   Avoid: [Click here](docs/manual.pdf) for the manual
   ```

2. **Use Appropriate Emphasis**

   ```markdown
   Good: The **most important** step is to save your work.
   Avoid: The **MOST** **IMPORTANT** **STEP** is to **SAVE** your work.
   ```

3. **Structure Lists Logically**

   ```markdown
   Good:

   1. Prepare the ingredients
   2. Mix the dry ingredients
   3. Add wet ingredients
   4. Bake for 30 minutes

   Avoid:

   - Bake for 30 minutes
   - Prepare ingredients
   - Add wet ingredients
   - Mix dry ingredients
   ```

### Technical Best Practices

1. **Always Test Your Markdown**
   - Preview your content before publishing
   - Check that all links work
   - Verify formatting appears as intended

2. **Use Proper Encoding**
   - Save files as UTF-8
   - Test with special characters
   - Be consistent with line endings

3. **Consider Your Audience**
   - Use appropriate technical level
   - Include necessary context
   - Provide examples for complex concepts

## Common Mistakes

### Formatting Errors

1. **Missing Spaces After Hash Symbols**

   ```markdown
   Wrong: #Header without space
   Correct: # Header with space
   ```

2. **Inconsistent List Formatting**

   ```markdown
   Wrong:
   -Item one

   - Item two
   - Item three

   Correct:

   - Item one
   - Item two
   - Item three
   ```

3. **Broken Emphasis Formatting**
   ```markdown
   Wrong: **Bold text without closing
   Wrong: _Italic text _ with space before closing
   Correct: **Bold text with proper closing\**
   Correct: *Italic text without spaces\*
   ```

### Structural Issues

1. **Skipping Header Levels**

   ```markdown
   Wrong:

   # Main Title

   ### Subsection (skipped H2)

   Correct:

   # Main Title

   ## Major Section

   ### Subsection
   ```

2. **No Blank Lines Around Lists**

   ```markdown
   Wrong:
   This is a paragraph.

   - List item
   - Another item
     This is another paragraph.

   Correct:
   This is a paragraph.

   - List item
   - Another item

   This is another paragraph.
   ```

3. **Mixing List Types Inconsistently**

   ```markdown
   Avoid:

   1. First item

   - Second item (different style)

   2. Third item

   Better:

   1. First item
   2. Second item
   3. Third item
   ```

### Content Issues

1. **Unclear Link Text**

   ```markdown
   Avoid: For more information, click [here](docs/info.md).
   Better: View the [detailed documentation](docs/info.md) for more information.
   ```

2. **Overusing Emphasis**
   ```markdown
   Avoid: This is **very** **important** and you **must** **remember** it.
   Better: This is **very important** and you must remember it.
   ```

## Advanced Usage

### Combining Elements

You can combine different markdown elements for rich content:

````markdown
# Project Status Report

## Overview

This report covers the **current status** of the project, including:

- Completed features
- Outstanding issues
- Next steps

## Completed Features

> All major functionality has been implemented and tested.

### User Authentication

The authentication system supports:

1. **Login/Logout** - Basic user authentication
2. **Registration** - New user account creation
3. **Password Reset** - Self-service password recovery

Code example for login:

```javascript
async function login(username, password) {
  const response = await fetch("/api/login", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ username, password }),
  });
  return response.json();
}
```
````

For more details, see the [API documentation](docs/api.md).

````

### Document Organization

For larger documents, consider this structure:

```markdown
# Document Title

## Table of Contents
- [Section 1](#section-1)
- [Section 2](#section-2)
- [Appendices](#appendices)

## Section 1

Content here...

### Subsection 1.1

More detailed content...

### Subsection 1.2

Additional content...

## Section 2

Second major section...

## Appendices

### Appendix A: Additional Resources

- [Resource 1](link1)
- [Resource 2](link2)
````

## Technical Implementation

### Supported Markdown Features

pg-press supports the following markdown elements:

| Element     | Syntax                   | Notes                 |
| ----------- | ------------------------ | --------------------- |
| Headers     | `# ## ###`               | Up to 6 levels        |
| Bold        | `**text**`               | Standard emphasis     |
| Italic      | `*text*`                 | Standard emphasis     |
| Underline   | `__text__`               | pg-press extension    |
| Inline Code | `` `code` ``             | Monospace formatting  |
| Lists       | `- item` or `1. item`    | Nested supported      |
| Blockquotes | `> quote`                | Multi-line supported  |
| Links       | `[text](url)`            | External and internal |
| Line Breaks | Two spaces or blank line | Paragraph separation  |

### Rendering Process

1. **Input Processing**: Raw markdown text is processed by the markdown engine
2. **HTML Generation**: Markdown syntax is converted to HTML elements
3. **Sanitization**: Generated HTML is sanitized for security
4. **Styling**: CSS classes are applied for consistent appearance
5. **Output**: Safe HTML is rendered in the browser

### Security Considerations

- All HTML output is sanitized to prevent XSS attacks
- Script tags and event handlers are automatically removed
- Only safe HTML elements are allowed in the output
- Link protocols are validated to prevent malicious links

### Performance

- Markdown processing is done on-demand during display
- Rendering performance is optimized for typical document sizes
- No performance impact when markdown is disabled
- Memory usage is minimal for standard documents

## Integration with pg-press Features

### Trouble Reports

When creating or editing trouble reports:

1. Check the "Markdown-Formatierung verwenden" checkbox
2. Use the markdown toolbar for common formatting
3. Preview your content before saving
4. Markdown content will be formatted in both web display and PDF exports

### Editor Features

The pg-press editor provides:

- **Toolbar Buttons**: Quick insertion of common markdown elements
- **Live Preview**: Real-time rendering of markdown content
- **Syntax Help**: Access to comprehensive markdown documentation
- **Keyboard Shortcuts**: Standard shortcuts for bold, italic, etc.

## Conclusion

Markdown in pg-press provides a powerful way to create rich, formatted content while maintaining simplicity and readability. By following the guidelines and best practices in this document, you can create professional-quality documentation and reports that are both visually appealing and easy to maintain.

### Key Takeaways

1. **Start Simple**: Use basic formatting first, then add complexity as needed
2. **Be Consistent**: Maintain consistent formatting throughout your documents
3. **Preview Often**: Always preview your content before publishing
4. **Follow Standards**: Use established markdown conventions for better compatibility
5. **Consider Your Audience**: Match your formatting complexity to your readers' needs

For additional help, use the interactive markdown guide available at `/help/markdown` in your pg-press installation.

---

**Document Version**: 1.0
**Last Updated**: 2024
**Compatibility**: pg-press v1.0+
