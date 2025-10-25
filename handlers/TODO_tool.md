## Assistant

Based on my analysis of the `tool.go` file, here are my recommendations for cleaning up this code:

## Main Issues Identified

1. **Single Responsibility Principle Violation**: The `Tool` struct is handling too many different concerns
2. **File Size**: Over 1000 lines in a single file makes it hard to maintain
3. **Repetitive Code Patterns**: Many similar HTTP handler patterns
4. **Mixed Concerns**: HTTP handling, business logic, and data formatting are mixed together

## Cleanup Recommendations

### 1. **Split by Domain/Feature**

Break the file into multiple smaller, focused files:

```pg-press/handlers/tool/tool.go#L1-20
package tool

type Handler struct {
	*handlers.Base
}

func NewHandler(db *services.Registry) *Handler {
	return &Handler{
		Base: handlers.NewBase(db, logger.NewComponentLogger("Tool")),
	}
}

func (h *Handler) RegisterRoutes(e *echo.Echo) {
	// Main routes registration
}
```

```pg-press/handlers/tool/cycles.go#L1-15
package tool

// All cycle-related handlers
func (h *Handler) HTMXGetCycles(c echo.Context) error { ... }
func (h *Handler) HTMXGetToolCycleEditDialog(c echo.Context) error { ... }
func (h *Handler) HTMXPostToolCycleEditDialog(c echo.Context) error { ... }
func (h *Handler) HTMXPutToolCycleEditDialog(c echo.Context) error { ... }
func (h *Handler) HTMXDeleteToolCycle(c echo.Context) error { ... }
```

```pg-press/handlers/tool/binding.go#L1-10
package tool

// All binding-related handlers
func (h *Handler) HTMXPatchToolBinding(c echo.Context) error { ... }
func (h *Handler) HTMXPatchToolUnBinding(c echo.Context) error { ... }
```

```pg-press/handlers/tool/regeneration.go#L1-10
package tool

// All regeneration-related handlers
func (h *Handler) HTMXGetEditRegeneration(c echo.Context) error { ... }
func (h *Handler) HTMXPutEditRegeneration(c echo.Context) error { ... }
func (h *Handler) HTMXDeleteRegeneration(c echo.Context) error { ... }
```

### 2. **Extract Service Layer**

Create dedicated services to handle business logic:

```pg-press/services/tool_cycles.go#L1-25
package services

type ToolCycleService struct {
	registry *Registry
	logger   *logger.ComponentLogger
}

func NewToolCycleService(registry *Registry) *ToolCycleService {
	return &ToolCycleService{
		registry: registry,
		logger:   logger.NewComponentLogger("ToolCycleService"),
	}
}

func (s *ToolCycleService) GetCyclesForTool(toolID int64) ([]*models.Cycle, error) {
	// Business logic for getting and filtering cycles
}

func (s *ToolCycleService) CreateCycle(form *ToolCycleEditDialogFormData, userID int64) error {
	// Business logic for creating cycles
}

func (s *ToolCycleService) UpdateCycle(cycleID int64, form *ToolCycleEditDialogFormData) error {
	// Business logic for updating cycles
}
```

### 3. **Create Form Validators**

Extract form validation logic:

```pg-press/handlers/tool/validators.go#L1-20
package tool

type CycleFormValidator struct{}

func (v *CycleFormValidator) ValidateAndParse(c echo.Context) (*ToolCycleEditDialogFormData, error) {
	form := &ToolCycleEditDialogFormData{}

	// Validation logic extracted from getCycleFormData
	if err := v.validatePressNumber(c, form); err != nil {
		return nil, err
	}

	if err := v.validateDate(c, form); err != nil {
		return nil, err
	}

	// etc...

	return form, nil
}
```

### 4. **Create Response Builders**

Extract template rendering logic:

```pg-press/handlers/tool/responses.go#L1-15
package tool

type ResponseBuilder struct {
	registry *services.Registry
}

func (r *ResponseBuilder) BuildCyclesResponse(user *models.User, tool *models.Tool, cycles []*models.Cycle) templ.Component {
	// Extract the component building logic
	return components.PageTool_Cycles(components.PageTool_CyclesProps{
		// props setup
	})
}
```

### 5. **Standardize Error Handling**

Create consistent error handling patterns:

```pg-press/handlers/tool/errors.go#L1-15
package tool

type ErrorHandler struct {
	logger *logger.ComponentLogger
}

func (e *ErrorHandler) HandleToolNotFound(toolID int64) error {
	return handlers.HandleBadRequest(
		fmt.Errorf("tool %d not found", toolID),
		"tool not found",
	)
}
```

### 6. **Create Constants File**

Extract route patterns and form field names:

```pg-press/handlers/tool/constants.go#L1-20
package tool

const (
	// Routes
	RouteToolPage         = "/tools/tool/:id"
	RouteHTMXGetCycles    = "/htmx/tools/cycles"
	RouteHTMXEditCycle    = "/htmx/tools/cycle/edit"

	// Form Fields
	FormFieldTotalCycles  = "total_cycles"
	FormFieldPressNumber  = "press_number"
	FormFieldOriginalDate = "original_date"
	FormFieldRegenerating = "regenerating"
	FormFieldToolID       = "tool_id"
)
```

### 7. **Improved Project Structure**

The final structure would look like:

```
handlers/
├── tool/
│   ├── handler.go       // Main handler and routes
│   ├── cycles.go        // Cycle-related handlers
│   ├── binding.go       // Tool binding handlers
│   ├── regeneration.go  // Regeneration handlers
│   ├── status.go        // Status-related handlers
│   ├── validators.go    // Form validation
│   ├── responses.go     // Response builders
│   ├── constants.go     // Constants and route definitions
│   └── errors.go        // Error handling utilities
└── tool.go             // Keep for backward compatibility, delegate to tool/handler.go
```

This refactoring would:

- **Improve maintainability** by separating concerns
- **Reduce cognitive load** with smaller, focused files
- **Enhance testability** by isolating business logic
- **Follow Go conventions** with package-level organization
- **Maintain backward compatibility** through delegation

Would you like me to help implement any of these refactoring steps?
