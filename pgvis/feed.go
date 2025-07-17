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

func NewTroubleReportAddFeed(report *TroubleReport) *Feed {
	return NewFeed(
		fmt.Sprintf(
			`
			    <p>New trouble report: #%d</p>
				<p>Last modified by: %s</p>
				<a href="/trouble-reports#feed%d">%s</a>
			`,
			report.ID, report.Modified.User.UserName, report.ID, report.Title,
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

func NewTroubleReportUpdateFeed(report *TroubleReport) *Feed {
	return NewFeed(
		fmt.Sprintf(
			`
			    <p>Trouble report #%d updated</p>
				<p>Last modified by: %s</p>
				<a href="/trouble-reports#feed%d">%s</a>
			`,
			report.ID, report.Modified.User.UserName, report.ID, report.Title,
		),
		nil,
	)
}
