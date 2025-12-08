package shared

import "fmt"

type TelegramID int64

func (id TelegramID) String() string {
	return fmt.Sprintf("%d", id)
}

// User represents a user entity with relevant information.
type User struct {
	TelegramID TelegramID `json:"telegram_id"` // Unique Telegram ID for the user
	Name       string     `json:"name"`        // User's display name
	ApiKey     string     `json:"api_key"`     // Unique API key for the user
	LastFeed   EntityID   `json:"last_feed"`   // ID of the last feed accessed by the user
}

// Cookie represents a user session with authentication information.
type Cookie struct {
	UserAgent  string     `json:"user_agent"`  // User agent string of the client
	Value      string     `json:"value"`       // Unique UUID cookie value
	TelegramID TelegramID `json:"telegram_id"` // Associated Telegram ID
	LastLogin  UnixMilly  `json:"last_login"`  // Last login timestamp in milliseconds
}

type Session struct {
	ID EntityID `json:"id"` // Unique session ID
}

var _ Entity = (*User)(nil)
var _ Entity = (*Cookie)(nil)
var _ Entity = (*Session)(nil)
