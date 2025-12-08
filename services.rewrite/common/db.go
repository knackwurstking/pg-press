package common

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/knackwurstking/pg-press/services.rewrite/shared"
	"github.com/knackwurstking/pg-press/services.rewrite/user"
	_ "github.com/mattn/go-sqlite3"
)

type UserDB struct {
	User    *user.UserService    `json:"user"`
	Cookie  *user.CookieService  `json:"cookie"`
	Session *user.SessionService `json:"session"`
}

func (udb *UserDB) Setup() {
	wg := &sync.WaitGroup{}
	errCh := make(chan error, 3)

	wg.Go(func() {
		if err := udb.User.Setup(); err != nil {
			errCh <- fmt.Errorf("user: %w", err)
		}
	})

	wg.Go(func() {
		if err := udb.Cookie.Setup(); err != nil {
			errCh <- fmt.Errorf("cookie: %w", err)
		}
	})

	wg.Go(func() {
		if err := udb.Session.Setup(); err != nil {
			errCh <- fmt.Errorf("session: %w", err)
		}
	})

	wg.Wait()
	close(errCh)

	for err := range errCh {
		slog.Error("Failed to setup user database service", "error", err)
	}
}

func (udb *UserDB) Close() {
	wg := &sync.WaitGroup{}
	errCh := make(chan error, 3)

	wg.Go(func() {
		if err := udb.User.Close(); err != nil {
			errCh <- fmt.Errorf("user: %w", err)
		}
	})

	wg.Go(func() {
		if err := udb.Cookie.Close(); err != nil {
			errCh <- fmt.Errorf("cookie: %w", err)
		}
	})

	wg.Go(func() {
		if err := udb.Session.Close(); err != nil {
			errCh <- fmt.Errorf("session: %w", err)
		}
	})

	wg.Wait()
	close(errCh)

	for err := range errCh {
		slog.Error("Failed to close user database service", "error", err)
	}
}

// DB holds and initializes all the services required
type DB struct {
	User *UserDB `json:"user"`
}

func NewDB(c *shared.Config) *DB {
	db := &DB{
		User: &UserDB{
			User:    user.NewUserService(c),
			Cookie:  user.NewCookieService(c),
			Session: user.NewSessionService(c),
		},
	}

	db.Setup()

	return db
}

func (db *DB) Setup() {
	wg := &sync.WaitGroup{}

	wg.Go(func() {
		db.User.Setup()
	})

	wg.Wait()
}

func (db *DB) Close() {
	wg := &sync.WaitGroup{}
	wg.Go(db.User.Close)
	wg.Wait()
}
