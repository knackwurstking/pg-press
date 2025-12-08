package user

import (
	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/services.rewrite/shared"
)

const (
	SQLCreateCookieTable string = ``
)

type CookieService struct {
	*shared.BaseService
}

func NewCookieService(c *shared.Config) *CookieService {
	return &CookieService{
		BaseService: &shared.BaseService{
			Config: c,
		},
	}
}

func (s *CookieService) TableName() string {
	return "cookies"
}

func (s *CookieService) Setup() *errors.MasterError {
	return s.BaseService.Setup(s.TableName(), SQLCreateCookieTable)
}

func (s *CookieService) Create(entity *shared.Cookie) *errors.MasterError
func (s *CookieService) Update(entity *shared.Cookie) *errors.MasterError
func (s *CookieService) GetByID(id string) (*shared.Cookie, *errors.MasterError)
func (s *CookieService) List() ([]*shared.Cookie, *errors.MasterError)
func (s *CookieService) Delete(id string) *errors.MasterError

// Service validation
var _ shared.Service[*shared.Cookie, string] = (*CookieService)(nil)
