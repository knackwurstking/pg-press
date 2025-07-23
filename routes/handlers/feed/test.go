package feed

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/labstack/echo/v4"
)

// handleCreateTestFeed creates a test feed entry for testing real-time notifications
func (h *Handler) handleCreateTestFeed(c echo.Context) error {
	// Get optional parameters from query string
	feedType := c.QueryParam("type")
	if feedType == "" {
		feedType = "user_add" // default type
	}

	message := c.QueryParam("message")
	if message == "" {
		message = "Test feed created"
	}

	// Get count parameter for creating multiple feeds
	countStr := c.QueryParam("count")
	count := 1
	if countStr != "" {
		if parsedCount, err := strconv.Atoi(countStr); err == nil && parsedCount > 0 && parsedCount <= 10 {
			count = parsedCount
		}
	}

	// Create test feeds
	createdFeeds := make([]*pgvis.Feed, 0, count)
	currentTime := time.Now().Unix()

	for i := 0; i < count; i++ {
		var feedData interface{}

		// Create different types of test data based on feedType
		switch feedType {
		case "user_add":
			feedData = &pgvis.FeedUserAdd{
				ID:   int64(1000 + i),
				Name: fmt.Sprintf("TestUser%d", i+1),
			}
		case "user_remove":
			feedData = &pgvis.FeedUserRemove{
				ID:   int64(2000 + i),
				Name: fmt.Sprintf("RemovedUser%d", i+1),
			}
		case "user_name_change":
			feedData = &pgvis.FeedUserNameChange{
				ID:  int64(3000 + i),
				Old: fmt.Sprintf("OldName%d", i+1),
				New: fmt.Sprintf("NewName%d", i+1),
			}
		case "trouble_report_add":
			feedData = &pgvis.FeedTroubleReportAdd{
				ID:    int64(4000 + i),
				Title: fmt.Sprintf("Test Report %d", i+1),
				ModifiedBy: &pgvis.User{
					UserName: "TestUser",
				},
			}
		case "trouble_report_remove":
			feedData = &pgvis.FeedTroubleReportRemove{
				ID:    int64(5000 + i),
				Title: fmt.Sprintf("Removed Report %d", i+1),
				ModifiedBy: &pgvis.User{
					UserName: "TestUser",
				},
			}
		default:
			// Generic test data
			feedData = map[string]interface{}{
				"message": fmt.Sprintf("%s - %d", message, i+1),
				"test":    true,
				"index":   i + 1,
			}
		}

		// Create the feed
		feed := &pgvis.Feed{
			Time:     currentTime + int64(i), // Slightly different timestamps
			DataType: feedType,
			Data:     feedData,
		}

		// Add to database (this will trigger the real-time notification)
		if err := h.db.Feeds.Add(feed); err != nil {
			return echo.NewHTTPError(
				http.StatusInternalServerError,
				fmt.Errorf("failed to create test feed %d: %w", i+1, err),
			)
		}

		createdFeeds = append(createdFeeds, feed)
	}

	// Return success response with created feed information
	response := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Successfully created %d test feed(s)", count),
		"feeds":   createdFeeds,
		"type":    feedType,
	}

	return c.JSON(http.StatusCreated, response)
}

// RegisterTestRoutes adds test routes for development and testing
func (h *Handler) RegisterTestRoutes(e *echo.Echo) {
	// Only register test routes in development mode
	// In production, you might want to add additional checks or disable this entirely
	e.POST(h.serverPathPrefix+"/test/feed/create", h.handleCreateTestFeed)
}

// Test helper function to create a feed with custom data
func (h *Handler) CreateTestFeedWithData(feedType string, data interface{}) error {
	feed := &pgvis.Feed{
		Time:     time.Now().Unix(),
		DataType: feedType,
		Data:     data,
	}

	return h.db.Feeds.Add(feed)
}

// Batch test feed creation
func (h *Handler) handleCreateBatchTestFeeds(c echo.Context) error {
	// Create various types of test feeds for comprehensive testing
	testFeeds := []struct {
		Type string
		Data interface{}
	}{
		{
			Type: "user_add",
			Data: &pgvis.FeedUserAdd{ID: 9001, Name: "BatchTestUser1"},
		},
		{
			Type: "user_remove",
			Data: &pgvis.FeedUserRemove{ID: 9002, Name: "BatchTestUser2"},
		},
		{
			Type: "trouble_report_add",
			Data: &pgvis.FeedTroubleReportAdd{
				ID:    9003,
				Title: "Batch Test Report",
				ModifiedBy: &pgvis.User{
					UserName: "BatchTester",
				},
			},
		},
	}

	created := 0
	currentTime := time.Now().Unix()

	for i, testFeed := range testFeeds {
		feed := &pgvis.Feed{
			Time:     currentTime + int64(i),
			DataType: testFeed.Type,
			Data:     testFeed.Data,
		}

		if err := h.db.Feeds.Add(feed); err != nil {
			return echo.NewHTTPError(
				http.StatusInternalServerError,
				fmt.Errorf("failed to create batch test feed %d: %w", i+1, err),
			)
		}
		created++
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Successfully created %d batch test feeds", created),
		"types":   []string{"user_add", "user_remove", "trouble_report_add"},
	})
}
