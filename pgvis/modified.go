package pgvis

type Modified struct {
	User *User `json:"user"`
}

func NewModified(user *User) *Modified {
	return &Modified{
		User: user,
	}
}
