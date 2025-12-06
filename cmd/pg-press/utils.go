package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/SuperPaintman/nice/cli"
	"github.com/knackwurstking/pg-press/env"
	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/services"
	"github.com/lmittmann/tint"

	_ "github.com/mattn/go-sqlite3"
)

func initializeLogging() {
	level := slog.LevelWarn

	switch env.LogLevel {
	case "DEBUG", "debug":
		level = slog.LevelDebug
	case "INFO", "info":
		level = slog.LevelInfo
	case "WARN", "warn":
		level = slog.LevelWarn
	case "ERROR", "error":
		level = slog.LevelError
	}

	var handler slog.Handler
	if env.LogFormat == "text" {
		handler = tint.NewHandler(os.Stderr, &tint.Options{
			AddSource:  true,
			Level:      level,
			TimeFormat: time.DateTime,
		})
	} else {
		handler = slog.NewJSONHandler(
			os.Stderr, &slog.HandlerOptions{
				AddSource: true,
				Level:     level,
			},
		)
	}

	slog.SetDefault(slog.New(handler))
}

func openDB(customPath string, logging bool) (*services.Registry, error) {
	path := filepath.Join(configPath, databaseFile)

	if customPath != "" {
		var err error
		if path, err = filepath.Abs(customPath); err != nil {
			return nil, err
		}
	}

	if logging {
		slog.Info("Database opened", "path", path)
	}

	// Configure SQLite connection string with parameters to prevent locking issues
	connectionString := path + "?_busy_timeout=30000&_journal_mode=WAL&_foreign_keys=on&_synchronous=NORMAL"

	r, err := sql.Open("sqlite3", connectionString)
	if err != nil {
		return nil, err
	}

	// Configure connection pool to prevent resource exhaustion
	r.SetMaxOpenConns(10)                 // Allow more concurrent connections
	r.SetMaxIdleConns(5)                  // Keep some connections alive
	r.SetConnMaxLifetime(5 * time.Minute) // Close connections after 5 minutes

	// Test the connection
	if err = r.Ping(); err != nil {
		r.Close()
		return nil, err
	}

	return services.NewRegistry(r), nil
}

// createDBPathOption creates a standardized database path CLI option
func createDBPathOption(cmd *cli.Command, usage string) *string {
	if usage == "" {
		usage = "Custom database path"
	}
	return cli.String(cmd, "db",
		cli.WithShort("d"),
		cli.Usage(usage),
		cli.Optional,
	)
}

// withDBOperation is a helper that handles common database operations
func withDBOperation(customDBPath *string, operation func(*services.Registry) error) error {
	r, err := openDB(*customDBPath, false)
	if err != nil {
		return err
	}

	return operation(r)
}

// createSimpleCommand creates a CLI command with standardized database access
func createSimpleCommand(name, usage string, action func(*services.Registry) error) cli.Command {
	return cli.Command{
		Name:  name,
		Usage: cli.Usage(usage),
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			customDBPath := createDBPathOption(cmd, "")

			return func(cmd *cli.Command) error {
				return withDBOperation(customDBPath, action)
			}
		}),
	}
}

// handleNotFoundError provides consistent handling for not found errors
func handleNotFoundError(err *errors.MasterError) {
	if err.Code == http.StatusNotFound {
		fmt.Fprintf(os.Stderr, "not found: %v\n", err)
		os.Exit(exitCodeNotFound)
	}
}

// handleGenericError provides consistent error formatting and exit
func handleGenericError(err error, message string) {
	fmt.Fprintf(os.Stderr, "%s: %v\n", message, err)
	os.Exit(exitCodeGeneric)
}
