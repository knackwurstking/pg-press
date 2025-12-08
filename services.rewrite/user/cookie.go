package user

import (
	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/services.rewrite/shared"
)

type CookieService[T *shared.Cookie, ID string] struct {
	setup *shared.Setup `json:"-"`
}

func NewCookieService[T *shared.Cookie, ID string](setup *shared.Setup) *CookieService[T, ID] {
	return &CookieService[T, ID]{
		setup: setup,
	}
}

func (s *CookieService[T, ID]) TableName() string

func (s *CookieService[T, ID]) Setup(setup *shared.Setup) *errors.MasterError

func (s *CookieService[T, ID]) Create(entity T) *errors.MasterError

func (s *CookieService[T, ID]) GetByID(id ID) (T, *errors.MasterError)

func (s *CookieService[T, ID]) Update(entity T) *errors.MasterError

func (s *CookieService[T, ID]) Delete(id ID) *errors.MasterError

func (s *CookieService[T, ID]) List() ([]T, *errors.MasterError)

// Service validation
var _ shared.Service[*shared.Cookie, string] = (*CookieService[*shared.Cookie, string])(nil)
