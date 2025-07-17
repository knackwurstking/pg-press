package pgvis

import (
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

func NewTroubleReportAddFeed(tr *TroubleReport) *Feed {
	return NewFeed(
		fmt.Sprintf(
			`
			    <p>New trouble report: #%d: %s</p>
				<p>Last modified by: %s</p>
				<a href="/trouble-reports#feed%d">View</a>
			`,
			tr.ID, tr.Title, tr.Modified.User.UserName, tr.ID,
		),
		nil,
	)
}

// TODO: Update... Just like the trouble report add feed
func NewTroubleReportRemoveFeed(troubleReportID int) *Feed {
	return NewFeed(
		fmt.Sprintf(`<p>Trouble report #%d removed</p>`, troubleReportID),
		nil,
	)
}

// TODO: Update... Just like the trouble report add feed
func NewTroubleReportUpdateFeed(troubleReportID int) *Feed {
	return NewFeed(
		fmt.Sprintf(`<p>Trouble report #%d updated</p>`, troubleReportID),
		nil,
	)
}
