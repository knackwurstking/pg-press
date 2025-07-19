# Stucture

## Old Structure

```
routes/
├── routes.go          # Main routing setup
├── feed/              # Feed-related handlers
├── nav/               # Navigation components
├── profile/           # Profile management
├── troublereports/    # Trouble reports
├── shared/            # Shared constants
├── static/            # Static assets (CSS, JS, images)
├── templates/         # HTML templates
└── utils/             # Utility functions
```

## New Structure

```
routes/
├── router.go              # Main router setup (rename from routes.go)
│
├── handlers/              # HTTP handlers organized by domain
│   ├── auth/             # Authentication handlers
│   │   ├── auth.go       # Login/logout handlers
│   │   └── middleware.go # Auth middleware
│   │
│   ├── feed/             # Feed handlers
│   │   ├── feed.go       # Main feed handlers
│   │   └── data.go       # Feed data handlers
│   │
│   ├── profile/          # Profile handlers
│   │   ├── profile.go    # Profile page handlers
│   │   └── cookies.go    # Cookie management
│   │
│   ├── troublereports/   # Trouble report handlers
│   │   ├── reports.go    # Main reports handlers
│   │   ├── crud.go       # Create/update/delete operations
│   │   └── dialog.go     # Dialog/modal handlers
│   │
│   └── api/              # API endpoints (if any)
│       └── v1/           # Versioned API
│
├── middleware/           # HTTP middleware
│   ├── auth.go          # Authentication middleware
│   ├── logging.go       # Request logging
│   └── cors.go          # CORS handling
│
├── components/          # Reusable UI components/handlers
│   ├── nav/             # Navigation components
│   └── shared/          # Shared UI components
│
├── internal/            # Internal packages (not exported)
│   ├── constants/       # Application constants
│   ├── validation/      # Input validation
│   └── utils/           # Internal utilities
│
├── assets/              # Static assets (rename from static/)
│   ├── css/
│   ├── js/
│   ├── images/
│   └── icons/
│
└── templates/           # HTML templates (keep as is)
    ├── layouts/         # Layout templates
    ├── pages/           # Page templates
    └── components/      # Partial templates
```

## Key Improvements

### 1. Clearer Separation of Concerns

- handlers/: All HTTP handlers grouped by domain/feature
- middleware/: Dedicated middleware package
- internal/: Non-exported packages for better encapsulation

### 2. Better Authentication Organization

Move login/logout from main `routes.go` to dedicated auth handlers:

```go
// handlers/auth/auth.go
package auth

import (
    "github.com/labstack/echo/v4"
    "github.com/knackwurstking/pg-vis/pgvis"
)

type Handler struct {
    db               *pgvis.DB
    serverPathPrefix string
}

func NewHandler(db *pgvis.DB, serverPathPrefix string) *Handler {
    return &Handler{
        db:               db,
        serverPathPrefix: serverPathPrefix,
    }
}

func (h *Handler) RegisterRoutes(e *echo.Echo) {
    e.GET(h.serverPathPrefix+"/login", h.handleLogin)
    e.GET(h.serverPathPrefix+"/logout", h.handleLogout)
}
```

### 3. Improved Template Organization

```
templates/
├── layouts/           # Base layouts
│   └── main.html
├── pages/             # Full page templates
│   ├── home.html
│   ├── login.html
│   ├── feed.html
│   └── profile.html
└── components/        # Reusable partials
    ├── nav/
    │   ├── main.html
    │   └── feed-counter.html
    └── forms/
        └── login.html
```

### 4. Constants and Configuration

```go
// internal/constants/routes.go
package constants

const (
    // Cookie configuration
    CookieName               = "pgvis-api-key"
    CookieExpirationDuration = time.Hour * 24 * 31 * 6

    // Form fields
    APIKeyFormField   = "api-key"
    UserNameFormField = "user-name"

    // Validation
    UserNameMinLength = 1
    UserNameMaxLength = 100
)
```

## 5. Handler Interface Pattern

Implement a consistent handler interface:

```go
// handlers/handler.go
package handlers

import (
    "github.com/labstack/echo/v4"
    "github.com/knackwurstking/pg-vis/pgvis"
)

type Handler interface {
    RegisterRoutes(e *echo.Echo)
}

type BaseHandler struct {
    DB               *pgvis.DB
    ServerPathPrefix string
}
```

### Migration Benefits

1. **Scalability**: Easy to add new features without cluttering
2. **Maintainability**: Clear separation makes debugging easier
3. **Testing**: Each handler can be tested independently
4. **Team Development**: Multiple developers can work on different handlers
5. **Security**: Better organization of authentication logic
6. **Performance**: Cleaner middleware chain organization

### Implementation Strategy

I recommend implementing this restructure gradually:

1. Start with moving authentication logic to `handlers/auth/`
2. Create the `middleware/` package and move auth middleware
3. Reorganize templates into the new structure
4. Move existing handlers to the new `handlers/` structure
5. Create `internal/` packages for shared utilities
