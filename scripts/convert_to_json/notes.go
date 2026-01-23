package main

import "time"

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
	Linked    string    `json:"linked,omitempty"` // Generic linked entity (e.g., "press_5", "tool_123")
}
