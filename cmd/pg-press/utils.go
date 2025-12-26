package main

import (
	"fmt"

	"github.com/knackwurstking/pg-press/internal/db"

	"github.com/SuperPaintman/nice/cli"

	_ "github.com/mattn/go-sqlite3"
)

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

// withDBOperation is a helper that handles common database operations
func withDBOperation(dbPath string, createMode bool, operation func() error) error {
	if err := db.Open(dbPath, createMode); err != nil {
		return fmt.Errorf(
			"Error: could not open database %s (createMode: %t): %v\n",
			dbPath, createMode, err,
		)
	}

	return operation()
}
