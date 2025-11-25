# PG Press Roadmap

This document outlines the planned features and improvements for the PG Press system.

## Bug Fixes

- [x] Tool total cycles incorrect when cycles were added from new to old (backwards)
- [x] Change tool regenerations table heading from "Grund" to "Bemerkung"
- [ ] Tools list does not reload when navigating backwards in history
- [ ] Notes management page needs improvements after editing or deleting a note

## Features

- [x] Resolve binding tool and update press and tool pages to enable regeneration counter
- [x] To keep it simple, remove (UI) tool cycles for "top cassette"
- [ ] Tools List: Group tools by state: active, available, dead, and regenerating
- [ ] Tools List: Add filter utils (youtube like) for filtering tools contains notes, regenerations, ... (or something like this)
- [ ] Need a dead press (nr. -1) (admin only) where I can store cycles where a press is no longer available, need to change the cycles handling, I can no longer calculate partial cycles based on press total cycles [WIP]
- [ ] Create a new page: "Probleme Checklist" (votes, attachments, comments, close/open)
- [ ] Create a new page: "Daten Sammlung" for storing dryer temperatures (Markdown editor support)

## Chore

- [x] Remove outdated server script "./scripts/server"

## Style Improvements

- [x] Tool List Item: Update item styles to reduce card-like appearance
- [x] Tool List Item: Display regeneration count
- [x] Tool List Item: Display notes count (all 3 levels)

## Refactoring

- [x] Move handlers error utils to the errors package
- [x] Move shared components to "/handlers/components", using subdirectories as needed
- [x] Services need to properly return the not found error type everywhere possible
- [x] Remove useless stuff from models (user, note, modification, feed, metalsheet, cookie, attachment)
- [x] Move all 'page-trouble-reports' related JavaScript to '/assets/js/page-trouble-reports.js'
- [x] Cleanup middleware (cache control)
- [ ] Create a `ResolvedTroubleReport` type and replace `TroubleReportWithAttachments` with this
- [ ] Migrate `Attachment.ID` from string to int64, also need to migrate the database table for this
