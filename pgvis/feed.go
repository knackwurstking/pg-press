package pgvis

import (
	"encoding/json"
	"fmt"
	"time"
)

type Feed struct {
	ID    int
	Time  int64 // Time contains an UNIX millisecond timestamp
	Main  string
	Cache any
}

func NewFeed(main string, cache any) *Feed {
	return &Feed{
		Time:  time.Now().UnixMilli(),
		Main:  main,
		Cache: cache,
	}
}

func (f *Feed) CacheToJSON() []byte {
	b, err := json.Marshal(f.Cache)
	if err != nil {
		panic(err)
	}
	return b
}

func (f *Feed) JSONToCache(b []byte) {
	err := json.Unmarshal(b, &f.Cache)
	if err != nil {
		panic(err)
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
