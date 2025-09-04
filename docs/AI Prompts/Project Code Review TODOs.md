## User

[@pg-vis](file:///Users/knackwurstking/Git/pg-vis/) List todo comments i left in this project an tell me what you think

## Assistant

Of course! I'll look for `TODO` comments in your project and share my thoughts on them.

## Assistant

Okay, I've found a number of `TODO` comments in the project. I'll go through them and give you my thoughts. I'll ignore the `bootstrap-icons.min.css` file since that's a third-party library.

Here's a summary of the TODOs and my analysis:

### 1. Refactor `cookie.Service` and `feed.Service` to fit `interfaces.DataOperations`

- **Files:** `internal/database/services/cookie/cookie.go` and `internal/database/services/feed/feed.go`
- **Comment:** `// TODO: Need to fix this type to fit the interfaces.DataOperations[*models.Cookie]`
- **My take:** This is a good idea for consistency. The `DataOperations` interface provides a standard way to interact with your database services. Both `cookie.Service` and `feed.Service` have custom methods that don't align with the `Add`, `Update`, and `Delete` methods in the interface. Refactoring them to implement this interface would make your data access layer more predictable and easier to work with.

### 2. Make `press` argument optional in `NewPressCycle`

- **Files:** `internal/database/models/press-cycles.go` and `internal/web/htmx/tools.go`
- **Comment:** `// TODO: I need to make the press argument optional, because i will allow editing tools not active`
- **My take:** This seems like a necessary change to support editing tools that are not currently active in a press. Making the `press` argument optional would allow you to create or modify a `PressCycle` without associating it with a specific press, which aligns with your goal of editing inactive tools.

### 3. Implement `DELETE /htmx/tools/cycle/delete`

- **File:** `internal/web/htmx/tools.go`
- **Comment:** `// TODO: Add "DELETE /htmx/tools/cycle/delete?cycle_id=%d"`
- **My take:** This is a missing piece of functionality in your `tools` HTMX handler. Implementing this would allow users to delete press cycles directly from the UI, which is a crucial part of managing tool data.

### 4. Improve form data parsing in `troublereports_dialog.go`

- **File:** `internal/web/htmx/troublereports_dialog.go`
- **Comment:** `// TODO: Do somehtings like the get*FormData method in "tools.go"`
- **My take:** Adopting a consistent approach to form data parsing, like the `get*FormData` method used in `tools.go`, would improve the structure and readability of your code. It would centralize the form parsing logic and make it easier to maintain.

### 5. Add a `MetalSheetList` type with sorting

- **File:** `internal/database/models/metal-sheets.go`
- **Comment:** `// TODO: Add a MetalSheetList type with sorting functionality`
- **My take:** This is a great idea for improving the usability of your metal sheet data. A `MetalSheetList` type with sorting capabilities would allow you to easily sort metal sheets by different criteria, which would be very useful for displaying them in the UI.

### 6. Add tool cycle limits

- **File:** `internal/database/models/tools.go`
- **Comment:** `// TODO: Max cycles: 800.000 (Orange) -> 1.000.000 (Red)`
- **My take:** This seems like a business rule that needs to be implemented. Adding these cycle limits to your `Tool` model would allow you to track the tool's lifecycle and provide visual indicators (like colors) when it's approaching its end of life.

### 8. Add a voting system for deleting trouble reports

- **Files:** `internal/web/templates/components/troublereports/list.templ`
- **Comment:** `<!-- TODO: Benutzer ist kein Administrator, Abstimmungssystem zum Löschen von Problemberichten hinzufügen, noch nicht implementiert -->`
- **My take:** This is a feature that would allow non-admin users to vote on deleting trouble reports. It's a good way to involve more users in the decision-making process.

### 9. Display tool notes icons

- **File:** `internal/web/templates/components/tools/list.templ`
- **Comment:** `// TODO: Add icons previewing notes (warning, error, info or whatever)`
- **My take:** This is a UI improvement that would make it easier for users to quickly see the status of a tool's notes. Adding icons for different note levels would provide a quick visual cue about the importance of the notes associated with a tool.

### 10. Implement sorting for metal sheets table

- **File:** `internal/web/templates/pages/tool-page.templ`
- **Comment:** `// TODO: Need some sorting logic here`
- **My take:** This is another UI improvement that would enhance the user experience. Adding sorting functionality to the metal sheets table would allow users to easily organize the data based on their needs.

### 11. Complete the `tools-press-page.templ`

- **File:** `internal/web/templates/pages/tools-press-page.templ`
- **Comment:** `<!-- TODO: ... -->`
- **My take:** This template is a work in progress. The TODOs outline the remaining work, which includes displaying metal sheets, notes, and cycle data for tools on a specific press. Completing this would provide a more detailed view of the tools currently in use.

Overall, these `TODO` comments highlight some great opportunities to improve your project's functionality, structure, and user experience. If you'd like me to help with any of these, just let me know
