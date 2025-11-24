# PG Press Roadmap

This document outlines the planned features and improvements for the PG Press system.

## Features

### Tools Management

- [ ] Group tools by state: active, available, dead, and regenerating
- [ ] Tools List Filter: Add checkbox to show already bounded "top cassette" tools

### New Pages

- [ ] Create a new page: "Probleme Checklist" (votes, attachments, comments, close/open)
- [ ] Create a new page: "Daten Sammlung" for storing dryer temperatures (Markdown editor support)

## Development

### Chore

- [x] Remove outdated server script "./scripts/server"

### Style Improvements

- [x] Tool List Item: Update item styles to reduce card-like appearance
- [x] Tool List Item: Display regeneration count
- [x] Tool List Item: Display notes count (all 3 levels)

### Refactoring

- [x] Move handlers error utils to the errors package
- [x] Move shared components to "/handlers/components", using subdirectories as needed
- [x] Services need to properly return the not found error type everywhere possible
- [x] Remove useless stuff from models (user, note, modification, feed, metalsheet, cookie, attachment)
- [x] Move all 'page-trouble-reports' related JavaScript to '/assets/js/page-trouble-reports.js'
- [x] Cleanup middleware (cache control)
- [ ] Create a `ResolvedTroubleReport` type and replace `TroubleReportWithAttachments` with this
- [ ] Migrate `Attachment.ID` from string to int64, also need to migrate the database table for this
