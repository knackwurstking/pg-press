package pgvis

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

type TroubleReports struct {
	db    *sql.DB
	feeds *Feeds
}

func NewTroubleReports(db *sql.DB, feeds *Feeds) *TroubleReports {
	query := `
		CREATE TABLE IF NOT EXISTS trouble_reports (
			id INTEGER NOT NULL,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			linked_attachments BLOB NOT NULL,
			modified BLOB NOT NULL,
			PRIMARY KEY("id" AUTOINCREMENT)
		);
	`

	if _, err := db.Exec(query); err != nil {
		panic(err)
	}

	return &TroubleReports{
		db:    db,
		feeds: feeds,
	}
}

func (db *TroubleReports) List() ([]*TroubleReport, error) {
	query := `SELECT * FROM trouble_reports ORDER BY id DESC`
	r, err := db.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	trs := []*TroubleReport{}

	for r.Next() {
		tr := TroubleReport{}

		linkedAttachments := []byte{}
		modified := []byte{}

		err = r.Scan(&tr.ID, &tr.Title, &tr.Content, &linkedAttachments, &modified)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(linkedAttachments, &tr.LinkedAttachments)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(modified, &tr.Modified)
		if err != nil {
			return nil, err
		}

		trs = append(trs, &tr)
	}

	return trs, nil
}

func (db *TroubleReports) Get(id int64) (*TroubleReport, error) {
	query := fmt.Sprintf(`SELECT * FROM trouble_reports WHERE id = %d`, id)
	r, err := db.db.Query(query)
	if err != nil {
		return nil, err
	}

	defer r.Close()

	if !r.Next() {
		return nil, ErrNotFound
	}

	tr := &TroubleReport{}

	linkedAttachments := []byte{}
	modified := []byte{}

	err = r.Scan(&tr.ID, &tr.Title, &tr.Content, &linkedAttachments, &modified)
	if err != nil {
		return nil, fmt.Errorf("scan data from database: %s", err.Error())
	}

	err = json.Unmarshal(linkedAttachments, &tr.LinkedAttachments)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(modified, &tr.Modified)
	if err != nil {
		return nil, err
	}

	return tr, nil
}

// TODO: Create a new feed if the trouble report is created successfully
func (db *TroubleReports) Add(tr *TroubleReport) error {
	linkedAttachments, err := json.Marshal(tr.LinkedAttachments)
	if err != nil {
		return err
	}

	modified, err := json.Marshal(tr.Modified)
	if err != nil {
		return err
	}

	_, err = db.db.Exec(
		`INSERT INTO trouble_reports (title, content, linked_attachments, modified) VALUES (?, ?, ?, ?)`,
		tr.Title, tr.Content, linkedAttachments, modified,
	)
	return err
}

// TODO: Create a new feed if the trouble report is updated successfully
func (db *TroubleReports) Update(id int64, tr *TroubleReport) error {
	linkedAttachments, err := json.Marshal(tr.LinkedAttachments)
	if err != nil {
		return err
	}

	modified, err := json.Marshal(tr.Modified)
	if err != nil {
		return err
	}

	_, err = db.db.Exec(
		"UPDATE trouble_reports SET title = ?, content = ?, linked_attachments = ?, modified = ? WHERE id = ?",
		tr.Title, tr.Content, linkedAttachments, modified, id,
	)
	return err
}

func (db *TroubleReports) Remove(id int64) error {
	query := fmt.Sprintf(
		`DELETE FROM trouble_reports WHERE id = %d`,
		id,
	)

	_, err := db.db.Exec(query)
	return err
}
