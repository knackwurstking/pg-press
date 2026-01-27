## General

- [ ] Check if hx-trigger are properly set (handlers & templates)
- [x] Need to take care of regenerations in cycle calculations (tools)
- [-] Verify actions for non admins
- [x] Implement a rule: cycles will not be tracked for cassettes with an empty "code"
- [x] Remove the unknown special tool and use the press and tool cycle offsets instead
- [ ] Make sure the tool cycles offset is used properly in total cycle calculations

## Tools Page

- [x] Sort presses from low to height
- [x] Sort tools in alphabetical order
- [x] Fix badges for all active tools (non cassettes)

## Press Page

- [ ] Add: Action button for adding MetalSheet
- [ ] Add: Action button for adding Cycle
- [ ] Show badges for tool anchors
- [ ] PDF Summary
- [x] Fix: Edit press removes all the active tools
- [x] Ignore non trackable cassettes in the cycles list and summary

## Tool Page

- [x] Binding: Cassettes already bound to a tool should be disabled (or removed)
- [x] Fix: Binding section always shows the first option as selected if binded
- [ ] Show badges for tool anchors

## Database

- [x] Script for database migration from an older SQLite database
