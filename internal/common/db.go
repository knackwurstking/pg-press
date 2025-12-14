package common

import (
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/knackwurstking/pg-press/internal/services/press"
	"github.com/knackwurstking/pg-press/internal/services/tool"
	"github.com/knackwurstking/pg-press/internal/services/user"
	"github.com/knackwurstking/pg-press/internal/shared"

	_ "github.com/mattn/go-sqlite3"
)

// DB holds and initializes all the services required
type DB struct {
	User  *UserDB  `json:"user"`
	Press *PressDB `json:"press"`
	Tool  *ToolDB  `json:"tool"`
}

// NewDB creates a new database instance with initialized services
func NewDB(c *shared.Config) *DB {
	return &DB{
		User: &UserDB{
			User:    user.NewUserService(c),
			Cookie:  user.NewCookieService(c),
			Session: user.NewSessionService(c),
		},
		Press: &PressDB{
			Press:        press.NewPressService(c),
			Cycle:        press.NewCycleService(c),
			Regeneration: press.NewPressRegenerationService(c),
		},
		Tool: &ToolDB{
			Tool:            tool.NewToolService(c),
			Regeneration:    tool.NewToolRegenerationService(c),
			Cassette:        tool.NewCassetteService(c),
			UpperMetalSheet: tool.NewUpperMetalSheetService(c),
			LowerMetalSheet: tool.NewLowerMetalSheetService(c),
		},
	}
}

// Setup initializes all database services
func (db *DB) Setup() error {
	wg := &sync.WaitGroup{}
	errCh := make(chan error, 3) // User, Press and Tool services

	wg.Go(func() {
		errCh <- db.User.Setup()
	})

	wg.Go(func() {
		errCh <- db.Press.Setup()
	})

	wg.Go(func() {
		errCh <- db.Tool.Setup()
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

// Close shuts down all database services
func (db *DB) Close() {
	wg := &sync.WaitGroup{}
	wg.Go(db.User.Close)
	wg.Go(db.Press.Close)
	wg.Go(db.Tool.Close)
	wg.Wait()
}

// UserDB holds user-related database services
type UserDB struct {
	User    shared.Service[*shared.User, shared.TelegramID]  `json:"user"`
	Cookie  shared.Service[*shared.Cookie, string]           `json:"cookie"`
	Session shared.Service[*shared.Session, shared.EntityID] `json:"session"`
}

// Setup initializes user database services
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

// Close shuts down user database services
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

// PressDB holds press-related database services
type PressDB struct {
	Press        shared.Service[*shared.Press, shared.PressNumber]          `json:"press"`
	Cycle        shared.Service[*shared.Cycle, shared.EntityID]             `json:"cycle"`
	Regeneration shared.Service[*shared.PressRegeneration, shared.EntityID] `json:"regeneration"`
}

// Setup initializes press database services
func (pdb *PressDB) Setup() error {
	wg := &sync.WaitGroup{}
	errCh := make(chan error, 3)

	wg.Go(func() {
		if err := pdb.Press.Setup(); err != nil {
			errCh <- fmt.Errorf("press: %w", err)
		}
	})

	wg.Go(func() {
		if err := pdb.Cycle.Setup(); err != nil {
			errCh <- fmt.Errorf("cycle: %w", err)
		}
	})

	wg.Go(func() {
		if err := pdb.Regeneration.Setup(); err != nil {
			errCh <- fmt.Errorf("regeneration: %w", err)
		}
	})

	wg.Wait()
	close(errCh)

	errs := []string{}
	for err := range errCh {
		errs = append(errs, err.Error())
		slog.Error("Failed to setup press database service", "error", err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to setup press database service:\n%s", strings.Join(errs, "\n -> "))
	}

	return nil
}

// Close shuts down press database services
func (pdb *PressDB) Close() {
	wg := &sync.WaitGroup{}
	errCh := make(chan error, 3)

	wg.Go(func() {
		if err := pdb.Press.Close(); err != nil {
			errCh <- fmt.Errorf("press: %w", err)
		}
	})

	wg.Go(func() {
		if err := pdb.Cycle.Close(); err != nil {
			errCh <- fmt.Errorf("cycle: %w", err)
		}
	})

	wg.Go(func() {
		if err := pdb.Regeneration.Close(); err != nil {
			errCh <- fmt.Errorf("regeneration: %w", err)
		}
	})

	wg.Wait()
	close(errCh)

	for err := range errCh {
		slog.Error("Failed to close press database service", "error", err)
	}
}

// ToolDB holds tool-related database services
type ToolDB struct {
	Tool            shared.Service[*shared.Tool, shared.EntityID]             `json:"tool"`
	Regeneration    shared.Service[*shared.ToolRegeneration, shared.EntityID] `json:"regeneration"`
	Cassette        shared.Service[*shared.Cassette, shared.EntityID]         `json:"cassette"`
	UpperMetalSheet shared.Service[*shared.UpperMetalSheet, shared.EntityID]  `json:"upper_metal_sheet"`
	LowerMetalSheet shared.Service[*shared.LowerMetalSheet, shared.EntityID]  `json:"lower_metal_sheet"`
}

// Setup initializes tool database services
func (tdb *ToolDB) Setup() error {
	wg := &sync.WaitGroup{}
	errCh := make(chan error, 4)

	wg.Go(func() {
		if err := tdb.Tool.Setup(); err != nil {
			errCh <- fmt.Errorf("tool: %w", err)
		}
	})

	wg.Go(func() {
		if err := tdb.Regeneration.Setup(); err != nil {
			errCh <- fmt.Errorf("regeneration: %w", err)
		}
	})

	wg.Go(func() {
		if err := tdb.Cassette.Setup(); err != nil {
			errCh <- fmt.Errorf("cassette: %w", err)
		}
	})

	wg.Go(func() {
		if err := tdb.UpperMetalSheet.Setup(); err != nil {
			errCh <- fmt.Errorf("upper_metal_sheet: %w", err)
		}
	})

	wg.Go(func() {
		if err := tdb.LowerMetalSheet.Setup(); err != nil {
			errCh <- fmt.Errorf("lower_metal_sheet: %w", err)
		}
	})

	wg.Wait()
	close(errCh)

	errs := []string{}
	for err := range errCh {
		errs = append(errs, err.Error())
		slog.Error("Failed to setup tool database service", "error", err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to setup tool database service:\n%s", strings.Join(errs, "\n -> "))
	}

	return nil
}

// Close shuts down tool database services
func (tdb *ToolDB) Close() {
	wg := &sync.WaitGroup{}
	errCh := make(chan error, 4)

	wg.Go(func() {
		if err := tdb.Tool.Close(); err != nil {
			errCh <- fmt.Errorf("tool: %w", err)
		}
	})

	wg.Go(func() {
		if err := tdb.Regeneration.Close(); err != nil {
			errCh <- fmt.Errorf("regeneration: %w", err)
		}
	})

	wg.Go(func() {
		if err := tdb.Cassette.Close(); err != nil {
			errCh <- fmt.Errorf("cassette: %w", err)
		}
	})

	wg.Go(func() {
		if err := tdb.UpperMetalSheet.Close(); err != nil {
			errCh <- fmt.Errorf("upper_metal_sheet: %w", err)
		}
	})

	wg.Go(func() {
		if err := tdb.LowerMetalSheet.Close(); err != nil {
			errCh <- fmt.Errorf("lower_metal_sheet: %w", err)
		}
	})

	wg.Wait()
	close(errCh)

	for err := range errCh {
		slog.Error("Failed to close tool database service", "error", err)
	}
}
