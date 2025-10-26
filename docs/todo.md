- [x] Replace the `datalist` input elements with `select` elements for the "umbau" page.
- [x] Also add some input handling for the "umbau" page select elements, which will disable tools not possible to change. For example, if the top tool is from a 100x100 format, the bottom tool must also be from a 100x100 format.

- [x] Improve model types for ID's
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

- [x] Create a new feed entry after a regeneration gets removed

- [x] Create helpers for resolved data types (tool, regeneration, ...)
  - ResolveRegeneration
  - ResolveTool

- [x] Create a new feed entry after editing a regeneration (reason)
- [x] Make the notes page reload content after editing or deleting a note.
- [x] When the filter input is in focus, try to use `scrollIntoView` on each input change.
- [x] Improve the notes flex grid, (grow, shrink settings)
- [ ] Create a new page: "Probleme Checklist" (votings, attachments, comments)
- [ ] Create a new page: "Daten Sammlung", this is a place where stuff like dryer temperatures can be stored (markdown editor support)
