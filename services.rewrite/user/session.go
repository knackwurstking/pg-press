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

func (s *SessionService) TableName() string
func (s *SessionService) Setup() *errors.MasterError
func (s *SessionService) Create(entity *shared.Session) *errors.MasterError
func (s *SessionService) Update(entity *shared.Session) *errors.MasterError
func (s *SessionService) GetByID(id shared.EntityID) (*shared.Session, *errors.MasterError)
func (s *SessionService) List() ([]*shared.Session, *errors.MasterError)
func (s *SessionService) Delete(id shared.EntityID) *errors.MasterError

// Service validation
var _ shared.Service[*shared.Session, shared.EntityID] = (*SessionService)(nil)
