# Changelog

## v2.1.15 [2025-03-01]

- Fixed revision 30 limit.
- Fixed pushing to gist repo if nothing has changed.
- Fixed files not removed from gist repo after push.

## v2.1.14 [2025-03-31]

- Fixed android system back button handling

## v2.1.13 [2025-03-30]

- Fixed app-bar hidden behind the android status bar on top

## v2.1.12 [2025-03-28]

- Fixed metal sheet print to pdf for android build
- Fixed flakes print to pdf for android build

## v2.1.11 [2025-03-28]

- Add android build (capacitor)

## v2.1.10 [2025-03-05]

**Fixed**:

- Missing fonts/ path in font assets path

## v2.1.9 [2025-03-05]

**Fixed**:

- Icon asset paths

## v2.1.8 [2025-03-04]

**Fixed**:

- Add Bootstrap icons and CSS to the project and update the service worker assets.

## v2.1.7 [2025-03-03]

- Fix integer parsing for flakes dialog input elements that allow floats.

## v2.1.6 [2025-03-03]

- Fixed misplaced resolve in cancel button callback for the vis-data-entry dialog

## v2.1.5 [2025-03-02]

- Fixed a bug introduced in v2.1.4

## v2.1.4 [2025-02-27]

- Fixed metal sheet filter data-index
- Removed unused code for cancel button in dialogs with a from

## v2.1.3 [2025-02-14]

- Fixed missing class name in "special-flakes-entry" dialog

## v2.1.2 [2025-02-13]

- Fixed bootstrap icons href

## v2.1.1 [2025-02-11]

- Fixed vis and vis-data dialog resolve, missing data in edit mode

## v2.1.0 [2025-02-10]

- Clean up dialogues
- Added new dialog for metal sheet filters
- Fixed flakes (special) dialog, switched inputs main - c1
- Fixed bookmarks store initialization, always resets even if available

## v2.0.4 [2025-02-07]

- Fixed metal sheet entry dialog form input for thickness
- Fixed vis conversion for thickness floats

## v2.0.3 [2025-02-07]

- Re-render metal sheet page after edit sheet dialog submission if list key remains unchanged.

## v2.0.2 [2025-02-07]

- Fixed wrong press initialization for metal sheet dialog

## v2.0.1 [2025-02-07]

- Updated confirmation message for updating
- Changed overflow behavior to only block the y axis

## v2.0.0 [2025-02-07]

- Updated UI library to [v2.0.0](https://github.com/knackwurstking/ui/tree/dev?tab=readme-ov-file)
- Kicked lit library

## v1.2.1 [2024-01-16]

**Fixed**:

- Fixed sorting of metal sheet list keys, active ones not always on top

## v1.2.0 [2024-01-16]

**Changed**:

- Changed metal sheet list sorting, active items will be moved to the top of the list. Need to
  refresh metal sheet lists data, just push the update button for metal sheets :)

**Fixed**:

- Fixed missing table headers after creating a new metal sheet list

## v1.1.3 — [2025-01-10]

**Fixed**:

- Fixed overwriting of “Metal Sheets” and “Vis Data” lists in some cases.

## v1.1.2 — [2024-11-20]

- [ui v1.1.1](https://github.com/knackwurstking/ui)

**Changed**:

- Changed input type from number to text

## v1.1.1 — [2024-11-18]

**Changed**:

- Changed Highlighting for flakes percentages

**Fixed**:

- Fixed missing search bar in safari, flakes page positioning
- Fixed sorting for flakes entries
- Fixed flakes data validation
- Fixed zero consumption on tower slots
- Fixed "flakes" dialog, using keyed for inputs on entry object

## v1.1.0 — [2024-11-18]

- [ui v1.1.0](https://github.com/knackwurstking/ui)

**TODOs**:

- [x] Change dev app icons
- [x] Replace "pg-vis-dev.github.io" with "pg-vis.github.io" before publishing

**General**:

- Switched code base to typescript using <https://lit.dev/>
- Take bookmarks data from the "Vis" if possible
- Renamed "Vis Listen" to "Vis Bookmarks"
- Changed "Vis Data": "title" field is now optional
- Changed gist handling

**Added**:

- Added a new theme "gruvbox" & a theme switcher
- Added vis data edit to the product page
- Added a search bar to the metal sheets page
- Added confirmation dialogs to delete action on drawer items
- Added new drawer group: "Spezial" for "Flakes"
- Added download (printer) button to metal sheets and flakes pages

## v1.0.0 — [2024-08-12]

- [ui version v0.2.16](https://github.com/knackwurstking/ui)

**General**:

- Removed "date" from types `AlertList`, `MetalSheetList`, `VisList` and `VisData`
- Improved error handling
- Added validation for `VisData` files
- Changed `<ul>` click handling for "alert-lists"
- Changed `<ul>` click handling for "vis"
- Changed `<ul>` click handling for "vis-lists"
- Changed `<ul>` click handling for "vis-data"
- Added some German translations
- Enabled edit icon for vis-data (rename lists)
- Added the vis product page to vis-data
- Added gist handler for the vis-data group (drawer)
- Checking for build numbers after each app start
- Added some preview for active vis-data "filters"
- Added sorting for vis-data (drawer and page)
- Changed vis filenames (download)
- Added drag and drop for vis-lists
- Code clean up

**Fixes**:

- Fixed bookmarks dialog sorting
- Fixed error handling
- Fixed `VisDataEntryPage` submit, original entries will now be removed
- Fixed product page rendering
- Fixed drag and drop for metal-sheets

## v0.0.32 — [2024-08-06]

**Added**:

- Drawer group "Vis Data"
- Page "Vis Data"

**Changed**:

- New button text (content) for metal-sheet-lists drawer group
- [ui version v0.2.15](https://github.com/knackwurstking/ui)

**Removed**:

- Import button from drawer group "Vis Listen"
- New button from drawer group "Vis Listen"

## v0.0.31 — [2024-08-05]

**Added**:

- Bookmarks (icon) button for the product page
- Drawer group "Vis Listen"
- New page `VisListsPage`

**Changed**:

- Code formatting
- [ui version v0.2.14](https://github.com/knackwurstking/ui)

**Fixed**:

- `PushDialog` constructor query select, moved to render method
- `NewVisDataDialog` constructor query select, moved to render method
- `SearchBar` height

## v0.0.30 — [2024-07-28]

**Added**:

- Fold and unfold drawer groups (initially folded)
- Spinner while vis list is loading
- Export handler for drawer vis group

**Changed**:

- `SearchBar` is now always visible for the Vis page
- [ui version v0.2.13](https://github.com/knackwurstking/ui)

**Fixed**:

- Titles for all meta-sheet related dialogs
- `SearchBar` auto focus
- Broken `ImportDialog` submit and dialog titles

**Removed**:

- Delete button from the metal-sheet-list page
- Share for the vis page

## v0.0.29 — [2024-07-27]

**Changed**:

- Code clean up
- [ui version v0.2.12](https://github.com/knackwurstking/ui)

## v0.0.28 — [2024-07-20]

**Changed**:

- Drawer will now open after go back to no page
- [ui version v0.2.11](https://github.com/knackwurstking/ui)

**Fixed**:

- Scroll (vis) list (hanging)
- Close drawer after each click

**Improved**:

- Search-bar open/close (Performance)

**Removed**:

- Refresh on pull down (overscroll-behavior: none)
- Pointer lock on scroll for drawer and vis list

## v0.0.27 — [2024-07-19]

**Added**:

- Ignore case sensitive for all regex searches
- Pointer events block on scroll for "vis" page and the "pg-drawer"

**Fixed**:

- PGDrawer component HTML, this should fix some performance issues
- VisPage list rendering, this should fix some performance issues

**Changed**:

- [ui version v0.2.10](https://github.com/knackwurstking/ui)

## v0.0.26 — [2024-07-19]

- Added a vis-search-bar
- [ui version v0.2.9](https://github.com/knackwurstking/ui)

## v0.0.25 — [2024-07-19]

- Fixed app-bar title (vis-page)
- Fixed git push

## v0.0.24 — [2024-07-18]

- Fixed typo in pull alert, "already up to..."
- Changed vis title when imported from a local text file
- Add vis data entries as secondary text
- [ui version v0.2.8](https://github.com/knackwurstking/ui)

## v0.0.23 — [2024-07-15 Quick Fix]

- Fixed vis import from gist

## v0.0.22 — [2024-07-15]

- Allow downgrade and "repull" on a (gist) pull
- Revision will raise up once after each successful (gist) push
- Added vis and product page
- [ui version v0.2.7](https://github.com/knackwurstking/ui)

## v0.0.21 — [2024-07-12]

- Fixed press 0 not showing for drawer items
- Zip metal-sheet files before downloading

## v0.0.20 — [2024-07-12]

- Moved press info to primary text (metal-sheet-lists)
- Changed cursor to pointer for metal-sheet table rows
- Changed metal-sheet table row role to "button"
- [ui version v0.2.6](https://github.com/knackwurstking/ui)

## v0.0.19 — [2024-07-11]

- Removed `gistID` from data: `MetalSheetList`, `AlertList`
- Fixed gist push, check differences and sort before
- Fixed press editing for metal-sheet-lists
- Data will be cleared before importing, or pull, a new gist revision
- Added icon-button alongside the import buttons, for exporting files
- [ui version v0.2.5](https://github.com/knackwurstking/ui)

## v0.0.18 — [2024-07-10]

- Added the missing delete button for metal-sheet-lists entries
- Added press info to the `appBarTitle` store for metal-sheet-lists pages

## v0.0.17 — [2024-07-09]

- Added gist pull and push handlers
- [ui version v0.2.4](https://github.com/knackwurstking/ui)

**TODO**:

- Add pull and push handler for Gist IDs
- Allow "push" to files instead of gist

## v0.0.16 — [2024-07-05]

- Added confirmation dialog before deleting a drawer item
- Kicked notes from the `MetalSheetList`
- Replaced date with a timestamp from (downloaded) filename
- Updated "date" for metal-sheet-lists data after a change
- Fixed data share where `canShare` does not exist
- Fixed edit dialog (submit) for metal-sheet-lists
- Add user input via `innerHTML` instead of `innerText`
- [ui version v0.2.3](https://github.com/knackwurstking/ui)

## v0.0.15 — [2024-07-03]

- Share the current metal-sheet-list

## v0.0.14 — [2024-07-03]

- Improved overflow handling for the metal-sheet-lists-page
- Fixed "Import" buttons (broken in v0.0.13)

## v0.0.13 — [2024-07-03]

- Create new metal-sheet-lists
- Edit current metal-sheet-list
- Add new entries to the metal-sheet-list
- Delete the current metal-sheet-list
- Edit a single entry from the metal-sheet-list table, just click on the table entry
- Drag'n'Drop for metal-sheet-list table entries
- [ui version v0.2.2](https://github.com/knackwurstking/ui)

## v0.0.12 — [2024-06-28]

- Fixed app-bar items for the "metal-sheet-lists" page
- Added revisions number for each drawer group
- Added drawer gist item for "alert-lists"

**Tweaks**:

- Reimport all (gists) data

## v0.0.11 — [2024-06-27]

- Fixed deletion for drawer group items
- Fixed the not implemented error

## v0.0.10 — [2024-06-27]

- PATCH gist database repo, needs a token with write access for gist repos
- Update drawer section: "Blech Listen", new item added (pull/push gist data)

## v0.0.9 — [2024-06-27]

- Fixed padding for "Blech Listen" page

## v0.0.8 — [2024-06-26]

- Enable drawer section: "Blech Listen"
- Added "Blech Listen" page(s) (READONLY for now, work-in-progress)
- Set html language to "de"
- [ui version v0.2.1](https://github.com/knackwurstking/ui)

## v0.0.7 — [2024-06-23]

- Added storage handling for the search-bar
- [ui version v0.2.0](https://github.com/knackwurstking/ui)

## v0.0.5 — [2024-06-23]

- Allow importing multiple files at once
- Use blur drawer backdrop filter

## v0.0.4 — [2024-06-23]

- Import dialog for downloading a gist file from GitHub

## v0.0.3

- Added history handling

## v0.0.2

- Changed svg icons
- Removed search-bar submit button

## v0.0.1

- Added "Alert List" page(s)
