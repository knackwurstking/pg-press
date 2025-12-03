# PG Press Roadmap

This document outlines the planned features and improvements for the PG Press
system.

## Fixes

- [x] Tools list does not reload when navigating backwards in history
- [x] Authentication security flaw - API key validation inconsistent with minimum length requirement
- [x] Cookie expiration logic error in user agent validation function
- [x] Database connection pooling may be too restrictive for concurrent access
- [x] Potential race conditions in cookie update middleware

## Features

- [x] Create a database migration system for managing schema changes
- [ ] Improve press regenerations page and add features
- [x] Add action buttons to press regenerations table (edit/delete)
- [x] Improve feed generation, and add missing ones (press regeneration add/delete)

## New Pages

- [ ] "Probleme Checklist" (troublechecklist)
- [ ] "Daten Sammlung" (datastorage)
    - Store dryer settings here, temperature, press speeds, etc.
- [ ] "Statistiken" (statistics dashboard)
- [ ] "Benutzerverwaltung" (user management)

## Refactoring

- Extract common UI components into reusable templates
  - [ ] Notes Card component (consolidate handler version with components/cards.templ)
  - [ ] Section container templates with consistent styling
  - [ ] Action bar components with edit/delete functionality
  - [ ] Standard form controls for status display and editing
  - [ ] Table components with action buttons

## Performance Improvements

- [ ] Optimize database queries with proper indexing, 
    - For now use the scripts for this [scripts](/scripts)
- [ ] Add caching layer for frequently accessed data

## Security Enhancements

- [ ] Implement CSRF protection for all forms
- [ ] Improve session management and authentication flow
- [ ] Add input sanitization for all user-provided data
