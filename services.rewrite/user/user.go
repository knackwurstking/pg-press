package user

import (
	"database/sql"
	"sync"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/services.rewrite/shared"
)

const (
	SQLCreateUserTable string = `
		CREATE TABLE IF NOT EXISTS :table_name (
			id 			INTEGER PRIMARY KEY NOT NULL,
			name 	TEXT NOT NULL,
			api_key 	TEXT NOT NULL UNIQUE,
			last_feed 	TEXT NOT NULL
		);
	`
	SQLCreateUser string = `
		INSERT INTO :table_name (name, api_key, last_feed) 
		VALUES (:name, :api_key, :last_feed);
	`
	SQLGetUserByID string = `
		SELECT id, name, api_key, last_feed 
		FROM :table_name
		WHERE id = :id;
	`
	SQLUpdateUser string = `
		UPDATE :table_name
		SET name 	= :name,
			api_key 	= :api_key,
			last_feed 	= :last_feed
		WHERE id = :id;
	`
	SQLDeleteUser string = `
		DELETE FROM :table_name
		WHERE id = :id;
	`
	SQLListUsers string = `
		SELECT id, name, api_key, last_feed 
		FROM :table_name;
	`
)

type UserService struct {
	*shared.BaseService

	mx *sync.Mutex `json:"-"`
}

func NewUserService(c *shared.Config) *UserService {
	return &UserService{
		BaseService: &shared.BaseService{
			Config: c,
		},

		mx: &sync.Mutex{},
	}
}

func (s *UserService) TableName() string {
	return "users"
}

func (s *UserService) Setup() *errors.MasterError {
	return s.BaseService.Setup(s.TableName(), SQLCreateUserTable)
}

func (s *UserService) Create(entity *shared.User) *errors.MasterError {
	verr := entity.Validate()
	if verr != nil {
		return verr.MasterError()
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	r, err := s.DB.Exec(SQLCreateUser,
		sql.Named("table_name", s.TableName()),
		sql.Named("name", entity.Name),
		sql.Named("api_key", entity.ApiKey),
		sql.Named("last_feed", entity.LastFeed),
	)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	// Store the inserted ID back into the entity
	id, _ := r.LastInsertId()
	entity.ID = shared.TelegramID(id)

	return nil
}

func (s *UserService) Update(entity *shared.User) *errors.MasterError {
	verr := entity.Validate()
	if verr != nil {
		return verr.MasterError()
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	_, err := s.DB.Exec(SQLUpdateUser,
		sql.Named("table_name", s.TableName()),
		sql.Named("id", entity.ID),
		sql.Named("name", entity.Name),
		sql.Named("api_key", entity.ApiKey),
		sql.Named("last_feed", entity.LastFeed),
	)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

func (s *UserService) GetByID(id shared.TelegramID) (*shared.User, *errors.MasterError) {
	if id <= 0 {
		return nil, errors.NewValidationError("invalid ID: %v", id).MasterError()
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	r := s.DB.QueryRow(SQLGetUserByID,
		sql.Named("table_name", s.TableName()),
		sql.Named("id", id),
	)

	// Scan row into user entity
	var u = &shared.User{}
	err := r.Scan(
		&u.ID,
		&u.Name,
		&u.ApiKey,
		&u.LastFeed,
	)
	if err != nil {
		return u, errors.NewMasterError(err, 0)
	}

	return u, nil
}

func (s *UserService) List() ([]*shared.User, *errors.MasterError) {
	s.mx.Lock()
	defer s.mx.Unlock()

	rows, err := s.DB.Query(SQLListUsers,
		sql.Named("table_name", s.TableName()),
	)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	defer rows.Close()

	users := []*shared.User{}
	for rows.Next() {
		u := &shared.User{}
		err := rows.Scan(
			&u.ID,
			&u.Name,
			&u.ApiKey,
			&u.LastFeed,
		)
		if err != nil {
			return nil, errors.NewMasterError(err, 0)
		}
		users = append(users, u)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.NewMasterError(err, 0)
	}

	return users, nil
}

func (s *UserService) Delete(id shared.TelegramID) *errors.MasterError {
	s.mx.Lock()
	defer s.mx.Unlock()

	_, err := s.DB.Exec(SQLDeleteUser,
		sql.Named("table_name", s.TableName()),
		sql.Named("id", id),
	)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

// Service validation
var _ shared.Service[*shared.User, shared.TelegramID] = (*UserService)(nil)
