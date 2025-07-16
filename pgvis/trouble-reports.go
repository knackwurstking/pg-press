package pgvis

import (
	"database/sql"
	"fmt"
)

type TroubleReports struct {
	db *sql.DB
}

func NewTroubleReports(db *sql.DB) *TroubleReports {
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
		db: db,
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

		tr.JSONToLinkedAttachments(linkedAttachments)
		tr.JSONToModified(modified)

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

	tr.JSONToLinkedAttachments(linkedAttachments)
	tr.JSONToModified(modified)

	return tr, nil
}

func (db *TroubleReports) Add(tr *TroubleReport) error {
	query := `INSERT INTO trouble_reports (title, content, linked_attachments, modified) VALUES (?, ?, ?, ?)`

	_, err := db.db.Exec(query, tr.Title, tr.Content, tr.LinkedAttachmentsToJSON(), tr.ModifiedToJSON())
	return err
}

func (db *TroubleReports) Update(id int64, tr *TroubleReport) error {
	query := fmt.Sprintf(
		`UPDATE trouble_reports SET title = ?, content = ?, linked_attachments = ?, modified = ? WHERE id=%d`,
		id,
	)

	_, err := db.db.Exec(query, tr.Title, tr.Content, tr.LinkedAttachmentsToJSON(), tr.ModifiedToJSON())
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
