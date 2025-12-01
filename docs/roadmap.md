# PG Press Roadmap

This document outlines the planned features and improvements for the PG Press
system.

## Fixes

- [ ] Tools list does not reload when navigating backwards in history(?)
- [ ] Notes management page needs improvements after editing or deleting a note(?)
- [x] Authentication security flaw - API key validation inconsistent with minimum length requirement
- [x] Cookie expiration logic error in user agent validation function
- [x] Database connection pooling may be too restrictive for concurrent access
- [x] Potential race conditions in cookie update middleware

## Features

- [x] Create a database migration system for managing schema changes
- [ ] Improve press regenerations page visibility and add features
- [ ] Add action buttons to press regenerations table (edit/delete)

### New Pages

- [ ] "Probleme Checklist" (troublechecklist)
- [ ] "Daten Sammlung" (datastorage)
    - Store dryer settings here, temperature, press speeds, etc.

## Refactoring

- [ ] Create a `ResolvedTroubleReport` type and replace
      `TroubleReportWithAttachments` with this
- [ ] ~Migrate `Attachment.ID` from string to int64, also need to migrate the
      database table for this~
