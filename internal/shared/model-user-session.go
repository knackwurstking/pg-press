package shared

import "github.com/knackwurstking/pg-press/internal/errors"

type Session struct {
	ID EntityID `json:"id"` // Unique session ID
}

func (e *Session) Validate() *errors.ValidationError {
	// Add validation logic here if needed
	return nil
}

func (e *Session) Clone() *Session {
	return &Session{
		ID: e.ID,
	}
}

func (e *Session) String() string {
	return "Session{ID:" + e.ID.String() + "}"
}

var _ Entity[*Session] = (*Session)(nil)
