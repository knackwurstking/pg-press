# Review

Based on the analysis, here are the main CSS class issues found:

## High Priority Issues:

1. Unused classes - `app-bar-center` in main-layout.css:7
2. Unused classes - `attachment-preview-icon`, `preview-thumbnail` in page-trouble-reports.css  
3. Temporary classes - `[uo]l-item`, `ol-item` in markdown.templ need cleanup

## Missing Custom Class Definitions:

- `attachment-name`, `attachment-size`, `cookies-details`, `cookies-summary`, `cookies-table`
- `editor-container`, `editor-form`, `empty-state`, `markdown-content`
- `note-date`, `note-importance`, `notes-list`, `press-link`, `tool-link`

## Potential Typos:

- `delete` should probably be `destructive`
- `hover:contrast` is suspicious
- Non-standard Tailwind classes like `gap-xs`, `mb-sm`, `mt-lg`

Recommendation: Add the missing custom class definitions and remove unused classes to clean up the CSS architecture.
