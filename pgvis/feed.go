// Package pgvis provides feed models for tracking system events and user actions.
package pgvis

import (
	"fmt"
	"html/template"
	"time"
)

// FeedCacheHandler defines the interface for rendering feed cache content.
type FeedCacheHandler interface {
	Render() template.HTML
}

// FeedUserAdd represents a user addition event.
type FeedUserAdd struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func (f *FeedUserAdd) Render() template.HTML {
	return template.HTML(fmt.Sprintf(
		`<div class="feed-item"><div class="feed-item-content">User %s was added.</div></div>`,
		f.Name,
	))
}

// FeedUserRemove represents a user removal event.
type FeedUserRemove struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func (f *FeedUserRemove) Render() template.HTML {
	return template.HTML(fmt.Sprintf(
		`<div class="feed-item"><div class="feed-item-content">User %s was removed.</div></div>`,
		f.Name,
	))
}

// FeedUserNameChange represents a user name change event.
type FeedUserNameChange struct {
	ID  int64  `json:"id"`
	Old string `json:"old"`
	New string `json:"new"`
}

func (f *FeedUserNameChange) Render() template.HTML {
	return template.HTML(fmt.Sprintf(
		`<div class="feed-item"><div class="feed-item-content">User %s changed their name to %s.</div></div>`,
		f.Old, f.New,
	))
}

// FeedTroubleReportAdd represents a trouble report creation event.
type FeedTroubleReportAdd struct {
	ID         int    `json:"id"`
	Title      string `json:"title"`
	ModifiedBy *User  `json:"modified_by"`
}

func (f *FeedTroubleReportAdd) Render() template.HTML {
	return template.HTML(fmt.Sprintf(
		`<div class="feed-item"><div class="feed-item-content">User %s added a new trouble report titled <a href="./trouble-reports/%d">%s</a>.</div></div>`,
		f.ModifiedBy.UserName, f.ID, f.Title,
	))
}

// FeedTroubleReportUpdate represents a trouble report update event.
type FeedTroubleReportUpdate struct {
	ID         int    `json:"id"`
	Title      string `json:"title"`
	ModifiedBy *User  `json:"modified_by"`
}

func (f *FeedTroubleReportUpdate) Render() template.HTML {
	return template.HTML(fmt.Sprintf(
		`<div class="feed-item"><div class="feed-item-content">User %s updated the trouble report titled <a href="./trouble-reports/%d">%s</a>.</div></div>`,
		f.ModifiedBy.UserName, f.ID, f.Title,
	))
}

// FeedTroubleReportRemove represents a trouble report removal event.
type FeedTroubleReportRemove struct {
	ID         int    `json:"id"`
	Title      string `json:"title"`
	ModifiedBy *User  `json:"modified_by"`
}

func (f *FeedTroubleReportRemove) Render() template.HTML {
	return template.HTML(fmt.Sprintf(
		`<div class="feed-item"><div class="feed-item-content">User %s removed the trouble report titled <a href="./trouble-reports/%d">"%s"</a>.</div></div>`,
		f.ModifiedBy.UserName, f.ID, f.Title,
	))
}

// Feed represents a feed entry in the system that tracks activity events.
type Feed struct {
	ID    int   `json:"id"`    // Unique identifier for the feed entry
	Time  int64 `json:"time"`  // UNIX millisecond timestamp when the event occurred
	Cache any   `json:"cache"` // Cached data related to the feed entry
}

// NewFeed creates a new feed entry with the current timestamp.
func NewFeed(cache any) *Feed {
	return &Feed{
		Time:  time.Now().UnixMilli(),
		Cache: cache,
	}
}

// NewFeedWithTime creates a new feed entry with a specific timestamp.
func NewFeedWithTime(cache any, timestamp int64) *Feed {
	return &Feed{
		Time:  timestamp,
		Cache: cache,
	}
}

// Render generates HTML for the feed entry.
func (f *Feed) Render() template.HTML {
	timeStr := f.GetTime().Format("2006-01-02 15:04:05")

	if h, ok := f.Cache.(FeedCacheHandler); ok {
		return template.HTML(fmt.Sprintf(
			`<div id="feed-%d" class="feed-entry" data-id="%d" data-time="%d"><div class="feed-time">%s</div><div class="feed-content">%s</div></div>`,
			f.ID, f.ID, f.Time, timeStr, h.Render(),
		))
	}

	return template.HTML(fmt.Sprintf(
		`<div id="feed-%d" class="feed-entry" data-id="%d" data-time="%d"><div class="feed-time">%s</div><div class="feed-content">%#v</div></div>`,
		f.ID, f.ID, f.Time, timeStr, f.Cache,
	))
}

// Validate checks if the feed has valid data.
func (f *Feed) Validate() error {
	if f.Cache == nil {
		return NewValidationError("cache", "cannot be nil", f.Cache)
	}
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
	return fmt.Sprintf("Feed{ID: %d, Time: %s, Cache: %#v}",
		f.ID, f.GetTime().Format("2006-01-02 15:04:05"), f.Cache)
}

// Clone creates a copy of the feed.
func (f *Feed) Clone() *Feed {
	return &Feed{
		ID:    f.ID,
		Time:  f.Time,
		Cache: f.Cache,
	}
}
