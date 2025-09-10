package feed

import (
	"fmt"
	"time"

	"github.com/knackwurstking/pgpress/pkg/utils"
)

// Feed represents a simple feed entry with just title, content, and user info.
type Feed struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	UserID    int64  `json:"user_id"`
	CreatedAt int64  `json:"created_at"`
}

// New creates a new feed entry with the current timestamp.
func New(title, content string, userID int64) *Feed {
	return &Feed{
		Title:     title,
		Content:   content,
		UserID:    userID,
		CreatedAt: time.Now().UnixMilli(),
	}
}

// Validate checks if the feed has valid data.
func (f *Feed) Validate() error {
	if f.Title == "" {
		return utils.NewValidationError("title: cannot be empty")
	}
	if len(f.Title) > 255 {
		return utils.NewValidationError("title: too long (max 255 characters)")
	}

	if f.Content == "" {
		return utils.NewValidationError("content: cannot be empty")
	}
	if len(f.Content) > 10000 {
		return utils.NewValidationError("content: too long (max 10000 characters)")
	}

	if f.UserID <= 0 {
		return utils.NewValidationError("user_id: must be positive")
	}

	if f.CreatedAt <= 0 {
		return utils.NewValidationError("created_at: must be positive")
	}

	return nil
}

// GetCreatedTime returns the feed creation time as a Go time.Time.
func (f *Feed) GetCreatedTime() time.Time {
	return time.UnixMilli(f.CreatedAt)
}

// Age returns the duration since the feed was created.
func (f *Feed) Age() time.Duration {
	return time.Since(f.GetCreatedTime())
}

// IsOlderThan checks if the feed is older than the specified duration.
func (f *Feed) IsOlderThan(duration time.Duration) bool {
	return f.Age() > duration
}

// String returns a string representation of the feed.
func (f *Feed) String() string {
	return fmt.Sprintf("Feed{ID: %d, Title: %q, UserID: %d, CreatedAt: %s}",
		f.ID, f.Title, f.UserID, f.GetCreatedTime().Format("2006-01-02 15:04:05"))
}

// Clone creates a copy of the feed.
func (f *Feed) Clone() *Feed {
	return &Feed{
		ID:        f.ID,
		Title:     f.Title,
		Content:   f.Content,
		UserID:    f.UserID,
		CreatedAt: f.CreatedAt,
	}
}
