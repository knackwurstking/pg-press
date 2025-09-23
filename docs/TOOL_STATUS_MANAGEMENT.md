# Tool Status Management

This document describes the tool status management feature that allows users to easily edit and update tool statuses directly from the interface.

## Overview

The tool status management system provides an intuitive way to change tool statuses between three states:

- **Available** - Tool is ready for use but not currently assigned to a press
- **Active** - Tool is currently installed and active on a specific press
- **Regenerating** - Tool is undergoing regeneration/maintenance and is unavailable

## Features

### Interactive Status Display

- Visual status badges with color coding:
    - **Available**: Blue info badge
    - **Active**: Green primary badge with press number
    - **Regenerating**: Orange warning badge
- Click-to-edit functionality with pencil icon

### Status Edit Form

- Dropdown selection for status change
- Dynamic press assignment (only shown when status is "Active")
- Save/Cancel buttons for confirmation
- Real-time UI updates without page refresh

## User Interface

### Tool List View

Each tool in the list displays:

- Current status with appropriate badge styling
- Edit button (pencil icon) to modify status
- Press information when tool is active

### Tool Detail Page

The tool detail page includes:

- Status section with current status display
- Edit functionality inline
- Link to press overview when tool is active on a press

### Status Edit Form

When editing status:

1. **Status Dropdown**: Select new status (Available/Active/Regenerating)
2. **Press Selection**: Only visible when "Active" is selected
    - Dropdown with valid press numbers (0, 2, 3, 4, 5)
    - Option for "Not assigned"
3. **Action Buttons**:
    - **Save**: Apply changes and update status
    - **Cancel**: Discard changes and return to display view

## Status Change Logic

### Changing to "Available"

- Sets `regenerating = false`
- Clears press assignment (`press = null`)
- Tool becomes available for assignment

### Changing to "Active"

- Sets `regenerating = false`
- Allows selection of press number (0, 2, 3, 4, 5)
- Tool is assigned to the selected press
- If no press selected, tool remains unassigned but not regenerating

### Changing to "Regenerating"

- Sets `regenerating = true`
- Automatically clears press assignment
- Tool becomes unavailable until regeneration is complete

## Technical Implementation

### Backend Endpoints

- `GET /htmx/tools/status-edit?id={tool_id}` - Get edit form
- `GET /htmx/tools/status-display?id={tool_id}` - Get display view
- `PUT /htmx/tools/status` - Update tool status

### Database Updates

- Updates `tools.regenerating` boolean field
- Updates `tools.press` integer field
- Creates audit trail entries via existing service methods
- Generates feed notifications for status changes

### HTMX Integration

- Uses HTMX for seamless form interactions
- In-place updates without page refresh
- Error handling with user feedback
- Progressive enhancement approach

## Usage Examples

### Making a Tool Available

1. Click the edit button (pencil icon) on any tool
2. Select "Available" from status dropdown
3. Click "Save"
4. Tool status updates to blue "Available" badge

### Assigning Tool to Press

1. Click edit button on an available tool
2. Select "Active" from status dropdown
3. Choose press number from press dropdown
4. Click "Save"
5. Tool shows green "Active" badge with press number

### Setting Tool for Regeneration

1. Click edit button on any tool
2. Select "Regenerating" from status dropdown
3. Click "Save"
4. Tool shows orange "Regenerating" badge
5. Press assignment is automatically cleared

## Security & Permissions

- All status changes require authenticated user
- Changes are logged with user attribution
- Admin permissions may be required for certain operations
- Audit trail maintained for all modifications

## Integration with Existing Features

### Press Management

- Active tools appear in press utilization reports
- Press assignment updates affect press-specific views
- Tool changes trigger press-related notifications

### Feed System

- Status changes generate feed notifications
- User attribution included in feed messages
- Real-time updates to connected clients

### Cycle Management

- Status changes may affect cycle tracking
- Active tools participate in cycle calculations
- Regenerating tools are excluded from active cycles

## Future Enhancements

Potential improvements to consider:

- Batch status updates for multiple tools
- Status change scheduling/automation
- Integration with maintenance schedules
- Enhanced status history tracking
- Status-based filtering and search
