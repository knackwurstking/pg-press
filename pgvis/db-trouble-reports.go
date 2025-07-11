package pgvis

import (
	"database/sql"
	"errors"
	"fmt"
)

type TroubleReport struct {
	ID       int                       `json:"id"`
	Title    string                    `json:"title"`
	Content  string                    `json:"content"`
	Modified *Modified[*TroubleReport] `json:"modified"`
}

func NewTroubleReport(m *Modified[*TroubleReport], title, content string) *TroubleReport {
	return &TroubleReport{
		Modified: m,
	}
}

type DBTroubleReports struct {
	db *sql.DB
}

func NewDBTroubleReports(db *sql.DB) *DBTroubleReports {
	query := `
		CREATE TABLE IF NOT EXISTS cookies (
			id INTEGER NOT NULL AUTOINCREMENT,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			modified BLOB NOT NULL,
			PRIMARY KEY("id")
		);
	`

	if _, err := db.Exec(query); err != nil {
		panic(err)
	}

	return &DBTroubleReports{
		db: db,
	}
}

func (db *DBTroubleReports) List(tr *TroubleReport) ([]*TroubleReport, error) {
	// TODO: ...

	return nil, errors.New("under construction")
}

func (db *DBTroubleReports) Get(id int64) (*TroubleReport, error) {
	// TODO: ...

	return nil, errors.New("under construction")
}

func (db *DBTroubleReports) Add(tr *TroubleReport) error {
	query := fmt.Sprintf(
		`INSERT INTO trouble-reports (title, content, modified) VALUES ("%s", "%s", ?)`,
		tr.Title, tr.Content,
	)

	_, err := db.db.Exec(query, tr.Modified.JSON())
	return err
}

func (db *DBTroubleReports) Update(id int64, tr *TroubleReport) error {
	// TODO: ...

	return errors.New("under construction")
}

func (db *DBTroubleReports) Remove(id int64) error {
	// TODO: ...

	return errors.New("under construction")
}
