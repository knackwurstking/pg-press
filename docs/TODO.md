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
- [ ] dialogs
    - [x] tool dialog (new/edit)
    - [-] cassette dialog (new/edit)

## Recommended Improvements

### 1. **Complete Refactoring Tasks**
- [ ] Finish all renaming tasks in `docs/TODO.md`
- [ ] Standardize naming conventions across models and services
- [ ] Implement missing attachment model functionality

### 2. **Documentation and Maintenance**
- [ ] Update README to better describe the application's purpose and functionality

### 3. **Performance Optimization**
- [ ] ~Implement database connection pooling if not already done~
- [ ] Add caching mechanisms where appropriate
- [ ] Optimize queries in services

### 4. **Security Considerations**
- [ ] Ensure proper input validation and sanitization
- [ ] Check for potential SQL injection vulnerabilities
