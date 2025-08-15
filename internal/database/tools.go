package database

import "database/sql"

const (
	createToolsTableQuery = `
		CREATE TABLE IF NOT EXISTS tools (
			id INTEGER NOT NULL,
			format TEXT NOT NULL,
			PRIMARY KEY("id" AUTOINCREMENT)
		);
	`

	// TODO: Implement the Tools struct.
)

// Tools represents a collection of tools in the database.
type Tools struct {
	db    *sql.DB
	feeds *Feeds
}

func NewTools(db *sql.DB, feeds *Feeds) *Tools {
	if _, err := db.Exec(createToolsTableQuery); err != nil {
		panic(
			NewDatabaseError(
				"create_table",
				"tools",
				"failed to create tools table",
				err,
			),
		)
	}

	return &Tools{
		db:    db,
		feeds: feeds,
	}
}

// TODO: Implement the Tools struct.
