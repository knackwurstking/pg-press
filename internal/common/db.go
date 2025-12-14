package common

import (
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/services/note"
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
	Note  *NoteDB  `json:"note"`
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
		Note: &NoteDB{
			Note: note.NewNoteService(c),
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
	return setupServices([]setupServicesProps{
		{"user", udb.User.Setup},
		{"cookie", udb.Cookie.Setup},
		{"session", udb.Session.Setup},
	}...)
}

// Close shuts down user database services
func (udb *UserDB) Close() {
	closeServices([]closeServicesProps{
		{"user", udb.User.Close},
		{"cookie", udb.Cookie.Close},
		{"session", udb.Session.Close},
	}...)
}

// PressDB holds press-related database services
type PressDB struct {
	Press        shared.Service[*shared.Press, shared.PressNumber]          `json:"press"`
	Cycle        shared.Service[*shared.Cycle, shared.EntityID]             `json:"cycle"`
	Regeneration shared.Service[*shared.PressRegeneration, shared.EntityID] `json:"regeneration"`
}

// Setup initializes press database services
func (pdb *PressDB) Setup() error {
	return setupServices([]setupServicesProps{
		{"press", pdb.Press.Setup},
		{"cycle", pdb.Cycle.Setup},
		{"regeneration", pdb.Regeneration.Setup},
	}...)
}

// Close shuts down press database services
func (pdb *PressDB) Close() {
	closeServices([]closeServicesProps{
		{"press", pdb.Press.Close},
		{"cycle", pdb.Cycle.Close},
		{"regeneration", pdb.Regeneration.Close},
	}...)
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
	return setupServices([]setupServicesProps{
		{"tool", tdb.Tool.Setup},
		{"regeneration", tdb.Regeneration.Setup},
		{"cassette", tdb.Cassette.Setup},
		{"upper_metal_sheet", tdb.UpperMetalSheet.Setup},
		{"lower_metal_sheet", tdb.LowerMetalSheet.Setup},
	}...)
}

// Close shuts down tool database services
func (tdb *ToolDB) Close() {
	closeServices([]closeServicesProps{
		{"tool", tdb.Tool.Close},
		{"regeneration", tdb.Regeneration.Close},
		{"cassette", tdb.Cassette.Close},
		{"upper_metal_sheet", tdb.UpperMetalSheet.Close},
		{"lower_metal_sheet", tdb.LowerMetalSheet.Close},
	}...)
}

// NoteDB holds tool-related database services
type NoteDB struct {
	Note shared.Service[*shared.Note, shared.EntityID] `json:"note"`
}

// Setup initializes tool database services
func (ndb *NoteDB) Setup() error {
	return setupServices([]setupServicesProps{
		{"note", ndb.Note.Setup},
	}...)
}

// Close shuts down tool database services
func (ndb *NoteDB) Close() {
	closeServices([]closeServicesProps{
		{"note", ndb.Note.Close},
	}...)
}

type setupServicesProps struct {
	name    string
	setupFn func() *errors.MasterError
}

func setupServices(services ...setupServicesProps) error {
	errCh := make(chan *errors.MasterError, len(services))

	wg := &sync.WaitGroup{}
	for _, s := range services {
		wg.Go(func() {
			if err := s.setupFn(); err != nil {
				errCh <- errors.NewMasterError(
					fmt.Errorf("%s: %w", s.name, err),
					0,
				)
			}
		})
	}
	wg.Wait()
	close(errCh)

	errs := []string{}
	for err := range errCh {
		errs = append(errs, err.Error())
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to setup user database service:\n%s", strings.Join(errs, "\n -> "))
	}

	return nil
}

type closeServicesProps struct {
	name    string
	closeFn func() *errors.MasterError
}

func closeServices(services ...closeServicesProps) {
	errCh := make(chan error, len(services))

	wg := &sync.WaitGroup{}
	for _, s := range services {
		wg.Go(func() {
			if err := s.closeFn(); err != nil {
				errCh <- fmt.Errorf("%s: %w", s.name, err)
			}
		})
	}
	wg.Wait()
	close(errCh)

	for err := range errCh {
		slog.Error("Failed to close database service(s)", "error", err)
	}
}
