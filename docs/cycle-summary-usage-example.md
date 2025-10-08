# Cycle Summary Service Usage Examples

This document provides examples of how to use the cycle summary methods added to the `PressCycles` service.

## Available Methods

### 1. GetCycleSummaryData

Retrieves complete cycle summary data for a press including cycles, tools map, and users map.

### 2. GetCycleSummaryStats

Calculates statistics from cycles data (total cycles, partial cycles, active tools count, entries count).

### 3. GetToolSummaries

Creates consolidated tool summaries with start/end dates and proper consolidation logic.

## Usage Examples

### Basic Cycle Summary Data Retrieval

```go
package main

import (
    "fmt"
    "log"

    "github.com/knackwurstking/pgpress/internal/database"
    "github.com/knackwurstking/pgpress/pkg/models"
)

func getCycleSummaryExample(db *database.DB, press models.PressNumber) error {
    // Get complete cycle summary data
    cycles, toolsMap, usersMap, err := db.PressCycles.GetCycleSummaryData(
        press,
        db.Tools,
        db.Users,
    )
    if err != nil {
        return fmt.Errorf("failed to get cycle summary data: %v", err)
    }

    fmt.Printf("Retrieved %d cycles for press %d\n", len(cycles), press)
    fmt.Printf("Tools map contains %d tools\n", len(toolsMap))
    fmt.Printf("Users map contains %d users\n", len(usersMap))

    return nil
}
```

### Calculate Statistics

```go
func calculateStatsExample(db *database.DB, press models.PressNumber) error {
    // First get the cycles data
    cycles, _, _, err := db.PressCycles.GetCycleSummaryData(
        press,
        db.Tools,
        db.Users,
    )
    if err != nil {
        return err
    }

    // Calculate statistics
    totalCycles, totalPartial, activeTools, entries := db.PressCycles.GetCycleSummaryStats(cycles)

    fmt.Printf("Press %d Statistics:\n", press)
    fmt.Printf("  Total Cycles: %d\n", totalCycles)
    fmt.Printf("  Total Partial Cycles: %d\n", totalPartial)
    fmt.Printf("  Active Tools: %d\n", activeTools)
    fmt.Printf("  Total Entries: %d\n", entries)

    return nil
}
```

### Generate Tool Summaries

```go
func generateToolSummariesExample(db *database.DB, press models.PressNumber) error {
    // Get cycles and tools data
    cycles, toolsMap, _, err := db.PressCycles.GetCycleSummaryData(
        press,
        db.Tools,
        db.Users,
    )
    if err != nil {
        return err
    }

    // Generate tool summaries
    summaries, err := db.PressCycles.GetToolSummaries(cycles, toolsMap)
    if err != nil {
        return fmt.Errorf("failed to generate tool summaries: %v", err)
    }

    fmt.Printf("Tool Summaries for Press %d:\n", press)
    for _, summary := range summaries {
        fmt.Printf("  Tool: %s, Position: %s\n",
            summary.ToolCode,
            summary.Position.GermanString())
        fmt.Printf("    Max Cycles: %d, Partial Cycles: %d\n",
            summary.MaxCycles,
            summary.TotalPartial)

        if summary.IsFirstAppearance {
            fmt.Printf("    Period: Unknown start - %s\n",
                summary.EndDate.Format("02.01.2006"))
        } else {
            fmt.Printf("    Period: %s - %s\n",
                summary.StartDate.Format("02.01.2006"),
                summary.EndDate.Format("02.01.2006"))
        }
        fmt.Println()
    }

    return nil
}
```

### Complete Example: Generate Custom Report

```go
func generateCustomReportExample(db *database.DB, press models.PressNumber) error {
    // Get all cycle summary data
    cycles, toolsMap, usersMap, err := db.PressCycles.GetCycleSummaryData(
        press,
        db.Tools,
        db.Users,
    )
    if err != nil {
        return err
    }

    // Calculate overall statistics
    totalCycles, totalPartial, activeTools, entries := db.PressCycles.GetCycleSummaryStats(cycles)

    // Generate detailed tool summaries
    summaries, err := db.PressCycles.GetToolSummaries(cycles, toolsMap)
    if err != nil {
        return err
    }

    // Create custom report
    fmt.Printf("=== PRESS %d CYCLE SUMMARY REPORT ===\n\n", press)

    // Overview section
    fmt.Println("OVERVIEW:")
    fmt.Printf("  Total Cycles: %d\n", totalCycles)
    fmt.Printf("  Active Tools: %d\n", activeTools)
    fmt.Printf("  Total Entries: %d\n", entries)
    fmt.Printf("  Partial Cycles Sum: %d\n", totalPartial)
    fmt.Println()

    // Tool details section
    fmt.Println("TOOL DETAILS:")
    for _, summary := range summaries {
        fmt.Printf("  %s (%s)\n", summary.ToolCode, summary.Position.GermanString())
        fmt.Printf("    Cycles: %d (Partial: %d)\n",
            summary.MaxCycles, summary.TotalPartial)

        if summary.IsFirstAppearance {
            fmt.Printf("    Active since: Unknown - %s\n",
                summary.EndDate.Format("02.01.2006"))
        } else {
            fmt.Printf("    Active period: %s - %s\n",
                summary.StartDate.Format("02.01.2006"),
                summary.EndDate.Format("02.01.2006"))
        }
        fmt.Println()
    }

    // Recent activity section
    fmt.Println("RECENT ACTIVITY:")
    recentLimit := 5
    if len(cycles) < recentLimit {
        recentLimit = len(cycles)
    }

    for i := 0; i < recentLimit; i++ {
        cycle := cycles[i]

        // Get tool info
        toolInfo := "Unknown Tool"
        if tool, exists := toolsMap[cycle.ToolID]; exists {
            toolInfo = fmt.Sprintf("%s %s", tool.Format.String(), tool.Code)
        }

        // Get user info
        userInfo := "Unknown User"
        if user, exists := usersMap[cycle.PerformedBy]; exists {
            userInfo = user.Name
        }

        fmt.Printf("  %s - %s (%s): %d cycles (Partial: %d) by %s\n",
            cycle.Date.Format("02.01.2006 15:04"),
            toolInfo,
            cycle.ToolPosition.GermanString(),
            cycle.TotalCycles,
            cycle.PartialCycles,
            userInfo)
    }

    return nil
}
```

### Usage in HTTP Handler

```go
func (h *Handler) HandleCycleSummaryAPI(c echo.Context) error {
    // Parse press number from request
    press, err := h.getPressNumberFromParam(c)
    if err != nil {
        return err
    }

    // Get cycle summary data using service methods
    cycles, toolsMap, usersMap, err := h.DB.PressCycles.GetCycleSummaryData(
        press,
        h.DB.Tools,
        h.DB.Users,
    )
    if err != nil {
        return h.HandleError(c, err, "failed to get cycle summary data")
    }

    // Calculate stats
    totalCycles, totalPartial, activeTools, entries := h.DB.PressCycles.GetCycleSummaryStats(cycles)

    // Generate tool summaries
    summaries, err := h.DB.PressCycles.GetToolSummaries(cycles, toolsMap)
    if err != nil {
        return h.HandleError(c, err, "failed to generate tool summaries")
    }

    // Return JSON response
    response := map[string]interface{}{
        "press": press,
        "stats": map[string]interface{}{
            "total_cycles":   totalCycles,
            "total_partial":  totalPartial,
            "active_tools":   activeTools,
            "entries_count":  entries,
        },
        "tool_summaries": summaries,
        "raw_cycles":     cycles,
    }

    return c.JSON(http.StatusOK, response)
}
```

## Benefits of Using These Service Methods

1. **Reusability**: These methods can be used across different parts of the application
2. **Consistency**: Ensures the same logic is applied everywhere for cycle summaries
3. **Maintainability**: Changes to cycle summary logic only need to be made in one place
4. **Testing**: Each method can be unit tested independently
5. **Performance**: Optimized database queries and data processing

## Integration Points

These methods are already integrated with:

- PDF generation (`internal/pdf/cycle-summary.go`)
- Press handlers (`internal/web/features/press/handlers.go`)

They can also be used for:

- API endpoints
- Report generation
- Dashboard statistics
- Data export functionality
- Custom analytics
