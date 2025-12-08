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
			user_name 	TEXT NOT NULL,
			api_key 	TEXT NOT NULL UNIQUE,
			last_feed 	TEXT NOT NULL
		);
	`
	SQLCreateUser string = `
		INSERT INTO :table_name (user_name, api_key, last_feed) 
		VALUES (:user_name, :api_key, :last_feed);
	`
	SQLGetUserByID string = `
		SELECT id, user_name, api_key, last_feed 
		FROM :table_name
		WHERE id = :id;
	`
)

type UserService[T *shared.User, ID shared.TelegramID] struct {
	setup *shared.Setup `json:"-"`
	mx    *sync.Mutex   `json:"-"`
}

func NewUserService[T *shared.User, ID shared.TelegramID](setup *shared.Setup) *UserService[T, ID] {
	return &UserService[T, ID]{
		setup: setup,
		mx:    &sync.Mutex{},
	}
}

func (s *UserService[T, ID]) TableName() string {
	return "users"
}

func (s *UserService[T, ID]) Setup(setup *shared.Setup) *errors.MasterError {
	if s.setup != nil {
		s.setup.Close()
	}
	s.setup = setup

	merr := s.setup.Open()
	if merr != nil {
		return merr
	}

	return s.createSQLTable()
}

func (s *UserService[T, ID]) Create(entity *shared.User) *errors.MasterError {
	verr := entity.Validate()
	if verr != nil {
		return verr.MasterError()
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	r, err := s.setup.DB.Exec(SQLCreateUser,
		sql.Named("table_name", s.TableName()),
		sql.Named("user_name", entity.Name),
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

func (s *UserService[T, ID]) GetByID(id shared.TelegramID) (*shared.User, *errors.MasterError) {
	if id <= 0 {
		return nil, errors.NewValidationError("invalid ID: %v", id).MasterError()
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	r := s.setup.DB.QueryRow(SQLGetUserByID,
		sql.Named("table_name", s.TableName()),
		sql.Named("id", id),
	)

	// Scan row into T
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

func (s *UserService[T, ID]) Update(entity T) *errors.MasterError

func (s *UserService[T, ID]) Delete(id ID) *errors.MasterError

func (s *UserService[T, ID]) List() ([]T, *errors.MasterError)

func (s *UserService[T, ID]) createSQLTable() *errors.MasterError {
	_, err := s.setup.DB.Exec(SQLCreateUserTable, sql.Named("table_name", s.TableName()))
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

// Service validation
var _ shared.Service[*shared.User, shared.TelegramID] = (*UserService[*shared.User, shared.TelegramID])(nil)
