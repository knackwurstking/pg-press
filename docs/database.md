# Database Notes

> I really need to simplify all this database stuff...

## Another idea would be to create multiple sql databases

1. Users and Cookies
2. Tools, cycles and regenerations for press and tools
3. Trouble Reports

But first i need to restructure the services package, Also remove old the models 
package, include stuff to the service/shared.

```text
services/
├── shared/
│   ├── base-service-config.go
│   ├── base-service.go
│   ├── model-...
│   ├── types.go
│   └── interfaces.go
├── user/
│   ├── cookie.go
│   ├── session.go
│   └── user.go
├── press/
│   ├── cycles.go
│   ├── press.go
│   └── regeneration.go
└── common/
    └── db.go
```

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

- common/
    - db.go
        - Main entry point for database services

- shared/
    - interfaces.go
        - `Service`
        - `Entity`
    - types.go
        - Common and simple type aliases used across services
    - model-...
        - Common models used across services
    - base-service.go
        - Base service struct implementing common functionality
    - base-service-config.go
        - Configuration struct for base service configuration

- user/
    - cookie.go
        - Index: `cookies(user_id)`
        - Index: `cookies(value)`
    - session.go
        - In-Memory Service, does not store stuf in the database
    - user.go
        - Index: `users(api_key)`
        - Contains api keys for users
        - Uses telegram ID as primary identifier

- press/
    - cycles.go
    - press.go
    - regeneration.go

