package user

import (
	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/services.rewrite/shared"
)

type SessionService struct {
	*shared.BaseService
}

func NewSessionService(c *shared.Config) *SessionService {
	return &SessionService{
		BaseService: &shared.BaseService{
			Config: c,
		},
	}
}

func (s *SessionService) TableName() string {
	return "sessions"
}

func (s *SessionService) Setup() *errors.MasterError {
	return nil // Only in-memory storage; no setup needed
}

// TODO: Implement session service methods (in-memory)
func (s *SessionService) Create(entity *shared.Session) *errors.MasterError
func (s *SessionService) Update(entity *shared.Session) *errors.MasterError
func (s *SessionService) GetByID(id shared.EntityID) (*shared.Session, *errors.MasterError)
func (s *SessionService) List() ([]*shared.Session, *errors.MasterError)
func (s *SessionService) Delete(id shared.EntityID) *errors.MasterError

// Service validation
var _ shared.Service[*shared.Session, shared.EntityID] = (*SessionService)(nil)
