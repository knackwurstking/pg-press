package user

import (
	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/services.rewrite/shared"
)

type SessionService struct {
	setup *shared.Setup `json:"-"`
}

func NewSessionService(setup *shared.Setup) *SessionService {
	return &SessionService{
		setup: setup,
	}
}

func (s *SessionService) TableName() string
func (s *SessionService) Setup(setup *shared.Setup) *errors.MasterError
func (s *SessionService) Create(entity *shared.Session) *errors.MasterError
func (s *SessionService) GetByID(id shared.EntityID) (*shared.Session, *errors.MasterError)
func (s *SessionService) Update(entity *shared.Session) *errors.MasterError
func (s *SessionService) Delete(id shared.EntityID) *errors.MasterError
func (s *SessionService) List() ([]*shared.Session, *errors.MasterError)

// Service validation
var _ shared.Service[*shared.Session, shared.EntityID] = (*SessionService)(nil)
