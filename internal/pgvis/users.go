package pgvis

type User struct {
	TelegramID int64  `json:"telegram_id"`
	UserName   string `json:"user_name"`
	ApiKey     string `json:"api_key"` // ApiKey is optional and can be nil
}

func NewUser(telegramID int64, userName string, apiKey string) *User {
	return &User{
		TelegramID: telegramID,
		UserName:   userName,
		ApiKey:     apiKey,
	}
}
