# Refactoring Tasks

## Shared contains all models, interfaces and types shared across services & more

- [ ] Attachment model missing

## Services

- note
  - note
    - [ ] Rename to notes
- press
  - cycles
  - press
    - [ ] Rename to presses
  - regeneration
    - [ ] Rename to regenerations
- tool
  - cassette
    - [ ] Rename to cassettes
  - lower-metal-sheet
    - [ ] Rename to metal-sheets-lower
  - metal-sheet
    - [ ] Rename to metal-sheets
  - regeneration
    - [ ] Rename to regenerations
  - tool
    - [ ] Rename to tools
  - upper-metal-sheet
    - [ ] Rename to metal-sheets-upper
- user
  - cookie
    - [ ] Rename to cookies
  - session
    - [ ] Rename to sessions
  - user
    - [ ] Rename to users
- attachment
  - attachments
    - [ ] Create service for attachments handling, no SQL, local file
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
    - [-] Regenerations edit
    - [-] Regenerations table
    - [x] Tool Binding
    - [ ] Cycles table, dialog missing
