package main

// CREATE TABLE IF NOT EXISTS users (
// 		telegram_id INTEGER NOT NULL,
// 		user_name TEXT NOT NULL,
// 		api_key TEXT NOT NULL UNIQUE,
// 		last_feed TEXT NOT NULL,
// 		PRIMARY KEY("telegram_id")
// 	);

type User struct {
	TelegramID int64  `json:"telegram_id"`
	Name       string `json:"user_name"`
	ApiKey     string `json:"api_key"`
	LastFeed   int64  `json:"last_feed"`
}
