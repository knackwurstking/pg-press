package user

import (
	"database/sql"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/services.rewrite/shared"
)

const (
	SQLCreateCookieTable string = ``
)

type CookieService struct {
	setup *shared.Setup `json:"-"`
}

func NewCookieService() *CookieService {
	return &CookieService{}
}

func (s *CookieService) TableName() string {
	return "cookies"
}

func (s *CookieService) Setup(setup *shared.Setup) *errors.MasterError
func (s *CookieService) Close() *errors.MasterError
func (s *CookieService) Create(entity *shared.Cookie) *errors.MasterError
func (s *CookieService) Update(entity *shared.Cookie) *errors.MasterError
func (s *CookieService) GetByID(id string) (*shared.Cookie, *errors.MasterError)
func (s *CookieService) List() ([]*shared.Cookie, *errors.MasterError)
func (s *CookieService) Delete(id string) *errors.MasterError

func (s *CookieService) createSQLTable() *errors.MasterError {
	_, err := s.setup.DB.Exec(SQLCreateCookieTable, sql.Named("table_name", s.TableName()))
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

func (s *CookieService) checkSetup() *errors.MasterError {
	if s.setup == nil || s.setup.DB == nil {
		return errors.NewValidationError("service not properly setup").MasterError()
	}
	return nil
}

// Service validation
var _ shared.Service[*shared.Cookie, string] = (*CookieService)(nil)
