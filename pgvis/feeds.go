package pgvis

import "database/sql"

type Feeds struct {
	db *sql.DB
}

func NewFeeds(db *sql.DB) *Feeds {
	query := `
		DROP TABLE IF EXISTS feeds;
		CREATE TABLE IF NOT EXISTS feeds (
			id INTEGER NOT NULL,
			time INTEGER NOT NULL,
			data BLOB NOT NULL,
			PRIMARY KEY("id" AUTOINCREMENT)
		);
	`

	if _, err := db.Exec(query); err != nil {
		panic(err)
	}

	return &Feeds{
		db: db,
	}
}

// TODO: List method, allow count (from .. to)
