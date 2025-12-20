package user

import (
	"database/sql"
	"sync"

	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"
)

const (
	SQLCreateCookieTable string = `
		CREATE TABLE IF NOT EXISTS cookies (
			user_agent 	TEXT NOT NULL,
			value 		TEXT NOT NULL,
			user_id 	INTEGER NOT NULL,
			last_login 	INTEGER NOT NULL,

			PRIMARY KEY("value")
		);

		-- Index to quickly find cookies by user_id

		CREATE INDEX IF NOT EXISTS idx_cookies_user_id
		ON cookies(user_id);
	`
	SQLCreateCookie string = `
		INSERT INTO cookies (user_agent, value, user_id, last_login) 
		VALUES (:user_agent, :value, :user_id, :last_login);
	`
	SQLUpdateCookie string = `
		UPDATE cookies
		SET user_agent 	= :user_agent,
			user_id 	= :user_id,
			last_login 	= :last_login
		WHERE value = :value;
	`
	SQLGetCookieByID string = `
		SELECT user_agent, value, user_id, last_login 
		FROM cookies
		WHERE value = :value;
	`
	SQLListCookies string = `
		SELECT user_agent, value, user_id, last_login 
		FROM cookies;
	`
	SQLDeleteCookie string = `
		DELETE FROM cookies
		WHERE value = :value;
	`
)

type CookiesService struct {
	*shared.BaseService

	mx *sync.Mutex `json:"-"`
}

func NewCookiesService(c *shared.Config) *CookiesService {
	return &CookiesService{
		BaseService: shared.NewBaseService(c, "Cookie"),
		mx:          &sync.Mutex{},
	}
}

func (s *CookiesService) Setup() *errors.MasterError {
	return s.BaseService.Setup(DBName, SQLCreateCookieTable)
}

func (s *CookiesService) Create(entity *shared.Cookie) (*shared.Cookie, *errors.MasterError) {
	verr := entity.Validate()
	if verr != nil {
		return nil, verr.MasterError()
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	_, err := s.DB().Exec(SQLCreateCookie,
		sql.Named("user_agent", entity.UserAgent),
		sql.Named("value", entity.Value),
		sql.Named("user_id", entity.UserID),
		sql.Named("last_login", entity.LastLogin),
	)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}

	return entity, nil
}

func (s *CookiesService) Update(entity *shared.Cookie) *errors.MasterError {
	verr := entity.Validate()
	if verr != nil {
		return verr.MasterError()
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	_, err := s.DB().Exec(SQLUpdateCookie,
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

func (s *CookiesService) GetByID(value string) (*shared.Cookie, *errors.MasterError) {
	if value == "" {
		return nil, errors.NewValidationError("cookie value is required").MasterError()
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	r := s.DB().QueryRow(SQLGetCookieByID,
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

func (s *CookiesService) List() ([]*shared.Cookie, *errors.MasterError) {
	s.mx.Lock()
	defer s.mx.Unlock()

	rows, err := s.DB().Query(SQLListCookies)
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

func (s *CookiesService) Delete(value string) *errors.MasterError {
	if value == "" {
		return errors.NewValidationError("cookie value is required").MasterError()
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	_, err := s.DB().Exec(SQLDeleteCookie,
		sql.Named("value", value),
	)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

// Service validation
var _ shared.Service[*shared.Cookie, string] = (*CookiesService)(nil)
