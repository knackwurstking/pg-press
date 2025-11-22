# PG Press Roadmap

This document outlines the planned features and improvements for the PG Press system.

## Feature Roadmap

### Tools Management
- [ ] Add tools pinning feature: pinned tools appear on top in a foldable section
- [ ] Group tools by state: active, available, dead, and regenerating
- [ ] Group tools by position: "top", "bottom", or "top cassette"

### New Pages
- [ ] Create a new page: "Probleme Checklist" (votings, attachments, comments, close/open)
- [ ] Create a new page: "Daten Sammlung" for storing dryer temperatures (markdown editor support)

## Development Plans

### Chore
- [ ] Remove outdated server script "./scripts/server"

### Feat
- [ ] Implement tool grouping functionality

### Style
- [ ] Change tools list item styles, showing regeneration count and notes count (for all 3 levels)

### Refactor
- [ ] Refactor code to improve maintainability
- [ ] Move shared components to "/handlers/components", using subdirectories as needed
- [ ] Tools list: Server-side search
