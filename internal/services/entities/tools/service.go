package tools

import (
	"database/sql"

	"github.com/knackwurstking/pgpress/internal/services/base"
	"github.com/knackwurstking/pgpress/pkg/models"
)

// NotesService defines the interface for notes service methods used by Tools
type NotesService interface {
	Add(note *models.Note) (int64, error)
	Delete(id int64, user *models.User) error
	GetByTool(toolID int64) ([]*models.Note, error)
}

type Service struct {
	*base.BaseService
	notes NotesService
}

func NewService(db *sql.DB, notes NotesService) *Service {
	baseService := base.NewBaseService(db, "Tools")

	t := &Service{
		BaseService: baseService,
		notes:       notes,
	}

	if err := t.CreateTable(); err != nil {
		panic(err)
	}

	return t
}

func (t *Service) CreateTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS tools (
			id INTEGER NOT NULL,
			position TEXT NOT NULL,
			format BLOB NOT NULL,
			type TEXT NOT NULL,
			code TEXT NOT NULL,
			regenerating INTEGER NOT NULL DEFAULT 0,
			is_dead INTEGER NOT NULL DEFAULT 0,
			press INTEGER,
			binding INTEGER,
			PRIMARY KEY("id" AUTOINCREMENT)
		);
	`
	return t.BaseService.CreateTable(query, "tools")
}
