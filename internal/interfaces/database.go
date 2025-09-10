package interfaces

// Scannable is an interface that abstracts sql.Row and sql.Rows for scanning.
type Scannable interface {
	Scan(dest ...any) error
}

type Broadcaster interface {
	Broadcast()
}
