## General

- [ ] Make sure the tool cycles offset is used properly in total cycle calculations
- [x] Fix: tool anchor binding cassette thickness if min thickness is zero
- [x] Fix cache control, especially for the editor page (trouble reports)

## Presses

```
Presse: 5
Type: SACMI
Code: PH5000

# Key: 5SACMIPH5000

Presse: 5
Type: SITI
Code: PH4000

# Key: 5SITIPH4000
```

- [x] Add to presses table: "code"
- [x] ~Make presses table columns unique (union) `<press-number>, <press-type>, <press-code>`~
- [x] Need to fix all conversion scripts after this

## Press Page

- [ ] Add: Action button for adding MetalSheet
- [ ] Add: Action button for adding Cycle
- [ ] Show badges for tool anchors
- [ ] PDF: cycle Summary
- [ ] PDF: press page summary

## Tool Page

- [ ] Show badges for tool anchors
- [ ] Add a revive tool/cassette button to the page actions bar
- [ ] PDF: tool summary(?)
