## General

- [ ] Make sure the tool cycles offset is used properly in total cycle calculations
- [-] Kick templui files from this project for now
- [x] Handle page back events properly, hx reload, using the global js `window.hxTriggers`
- [x] Remove `shared.AllPressNumbers` and use `db.ListPress` or `db.ListPressNumbers`
- [x] Need to take care of regenerations in cycle calculations (tools)
- [x] Implement a rule: cycles will not be tracked for cassettes with an empty "code"
- [x] Remove the unknown special tool and use the press and tool cycle offsets instead

## Tools Page

- [x] Fix tools filtering on page load if query tools_filter is set
- [x] Sort presses from low to high
- [x] Sort tools in alphabetical order
- [x] Fix badges for all active tools (non cassettes)

## Press Page

- [ ] Add: Action button for adding MetalSheet
- [ ] Add: Action button for adding Cycle
- [ ] Show badges for tool anchors
- [ ] PDF Summary
- [x] Add a proper redirect after press deletion
- [x] Fix missing presses
- [x] Fix: Edit press removes all the active tools
- [x] Ignore non trackable cassettes in the cycles list and summary

## Tool Page

- [ ] Show badges for tool anchors
- [x] Binding: Cassettes already bound to a tool should be disabled (or removed)
- [x] Fix: Binding section always shows the first option as selected if bound

## Database

- [x] Script for database migration from an older SQLite database
