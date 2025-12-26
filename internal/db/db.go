package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/knackwurstking/pg-press/internal/errors"
)

type Scannable interface {
	Scan(dest ...any) error
}

var (
	dbTool  *sql.DB
	dbPress *sql.DB
	dbNote  *sql.DB
	dbUser  *sql.DB
)

func Open(path string, allowCreate bool) error {
	if err := os.MkdirAll(path, 0700); err != nil {
		return fmt.Errorf("failed to create database directory: %v", err)
	}

	mode := "rw"
	if allowCreate {
		mode = "rwc"
	}

	wg := &sync.WaitGroup{}
	chErr := make(chan error, 4)
	for _, name := range []string{"tool", "press", "note", "user"} {
		wg.Go(func() {
			path = fmt.Sprintf(
				"file:%s.sqlite?cache=shared&mode=%s&_journal=WAL&_sync=0",
				filepath.Join(path, name), mode,
			)
			db, err := sql.Open("sqlite3", path)
			if err != nil {
				chErr <- fmt.Errorf("failed to open %s database: %v", name, err)
				return
			}

			// Configure connection pool to prevent resource exhaustion
			db.SetMaxOpenConns(10)                 // Allow more concurrent connections
			db.SetMaxIdleConns(5)                  // Keep some connections alive
			db.SetConnMaxLifetime(5 * time.Minute) // Close connections after 5 minutes

			switch name {
			case "tool":
				dbTool = db
				if err = createTable(db, sqlCreateMetalSheetsTable); err != nil {
					chErr <- errors.Wrap(err, "failed to create metal_sheets table")
					return
				}
				if err = createTable(db, sqlCreateToolRegenerationsTable); err != nil {
					chErr <- errors.Wrap(err, "failed to create tool_regenerations table")
					return
				}
				if err = createTable(db, sqlCreateToolsTable); err != nil {
					chErr <- errors.Wrap(err, "failed to create tools table")
					return
				}

			case "press":
				dbPress = db
				if err = createTable(db, sqlCreateCyclesTable); err != nil {
					chErr <- errors.Wrap(err, "failed to create cycles table")
					return
				}
				if err = createTable(db, sqlCreatePressesTable); err != nil {
					chErr <- errors.Wrap(err, "failed to create presses table")
					return
				}
				if err = createTable(db, sqlCreatePressRegenerationsTable); err != nil {
					chErr <- errors.Wrap(err, "failed to create press_regenerations table")
					return
				}

			case "note":
				dbNote = db
				if err := createTable(db, sqlCreateNotesTable); err != nil {
					chErr <- errors.Wrap(err, "failed to create notes table")
					return
				}

			case "user":
				dbUser = db
				if err := createTable(db, sqlCreateCookiesTable); err != nil {
					chErr <- errors.Wrap(err, "failed to create cookies table")
					return
				}
				if err := createTable(db, sqlCreateUsersTable); err != nil {
					chErr <- errors.Wrap(err, "failed to create users table")
					return
				}
			}

			chErr <- nil
		})
	}
	wg.Wait()
	close(chErr)

	var errs []string
	for e := range chErr {
		if e != nil {
			errs = append(errs, e.Error())
		}
	}

	return fmt.Errorf("%s", strings.Join(errs, "\n"))
}

func Close() {
	for _, db := range []*sql.DB{dbTool, dbPress, dbNote, dbUser} {
		if db != nil {
			db.Close()
		}
	}
}

func createTable(db *sql.DB, query string) error {
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create table: %v", err)
	}
	return nil
}
