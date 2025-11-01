- [x] Replace `datalist` input elements with `select` elements for the "umbau" page.
  - Add input handling to disable tools not possible to change. For example, if the top tool is from a 100x100 format, the bottom tool must also be from a 100x100 format.

- [x] Improve model types for IDs
  - Attachment IDs
  - Cycle IDs
  - Feed IDs
  - Metal Sheet IDs
  - Modification IDs
  - Note IDs
  - Regeneration IDs
  - Tool IDs
  - Trouble Report IDs
  - User IDs

- [x] Create a new feed entry after a regeneration is removed.

- [x] Create helpers for resolved data types (tool, regeneration, ...)
  - `ResolveRegeneration`
  - `ResolveTool`

- [x] Create a new feed entry after editing a regeneration (reason).

- [x] Make the notes page reload content after editing or deleting a note.

- [x] When the filter input is in focus, use `scrollIntoView` on each input change.

- [x] Improve the notes flex grid settings (grow, shrink).

- [x] Rename this module from `/pgpress` to `/pg-press`.

- [ ] Create a new page: "Probleme Checklist" (votings, attachments, comments).

- [ ] Create a new page: "Daten Sammlung", for storing stuff like dryer temperatures (markdown editor support).

- [x] Remove the logging package and use logger just like in the picow-led project.

- [ ] Modify the logging flags (env), remove the DEBUG flag and replace it with a LOG_LEVEL flag.

- [ ] Improve/Simplify error handling(?), see picow-led project.

- [ ] Change the base dialogs' error handling, see picow-led project.
