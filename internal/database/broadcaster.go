package database

// Broadcaster interface, ex: the `*wshandler.FeedHandler` implements this
type Broadcaster interface {
	Broadcast()
}
