package user

import (
	"database/sql"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/services.rewrite/shared"
)

const (
	SQLCreateUserTable string = `
		CREATE TABLE IF NOT EXISTS :table_name (
			telegram_id INTEGER PRIMARY KEY NOT NULL,
			user_name 	TEXT NOT NULL,
			api_key 	TEXT NOT NULL UNIQUE,
			last_feed 	TEXT NOT NULL
		);
	`
)

type UserService[T *shared.User, ID shared.TelegramID] struct {
	setup *shared.Setup `json:"-"`
	db    *sql.DB       `json:"-"`
}

func NewUserService[T *shared.User, ID shared.TelegramID](setup *shared.Setup) *UserService[T, ID] {
	return &UserService[T, ID]{
		setup: setup,
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

// TODO: Implement CRUD methods
func (s *UserService[T, ID]) Create(entity T) *errors.MasterError

func (s *UserService[T, ID]) GetByID(id ID) (T, *errors.MasterError)

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
