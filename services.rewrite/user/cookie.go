package user

import (
	"database/sql"
	"sync"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/services.rewrite/shared"
)

const (
	SQLCreateCookieTable string = `
		CREATE TABLE IF NOT EXISTS :table_name (
			user_agent 	TEXT NOT NULL,
			value 		TEXT PRIMARY KEY NOT NULL,
			user_id 	INTEGER NOT NULL,
			last_login 	INTEGER NOT NULL,
		);

		-- Index to quickly find cookies by user_id

		CREATE INDEX IF NOT EXISTS idx_(:table_name)_user_id
		ON :table_name(user_id);

		-- Index to quickly find cookies by value

		create index if not exists idx_(:table_name)_value
		on :table_name(value);
	`
	SQLCreateCookie string = `
		INSERT INTO :table_name (user_agent, value, user_id, last_login) 
		VALUES (:user_agent, :value, :user_id, :last_login);
	`
	SQLUpdateCookie string = `
		UPDATE :table_name
		SET user_agent 	= :user_agent,
			user_id 	= :user_id,
			last_login 	= :last_login
		WHERE value = :value;
	`
	SQLGetCookieByID string = `
		SELECT user_agent, value, user_id, last_login 
		FROM :table_name
		WHERE value = :value;
	`
	SQLListCookies string = `
		SELECT user_agent, value, user_id, last_login 
		FROM :table_name;
	`
	SQLDeleteCookie string = `
		DELETE FROM :table_name
		WHERE value = :value;
	`
)

type CookieService struct {
	*shared.BaseService

	mx *sync.Mutex `json:"-"`
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

func (s *CookieService) Create(entity *shared.Cookie) *errors.MasterError {
	verr := entity.Validate()
	if verr != nil {
		return verr.MasterError()
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	_, err := s.DB.Exec(SQLCreateCookie,
		sql.Named("user_agent", entity.UserAgent),
		sql.Named("value", entity.Value),
		sql.Named("user_id", entity.UserID),
		sql.Named("last_login", entity.LastLogin),
	)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

func (s *CookieService) Update(entity *shared.Cookie) *errors.MasterError {
	verr := entity.Validate()
	if verr != nil {
		return verr.MasterError()
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	_, err := s.DB.Exec(SQLUpdateCookie,
		sql.Named("user_agent", entity.UserAgent),
		sql.Named("value", entity.Value),
		sql.Named("user_id", entity.UserID),
		sql.Named("last_login", entity.LastLogin),
	)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

func (s *CookieService) GetByID(value string) (*shared.Cookie, *errors.MasterError) {
	if value == "" {
		return nil, errors.NewValidationError("cookie value is required").MasterError()
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	r := s.DB.QueryRow(SQLGetCookieByID,
		sql.Named("table_name", s.TableName()),
		sql.Named("value", value),
	)

	// Scan row into user entity
	var c = &shared.Cookie{}
	err := r.Scan(&c.UserAgent, &c.Value, &c.UserID, &c.LastLogin)
	if err != nil {
		return c, errors.NewMasterError(err, 0)
	}

	return c, nil
}

func (s *CookieService) List() ([]*shared.Cookie, *errors.MasterError) {
	s.mx.Lock()
	defer s.mx.Unlock()

	rows, err := s.DB.Query(SQLListCookies,
		sql.Named("table_name", s.TableName()),
	)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	defer rows.Close()

	cookies := []*shared.Cookie{}
	for rows.Next() {
		c := &shared.Cookie{}
		err := rows.Scan(&c.UserAgent, &c.Value, &c.UserID, &c.LastLogin)
		if err != nil {
			return nil, errors.NewMasterError(err, 0)
		}
		cookies = append(cookies, c)
	}

	return cookies, nil
}

func (s *CookieService) Delete(value string) *errors.MasterError {
	if value == "" {
		return errors.NewValidationError("cookie value is required").MasterError()
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	_, err := s.DB.Exec(SQLDeleteCookie,
		sql.Named("table_name", s.TableName()),
		sql.Named("value", value),
	)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

// Service validation
var _ shared.Service[*shared.Cookie, string] = (*CookieService)(nil)
