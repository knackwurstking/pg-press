package main

import (
	"database/sql"
	"os"

	_ "github.com/mattn/go-sqlite3"

	"github.com/knackwurstking/pg-vis/pkg/pgvis"
)

func openDB(path string) (*pgvis.DB, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	return pgvis.NewDB(db), nil
}
