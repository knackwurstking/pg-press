package pgvis

import (
	"fmt"
	"time"
)

// TODO: Need to find a way to handle feeds viewed per user
type Feed struct {
	ID       int
	Time     int64 // Time contains an UNIX millisecond timestamp
	Main     string
	ViewedBy []int64
	Cache    any
}

func NewFeed(main string, cache any) *Feed {
	return &Feed{
		Time:  time.Now().UnixMilli(),
		Main:  main,
		Cache: cache,
	}
}

func NewUserAddFeed(userName string) *Feed {
	return NewFeed(
		fmt.Sprintf(`<p>New user: %s</p>`, userName),
		nil,
	)
}

func NewUserRemoveFeed(userName string) *Feed {
	return NewFeed(
		fmt.Sprintf(`<p>%s Kicked!</p>`, userName),
		nil,
	)
}

func NewUserNameChangeFeed(old, new string) *Feed {
	return NewFeed(
		fmt.Sprintf(`<p>User name changed from %s to %s</p>`, old, new),
		nil,
	)
}
