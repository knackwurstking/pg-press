package user

import (
	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/services.rewrite/shared"
)

type SessionService[T *shared.Session, ID shared.EntityID] struct {
	setup *shared.Setup `json:"-"`
}

func NewSessionService[T *shared.Session, ID shared.EntityID](setup *shared.Setup) *SessionService[T, ID] {
	return &SessionService[T, ID]{
		setup: setup,
	}
}

func (s *SessionService[T, ID]) TableName() string

func (s *SessionService[T, ID]) Setup(setup *shared.Setup) *errors.MasterError

func (s *SessionService[T, ID]) Create(entity T) *errors.MasterError

func (s *SessionService[T, ID]) GetByID(id ID) (T, *errors.MasterError)

func (s *SessionService[T, ID]) Update(entity T) *errors.MasterError

func (s *SessionService[T, ID]) Delete(id ID) *errors.MasterError

func (s *SessionService[T, ID]) List() ([]T, *errors.MasterError)

// Service validation
var _ shared.Service[*shared.Session, shared.EntityID] = (*SessionService[*shared.Session, shared.EntityID])(nil)
