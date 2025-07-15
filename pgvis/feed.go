package pgvis

type Feed struct {
	ID      int
	Time    int64 // Time contains an UNIX millisecond timestamp
	Content string
}
