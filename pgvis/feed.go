// Package pgvis feed models.
//
// This file defines the Feed data structure and its associated
// validation and utility methods. Feeds represent activity entries
// that track system events and user actions.
package pgvis

import (
	"fmt"
	"html/template"
	"time"
)

const (
	// Validation constants for feeds
	MinFeedMainLength = 1
	MaxFeedMainLength = 10000
)

type FeedCacheHandler interface {
	Render() template.HTML
}

type FeedUserAdd struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func (f *FeedUserAdd) Render() template.HTML {
	return template.HTML(
		fmt.Sprintf(
			`
			<div class="feed-item">
				<div class="feed-item-content">
					User %s was added.
				</div>
			</div>
			`,
			f.Name),
	)
}

type FeedUserRemove struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func (f *FeedUserRemove) Render() template.HTML {
	return template.HTML(fmt.Sprintf(
		`
			<div class="feed-item">
				<div class="feed-item-content">
					User %s was removed.
				</div>
			</div>
			`,
		f.Name),
	)
}

type FeedUserNameChange struct {
	ID  int64  `json:"id"`
	Old string `json:"old"`
	New string `json:"new"`
}

func (f *FeedUserNameChange) Render() template.HTML {
	return template.HTML(fmt.Sprintf(
		`
			<div class="feed-item">
				<div class="feed-item-content">
					User %s changed their name to %s.
				</div>
			</div>
			`,
		f.Old, f.New),
	)
}

type FeedTroubleReportAdd struct {
	ID         int    `json:"id"`
	Title      string `json:"title"`
	ModifiedBy *User  `json:"modified_by"`
}

func (f *FeedTroubleReportAdd) Render() template.HTML {
	return template.HTML(fmt.Sprintf(
		`
			<div class="feed-item">
				<div class="feed-item-content">
					User %s added a new trouble report titled <a href="./trouble-reports/%d">%s</a>.
				</div>
			</div>
			`,
		f.ModifiedBy.UserName, f.ID, f.Title),
	)
}

type FeedTroubleReportUpdate struct {
	ID         int    `json:"id"`
	Title      string `json:"title"`
	ModifiedBy *User  `json:"modified_by"`
}

func (f *FeedTroubleReportUpdate) Render() template.HTML {
	return template.HTML(fmt.Sprintf(
		`
			<div class="feed-item">
				<div class="feed-item-content">
					User %s updated the trouble report titled <a href="./trouble-reports/%d">%s</a>.
				</div>
			</div>
			`,
		f.ModifiedBy.UserName, f.ID, f.Title),
	)
}

type FeedTroubleReportRemove struct {
	ID         int    `json:"id"`
	Title      string `json:"title"`
	ModifiedBy *User  `json:"modified_by"`
}

func (f *FeedTroubleReportRemove) Render() template.HTML {
	return template.HTML(fmt.Sprintf(
		`
			<div class="feed-item">
				<div class="feed-item-content">
					User %s removed the trouble report titled <a href="./trouble-reports/%d">"%s"</a>.
				</div>
			</div>
			`,
		f.ModifiedBy.UserName, f.ID, f.Title),
	)
}

// Feed represents a feed entry in the system.
// It contains activity information and events that have occurred in the system.
type Feed struct {
	// ID is the unique identifier for the feed entry
	ID int `json:"id"`
	// Time is the UNIX millisecond timestamp when the event occurred
	Time int64 `json:"time"`
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
func NewFeed(cache any) *Feed {
	return &Feed{
		Time:  time.Now().UnixMilli(),
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
func NewFeedWithTime(main template.HTML, cache any, timestamp int64) *Feed {
	return &Feed{
		Time:  timestamp,
		Cache: cache,
	}
}

// Render generates HTML for the feed entry.
//
// Returns:
//   - template.HTML: The rendered HTML content for the feed entry
func (f *Feed) Render() template.HTML {
	// Type assert through all Feed types
	if h, ok := f.Cache.(FeedCacheHandler); ok {
		return template.HTML(
			fmt.Sprintf(
				`
         			<div id="feed-%d" class="feed-entry" data-id="%d" data-time="%d">
                 		<div class="feed-time">%s</div>
                 		<div class="feed-content">%s</div>
                  	</div>
               `,
				f.ID, f.ID, f.Time, f.GetTime().Format("2006-01-02 15:04:05"), h.Render()),
		)
	}

	return template.HTML(
		fmt.Sprintf(
			`
         			<div id="feed-%d" class="feed-entry" data-id="%d" data-time="%d">
                 		<div class="feed-time">%s</div>
                 		<div class="feed-content">%#v</div>
                  	</div>
               `,
			f.ID, f.ID, f.Time, f.GetTime().Format("2006-01-02 15:04:05"), f.Cache),
	)
}

// Validate checks if the feed has valid data.
//
// Returns:
//   - error: ValidationError for the first validation failure, or nil if valid
func (f *Feed) Validate() error {
	if f.Cache == nil {
		return NewValidationError("cache", "cannot be nil", f.Cache)
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
	return fmt.Sprintf("Feed{ID: %d, Time: %s, Cache: %#v}",
		f.ID, f.GetTime().Format("2006-01-02 15:04:05"), f.Cache)
}

// Clone creates a deep copy of the feed.
func (f *Feed) Clone() *Feed {
	return &Feed{
		ID:    f.ID,
		Time:  f.Time,
		Cache: f.Cache,
	}
}
