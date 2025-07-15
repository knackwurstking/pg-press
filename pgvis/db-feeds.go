package pgvis

import "database/sql"

type DBFeeds struct {
	db *sql.DB
}

func NewDBFeeds(db *sql.DB) *DBFeeds {
	query := `
		CREATE TABLE IF NOT EXISTS trouble_reports (
			id INTEGER NOT NULL,
			time INTEGER NOT NULL,
			content TEXT NOT NULL,
			PRIMARY KEY("id" AUTOINCREMENT)
		);
	`

	if _, err := db.Exec(query); err != nil {
		panic(err)
	}

	return &DBFeeds{
		db: db,
	}
}
