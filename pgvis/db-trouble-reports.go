package pgvis

import (
	"database/sql"
	"encoding/json"
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
		Title:             title,
		Content:           content,
		LinkedAttachments: make([]*Attachment, 0),
		Modified:          m,
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

		linkedAttachments := []byte{}
		modified := []byte{}

		err = r.Scan(&tr.ID, &tr.Title, &tr.Content, &linkedAttachments, &modified)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(linkedAttachments, &tr.LinkedAttachments); err != nil {
			return nil, fmt.Errorf("unmarshal \"linked_attachments\" failed: %s", err.Error())
		}

		if err := json.Unmarshal(modified, &tr.Modified); err != nil {
			return nil, fmt.Errorf("unmarshal \"modified\" failed: %s", err.Error())
		}

		trs = append(trs, &tr)
	}

	return trs, nil
}

func (db *DBTroubleReports) Get(id int64) (*TroubleReport, error) {
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
		return nil, fmt.Errorf("unmarshal \"linked_attachments\" from database: %s", err.Error())
	}

	err = json.Unmarshal(modified, &tr.Modified)
	if err != nil {
		return nil, fmt.Errorf("unmarshal \"modified\" from database: %s", err.Error())
	}

	return tr, nil
}

func (db *DBTroubleReports) Add(tr *TroubleReport) error {
	query := `INSERT INTO trouble_reports (title, content, linked_attachments, modified) VALUES (?, ?, ?, ?)`

	modified, err := json.Marshal(tr.Modified)
	if err != nil {
		return fmt.Errorf("marshal \"modified\" failed: %s", err.Error())
	}

	linkedAttachments, err := json.Marshal(tr.LinkedAttachments)
	if err != nil {
		return fmt.Errorf("marshal \"linked_attachments\" failed: %s", err.Error())
	}

	_, err = db.db.Exec(query, tr.Title, tr.Content, linkedAttachments, modified)
	return err
}

func (db *DBTroubleReports) Update(id int64, tr *TroubleReport) error {
	query := fmt.Sprintf(
		`UPDATE ... SET title = ?, content = ?, linked_attachments = ?, mondified = ? WHERE id=%d`,
		tr.ID,
	)

	modified, err := json.Marshal(tr.Modified)
	if err != nil {
		return fmt.Errorf("marshal \"modified\" to JSON: %s", err.Error())
	}

	linkedAttachments, err := json.Marshal(tr.LinkedAttachments)
	if err != nil {
		return fmt.Errorf("marshal \"linked_attachments\" to JSON: %s", err.Error())
	}

	_, err = db.db.Exec(query, tr.Title, tr.Content, linkedAttachments, modified)
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
