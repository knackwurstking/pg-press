package shared

import "github.com/knackwurstking/pg-press/errors"

type Session struct {
	ID EntityID `json:"id"` // Unique session ID
}

func (e *Session) Validate() *errors.ValidationError

func (e *Session) Clone() *Session {
	return &Session{
		ID: e.ID,
	}
}

var _ Entity[*Session] = (*Session)(nil)
