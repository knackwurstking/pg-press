package common

import (
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/knackwurstking/pg-press/services/shared"
	"github.com/knackwurstking/pg-press/services/user"

	_ "github.com/mattn/go-sqlite3"
)

type UserDB struct {
	User    *user.UserService    `json:"user"`
	Cookie  *user.CookieService  `json:"cookie"`
	Session *user.SessionService `json:"session"`
}

func (udb *UserDB) Setup() error {
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

	errs := []string{}
	for err := range errCh {
		errs = append(errs, err.Error())
		slog.Error("Failed to setup user database service", "error", err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to setup user database service:\n%s", strings.Join(errs, "\n -> "))
	}

	return nil
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
	return &DB{
		User: &UserDB{
			User:    user.NewUserService(c),
			Cookie:  user.NewCookieService(c),
			Session: user.NewSessionService(c),
		},
	}
}

func (db *DB) Setup() error {
	wg := &sync.WaitGroup{}
	errCh := make(chan error, 1)

	wg.Go(func() {
		errCh <- db.User.Setup()
	})

	wg.Wait()
	close(errCh)

	errs := []string{}
	for err := range errCh {
		if err != nil {
			errs = append(errs, err.Error())
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to setup DB:\n%s", strings.Join(errs, "\n"))
	}

	return nil
}

func (db *DB) Close() {
	wg := &sync.WaitGroup{}
	wg.Go(db.User.Close)
	wg.Wait()
}
