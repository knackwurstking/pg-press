package services

type Scannable interface {
	Scan(dest ...any) error
}

type Broadcaster interface {
	Broadcast()
}
