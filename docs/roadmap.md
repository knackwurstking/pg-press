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
- [ ] Add action buttons to press regenerations table (edit/delete) [WIP]
- [ ] Improve feed generation, and add missing ones (press regeneration add/delete)

## New Pages

- [ ] "Probleme Checklist" (troublechecklist)
- [ ] "Daten Sammlung" (datastorage)
    - Store dryer settings here, temperature, press speeds, etc.
- [ ] "Statistiken" (statistics dashboard)
- [ ] "Benutzerverwaltung" (user management)

## Refactoring

- [ ] Extract common UI components into reusable templates

## Performance Improvements

- [ ] Optimize database queries with proper indexing
- [ ] Add caching layer for frequently accessed data

## Security Enhancements

- [ ] Implement CSRF protection for all forms
- [ ] Improve session management and authentication flow
- [ ] Add input sanitization for all user-provided data

## Testing

- [ ] Add unit tests for all service functions
