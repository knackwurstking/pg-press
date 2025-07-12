package pgvis

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
)

type TroubleReport struct {
	ID                int                       `json:"id"`
	Title             string                    `json:"title"`
	Content           string                    `json:"content"`
	LinkedAttachments []*Attachment             `json:"linked_attachments"`
	Modified          *Modified[*TroubleReport] `json:"modified"`
}

func NewTroubleReport(m *Modified[*TroubleReport], title, content string) *TroubleReport {
	return &TroubleReport{
		LinkedAttachments: make([]*Attachment, 0),
		Modified: m,
	}
}

type DBTroubleReports struct {
	db *sql.DB
}

func NewDBTroubleReports(db *sql.DB) *DBTroubleReports {
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

	return &DBTroubleReports{
		db: db,
	}
}

func (db *DBTroubleReports) List() ([]*TroubleReport, error) {
	query := `SELECT * FROM trouble_reports ORDER BY id ASC`
	r, err := db.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	trs := []*TroubleReport{}

	for r.Next() {
		tr := TroubleReport{}

		la := []byte{}
		m := []byte{}

		err = r.Scan(&tr.ID, &tr.Title, &tr.Content, &la, &m)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(m, &tr.LinkedAttachments); err != nil {
			return nil, fmt.Errorf("unmarshal \"linked_attachments\" failed: %s", err.Error())
		}

		if err := json.Unmarshal(m, &tr.Modified); err != nil {
			return nil, fmt.Errorf("unmarshal \"modified\" failed: %s", err.Error())
		}

		trs = append(trs, &tr)
	}

	return trs, nil
}

func (db *DBTroubleReports) Get(id int64) (*TroubleReport, error) {
	// TODO: Continue here

	return nil, errors.New("under construction")
}

func (db *DBTroubleReports) Add(tr *TroubleReport) error {
	query := `INSERT INTO trouble_reports (title, content, linked_attachments, modified) VALUES (?, ?, ?, ?)`

	modifiedBytes, err := json.Marshal(tr.Modified)
	if err != nil {
		return fmt.Errorf("marshal \"modified\" failed: %s", err.Error())
	}

	linkedAttachments, err := json.Marshal(tr.LinkedAttachments)
	if err != nil {
		return fmt.Errorf("marshal \"linked_attachments\" failed: %s", err.Error())
	}

	_, err = db.db.Exec(query, tr.Title, tr.Content, linkedAttachments, modifiedBytes)
	return err
}

func (db *DBTroubleReports) Update(id int64, tr *TroubleReport) error {
	query := fmt.Sprintf(
		`UPDATE ... SET title = ?, content = ?, linked_attachments = ?, mondified = ? WHERE id=%d`,
		tr.ID,
	)

	md, err := json.Marshal(tr.Modified)
	if err != nil {
		return fmt.Errorf("marshal \"modified\" to JSON failed")
	}

	la, err := json.Marshal(tr.LinkedAttachments)
	if err != nil {
		return fmt.Errorf("marshal \"modified\" to JSON failed")
	}

	_, err = db.db.Exec(query, tr.Title, tr.Content, la, md)
	return err
}

func (db *DBTroubleReports) Remove(id int64) error {
	query := fmt.Sprintf(
		`DELETE FROM trouble_reports WHERE id = %d`,
		id,
	)

	_, err := db.db.Exec(query)
	return err
}
