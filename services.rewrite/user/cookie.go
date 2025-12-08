package user

import (
	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/services.rewrite/shared"
)

type CookieService struct {
	setup *shared.Setup `json:"-"`
}

func NewCookieService(setup *shared.Setup) *CookieService {
	return &CookieService{
		setup: setup,
	}
}

func (s *CookieService) TableName() string
func (s *CookieService) Setup(setup *shared.Setup) *errors.MasterError
func (s *CookieService) Create(entity *shared.Cookie) *errors.MasterError
func (s *CookieService) GetByID(id string) (*shared.Cookie, *errors.MasterError)
func (s *CookieService) Update(entity *shared.Cookie) *errors.MasterError
func (s *CookieService) Delete(id string) *errors.MasterError
func (s *CookieService) List() ([]*shared.Cookie, *errors.MasterError)

// Service validation
var _ shared.Service[*shared.Cookie, string] = (*CookieService)(nil)
