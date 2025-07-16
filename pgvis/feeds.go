package pgvis

import (
	"database/sql"
)

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

func (f *Feeds) List() ([]*Feed, error) {
	r, err := f.db.Query(
		"SELECT id, time, data FROM feeds ORDER BY id DESC",
	)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	return f.scanAllRows(r)
}

func (f *Feeds) ListRange(from int, count int) ([]*Feed, error) {
	r, err := f.db.Query(
		"SELECT id, time, data FROM feeds ORDER BY id DESC LIMIT ? OFFSET ?",
		count, from,
	)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	return f.scanAllRows(r)
}

func (f *Feeds) scanAllRows(r *sql.Rows) (feeds []*Feed, err error) {
	feeds = []*Feed{}
	for r.Next() {
		f := &Feed{}
		data := []byte{}

		if err = r.Scan(&f.ID, &f.Time, &data); err != nil {
			return nil, err
		}

		f.JSONToData(data)
		feeds = append(feeds, f)
	}

	return feeds, nil
}
