package pgvis

type Feed[T any] struct {
	ID      int
	Time    int64 // Time contains an UNIX millisecond timestamp
	Content string
	Data    T
}
