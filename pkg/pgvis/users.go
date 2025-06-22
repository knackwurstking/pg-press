package types

type User struct {
	TelegramID int64  `json:"telegram_id"`
	UserName   string `json:"user_name"`
	ApiKey     string `json:"api_key"`
}

func NewUser() *User {
	return &User{}
}

type Users []*User
