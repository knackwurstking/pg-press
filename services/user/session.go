package user

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/services/shared"
)

type SessionService struct {
	*shared.BaseService

	sessions map[shared.EntityID]*shared.Session
	mx       *sync.Mutex
}

func NewSessionService(c *shared.Config) *SessionService {
	return &SessionService{
		BaseService: &shared.BaseService{
			Config: c,
		},

		sessions: make(map[shared.EntityID]*shared.Session),
		mx:       &sync.Mutex{},
	}
}

func (s *SessionService) TableName() string {
	return "sessions"
}

func (s *SessionService) Setup() *errors.MasterError {
	return nil // Only in-memory storage; no setup needed
}

// Implement session service methods (in-memory)
func (s *SessionService) Create(entity *shared.Session) *errors.MasterError {
	verr := entity.Validate()
	if verr != nil {
		return verr.MasterError()
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	if _, ok := s.sessions[entity.ID]; ok {
		return errors.NewExistsError("id", entity.ID).MasterError()
	}

	// Add to in-memory storage
	s.sessions[entity.ID] = entity.Clone()

	return nil
}

// NOTE: I need to overwrite the Close method from the BaseService here
func (s *SessionService) Close() *errors.MasterError {
	// No resources to close for in-memory storage
	return nil
}

func (s *SessionService) Update(entity *shared.Session) *errors.MasterError {
	verr := entity.Validate()
	if verr != nil {
		return verr.MasterError()
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	// Update in-memory storage
	if _, ok := s.sessions[entity.ID]; !ok {
		return errors.NewMasterError(
			fmt.Errorf("session %d not found", entity.ID),
			http.StatusNotFound,
		)
	}
	s.sessions[entity.ID] = entity.Clone()

	return nil
}

func (s *SessionService) GetByID(id shared.EntityID) (*shared.Session, *errors.MasterError) {
	s.mx.Lock()
	defer s.mx.Unlock()

	session, ok := s.sessions[id]
	if !ok {
		return nil, errors.NewMasterError(
			fmt.Errorf("session %d not found", id),
			http.StatusNotFound,
		)
	}

	return session.Clone(), nil
}

func (s *SessionService) List() ([]*shared.Session, *errors.MasterError) {
	s.mx.Lock()
	defer s.mx.Unlock()

	sessions := make([]*shared.Session, 0, len(s.sessions))
	for _, session := range s.sessions {
		sessions = append(sessions, session.Clone())
	}

	return sessions, nil
}

func (s *SessionService) Delete(id shared.EntityID) *errors.MasterError {
	s.mx.Lock()
	defer s.mx.Unlock()

	if _, ok := s.sessions[id]; !ok {
		return errors.NewMasterError(
			fmt.Errorf("session %d not found", id),
			http.StatusNotFound,
		)
	}

	delete(s.sessions, id)
	return nil
}

// Service validation
var _ shared.Service[*shared.Session, shared.EntityID] = (*SessionService)(nil)
