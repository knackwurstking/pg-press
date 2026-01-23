package main

// CREATE TABLE IF NOT EXISTS tools (
// 		id INTEGER NOT NULL,
// 		position TEXT NOT NULL,
// 		format BLOB NOT NULL,
// 		type TEXT NOT NULL,
// 		code TEXT NOT NULL,
// 		regenerating INTEGER NOT NULL DEFAULT 0,
// 		is_dead INTEGER NOT NULL DEFAULT 0,
// 		press INTEGER,
// 		binding INTEGER,
// 		PRIMARY KEY("id" AUTOINCREMENT)
// 	)`, TableNameTools)

type Format struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type Tool struct {
	ID           int64    `json:"id"`
	Position     Position `json:"position"`
	Format       Format   `json:"format"`
	Type         string   `json:"type"` // Ex: FC, GTC, MASS
	Code         string   `json:"code"` // Ex: G01, G02, ...
	Regenerating bool     `json:"regenerating"`
	IsDead       bool     `json:"is_dead"`
	Press        *int8    `json:"press"` // Press number (0-5) when status is active
	Binding      *int64   `json:"binding"`
}
