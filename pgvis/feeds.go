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
		"SELECT id, time, main, cache FROM feeds ORDER BY id DESC",
	)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	return f.scanAllRows(r)
}

func (f *Feeds) ListRange(from int, count int) ([]*Feed, error) {
	r, err := f.db.Query(
		"SELECT id, time, main, cache FROM feeds ORDER BY id DESC LIMIT ? OFFSET ?",
		count, from,
	)
	if err != nil {
		return nil, err
	}

	defer r.Close()
	return f.scanAllRows(r)
}

func (f *Feeds) Add(feed *Feed) error {
	if r, err := f.db.Query("select * from feeds where id = ?", feed.ID); err == nil {
		r.Close()
		if !r.Next() {
			return ErrAlreadyExists
		}
	}

	_, err := f.db.Exec(
		"INSERT INTO feeds (id, time, main, cache) VALUES (?, ?, ?, ?)",
		feed.ID, feed.Time, feed.Main, feed.Cache,
	)
	return err
}

func (f *Feeds) scanAllRows(r *sql.Rows) (feeds []*Feed, err error) {
	feeds = []*Feed{}
	for r.Next() {
		f := &Feed{}
		cache := []byte{}

		if err = r.Scan(&f.ID, &f.Time, &f.Main, &cache); err != nil {
			return nil, err
		}

		f.JSONToCache(cache)
		feeds = append(feeds, f)
	}

	return feeds, nil
}
