package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/SuperPaintman/nice/cli"
	"github.com/knackwurstking/pgpress/errors"
	"github.com/knackwurstking/pgpress/logger"
	"github.com/knackwurstking/pgpress/services"

	_ "github.com/mattn/go-sqlite3"
)

func log() *logger.Logger {
	return logger.GetComponentLogger("Server")
}

func clog(component string) *logger.Logger {
	return logger.GetComponentLogger(component)
}

func openDB(customPath string) (*services.Registry, error) {
	path := filepath.Join(configPath, databaseFile)
	log().Debug("Database path: %s", path)

	if customPath != "" {
		var err error
		if path, err = filepath.Abs(customPath); err != nil {
			return nil, err
		}
	}

	// Configure SQLite connection string with parameters to prevent locking issues
	connectionString := path + "?_busy_timeout=30000&_journal_mode=WAL&_foreign_keys=on&_synchronous=NORMAL"

	r, err := sql.Open("sqlite3", connectionString)
	if err != nil {
		return nil, err
	}

	// Configure connection pool to prevent resource exhaustion
	r.SetMaxOpenConns(1)    // SQLite works best with single writer
	r.SetMaxIdleConns(1)    // Keep one connection alive
	r.SetConnMaxLifetime(0) // No maximum lifetime

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
	r, err := openDB(*customDBPath)
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
func handleNotFoundError(err error) {
	if errors.IsNotFoundError(err) {
		fmt.Fprintf(os.Stderr, "not found: %s\n", err.Error())
		os.Exit(exitCodeNotFound)
	}
}

// handleGenericError provides consistent error formatting and exit
func handleGenericError(err error, message string) {
	fmt.Fprintf(os.Stderr, "%s: %s\n", message, err.Error())
	os.Exit(exitCodeGeneric)
}
