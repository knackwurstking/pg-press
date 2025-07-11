package pgvis

type Modified[T any] struct {
	User       *User `json:"user"`
	Time       int64 `json:"time"`
	DataBefore T     `json:"data"`
}

func NewModified[T any](user *User) *Modified[T] {
	return &Modified[T]{
		User: user,
	}
}
