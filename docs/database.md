# Database Notes

> I really need to simplify all this database stuff...

## Another idea would be to create multiple sql databases

1. Users and Cookies
2. Tools, cycles and regenerations for press and tools
3. Trouble Reports

But first i need to restructure the services package

```text
services/
├── shared/
│   └── interfaces.go
├── user/
│   ├── user_service.go
│   └── auth_service.go
├── press/
│   ├── press_service.go
│   ├── regeneration_service.go
│   └── cycle_service.go
├── tool/
│   ├── tool_service.go
│   └── regeneration_service.go
├── troublereport/
│   ├── modification_service.go
│   └── trouble_report_service.go
└── common/
    └── db.go
```

In memory database wich saves changes to the database in the background using
some kind of queue to prevent concurrency issues

Design it in a way so i could swap out the sqlite database with something else later on


### Generic Repository Pattern

Use Go generics to create reusable database operations:

```go
type Repository[T any] struct {
    db *sql.DB
    tableName string
}
func (r *Repository[T]) Get(id int64) (*T, error) {
    // Generic implementation
}
func (r *Repository[T]) Save(item *T) error {
    // Generic implementation
}
```

Struct-based Parameters

Group function parameters instead of long argument lists:

```go
type CycleCreateParams struct {
    PressNumber   models.PressNumber
    ToolID        models.ToolID
    ToolPosition  models.Position
    TotalCycles   int64
    PerformedBy   models.TelegramID
}
```

### Random Stuff

> Models should be moved to their respective packages

I still need to find an answer to the question when to save on-memory changes
to the database. Maybe a background goroutine that flushes changes every time 
something changes? And loads them on periodically and on startup?

> Do not forget to add json tags

- common/
    - db.go
        - Contains the main database object
        - Pass the SQL database here (optional), runs on-memory otherwise

- press/
    - cycles.go
    - press.go
    - regeneration.go

- shared/
    - interfaces.go
        - `type Repository[T any] struct` 

- tool/
    - tools.go
    - regeneration.go

- troublereport/
    - reports.go
    - modification.go
        - Modifications are linked to Reports via the Report `ID`

- user/
    - users.go
    - cookie.go
        - Cookies are linked to Users via the User `ID` 
    - session.go
        - This will be on-memory only
        - Shows current active user sessions
        - Handles all websocket connections for user updates (feed counter)
