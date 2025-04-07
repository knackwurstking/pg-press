package types

import "fmt"

type User struct {
	User   string `json:"user"`
	UserID int    `json:"user_id"`
}

func (u *User) Key() string {
	return fmt.Sprintf("%s:%d", u.User, u.UserID)
}
