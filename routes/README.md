# Routes Package

This package provides HTTP route handlers and web interface for the pgvis application. The structure has been redesigned to improve maintainability, scalability, and developer experience.

## Structure Overview

```
routes/
├── router.go           # Main router setup and configuration
├── handlers/           # HTTP handlers organized by domain
│   ├── auth/           # Authentication (login/logout)
│   ├── home/           # Home page
│   ├── nav/            # Navigation components
│   ├── feed/           # Feed management
│   ├── profile/        # User profile management
│   └── troublereports/ # Trouble report management
├── constants/          # Public constants package
├── internal/           # Internal utilities (routes package only)
├── assets/             # Static assets (CSS, JS, images)
└── templates/          # HTML templates
```

## Key Principles

### 1. Domain-Based Organization

Handlers are grouped by feature/domain rather than technical function, making it easier to:

- Locate related functionality
- Maintain and modify features
- Add new capabilities
- Work in teams

### 2. Consistent Handler Pattern

All handlers follow a standard pattern:

```go
type Handler struct {
    db               *pgvis.DB
    serverPathPrefix string
    templates        embed.FS
}

func NewHandler(db *pgvis.DB, serverPathPrefix string, templates embed.FS) *Handler {
    return &Handler{...}
}

func (h *Handler) RegisterRoutes(e *echo.Echo) {
    // Register routes here
}
```

### 3. Clean Separation of Concerns

- **Handlers**: HTTP request/response logic
- **Constants**: Shared configuration and constants
- **Internal**: Package-private utilities
- **Assets**: Static files (CSS, JS, images)
- **Templates**: HTML templates

## Usage

### Adding a New Handler

1. Create handler directory under `handlers/`:

```bash
mkdir handlers/newfeature
```

2. Implement the handler:

```go
// handlers/newfeature/newfeature.go
package newfeature

type Handler struct {
    db               *pgvis.DB
    serverPathPrefix string
    templates        embed.FS
}

func NewHandler(db *pgvis.DB, serverPathPrefix string, templates embed.FS) *Handler {
    return &Handler{
        db:               db,
        serverPathPrefix: serverPathPrefix,
        templates:        templates,
    }
}

func (h *Handler) RegisterRoutes(e *echo.Echo) {
    e.GET(h.serverPathPrefix+"/newfeature", h.handleNewFeature)
}

func (h *Handler) handleNewFeature(c echo.Context) error {
    // Implementation here
}
```

3. Register in main router:

```go
// router.go
newFeatureHandler := newfeature.NewHandler(o.DB, o.ServerPathPrefix, templates)
newFeatureHandler.RegisterRoutes(e)
```

### Using Constants

```go
import "github.com/knackwurstking/pg-vis/routes/constants"

// Use predefined constants
formValue := c.FormValue(constants.APIKeyFormField)
cookie := &http.Cookie{
    Name:    constants.CookieName,
    Expires: time.Now().Add(constants.CookieExpirationDuration),
}
```

### Using Internal Utilities

```go
import "github.com/knackwurstking/pg-vis/routes/internal/utils"

// Validate user from context
user, err := utils.GetUserFromContext(c)
if err != nil {
    return err
}

// Handle templates
return utils.HandleTemplate(c, data, templates, templatePaths)
```

## Migration Notes

This structure replaces the previous flat organization. Key changes:

- `routes.go` → `router.go` (authentication logic extracted)
- `shared/` → `constants/` (public constants)
- `utils/` → `internal/utils/` (package-private utilities)
- `static/` → `assets/` (better semantic naming)
- Domain handlers moved to `handlers/` subdirectories

All existing functionality is preserved with improved organization.

## Benefits

- **Maintainability**: Clear structure makes code easy to navigate
- **Scalability**: Simple to add new features without cluttering
- **Testing**: Handlers can be tested independently
- **Team Development**: Multiple developers can work on different domains
- **Security**: Proper encapsulation with internal packages
- **Consistency**: Standard patterns across all handlers

## Dependencies

- [Echo](https://echo.labstack.com/) - Web framework
- [pgvis](../pgvis/) - Core business logic and database operations

## Templates

HTML templates are organized by feature and include:

- Page templates for full pages
- Component templates for reusable parts
- Layout templates for common structure

Template paths are managed through the `constants` package for consistency.
