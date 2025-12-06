# PG Press Roadmap

This document outlines the planned features and improvements for the PG Press
system.

## New Pages

- [ ] "Probleme Checklist" (troublechecklist)
- [ ] "Daten Sammlung" (datastorage)
  - Store dryer settings here, temperature, press speeds, etc.
- [ ] "Statistiken" (statistics dashboard)
- [ ] "Benutzerverwaltung" (user management)

## Performance Improvements

- [ ] Optimize database queries with proper indexing,
  - For now use the scripts for this [scripts](/scripts)

## Refactoring

- [ ] Split the dialog handler, on file per dialog
- [ ] Implement a validation error, update all validation functions/methods
- [ ] Implement a already exists error (bad request), and update the already exists check in `commands-user.go`
