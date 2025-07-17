package pgvis

import (
	"fmt"
	"html"
	"time"
)

const (
	// HTML templates for different feed types
	userAddTemplate        = `<p>New user: %s</p>`
	userRemoveTemplate     = `<p>%s Kicked!</p>`
	userNameChangeTemplate = `<p>User name changed from %s to %s</p>`

	troubleReportAddTemplate = `<p>New trouble report: #%d</p>
<p>Last modified by: %s</p>
<a href="/trouble-reports#feed%d">%s</a>`

	troubleReportRemoveTemplate = `<p>Trouble report #%d removed</p>`

	troubleReportUpdateTemplate = `<p>Trouble report #%d updated</p>
<p>Last modified by: %s</p>
<a href="/trouble-reports#feed%d">%s</a>`
)

// Feed represents a feed entry in the system
type Feed struct {
	ID    int    `json:"id"`
	Time  int64  `json:"time"`  // UNIX millisecond timestamp
	Main  string `json:"main"`  // HTML content for the feed
	Cache any    `json:"cache"` // Additional cached data
}

// NewFeed creates a new feed entry with the current timestamp
func NewFeed(main string, cache any) *Feed {
	return &Feed{
		Time:  time.Now().UnixMilli(),
		Main:  main,
		Cache: cache,
	}
}

// User-related feed creators

// NewUserAddFeed creates a feed entry for when a new user is added
func NewUserAddFeed(userName string) *Feed {
	escapedUserName := html.EscapeString(userName)
	main := fmt.Sprintf(userAddTemplate, escapedUserName)
	return NewFeed(main, nil)
}

// NewUserRemoveFeed creates a feed entry for when a user is removed
func NewUserRemoveFeed(userName string) *Feed {
	escapedUserName := html.EscapeString(userName)
	main := fmt.Sprintf(userRemoveTemplate, escapedUserName)
	return NewFeed(main, nil)
}

// NewUserNameChangeFeed creates a feed entry for when a user changes their name
func NewUserNameChangeFeed(oldName, newName string) *Feed {
	escapedOldName := html.EscapeString(oldName)
	escapedNewName := html.EscapeString(newName)
	main := fmt.Sprintf(userNameChangeTemplate, escapedOldName, escapedNewName)
	return NewFeed(main, nil)
}

// Trouble report-related feed creators

// NewTroubleReportAddFeed creates a feed entry for when a new trouble report is added
func NewTroubleReportAddFeed(report *TroubleReport) *Feed {
	if report == nil {
		return NewFeed("<p>New trouble report added</p>", nil)
	}

	var modifiedBy string
	if report.Modified.User != nil {
		modifiedBy = html.EscapeString(report.Modified.User.UserName)
	} else {
		modifiedBy = "Unknown"
	}

	escapedTitle := html.EscapeString(report.Title)
	main := fmt.Sprintf(
		troubleReportAddTemplate,
		report.ID,
		modifiedBy,
		report.ID,
		escapedTitle,
	)

	return NewFeed(main, nil)
}

// NewTroubleReportRemoveFeed creates a feed entry for when a trouble report is removed
func NewTroubleReportRemoveFeed(report *TroubleReport) *Feed {
	if report == nil {
		return NewFeed("<p>Trouble report removed</p>", nil)
	}

	main := fmt.Sprintf(troubleReportRemoveTemplate, report.ID)
	return NewFeed(main, nil)
}

// NewTroubleReportUpdateFeed creates a feed entry for when a trouble report is updated
func NewTroubleReportUpdateFeed(report *TroubleReport) *Feed {
	if report == nil {
		return NewFeed("<p>Trouble report updated</p>", nil)
	}

	var modifiedBy string
	if report.Modified.User != nil {
		modifiedBy = html.EscapeString(report.Modified.User.UserName)
	} else {
		modifiedBy = "Unknown"
	}

	escapedTitle := html.EscapeString(report.Title)
	main := fmt.Sprintf(
		troubleReportUpdateTemplate,
		report.ID,
		modifiedBy,
		report.ID,
		escapedTitle,
	)

	return NewFeed(main, nil)
}
