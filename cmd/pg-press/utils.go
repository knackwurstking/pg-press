package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/knackwurstking/pg-press/internal/env"
	"github.com/knackwurstking/pg-press/internal/common"
	"github.com/knackwurstking/pg-press/internal/shared"

	"github.com/SuperPaintman/nice/cli"
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

func openDB(dbPath string, logging bool) (*common.DB, error) {
	if logging {
		slog.Info("Database opened", "path", dbPath)
	}

	return common.NewDB(&shared.Config{
		DriverName:       "sqlite3",
		DatabaseLocation: dbPath,
	}), nil
}

// createDBPathOption creates a standardized database path CLI option
func createDBPathOption(cmd *cli.Command) *string {
	db := cli.String(cmd, "db",
		cli.WithShort("d"),
		cli.Usage("Custom database path"),
		cli.Optional,
	)
	*db = configPath
	return db
}

// createSimpleCommand creates a CLI command with standardized database access
func createSimpleCommand(name, usage string, action func(*common.DB) error) cli.Command {
	return cli.Command{
		Name:  name,
		Usage: cli.Usage(usage),
		Action: cli.ActionFunc(func(cmd *cli.Command) cli.ActionRunner {
			customDBPath := createDBPathOption(cmd)

			return func(cmd *cli.Command) error {
				return withDBOperation(customDBPath, action)
			}
		}),
	}
}

// withDBOperation is a helper that handles common database operations
func withDBOperation(customDBPath *string, operation func(*common.DB) error) error {
	r, err := openDB(*customDBPath, false) // TODO: Continue here...
	if err != nil {
		return err
	}

	return operation(r)
}
