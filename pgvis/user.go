package pgvis

import (
	"os"
	"slices"
	"strconv"
	"strings"
)

type User struct {
	TelegramID int64
	UserName   string
	ApiKey     string // ApiKey is optional and can be nil
	LastFeed   int    // LastFeedViewed contains the Feed ID from the last feed the user viewed
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
