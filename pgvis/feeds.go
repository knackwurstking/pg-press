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
			main TEXT NOT NULL,
			viewed_by BLOB NOT NULL,
			cache BLOB NOT NULL,
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
		"SELECT id, time, main, viewed_by, cache FROM feeds ORDER BY id DESC",
	)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	return f.scanAllRows(r)
}

func (f *Feeds) ListRange(from int, count int) ([]*Feed, error) {
	r, err := f.db.Query(
		"SELECT id, time, main, viewed_by, cache FROM feeds ORDER BY id DESC LIMIT ? OFFSET ?",
		count, from,
	)
	if err != nil {
		return nil, err
	}

	defer r.Close()
	return f.scanAllRows(r)
}

func (f *Feeds) Add(feed *Feed) error {
	_, err := f.db.Exec(
		"INSERT INTO feeds (time, main, viewed_by, cache) VALUES (?, ?, ?, ?)",
		feed.Time, feed.Main, feed.ViewedByToJSON(), feed.CacheToJSON(),
	)
	return err
}

func (f *Feeds) scanAllRows(r *sql.Rows) (feeds []*Feed, err error) {
	feeds = []*Feed{}
	for r.Next() {
		f := &Feed{}
		cache := []byte{}
		viewedBy := []byte{}

		if err = r.Scan(&f.ID, &f.Time, &f.Main, &viewedBy, &cache); err != nil {
			return nil, err
		}

		f.JSONToViewedBy(viewedBy)
		f.JSONToCache(cache)
		feeds = append(feeds, f)
	}

	return feeds, nil
}
