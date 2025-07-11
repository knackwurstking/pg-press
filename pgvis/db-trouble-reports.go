package pgvis

import "database/sql"

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

// TODO: Add methods: List, Add, Update, Remove
