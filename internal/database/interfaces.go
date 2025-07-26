package database

// Broadcaster interface for handling feed update notifications
type Broadcaster interface {
	Broadcast()
}
