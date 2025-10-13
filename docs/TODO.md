# TODOs

- [x] Use templ.URL everywhere
- [x] Change tool add/edit form validations
- [x] Improve cycles handling for cassettes

- Implement a cassette binding system, allow binding to top tool
  - [x] Implement binding to the tools service
  - [x] Update the tool page and allow bind a tool of type (position top cassette) to a top tool, things need to match: "format"
  - [x] Removed top cassette from the "umbau" page for now
  - [x] Auto update press if a binding changed
  - [ ] Disallow double bindings, which is possible right now
  - [ ] Handle cycle tracking, currently ignored, need to manually add this for the cassette if needed
  - [ ] Throw error if a tool already is avaible at a position for a press (edit tool dialog)
  - [ ] Filter bindable tools list for format

- [ ] Add regenerations list management to tools for deletion (admin only)
- [ ] Using select for tools page filters
