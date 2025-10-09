# Admin Tools Section

## Overview

The Admin Tools Section is a new administrative feature in the PG Press tools management system that provides system-wide analysis capabilities. It is visible to all users and appears at the top of the main tools page (`/tools`).

## Features

### Overlapping Tools Detection

The primary feature of the admin section is the detection of tool usage overlaps across all presses. The system analyzes tool usage history across all presses (0, 2, 3, 4, 5) to identify when the same tool ID appears on multiple presses simultaneously. This is physically impossible and indicates critical data entry errors or system issues.

#### How It Works

1. **Data Collection**: The system retrieves cycle summary data for all active presses
2. **Tool Period Analysis**: For each press, it creates consolidated tool usage periods with start and end dates
3. **Overlap Detection**: It compares tool usage periods across different presses to identify overlaps
4. **Report Generation**: Found overlaps are presented in a user-friendly format

#### What Gets Detected

- **Same Tool ID on Multiple Presses**: When a tool appears to be active on two or more presses during overlapping time periods
- **Time Period Overlaps**: Any overlap in usage periods, even partial overlaps
- **Cross-Press Conflicts**: Tools that should physically only exist on one press at a time

## User Interface

### Access

The admin section appears as a collapsible details section at the top of the main tools page with the title "Admin: Werkzeug-Überschneidungen" (Admin: Tool Overlaps).

### Status Indicators

#### No Overlaps Found

- **Green Success Card**: Shows a check mark icon
- **Message**: "Keine Überschneidungen gefunden" (No overlaps found)
- **Subtitle**: Confirms all tools are correctly restricted to individual presses

#### Overlaps Detected

- **Warning Card**: Yellow warning with triangle icon
- **Message**: Shows the number of overlapping tools detected
- **Detailed List**: Each overlapping tool is shown in its own error card

### Overlap Details

For each overlapping tool, the system displays:

#### Tool Information

- **Tool Code**: Format and identification code
- **Tool ID**: Database identifier
- **Overall Time Range**: Complete period during which overlaps occurred

#### Press Instance Details

- **Press Number**: Which press the tool appears on
- **Position**: Tool position (Top, Top Cassette, Bottom)
- **Start/End Dates**: Exact time period for that press
- **Duration**: Calculated duration of usage

## Technical Implementation

### Backend Components

#### Services

- `PressCycles.GetOverlappingTools()`: Main detection function
- `PressCycles.GetCycleSummaryData()`: Retrieves cycles, tools, and users data
- `GetToolSummaries()`: Creates consolidated tool usage periods
- `GetOverlappingTools()`: Detects and reports overlapping periods

#### Data Structures

```go
type OverlappingTool struct {
    ToolID    int64
    ToolCode  string
    Overlaps  []*OverlappingToolInstance
    StartDate time.Time
    EndDate   time.Time
}

type OverlappingToolInstance struct {
    PressNumber models.PressNumber
    Position    models.Position
    StartDate   time.Time
    EndDate     time.Time
}
```

#### Handlers

- `HTMXGetAdminOverlappingTools`: HTMX endpoint for loading the admin section
- Route: `GET /htmx/tools/admin/overlapping-tools`

### Frontend Components

#### Template

- `AdminOverlappingTools()`: Main template for rendering overlap results
- Located in: `internal/web/features/tools/templates/sections.templ`

#### Integration

- Added to main tools page as collapsible section
- Uses HTMX for lazy loading when expanded
- Automatic error handling and user feedback

## Usage Scenarios

### Regular Monitoring

- Operations staff can quickly check for data integrity issues
- Part of routine system health checks
- Proactive identification of potential problems

### Troubleshooting

- When production issues occur, check for tool conflicts
- Verify data consistency after system maintenance
- Investigate unusual cycle data

### Data Quality Assurance

- Regular auditing of tool assignment data
- Verification after bulk data imports
- Quality control for press scheduling

## Performance Considerations

- **Lazy Loading**: Admin section only loads when explicitly opened
- **Efficient Queries**: Uses existing cycle summary infrastructure
- **Caching**: Leverages existing database optimization
- **Error Handling**: Continues analysis even if individual presses fail

## Future Enhancements

Potential future additions to the admin section could include:

- **Historical Overlap Reports**: Track overlaps over time
- **Automatic Notifications**: Alert when new overlaps are detected
- **Resolution Suggestions**: Proposed fixes for detected issues
- **Export Capabilities**: Download overlap reports for analysis
- **Integration Warnings**: Alerts during tool assignment operations

## Troubleshooting

### Common Issues

**Section Won't Load**

- Check server logs for database connection issues
- Verify press cycles data exists
- Ensure proper user authentication

**No Results When Expected**

- Verify cycle data exists for all presses
- Check date ranges of tool usage
- Confirm tool IDs are properly recorded

**Performance Issues**

- Large datasets may cause slower loading
- Consider pagination for extensive historical data
- Monitor database query performance
