# Attachment Clearing on Validation Errors

## Overview

This document describes the changes made to properly handle attachment clearing when form validation fails in the trouble report dialog edit component. When users submit a form with validation errors (invalid title or content), the attachments are now properly cleared to prevent confusion and potential data issues.

## Problem Statement

Previously, when users uploaded attachments and then submitted a form with validation errors (e.g., empty title or content), the attachment data would persist in the form state. This could lead to:

- Confusion about whether attachments were actually saved
- Potential data inconsistencies
- Poor user experience when form validation fails
- Uploaded files appearing to be "stuck" in the form

## Solution

The solution involves clearing attachment data when validation fails and providing clear visual feedback to users about what happened.

## Changes Made

### 1. Handler Logic Updates (`dialog-edit.go`)

#### POST Handler Changes

```go
// Before: Always set attachments regardless of validation
dialogEditData.LinkedAttachments = attachments

// After: Only set attachments when validation passes
if !dialogEditData.InvalidTitle && !dialogEditData.InvalidContent {
    dialogEditData.LinkedAttachments = attachments
    // ... proceed with saving
} else {
    dialogEditData.Submitted = false
    // LinkedAttachments remains nil when validation fails
}
```

#### PUT Handler Changes

```go
// Restructured to only set attachments when validation succeeds
dialogEditData := &DialogEditTemplateData{
    Submitted:      true,
    ID:             int(id),
    Title:          title,
    Content:        content,
    InvalidTitle:   title == "",
    InvalidContent: content == "",
    // LinkedAttachments not set here
}

if !dialogEditData.InvalidTitle && !dialogEditData.InvalidContent {
    dialogEditData.LinkedAttachments = attachments
    // ... proceed with update
} else {
    dialogEditData.Submitted = false
    // LinkedAttachments remains nil
}
```

### 2. Template Updates (`dialog-edit.html`)

#### Enhanced Attachment Section Condition

```html
<!-- Before: Simple check for existence -->
{{if .LinkedAttachments}}

<!-- After: Check for both existence and non-empty -->
{{if and .LinkedAttachments (gt (len .LinkedAttachments) 0)}}
```

#### Added Validation Error Message

```html
{{if or .InvalidTitle .InvalidContent}}
<div class="attachment-error">
    <i class="bi bi-exclamation-triangle"></i>
    Anhänge wurden aufgrund von Validierungsfehlern entfernt. Bitte korrigieren
    Sie die Fehler und laden Sie die Dateien erneut hoch.
</div>
{{end}}
```

### 3. JavaScript Updates (`dialog-edit.js`)

#### Added Validation Error Detection

```javascript
// Check if there are validation errors in the form
function hasValidationErrors() {
    const titleInput = document.getElementById("title");
    const contentInput = document.getElementById("content");

    return (
        (titleInput && titleInput.getAttribute("aria-invalid") === "true") ||
        (contentInput && contentInput.getAttribute("aria-invalid") === "true")
    );
}
```

#### Enhanced File State Reset

```javascript
// Reset file state on load, especially when validation errors occur
if (hasValidationErrors()) {
    resetFileState();
} else {
    resetFileState();
}
initializeAttachmentOrder();
```

## User Experience Flow

### Successful Submission

1. User fills form with valid title and content
2. User uploads attachments
3. User submits form
4. Form validates successfully
5. Attachments are preserved and saved
6. Dialog closes and data refreshes

### Failed Validation

1. User fills form with invalid data (empty title/content)
2. User uploads attachments
3. User submits form
4. Validation fails on server
5. **NEW**: Attachments are cleared from form state
6. **NEW**: User sees validation error message about cleared attachments
7. **NEW**: File input is reset via JavaScript
8. User must fix validation errors and re-upload attachments

## Benefits

### Data Integrity

- Prevents inconsistent state where attachments appear uploaded but aren't saved
- Ensures clean form state after validation failures
- Eliminates confusion about attachment status

### User Experience

- Clear feedback about what happened to uploaded files
- Consistent behavior across all validation scenarios
- Prevents users from thinking their files are "stuck"

### Developer Experience

- Cleaner code logic with explicit attachment handling
- Easier debugging of form state issues
- More predictable behavior in error scenarios

## Implementation Details

### Server-Side Logic

- Attachments are only processed and stored in `DialogEditTemplateData` when validation passes
- When validation fails, `LinkedAttachments` remains `nil`
- Template conditionally renders attachment section based on data presence

### Client-Side Logic

- JavaScript detects validation errors via `aria-invalid` attributes
- File input and preview are reset when validation errors are detected
- Attachment order tracking is reinitialized properly

### Visual Feedback

- Warning message appears when validation fails, explaining attachment clearing
- Uses Bootstrap icon for visual emphasis
- Message is only shown when relevant (validation errors present)

## Edge Cases Handled

### Multiple Validation Attempts

- Attachments are cleared on each validation failure
- User must re-upload files after fixing validation issues
- Consistent behavior regardless of number of attempts

### Mixed Valid/Invalid States

- Handles cases where only title OR content is invalid
- Both fields must be valid for attachments to be preserved
- Clear error messaging regardless of which field fails

### Browser Back/Forward

- File input state is properly reset on page navigation
- No stale file data remains in browser memory
- Consistent state across navigation events

## Testing Recommendations

### Manual Testing

1. Upload files with invalid title → verify files cleared and message shown
2. Upload files with invalid content → verify files cleared and message shown
3. Upload files with both invalid → verify files cleared and message shown
4. Upload files with valid data → verify files preserved and saved
5. Test multiple submission attempts → verify consistent clearing behavior

### Automated Testing

- Test handler logic with various validation states
- Verify template renders correctly with nil attachments
- Test JavaScript validation error detection
- Verify file input reset functionality

## Future Enhancements

### Potential Improvements

- Auto-save draft attachments to prevent total loss
- More granular validation (per-field attachment preservation)
- Progress indication for re-upload process
- Batch validation with partial success handling

### Monitoring

- Track validation failure rates with attachments
- Monitor user behavior after attachment clearing
- Measure re-upload completion rates
- Identify common validation error patterns

---

**Status**: Implemented and Ready for Testing
**Version**: 1.0
**Last Updated**: Current Date
