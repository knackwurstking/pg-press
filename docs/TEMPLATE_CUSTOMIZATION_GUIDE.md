# Template Customization Guide for pg-press

## Overview

This guide covers how to customize and extend templates in pg-press using the integrated `ui.min.css` utility framework. The help system serves as an excellent example of modern utility-first CSS implementation, demonstrating how to create professional, maintainable templates without writing custom CSS.

## Table of Contents

- [ui.min.css Utility System](#uimincss-utility-system)
- [Color System](#color-system)
- [Layout Utilities](#layout-utilities)
- [Typography System](#typography-system)
- [Component Classes](#component-classes)
- [Spacing System](#spacing-system)
- [Help Template Case Study](#help-template-case-study)
- [Customization Examples](#customization-examples)
- [Best Practices](#best-practices)
- [Advanced Techniques](#advanced-techniques)

## ui.min.css Utility System

The `ui.min.css` framework provides a comprehensive set of utility classes that enable rapid template development without writing custom CSS. The system is designed around utility-first principles, where classes do one thing well and can be composed together.

### Core Principles

1. **Utility-First**: Small, single-purpose classes
2. **Composable**: Classes work together seamlessly
3. **Consistent**: Predictable naming and behavior
4. **Semantic**: Color-based meaning system
5. **Responsive**: Built-in responsive design support

### Framework Benefits

- **Faster Development**: No need to write custom CSS
- **Consistent Design**: Unified design system
- **Maintainable**: Changes through class modifications
- **Smaller Bundle Size**: Minimal custom CSS required

## Color System

### Semantic Color Classes

The framework provides semantic colors that convey meaning:

```html
<!-- Text Colors -->
<h1 class="primary">Primary heading</h1>
<p class="secondary">Secondary text</p>
<span class="success">Success message</span>
<div class="warning">Warning notification</div>
<p class="destructive">Error message</p>
<small class="info">Informational text</small>
<span class="muted">Subtle text</span>
<strong class="contrast">High contrast text</strong>
```

### Color Modifiers

Colors can be modified with additional classes:

```html
<!-- Ghost: Text-only coloring without background -->
<h2 class="ghost primary">Primary text only</h2>
<button class="ghost success">Ghost success button</button>

<!-- Outline: Border coloring without fill -->
<div class="card outline info">Info card with border</div>
<button class="outline warning">Warning outline button</button>
```

### Background and Border Colors

```html
<!-- Background colors are applied automatically with semantic classes -->
<div class="primary">Primary background with white text</div>
<div class="success outline">Success border with transparent background</div>
<div class="warning ghost">Warning text color only</div>
```

## Layout Utilities

### Flexbox System

```html
<!-- Basic flex container -->
<div class="flex">
  <div>Item 1</div>
  <div>Item 2</div>
</div>

<!-- Flex with alignment -->
<div class="flex items-center justify-between">
  <div>Left content</div>
  <div>Right content</div>
</div>

<!-- Flex direction and wrapping -->
<div class="flex flex-col gap-md">
  <div>Stacked item 1</div>
  <div>Stacked item 2</div>
</div>

<!-- Common flex patterns from help template -->
<h2 class="text-xl text-bold mb-md flex items-center ghost primary">
  <i class="bi bi-lightning mr-sm"></i>
  Section Title
</h2>
```

### Grid System

```html
<!-- Basic grid -->
<div class="grid gap-md">
  <div>Item 1</div>
  <div>Item 2</div>
</div>

<!-- Example grid -->
<div class="example-grid gap-md">
  <div class="muted outline rounded p-sm">
    <code class="text-sm">**Bold text**</code>
  </div>
  <div class="border rounded p-sm">
    <strong>Bold text</strong>
  </div>
</div>
```

### Responsive Utilities

```html
<!-- Responsive grid that stacks on mobile -->
<div class="grid grid-cols-1 md:grid-cols-2 gap-md">
  <div>Column 1</div>
  <div>Column 2</div>
</div>

<!-- Hide/show on different screen sizes -->
<span class="hidden sm:inline">Desktop only text</span>
<div class="block md:hidden">Mobile only content</div>
```

## Typography System

### Text Sizing

```html
<h1 class="text-2xl">Extra large heading</h1>
<h2 class="text-xl">Large heading</h2>
<h3 class="text-lg">Medium heading</h3>
<p class="text-base">Base body text</p>
<small class="text-sm">Small text</small>
<span class="text-xs">Extra small text</span>
```

### Text Weight and Style

```html
<p class="text-light">Light weight text</p>
<p class="text-normal">Normal weight text</p>
<p class="text-medium">Medium weight text</p>
<p class="text-bold">Bold weight text</p>
<em class="italic">Italic text</em>
<span class="underline">Underlined text</span>
```

### Text Alignment

```html
<p class="text-left">Left aligned</p>
<p class="text-center">Center aligned</p>
<p class="text-right">Right aligned</p>
<p class="text-justify">Justified text</p>
```

## Component Classes

### Cards

```html
<!-- Basic card -->
<div class="card p-lg">
  <h3>Card title</h3>
  <p>Card content</p>
</div>

<!-- Elevated card with semantic color -->
<div class="card elevated success outline p-lg">
  <h3>Success card</h3>
  <p>This card has elevation and success styling</p>
</div>

<!-- Help template card example -->
<section class="card elevated p-lg">
  <h2 class="text-xl text-bold mb-md flex items-center ghost success">
    <i class="bi bi-play-circle mr-sm"></i>
    Interactive Demo
  </h2>
  <p class="muted mb-md">Content description here</p>
  <!-- Card content -->
</section>
```

### Buttons

```html
<!-- Basic buttons -->
<button class="primary">Primary button</button>
<button class="secondary">Secondary button</button>
<button class="success">Success button</button>

<!-- Button modifiers -->
<button class="outline warning">Warning outline</button>
<button class="ghost info">Info ghost button</button>
<button class="small primary">Small button</button>
<button class="large destructive">Large button</button>
```

### Form Elements

```html
<!-- Input styling -->
<input
  type="text"
  class="w-full border rounded p-sm"
  placeholder="Text input"
/>
<textarea class="w-full border rounded p-sm resize-y" rows="4"></textarea>
<select class="border rounded p-sm">
  <option>Select option</option>
</select>

<!-- Form layout -->
<div class="flex flex-col gap-sm">
  <label class="text-sm text-medium">Label</label>
  <input type="text" class="border rounded p-sm" />
</div>
```

## Spacing System

### Margin and Padding

```html
<!-- Margin classes -->
<div class="m-0">No margin</div>
<div class="m-sm">Small margin</div>
<div class="m-md">Medium margin (default)</div>
<div class="m-lg">Large margin</div>
<div class="m-xl">Extra large margin</div>

<!-- Directional margins -->
<div class="mt-lg">Margin top large</div>
<div class="mr-sm">Margin right small</div>
<div class="mb-md">Margin bottom medium</div>
<div class="ml-xs">Margin left extra small</div>

<!-- Horizontal/Vertical margins -->
<div class="mx-auto">Centered with auto margins</div>
<div class="my-lg">Vertical margin large</div>

<!-- Padding follows same pattern -->
<div class="p-lg">Large padding all around</div>
<div class="px-sm py-md">Small horizontal, medium vertical</div>
```

### Gap System

```html
<!-- Flexbox gaps -->
<div class="flex gap-sm">
  <div>Item with small gap</div>
  <div>Another item</div>
</div>

<!-- Grid gaps -->
<div class="grid gap-lg">
  <div>Grid item with large gap</div>
  <div>Another grid item</div>
</div>
```

### Custom Spacing Utilities

For consistent spacing in templates, you can define custom utilities:

```css
/* Custom spacing utilities (add to template styles) */
.space-y-lg > * + * {
  margin-top: 2rem;
}

.space-y-md > * + * {
  margin-top: 1rem;
}

.space-y-sm > * + * {
  margin-top: 0.5rem;
}
```

## Help Template Case Study

The markdown help template demonstrates excellent utility usage:

### Header Structure

```html
<header class="mb-xl">
  <h1 class="text-2xl text-bold mb-sm flex items-center">
    <i class="bi bi-markdown mr-sm ghost primary"></i>
    Markdown Hilfe
  </h1>
  <p class="text-lg muted">
    Eine vollständige Übersicht aller unterstützten Markdown-Funktionen
  </p>
</header>
```

**Utility breakdown:**

- `mb-xl`: Extra large bottom margin for section separation
- `text-2xl text-bold`: Large, bold text for main heading
- `mb-sm`: Small margin between title and subtitle
- `flex items-center`: Horizontal layout with vertical centering
- `ghost primary`: Primary color text without background
- `mr-sm`: Right margin for icon spacing
- `text-lg muted`: Large, muted text for subtitle

### Card Components

```html
<section class="card elevated p-lg">
  <h2 class="text-xl text-bold mb-md flex items-center ghost primary">
    <i class="bi bi-lightning mr-sm"></i>
    Schnellreferenz
  </h2>
  <div class="grid gap-sm">
    <div class="flex items-center justify-between p-sm border rounded">
      <code class="contrast px-sm py-xs rounded text-sm">**Fett**</code>
      <span class="muted">→</span>
      <strong>Fett</strong>
    </div>
  </div>
</section>
```

**Utility breakdown:**

- `card elevated p-lg`: Elevated card with large padding
- `text-xl text-bold mb-md`: Large bold heading with medium bottom margin
- `ghost primary`: Primary color without background
- `grid gap-sm`: Grid layout with small gaps
- `flex items-center justify-between`: Horizontal layout with space distribution
- `contrast px-sm py-xs rounded`: High contrast background with padding and rounding
- `border rounded`: Border with rounded corners

### Interactive Elements

```html
<div class="demo-grid gap-lg">
  <div>
    <label class="text-sm text-medium mb-sm block">Markdown eingeben:</label>
    <textarea class="w-full border rounded p-sm text-sm resize-y" rows="8">
      # Beispiel Markdown
    </textarea>
  </div>
  <div>
    <label class="text-sm text-medium mb-sm block">Vorschau:</label>
    <div class="border rounded p-sm overflow-y-auto markdown-content">
      <!-- Preview content -->
    </div>
  </div>
</div>
```

## Customization Examples

### Creating a Status Card

```html
<div class="card success outline p-md">
  <div class="flex items-center gap-sm mb-sm">
    <i class="bi bi-check-circle-fill success text-lg"></i>
    <h3 class="text-lg text-medium success">Task Completed</h3>
  </div>
  <p class="text-sm">
    Your task has been successfully completed and is ready for review.
  </p>
  <div class="flex justify-end mt-md">
    <button class="small success">View Details</button>
  </div>
</div>
```

### Warning Banner

```html
<div class="warning outline rounded p-md mb-lg">
  <div class="flex items-start gap-sm">
    <i class="bi bi-exclamation-triangle warning text-lg mt-xs"></i>
    <div class="flex-1">
      <strong class="text-medium warning">Important Notice</strong>
      <p class="text-sm mt-xs">
        This action cannot be undone. Please review carefully before proceeding.
      </p>
    </div>
  </div>
</div>
```

### Information Grid

```html
<div class="grid gap-md">
  <div class="info outline rounded p-md">
    <div class="flex items-center gap-sm mb-sm">
      <i class="bi bi-info-circle info"></i>
      <strong class="text-medium">Information</strong>
    </div>
    <p class="text-sm">Helpful information for users.</p>
  </div>

  <div class="secondary outline rounded p-md">
    <div class="flex items-center gap-sm mb-sm">
      <i class="bi bi-gear secondary"></i>
      <strong class="text-medium">Configuration</strong>
    </div>
    <p class="text-sm">System configuration details.</p>
  </div>
</div>
```

### Form Layout

```html
<form class="space-y-md">
  <div class="flex flex-col gap-xs">
    <label class="text-sm text-medium">Title</label>
    <input type="text" class="border rounded p-sm" placeholder="Enter title" />
  </div>

  <div class="flex flex-col gap-xs">
    <label class="text-sm text-medium">Description</label>
    <textarea
      class="border rounded p-sm resize-y"
      rows="4"
      placeholder="Enter description"
    ></textarea>
  </div>

  <div class="flex items-center gap-sm">
    <input type="checkbox" id="urgent" class="rounded" />
    <label for="urgent" class="text-sm">Mark as urgent</label>
  </div>

  <div class="flex justify-end gap-sm">
    <button type="button" class="ghost secondary">Cancel</button>
    <button type="submit" class="primary">Save</button>
  </div>
</form>
```

## Best Practices

### Utility Class Organization

1. **Layout First**: Start with layout utilities (flex, grid, positioning)
2. **Sizing**: Add width, height, and spacing utilities
3. **Typography**: Apply text sizing, weight, and color
4. **Background/Border**: Add visual styling utilities
5. **State**: Add hover, focus, and interactive states

```html
<!-- Good organization -->
<div class="flex items-center justify-between p-md border rounded bg-subtle">
  <h3 class="text-lg text-bold primary">Title</h3>
  <button class="small outline primary">Action</button>
</div>
```

### Semantic HTML with Utility Classes

Always use appropriate HTML elements with utility classes:

```html
<!-- Good: Semantic HTML with utilities -->
<article class="card p-lg">
  <header class="mb-md">
    <h2 class="text-xl text-bold">Article Title</h2>
    <time class="text-sm muted">Published 2024-01-01</time>
  </header>
  <main class="space-y-sm">
    <p>Article content...</p>
  </main>
</article>

<!-- Avoid: Non-semantic with utilities -->
<div class="card p-lg">
  <div class="mb-md">
    <div class="text-xl text-bold">Article Title</div>
    <div class="text-sm muted">Published 2024-01-01</div>
  </div>
</div>
```

### Responsive Design

Use responsive utilities for mobile-first design:

```html
<!-- Mobile-first responsive design -->
<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-md">
  <div class="card p-md">
    Mobile: 1 column, Tablet: 2 columns, Desktop: 3 columns
  </div>
</div>

<!-- Responsive text sizing -->
<h1 class="text-lg md:text-xl lg:text-2xl">Responsive heading</h1>

<!-- Responsive spacing -->
<section class="p-sm md:p-md lg:p-lg">Responsive padding</section>
```

### Color Usage Guidelines

1. **Primary**: Main brand color, call-to-action buttons
2. **Secondary**: Supporting content, less important actions
3. **Success**: Completed tasks, positive feedback
4. **Warning**: Caution, potential issues
5. **Destructive**: Errors, dangerous actions
6. **Info**: Neutral information, help text
7. **Muted**: Subtle text, secondary information
8. **Contrast**: High emphasis, important text

### Accessibility Considerations

```html
<!-- Ensure sufficient contrast -->
<div class="contrast bg-muted p-md rounded">
  High contrast text on muted background
</div>

<!-- Use semantic colors meaningfully -->
<div class="destructive outline p-sm rounded">
  <strong class="text-medium">Error:</strong>
  <span class="text-sm">Please fix the following issues</span>
</div>

<!-- Provide visual and text indicators -->
<button class="success flex items-center gap-sm">
  <i class="bi bi-check-circle" aria-hidden="true"></i>
  <span>Complete Task</span>
</button>
```

## Advanced Techniques

### Custom Component Classes

When utilities aren't enough, create reusable component classes:

```css
/* Add to your template styles */
.demo-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}

@media (max-width: 768px) {
  .demo-grid {
    grid-template-columns: 1fr;
  }
}

.example-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  align-items: start;
  gap: 1rem;
}

@media (max-width: 768px) {
  .example-grid {
    grid-template-columns: 1fr;
  }
}
```

### CSS Custom Properties Integration

Utilize CSS custom properties for theming:

```css
.custom-card {
  background: var(--ui-bg);
  border: 1px solid var(--ui-border);
  border-radius: var(--ui-radius);
  padding: var(--ui-spacing-md);
}

.custom-button {
  background: var(--ui-primary);
  color: var(--ui-primary-text);
  border-radius: var(--ui-radius);
  padding: var(--ui-spacing-sm) var(--ui-spacing-md);
}
```

### Template Inheritance Patterns

Create reusable template patterns:

```go
// Base card template
templ BaseCard(title string, variant string) {
    <div class={ "card p-lg", variant }>
        if title != "" {
            <h3 class="text-lg text-bold mb-sm">{ title }</h3>
        }
        { children... }
    </div>
}

// Usage
templ SuccessCard(title string) {
    @BaseCard(title, "success outline") {
        <p class="success">Success message content</p>
    }
}
```

### Performance Optimization

1. **Use utility classes instead of custom CSS** when possible
2. **Group related utilities** for better readability
3. **Avoid redundant utilities** (don't repeat defaults)
4. **Use semantic colors** for consistent theming

```html
<!-- Good: Concise and clear -->
<div class="flex items-center gap-sm p-md border rounded">
  <i class="primary"></i>
  <span>Content</span>
</div>

<!-- Avoid: Redundant utilities -->
<div
  class="flex flex-row items-center justify-start gap-sm p-md m-0 border border-solid rounded"
>
  <i class="primary"></i>
  <span class="text-base">Content</span>
</div>
```

## Conclusion

The ui.min.css utility framework provides a powerful foundation for creating consistent, maintainable templates in pg-press. By following the patterns established in the help template and adhering to best practices, you can create professional-quality interfaces without writing custom CSS.

### Key Takeaways

1. **Utility-First**: Compose interfaces using utility classes
2. **Semantic Colors**: Use meaningful color classes for consistent theming
3. **Responsive Design**: Build mobile-first with responsive utilities
4. **Component Patterns**: Create reusable patterns with utilities
5. **Performance**: Minimize custom CSS by leveraging the framework

The help template serves as an excellent reference implementation, demonstrating how to combine utilities effectively for complex interfaces while maintaining clean, readable code.

---

**Document Version**: 1.0
**Framework Compatibility**: ui.min.css v1.0+
**Last Updated**: 2024
