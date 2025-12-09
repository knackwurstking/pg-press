package models

import (
	"fmt"
	"time"

	"github.com/knackwurstking/pg-press/errors"
)

type FeedID int64

// Feed represents a simple feed entry with just title, content, and user info.
type Feed struct {
	ID        FeedID     `json:"id"`
	Title     string     `json:"title"`
	Content   string     `json:"content"`
	UserID    TelegramID `json:"user_id"`
	CreatedAt int64      `json:"created_at"`
}

// New creates a new feed entry with the current timestamp.
func NewFeed(title, content string, userID TelegramID) *Feed {
	return &Feed{
		Title:     title,
		Content:   content,
		UserID:    userID,
		CreatedAt: time.Now().UnixMilli(),
	}
}

// Validate checks if the feed has valid data.
func (f *Feed) Validate() *errors.ValidationError {
	if f.Title == "" {
		return errors.NewValidationError("missing title")
	}
	if len(f.Title) > 255 {
		return errors.NewValidationError("title exceeds 255 chars: %d", len(f.Title))
	}

	if f.Content == "" {
		return errors.NewValidationError("missing content")
	}
	if len(f.Content) > 10000 {
		return errors.NewValidationError("content exceeds 10000 chars: %d", len(f.Content))
	}

	if f.UserID <= 0 {
		return errors.NewValidationError("invalid user id %d", f.UserID)
	}

	if f.CreatedAt <= 0 {
		return errors.NewValidationError("invalid created_at timestamp: %d", f.CreatedAt)
	}

	return nil
}

// GetCreatedTime returns the feed creation time as a Go time.Time.
func (f *Feed) GetCreatedAt() time.Time {
	return time.UnixMilli(f.CreatedAt)
}

// Age returns the duration since the feed was created.
func (f *Feed) Age() time.Duration {
	return time.Since(f.GetCreatedAt())
}

// String returns a string representation of the feed.
func (f *Feed) String() string {
	return fmt.Sprintf("Feed{ID: %d, Title: %q, UserID: %d, CreatedAt: %s}",
		f.ID, f.Title, f.UserID, f.GetCreatedAt().Format("2006-01-02 15:04:05"))
}
