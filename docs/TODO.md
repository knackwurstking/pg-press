# Refactoring Tasks

## Shared contains all models, interfaces and types shared across services & more

- [ ] Attachment model missing
  - Create service for attachments handling, no SQL, local file
    system storage? (@ /var/www/pg-press/attachments)

## Handlers

Fix all handler and templates

- [x] home
- [x] auth
- [x] profile
- [x] tools
- [-] dialogs
  - [x] tool dialog (new/edit)
  - [x] cassette dialog (new/edit)
  - [ ] metal sheet dialog (new/edit)
  - [ ] cycle dialog (new/edit)
  - [ ] note dialog (new/edit)
- [x] tool
  - [ ] Notes, dialog missing
  - [ ] Metal Sheets, dialog missing
  - [x] Regenerations edit
  - [x] Regenerations table
  - [x] Tool Binding
  - [ ] Cycles table, dialog missing
