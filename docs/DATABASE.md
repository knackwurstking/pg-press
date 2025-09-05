# Database

The database is a SQLite database, and the schema is defined in the `internal/database` directory.

### Tables

#### `attachments`

Stores file attachments.

| Column    | Type    | Description                                      |
| --------- | ------- | ------------------------------------------------ |
| id        | INTEGER | The attachment ID (Primary Key, Auto-increment). |
| mime_type | TEXT    | The MIME type of the attachment (Not Null).      |
| data      | BLOB    | The attachment data (Not Null).                  |

#### `cookies`

Stores user session cookies.

| Column     | Type    | Description                                        |
| ---------- | ------- | -------------------------------------------------- |
| user_agent | TEXT    | The user agent of the client (Not Null).           |
| value      | TEXT    | The cookie value (Primary Key, Not Null).          |
| api_key    | TEXT    | The API key associated with the cookie (Not Null). |
| last_login | INTEGER | The timestamp of the last login (Not Null).        |

#### `feeds`

Stores activity feed entries.

| Column    | Type    | Description                                      |
| --------- | ------- | ------------------------------------------------ |
| id        | INTEGER | The feed entry ID (Primary Key, Auto-increment). |
| time      | INTEGER | The timestamp of the event (Not Null).           |
| data_type | TEXT    | The type of the event (Not Null).                |
| data      | BLOB    | The event data (Not Null).                       |

#### `metal_sheets`

Stores information about metal sheets.

| Column       | Type     | Description                                                                                  |
| ------------ | -------- | -------------------------------------------------------------------------------------------- |
| id           | INTEGER  | The metal sheet ID (Primary Key, Auto-increment).                                            |
| tile_height  | REAL     | The tile height (Not Null).                                                                  |
| value        | REAL     | The value (Not Null).                                                                        |
| marke_height | INTEGER  | The marke height (Not Null).                                                                 |
| stf          | REAL     | The STF value (Not Null).                                                                    |
| stf_max      | REAL     | The maximum STF value (Not Null).                                                            |
| tool_id      | INTEGER  | The ID of the tool the sheet is assigned to (Foreign Key to `tools.id`, On Delete Set Null). |
| notes        | BLOB     | Linked notes (Not Null).                                                                     |
| mods         | BLOB     | Modification history (Not Null).                                                             |
| created_at   | DATETIME | The timestamp of creation (Default: CURRENT_TIMESTAMP).                                      |
| updated_at   | DATETIME | The timestamp of the last update (Default: CURRENT_TIMESTAMP).                               |

#### `notes`

Stores notes that can be linked to other items.

| Column     | Type     | Description                                                |
| ---------- | -------- | ---------------------------------------------------------- |
| id         | INTEGER  | The note ID (Primary Key, Auto-increment).                 |
| level      | INTEGER  | The note level (e.g., INFO, ATTENTION, BROKEN) (Not Null). |
| content    | TEXT     | The note content (Not Null).                               |
| created_at | DATETIME | The timestamp of creation (Default: CURRENT_TIMESTAMP).    |

#### `press_cycles`

Stores press cycle information for tools.

| Column       | Type     | Description                                                                                            |
| ------------ | -------- | ------------------------------------------------------------------------------------------------------ |
| id           | INTEGER  | The press cycle ID (Primary Key, Auto-increment).                                                      |
| press_number | INTEGER  | The press number (Not Null, Check: 0-5).                                                               |
| tool_id      | INTEGER  | The tool ID (Not Null, Foreign Key to `tools.id`).                                                     |
| date         | DATETIME | The date of the cycle (Not Null).                                                                      |
| total_cycles | INTEGER  | The total number of cycles (Not Null, Default: 0).                                                     |
| performed_by | INTEGER  | The ID of the user who performed the action (Not Null, Foreign Key to `users.id`, On Delete Set Null). |

#### `tool_regenerations`

Stores tool regeneration history.

| Column       | Type    | Description                                                                                               |
| ------------ | ------- | --------------------------------------------------------------------------------------------------------- |
| id           | INTEGER | The regeneration ID (Primary Key, Auto-increment).                                                        |
| tool_id      | INTEGER | The tool ID (Not Null, Foreign Key to `tools.id`, On Delete Cascade).                                     |
| cycle_id     | INTEGER | The cycle ID at the time of regeneration (Not Null, Foreign Key to `press_cycles.id`, On Delete Cascade). |
| reason       | TEXT    | The reason for regeneration.                                                                              |
| performed_by | INTEGER | The ID of the user who performed the action (Foreign Key to `users.id`, On Delete Set Null).              |

#### `tools`

Stores tool information.

| Column       | Type    | Description                                              |
| ------------ | ------- | -------------------------------------------------------- |
| id           | INTEGER | The tool ID (Primary Key, Auto-increment).               |
| position     | TEXT    | The tool position (e.g., 'top', 'bottom') (Not Null).    |
| format       | BLOB    | The tool format (Not Null).                              |
| type         | TEXT    | The tool type (Not Null).                                |
| code         | TEXT    | The tool code (Not Null).                                |
| regenerating | BOOLEAN | Whether the tool is regenerating (Not Null, Default: 0). |
| press        | INTEGER | The press number the tool is on.                         |
| notes        | BLOB    | Linked notes (Not Null).                                 |
| mods         | BLOB    | Modification history (Not Null).                         |

#### `trouble_reports`

Stores trouble reports.

| Column             | Type    | Description                                          |
| ------------------ | ------- | ---------------------------------------------------- |
| id                 | INTEGER | The trouble report ID (Primary Key, Auto-increment). |
| title              | TEXT    | The title of the report (Not Null).                  |
| content            | TEXT    | The content of the report (Not Null).                |
| linked_attachments | TEXT    | Linked attachments (Not Null).                       |
| mods               | BLOB    | Modification history (Not Null).                     |

#### `users`

Stores user information.

| Column      | Type    | Description                                                 |
| ----------- | ------- | ----------------------------------------------------------- |
| telegram_id | INTEGER | The user's Telegram ID (Primary Key, Not Null).             |
| user_name   | TEXT    | The user's name (Not Null).                                 |
| api_key     | TEXT    | The user's API key (Not Null, Unique).                      |
| last_feed   | TEXT    | The ID of the last feed entry the user has seen (Not Null). |
