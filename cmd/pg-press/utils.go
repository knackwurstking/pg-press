package main

import (
	"github.com/knackwurstking/pg-press/internal/common"
	"github.com/knackwurstking/pg-press/internal/shared"

	"github.com/SuperPaintman/nice/cli"

	_ "github.com/mattn/go-sqlite3"
)

func openDB(dbPath string, createMode bool) (*common.DB, error) {
	log.Debug("Opening database: %#v", dbPath)

	db := common.NewDB(&shared.Config{
		DriverName:       "sqlite3",
		DatabaseLocation: dbPath,
		CreateMode:       createMode,
	})

	return db, db.Setup()
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
				return withDBOperation(*customDBPath, action)
			}
		}),
	}
}

// withDBOperation is a helper that handles common database operations
func withDBOperation(customDBPath string, operation func(*common.DB) error) error {
	r, err := openDB(customDBPath, false)
	if err != nil {
		return err
	}

	return operation(r)
}
