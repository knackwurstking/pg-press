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

type UserService struct {
	Data map[shared.TelegramID]*shared.User `json:"data"`

	setup *shared.Setup `json:"-"`
	mx    *sync.Mutex   `json:"-"`
}

func NewUserService(setup *shared.Setup) *UserService {
	return &UserService{
		setup: setup,
		mx:    &sync.Mutex{},
	}
}

func (s *UserService) TableName() string {
	return "users"
}

func (s *UserService) Memory() map[shared.TelegramID]*shared.User {
	if s.Data == nil {
		s.Data = make(map[shared.TelegramID]*shared.User)
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	// Create a deep copy to avoid external modifications
	data := make(map[shared.TelegramID]*shared.User)
	for k, v := range s.Data {
		if v != nil {
			data[k] = v.Clone()
		} else {
			data[k] = nil
		}
	}

	return data
}

func (s *UserService) Setup(setup *shared.Setup) *errors.MasterError {
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

func (s *UserService) Create(entity *shared.User) *errors.MasterError {
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

func (s *UserService) GetByID(id shared.TelegramID) (*shared.User, *errors.MasterError) {
	if id <= 0 {
		return nil, errors.NewValidationError("invalid ID: %v", id).MasterError()
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	r := s.setup.DB.QueryRow(SQLGetUserByID,
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

func (s *UserService) Update(entity *shared.User) *errors.MasterError

func (s *UserService) Delete(id shared.TelegramID) *errors.MasterError

func (s *UserService) List() ([]*shared.User, *errors.MasterError)

func (s *UserService) createSQLTable() *errors.MasterError {
	_, err := s.setup.DB.Exec(SQLCreateUserTable, sql.Named("table_name", s.TableName()))
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

// Service validation
var _ shared.Service[*shared.User, shared.TelegramID] = (*UserService)(nil)
