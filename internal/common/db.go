package common

import (
	"fmt"
	"strings"
	"sync"

	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/logger"
	"github.com/knackwurstking/pg-press/internal/services/note"
	"github.com/knackwurstking/pg-press/internal/services/press"
	"github.com/knackwurstking/pg-press/internal/services/tool"
	"github.com/knackwurstking/pg-press/internal/services/user"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/ui/ui-templ"

	_ "github.com/mattn/go-sqlite3"
)

var (
	log *ui.Logger
)

func init() {
	log = logger.New("common")
}

// DB holds and initializes all the services required
type DB struct {
	User  *UserDB                                       `json:"user"`
	Press *PressDB                                      `json:"press"`
	Tool  *ToolDB                                       `json:"tool"`
	Notes shared.Service[*shared.Note, shared.EntityID] `json:"note"`
}

// NewDB creates a new database instance with initialized services
func NewDB(c *shared.Config) *DB {
	return &DB{
		User: &UserDB{
			Users:    user.NewUsersService(c),
			Cookies:  user.NewCookiesService(c),
			Sessions: user.NewSessionsService(c),
		},
		Press: &PressDB{
			Presses:       press.NewPressesService(c),
			Cycles:        press.NewPressCyclesService(c),
			Regenerations: press.NewPressRegenerationsService(c),
		},
		Tool: &ToolDB{
			Tools:            tool.NewToolsService(c),
			Regenerations:    tool.NewToolRegenerationsService(c),
			Cassettes:        tool.NewCassettesService(c),
			UpperMetalSheets: tool.NewUpperMetalSheetsService(c),
			LowerMetalSheets: tool.NewLowerMetalSheetsService(c),
		},
		Notes: note.NewNotesService(c),
	}
}

// Setup initializes all database services
func (db *DB) Setup() error {
	wg := &sync.WaitGroup{}
	errCh := make(chan error, 4) // User, Pres, Tool and Note services

	wg.Go(func() {
		errCh <- db.User.Setup()
	})

	wg.Go(func() {
		errCh <- db.Press.Setup()
	})

	wg.Go(func() {
		errCh <- db.Tool.Setup()
	})

	wg.Go(func() {
		errCh <- setupServices([]setupServicesProps{
			{"note", db.Notes.Setup},
		}...)
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
		return fmt.Errorf("%s", strings.Join(errs, "\n"))
	}

	return nil
}

// Close shuts down all database services
func (db *DB) Close() {
	wg := &sync.WaitGroup{}
	wg.Go(db.User.Close)
	wg.Go(db.Press.Close)
	wg.Go(db.Tool.Close)
	wg.Go(func() {
		closeServices([]closeServicesProps{
			{"note", db.Notes.Close},
		}...)
	})
	wg.Wait()
}

// UserDB holds user-related database services
type UserDB struct {
	Users    shared.Service[*shared.User, shared.TelegramID]  `json:"user"`
	Cookies  shared.Service[*shared.Cookie, string]           `json:"cookie"`
	Sessions shared.Service[*shared.Session, shared.EntityID] `json:"session"`
}

// Setup initializes user database services
func (udb *UserDB) Setup() error {
	return setupServices([]setupServicesProps{
		{"user", udb.Users.Setup},
		{"cookie", udb.Cookies.Setup},
		{"session", udb.Sessions.Setup},
	}...)
}

// Close shuts down user database services
func (udb *UserDB) Close() {
	closeServices([]closeServicesProps{
		{"user", udb.Users.Close},
		{"cookie", udb.Cookies.Close},
		{"session", udb.Sessions.Close},
	}...)
}

// PressDB holds press-related database services
type PressDB struct {
	Presses       shared.Service[*shared.Press, shared.PressNumber]          `json:"press"`
	Cycles        shared.Service[*shared.Cycle, shared.EntityID]             `json:"cycle"`
	Regenerations shared.Service[*shared.PressRegeneration, shared.EntityID] `json:"regeneration"`
}

// Setup initializes press database services
func (pdb *PressDB) Setup() error {
	return setupServices([]setupServicesProps{
		{"press", pdb.Presses.Setup},
		{"cycle", pdb.Cycles.Setup},
		{"regeneration", pdb.Regenerations.Setup},
	}...)
}

// Close shuts down press database services
func (pdb *PressDB) Close() {
	closeServices([]closeServicesProps{
		{"press", pdb.Presses.Close},
		{"cycle", pdb.Cycles.Close},
		{"regeneration", pdb.Regenerations.Close},
	}...)
}

// ToolDB holds tool-related database services
type ToolDB struct {
	Tools            shared.Service[*shared.Tool, shared.EntityID]             `json:"tool"`
	Regenerations    shared.Service[*shared.ToolRegeneration, shared.EntityID] `json:"regeneration"`
	Cassettes        shared.Service[*shared.Cassette, shared.EntityID]         `json:"cassette"`
	UpperMetalSheets shared.Service[*shared.UpperMetalSheet, shared.EntityID]  `json:"upper_metal_sheet"`
	LowerMetalSheets shared.Service[*shared.LowerMetalSheet, shared.EntityID]  `json:"lower_metal_sheet"`
}

// Setup initializes tool database services
func (tdb *ToolDB) Setup() error {
	return setupServices([]setupServicesProps{
		{"tool", tdb.Tools.Setup},
		{"regeneration", tdb.Regenerations.Setup},
		{"cassette", tdb.Cassettes.Setup},
		{"upper_metal_sheet", tdb.UpperMetalSheets.Setup},
		{"lower_metal_sheet", tdb.LowerMetalSheets.Setup},
	}...)
}

// Close shuts down tool database services
func (tdb *ToolDB) Close() {
	closeServices([]closeServicesProps{
		{"tool", tdb.Tools.Close},
		{"regeneration", tdb.Regenerations.Close},
		{"cassette", tdb.Cassettes.Close},
		{"upper_metal_sheet", tdb.UpperMetalSheets.Close},
		{"lower_metal_sheet", tdb.LowerMetalSheets.Close},
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
			if merr := s.setupFn(); merr != nil {
				errCh <- merr.Wrap("%s", s.name)
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
		return fmt.Errorf("failed to setup user database service:\n\t%s", strings.Join(errs, "\n\t -> "))
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
		log.Error("Failed to close database service: %v", err)
	}
}
