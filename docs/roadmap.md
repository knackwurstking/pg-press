# PG Press Roadmap

This document outlines the planned features and improvements for the PG Press
system.

## Fixes

- [ ] Tools list does not reload when navigating backwards in history(?)
- [ ] Notes management page needs improvements after editing or deleting a note(?)
- [ ] Authentication security flaw - API key validation inconsistent with minimum length requirement
- [ ] Cookie expiration logic error in user agent validation function
- [ ] Database connection pooling may be too restrictive for concurrent access
- [ ] Potential race conditions in cookie update middleware

## Features

### New Pages

- [ ] Create a new page: "Probleme Checklist" (votes, attachments, comments, close/open)
- [ ] Create a new page: "Daten Sammlung" for storing dryer temperatures
      (Markdown editor support)

### Press Regeneration

Implement the `PressRegenerations` system, which will reset the press total
cycles back to zero just like the `ToolRegenerations` system just with presses

- [x] Rename the current `Regeneration` model to `ToolRegeneration`
- [x] Add a `PressRegeneration` model
- [x] Add a `PressRegenerations` service
- [x] Remove the dead press (-1) stuff again
- [x] Press Cycles ordering needs to be changed; I need to sort by date,
      not total cycles (find "ORDER BY total_cycles")
- [x] Fix to total cycles calculation
- [x] UI: Submit a press regeneration
- [x] UI: Add press regeneration section to the press page
- [x] Fix the `GetPartialCycles` press-cycles method, need to handle the press regenerations 

## Refactoring

- [ ] Create a `ResolvedTroubleReport` type and replace
      `TroubleReportWithAttachments` with this
- [ ] ~Migrate `Attachment.ID` from string to int64, also need to migrate the
      database table for this~
- [x] Change url builders in utils package, also update handlers
- [ ] Refactor the handler+components structure, also rename the components directory to templates [WIP]
    - [x] Auth
    - [ ] Dialogs
        - [x] cycle dialog
        - [ ] metal-sheet dialog
        - [ ] note dialog
        - [ ] tool dialog
        - [ ] tool-regeneration dialog
    - [ ] Editor
    - [ ] Feed
    - [ ] Help
    - [ ] Home
    - [ ] MetalSheets
    - [ ] Nav
    - [ ] Notes
    - [x] Press
    - [ ] PressRegenerations
    - [ ] Profile
    - [ ] Tool
    - [ ] Tools
    - [ ] TroubleReports
    - [ ] Umbau
    - [ ] WSFeed
