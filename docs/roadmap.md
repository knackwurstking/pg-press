# PG Press Roadmap

This document outlines the planned features and improvements for the PG Press
system.

## Fixes

- [ ] Tools list does not reload when navigating backwards in history
- [ ] Notes management page needs improvements after editing or deleting a note

## Features

### New Pages

- [ ] Create a new page: "Probleme Checklist" (votes, attachments, comments, close/open)
- [ ] Create a new page: "Daten Sammlung" for storing dryer temperatures
      (Markdown editor support)

### Press Regeneration

Implement the `PressRegenerations` system, which will reset the press total
cycles back to zero just like the `ToolRegenerations` system just with presses

- [x] Rename the current `Regeneration` model to `ToolRegeneration`
- [x] Add a `PressRegeneration` model
- [x] Add a `PressRegenerations` service
- [x] Remove the dead press (-1) stuff again
- [x] Press Cycles ordering needs to be changed; I need to sort by date,
      not total cycles (find "ORDER BY total_cycles")
- [x] Fix to total cycles calculation
- [ ] UI: Add press regeneration section where the admin can start/stop a
      regeneration, or retroactively add a regeneration [WIP: ready to go,
      needs testing]
- [ ] Fix the `getTotalCycles tool handler method to check the last tool regeneration

## Refactoring

- [ ] Create a `ResolvedTroubleReport` type and replace
      `TroubleReportWithAttachments` with this
- [ ] Migrate `Attachment.ID` from string to int64, also need to migrate the
      database table for this
