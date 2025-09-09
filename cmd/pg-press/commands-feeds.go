package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	database "github.com/knackwurstking/pgpress/internal/database/core"
	"github.com/knackwurstking/pgpress/internal/database/dberror"
	feedmodels "github.com/knackwurstking/pgpress/internal/database/models/feed"
	"github.com/knackwurstking/pgpress/internal/logger"

	"github.com/SuperPaintman/nice/cli"
	"github.com/jedib0t/go-pretty/v6/table"
)

// listFeedsCommand creates a CLI command for listing feeds from the database.
//
// Usage examples:
//
//	pgpress feeds list                                    # List all feeds
//	pgpress feeds list --limit 10                         # List first 10 feeds
//	pgpress feeds list --limit 5 --offset 20              # List 5 feeds starting from offset 20
//	pgpress feeds list --since "2025-07-25"               # List feeds since July 25, 2025
//	pgpress feeds list --before "2025-07-26 15:30:00"     # List feeds before specific date/time
//	pgpress feeds list --since "2025-07-25" --limit 10    # Combined filtering and pagination
func listFeedsCommand() cli.Command {
	return cli.Command{
		Name:  "list",
		Usage: cli.Usage("List all feeds in the database"),
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			customDBPath := cli.String(cmd, "db",
				cli.WithShort("d"),
				cli.Usage("Custom database path"),
				cli.Optional,
			)

			limit := cli.Int(cmd, "limit",
				cli.WithShort("l"),
				cli.Usage("Limit number of results (default: all)"),
				cli.Optional,
			)

			offset := cli.Int(cmd, "offset",
				cli.WithShort("o"),
				cli.Usage("Offset for pagination (default: 0)"),
				cli.Optional,
			)

			since := cli.String(cmd, "since",
				cli.Usage("Show feeds since date (format: 2006-01-02 or 2006-01-02 15:04:05)"),
				cli.Optional,
			)

			before := cli.String(cmd, "before",
				cli.Usage("Show feeds before date (format: 2006-01-02 or 2006-01-02 15:04:05)"),
				cli.Optional,
			)

			return func(cmd *cli.Command) error {
				db, err := openDB(*customDBPath)
				if err != nil {
					return err
				}

				var feeds []*feedmodels.Feed

				// Get feeds based on parameters
				if *limit > 0 {
					feeds, err = db.Feeds.ListRange(*offset, *limit)
				} else {
					feeds, err = db.Feeds.List()
				}

				if err != nil {
					return err
				}

				// Filter by date if specified
				if *since != "" || *before != "" {
					feeds = filterFeedsByDate(feeds, *since, *before)
				}

				if len(feeds) == 0 {
					logger.AppLogger.Info("No feeds found")
					return nil
				}

				// Create table
				t := table.NewWriter()
				t.SetOutputMirror(os.Stdout)
				t.AppendHeader(table.Row{"ID", "Time", "Age", "Title", "Content"})

				// Add rows
				for _, feed := range feeds {
					age := formatAge(feed.Age())
					t.AppendRow(table.Row{
						feed.ID,
						feed.GetCreatedTime().Format("2006-01-02 15:04:05"),
						age,
						feed.Title,
						feed.Content,
					})
				}

				t.SetStyle(table.StyleLight)
				t.Render()

				fmt.Printf("\nTotal: %d feed(s)\n", len(feeds))

				return nil
			}
		}),
	}
}

// removeFeedsCommand creates a CLI command for removing feeds from the database.
//
// Usage examples:
//
//	pgpress feeds remove 1,2,3                           # Remove feeds with specific IDs
//	pgpress feeds remove --older-than 7d                 # Remove feeds older than 7 days
//	pgpress feeds remove --older-than 24h                # Remove feeds older than 24 hours
//	pgpress feeds remove --before "2025-07-25"           # Remove feeds before July 25, 2025
//	pgpress feeds remove --before "2025-07-25 15:30:00"  # Remove feeds before specific date/time
func removeFeedsCommand() cli.Command {
	return cli.Command{
		Name:  "remove",
		Usage: cli.Usage("Remove feeds from the database"),
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			customDBPath := cli.String(cmd, "db",
				cli.WithShort("d"),
				cli.Usage("Custom database path"),
				cli.Optional,
			)

			olderThan := cli.String(cmd, "older-than",
				cli.Usage("Remove feeds older than duration (e.g., 24h, 7d, 30d)"),
				cli.Optional,
			)

			before := cli.String(cmd, "before",
				cli.Usage("Remove feeds before date (format: 2006-01-02 or 2006-01-02 15:04:05)"),
				cli.Optional,
			)

			idsStr := cli.StringArg(cmd, "ids",
				cli.Usage("Feed IDs to remove (comma-separated, e.g., 1,2,3)"),
				cli.Optional,
			)

			return func(cmd *cli.Command) error {
				db, err := openDB(*customDBPath)
				if err != nil {
					return err
				}

				// Remove by IDs
				if *idsStr != "" {
					ids := strings.Split(*idsStr, ",")
					// Trim whitespace from each ID
					for i, id := range ids {
						ids[i] = strings.TrimSpace(id)
					}
					return removeFeedsByIDs(db, ids)
				}

				// Remove by duration
				if *olderThan != "" {
					return removeFeedsByDuration(db, *olderThan)
				}

				// Remove by date
				if *before != "" {
					return removeFeedsByDate(db, *before)
				}

				return fmt.Errorf("must specify either feed IDs, --older-than, or --before")
			}
		}),
	}
}

// Helper functions

func filterFeedsByDate(feeds []*feedmodels.Feed, since, before string) []*feedmodels.Feed {
	var filtered []*feedmodels.Feed

	var sinceTime, beforeTime time.Time
	var err error

	if since != "" {
		sinceTime, err = parseDateTime(since)
		if err != nil {
			logger.AppLogger.Warn("Invalid since date format: %s", err)
			return feeds
		}
	}

	if before != "" {
		beforeTime, err = parseDateTime(before)
		if err != nil {
			logger.AppLogger.Warn("Invalid before date format: %s", err)
			return feeds
		}
	}

	for _, feed := range feeds {
		feedTime := feed.GetCreatedTime()

		if since != "" && feedTime.Before(sinceTime) {
			continue
		}

		if before != "" && feedTime.After(beforeTime) {
			continue
		}

		filtered = append(filtered, feed)
	}

	return filtered
}

func formatAge(duration time.Duration) string {
	days := int(duration.Hours() / 24)
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh", days, hours)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

// This function is no longer needed with the simplified feed structure
// Content is now directly accessible as feed.Content

func removeFeedsByIDs(db *database.DB, ids []string) error {
	var failed []string
	var removed int

	for _, idStr := range ids {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			failed = append(failed, fmt.Sprintf("invalid ID '%s': %s", idStr, err))
			continue
		}

		err = db.Feeds.Delete(int64(id))
		if err != nil {
			if errors.Is(err, dberror.ErrNotFound) {
				failed = append(failed, fmt.Sprintf("feed ID %d not found", id))
			} else {
				failed = append(failed, fmt.Sprintf("failed to remove feed ID %d: %s", id, err))
			}
			continue
		}

		removed++
		fmt.Printf("Removed feed ID: %d\n", id)
	}

	if len(failed) > 0 {
		fmt.Fprintf(os.Stderr, "\nErrors:\n")
		for _, errMsg := range failed {
			fmt.Fprintf(os.Stderr, "  - %s\n", errMsg)
		}
	}

	fmt.Printf("\nSummary: Removed %d feed(s), %d error(s)\n", removed, len(failed))

	if len(failed) > 0 {
		os.Exit(exitCodeGeneric)
	}

	return nil
}

func removeFeedsByDuration(db *database.DB, durationStr string) error {
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		// Try parsing as days if direct parsing fails
		if strings.HasSuffix(durationStr, "d") {
			days, parseErr := strconv.Atoi(strings.TrimSuffix(durationStr, "d"))
			if parseErr != nil {
				return fmt.Errorf("invalid duration format: %s (use format like '24h', '7d', '30d')", durationStr)
			}
			duration = time.Duration(days) * 24 * time.Hour
		} else {
			return fmt.Errorf("invalid duration format: %s (use format like '24h', '7d', '30d')", durationStr)
		}
	}

	cutoffTime := time.Now().Add(-duration)
	timestamp := cutoffTime.UnixMilli()

	rowsAffected, err := db.Feeds.DeleteBefore(timestamp)
	if err != nil {
		return fmt.Errorf("failed to remove feeds: %s", err)
	}

	fmt.Printf("Removed %d feed(s) older than %s (before %s)\n",
		rowsAffected, durationStr, cutoffTime.Format("2006-01-02 15:04:05"))

	return nil
}

func removeFeedsByDate(db *database.DB, dateStr string) error {
	cutoffTime, err := parseDateTime(dateStr)
	if err != nil {
		return fmt.Errorf("invalid date format: %s (use format '2006-01-02' or '2006-01-02 15:04:05')", dateStr)
	}

	timestamp := cutoffTime.UnixMilli()

	rowsAffected, err := db.Feeds.DeleteBefore(timestamp)
	if err != nil {
		return fmt.Errorf("failed to remove feeds: %s", err)
	}

	fmt.Printf("Removed %d feed(s) before %s\n", rowsAffected, cutoffTime.Format("2006-01-02 15:04:05"))

	return nil
}

func parseDateTime(dateStr string) (time.Time, error) {
	// Try different formats
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}
