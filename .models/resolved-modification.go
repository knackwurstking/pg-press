package models

type ResolvedModification[T any] struct {
	*Modification[T]
	user *User
}

func NewResolvedModification[T any](m *Modification[T], user *User) *ResolvedModification[T] {
	return &ResolvedModification[T]{
		Modification: m,
		user:         user,
	}
}

func (rm *ResolvedModification[T]) GetUser() *User {
	return rm.user
}
