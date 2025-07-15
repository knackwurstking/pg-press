package pgvis

import (
	"fmt"
	"time"
)

type Cookie struct {
	UserAgent string `json:"user_agent"`
	Value     string `json:"value"`
	ApiKey    string `json:"api_key"`
	LastLogin int64  `json:"last_login"` // LastLogin, (Unix) Milliseconds
}

func (c *Cookie) TimeString() string {
	t := time.UnixMilli(c.LastLogin)
	return fmt.Sprintf(
		"%04d/%02d/%02d %02d:%02d:%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second(),
	)
}
