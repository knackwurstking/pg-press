package shared

type User struct {
	TelegramID int64  `json:"telegram_id"`
	Name       string `json:"name"`
	ApiKey     string `json:"api_key"`
	LastFeed   int64  `json:"last_feed"`
}
