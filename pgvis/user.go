package pgvis

import (
	"os"
	"slices"
	"strconv"
	"strings"
)

type User struct {
	TelegramID   int64  `json:"telegram_id"`
	UserName     string `json:"user_name"`
	ApiKey       string `json:"api_key"` // ApiKey is optional and can be nil
}

func NewUser(telegramID int64, userName string, apiKey string) *User {
	return &User{
		TelegramID: telegramID,
		UserName:   userName,
		ApiKey:     apiKey,
	}
}

func (u *User) IsAdmin() bool {
	return slices.Contains(
		strings.Split(
			os.Getenv("ADMINS"), ",",
		),
		strconv.Itoa(int(u.TelegramID)),
	)
}
