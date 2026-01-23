package main

import "time"

// CREATE TABLE IF NOT EXISTS notes (
// 		id INTEGER NOT NULL,
// 		level INTEGER NOT NULL,
// 		content TEXT NOT NULL,
// 		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
// 		linked TEXT,
// 		PRIMARY KEY("id" AUTOINCREMENT)
// 	)

const (
	LevelInfo Level = iota
	LevelAttention
	LevelBroken
)

type Level int

type Note struct {
	ID        int64     `json:"id"`
	Level     Level     `json:"level"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	Linked    string    `json:"linked,omitempty"`
}
