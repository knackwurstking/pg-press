# Refactoring Tasks

- [ ] Add to env: htmx trigger values, just grep for `urlb.SetHXTrigger` (and `hx-trigger`)

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
- [x] dialogs
  - [x] tool dialog (new/edit)
  - [x] cassette dialog (new/edit)
  - [ ] metal sheet dialog (new/edit)
  - [ ] cycle dialog (new/edit)
  - [x] note dialog (new/edit)
- [x] tool
  - [-] Notes: add, rendering, reload
  - [ ] Metal Sheets, dialog missing
  - [x] Regenerations edit
  - [x] Regenerations table
  - [x] Tool Binding
  - [ ] Cycles table, dialog missing
