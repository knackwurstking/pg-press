// Package pgvis feed models.
//
// This file defines the Feed data structure and its associated
// validation and utility methods. Feeds represent activity entries
// that track system events and user actions.
package pgvis

import (
	"fmt"
	"html"
	"strings"
	"time"
)

const (
	// Validation constants for feeds
	MinFeedMainLength = 1
	MaxFeedMainLength = 10000

	// HTML templates for different feed types
	userAddTemplate        = `<p>New user: %s</p>`                    // [%(user-name)]
	userRemoveTemplate     = `<p>%s Kicked!</p>`                      // [%(user-name)]
	userNameChangeTemplate = `<p>User name changed from %s to %s</p>` // [%(old-user-name), %(new-user-name)]

	// Trouble report templates with improved formatting
	troubleReportAddTemplate = `
	<p>
    	New trouble report: <a href="/trouble-reports#trouble-report-%d">#%d - %s</a> <br />
        Last modified by: %s
    </p>
` // [%(id), %(id), %(title), %(modified)]

	troubleReportRemoveTemplate = `<p>Trouble report #%d removed</p>` // [%(id)]

	troubleReportUpdateTemplate = `
	<p>
    	Trouble report <a href="/trouble-reports#trouble-report-%d">#%d - %s</a> updated <br />
        Last modified by: %s
    </p>
` // [%(id), %(id), %(title), %(modified)]
)

// Feed represents a feed entry in the system.
// It contains activity information and events that have occurred in the system.
type Feed struct {
	// ID is the unique identifier for the feed entry
	ID int `json:"id"`
	// Time is the UNIX millisecond timestamp when the event occurred
	Time int64 `json:"time"`
	// Main contains the HTML content for displaying the feed entry
	Main string `json:"main"`
	// Cache contains additional cached data related to the feed entry
	Cache any `json:"cache"`
}

// NewFeed creates a new feed entry with the current timestamp.
//
// Parameters:
//   - main: HTML content for the feed entry
//   - cache: Optional cached data
//
// Returns:
//   - *Feed: The newly created feed entry
func NewFeed(main string, cache any) *Feed {
	return &Feed{
		Time:  time.Now().UnixMilli(),
		Main:  strings.TrimSpace(main),
		Cache: cache,
	}
}

// NewFeedWithTime creates a new feed entry with a specific timestamp.
//
// Parameters:
//   - main: HTML content for the feed entry
//   - cache: Optional cached data
//   - timestamp: Unix millisecond timestamp
//
// Returns:
//   - *Feed: The newly created feed entry
func NewFeedWithTime(main string, cache any, timestamp int64) *Feed {
	return &Feed{
		Time:  timestamp,
		Main:  strings.TrimSpace(main),
		Cache: cache,
	}
}

// Validate checks if the feed has valid data.
//
// Returns:
//   - error: ValidationError for the first validation failure, or nil if valid
func (f *Feed) Validate() error {
	// Validate main content
	if f.Main == "" {
		return NewValidationError("main", "cannot be empty", f.Main)
	}
	if len(f.Main) < MinFeedMainLength {
		return NewValidationError("main", "too short", len(f.Main))
	}
	if len(f.Main) > MaxFeedMainLength {
		return NewValidationError("main", "too long", len(f.Main))
	}

	// Validate timestamp
	if f.Time <= 0 {
		return NewValidationError("time", "must be positive", f.Time)
	}

	return nil
}

// GetTime returns the feed time as a Go time.Time.
func (f *Feed) GetTime() time.Time {
	return time.UnixMilli(f.Time)
}

// Age returns the duration since the feed was created.
func (f *Feed) Age() time.Duration {
	return time.Since(f.GetTime())
}

// IsOlderThan checks if the feed is older than the specified duration.
func (f *Feed) IsOlderThan(duration time.Duration) bool {
	return f.Age() > duration
}

// String returns a string representation of the feed.
func (f *Feed) String() string {
	return fmt.Sprintf("Feed{ID: %d, Time: %s, Content: %.50s...}",
		f.ID, f.GetTime().Format("2006-01-02 15:04:05"), f.Main)
}

// Clone creates a deep copy of the feed.
func (f *Feed) Clone() *Feed {
	return &Feed{
		ID:    f.ID,
		Time:  f.Time,
		Main:  f.Main,
		Cache: f.Cache,
	}
}

// User-related feed creators

// NewUserAddFeed creates a feed entry for when a new user is added
func NewUserAddFeed(userName string) *Feed {
	if userName == "" {
		userName = "Unknown User"
	}
	escapedUserName := html.EscapeString(userName)
	main := fmt.Sprintf(userAddTemplate, escapedUserName)
	return NewFeed(main, map[string]any{
		"type":      "user_add",
		"user_name": userName,
	})
}

// NewUserRemoveFeed creates a feed entry for when a user is removed
func NewUserRemoveFeed(userName string) *Feed {
	if userName == "" {
		userName = "Unknown User"
	}
	escapedUserName := html.EscapeString(userName)
	main := fmt.Sprintf(userRemoveTemplate, escapedUserName)
	return NewFeed(main, map[string]any{
		"type":      "user_remove",
		"user_name": userName,
	})
}

// NewUserNameChangeFeed creates a feed entry for when a user changes their name
func NewUserNameChangeFeed(oldName, newName string) *Feed {
	if oldName == "" {
		oldName = "Unknown"
	}
	if newName == "" {
		newName = "Unknown"
	}
	escapedOldName := html.EscapeString(oldName)
	escapedNewName := html.EscapeString(newName)
	main := fmt.Sprintf(userNameChangeTemplate, escapedOldName, escapedNewName)
	return NewFeed(main, map[string]interface{}{
		"type":     "user_name_change",
		"old_name": oldName,
		"new_name": newName,
	})
}

// Trouble report-related feed creators

// NewTroubleReportAddFeed creates a feed entry for when a new trouble report is added
func NewTroubleReportAddFeed(report *TroubleReport) *Feed {
	if report == nil {
		return NewFeed("<p>New trouble report added</p>", map[string]any{
			"type": "trouble_report_add",
		})
	}

	var modifiedBy string
	if report.Modified != nil && report.Modified.User != nil {
		modifiedBy = html.EscapeString(report.Modified.User.UserName)
	} else {
		modifiedBy = "Unknown"
	}

	escapedTitle := html.EscapeString(report.Title)
	main := fmt.Sprintf(
		troubleReportAddTemplate,
		report.ID,
		report.ID,
		escapedTitle,
		modifiedBy,
	)

	return NewFeed(main, map[string]any{
		"type":                 "trouble_report_add",
		"trouble_report_id":    report.ID,
		"trouble_report_title": report.Title,
		"modified_by":          modifiedBy,
	})
}

// NewTroubleReportRemoveFeed creates a feed entry for when a trouble report is removed
func NewTroubleReportRemoveFeed(report *TroubleReport) *Feed {
	if report == nil {
		return NewFeed("<p>Trouble report removed</p>", map[string]interface{}{
			"type": "trouble_report_remove",
		})
	}

	main := fmt.Sprintf(troubleReportRemoveTemplate, report.ID)
	return NewFeed(main, map[string]interface{}{
		"type":                 "trouble_report_remove",
		"trouble_report_id":    report.ID,
		"trouble_report_title": report.Title,
	})
}

// NewTroubleReportUpdateFeed creates a feed entry for when a trouble report is updated
func NewTroubleReportUpdateFeed(report *TroubleReport) *Feed {
	if report == nil {
		return NewFeed("<p>Trouble report updated</p>", map[string]interface{}{
			"type": "trouble_report_update",
		})
	}

	var modifiedBy string
	if report.Modified != nil && report.Modified.User != nil {
		modifiedBy = html.EscapeString(report.Modified.User.UserName)
	} else {
		modifiedBy = "Unknown"
	}

	escapedTitle := html.EscapeString(report.Title)
	main := fmt.Sprintf(
		troubleReportUpdateTemplate,
		report.ID,
		report.ID,
		escapedTitle,
		modifiedBy,
	)

	return NewFeed(main, map[string]interface{}{
		"type":                 "trouble_report_update",
		"trouble_report_id":    report.ID,
		"trouble_report_title": report.Title,
		"modified_by":          modifiedBy,
	})
}
