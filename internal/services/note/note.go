package note

import (
	"database/sql"
	"sync"

	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"
)

const DBName string = "note"

const (
	SQLCreateNoteTable string = `
		CREATE TABLE IF NOT EXISTS notes (
			id 			INTEGER NOT NULL,
			level 		INTEGER NOT NULL,
			content 	TEXT NOT NULL,
			created_at 	INTEGER NOT NULL,
			linked 		TEXT NOT NULL,

			PRIMARY KEY("id" AUTOINCREMENT)
		);
	`
	SQLCreateNote string = `
		INSERT INTO notes (level, content, created_at, linked)
		VALUES (:level, :content, :created_at, :linked);
	`
	SQLGetNoteByID string = `
		SELECT id, level, content, created_at, linked
		FROM notes
		WHERE id = :id;
	`
	SQLUpdateNote string = `
		UPDATE notes
		SET level 		= :level,
			content 	= :content,
			created_at 	= :created_at,
			linked 		= :linked
		WHERE id = :id;
	`
	SQLDeleteNote string = `
		DELETE FROM notes
		WHERE id = :id;
	`
	SQLListNotes string = `
		SELECT id, level, content, created_at, linked
		FROM notes;
	`
)

type NoteService struct {
	*shared.BaseService

	mx *sync.Mutex `json:"-"`
}

func NewNoteService(c *shared.Config) *NoteService {
	return &NoteService{
		BaseService: &shared.BaseService{
			Config: c,
		},

		mx: &sync.Mutex{},
	}
}

func (s *NoteService) Setup() *errors.MasterError {
	return s.BaseService.Setup(DBName, SQLCreateNoteTable)
}

func (s *NoteService) Create(entity *shared.Note) *errors.MasterError {
	verr := entity.Validate()
	if verr != nil {
		return verr.MasterError()
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	r, err := s.DB().Exec(SQLCreateNote,
		sql.Named("level", entity.Level),
		sql.Named("content", entity.Content),
		sql.Named("created_at", entity.CreatedAt),
		sql.Named("linked", entity.Linked),
	)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	// Store the inserted ID back into the entity
	id, err := r.LastInsertId()
	if err != nil {
		return errors.NewMasterError(err, 0)
	}
	if id <= 0 {
		return errors.NewMasterError(
			errors.NewValidationError("invalid ID returned after insert: %v", id), 0)
	}

	entity.ID = shared.EntityID(id)

	return nil
}

func (s *NoteService) Update(entity *shared.Note) *errors.MasterError {
	verr := entity.Validate()
	if verr != nil {
		return verr.MasterError()
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	_, err := s.DB().Exec(SQLUpdateNote,
		sql.Named("id", entity.ID),
		sql.Named("level", entity.Level),
		sql.Named("content", entity.Content),
		sql.Named("created_at", entity.CreatedAt),
		sql.Named("linked", entity.Linked),
	)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

func (s *NoteService) GetByID(id shared.EntityID) (*shared.Note, *errors.MasterError) {
	if id <= 0 {
		return nil, errors.NewValidationError("invalid ID: %v", id).MasterError()
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	r := s.DB().QueryRow(SQLGetNoteByID,
		sql.Named("id", id),
	)

	// Scan row into note entity
	var n = &shared.Note{}
	err := r.Scan(
		&n.ID,
		&n.Level,
		&n.Content,
		&n.CreatedAt,
		&n.Linked,
	)
	if err != nil {
		return n, errors.NewMasterError(err, 0)
	}

	return n, nil
}

func (s *NoteService) List() ([]*shared.Note, *errors.MasterError) {
	s.mx.Lock()
	defer s.mx.Unlock()

	rows, err := s.DB().Query(SQLListNotes)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	defer rows.Close()

	notes := []*shared.Note{}
	for rows.Next() {
		n := &shared.Note{}
		err := rows.Scan(
			&n.ID,
			&n.Level,
			&n.Content,
			&n.CreatedAt,
			&n.Linked,
		)
		if err != nil {
			return nil, errors.NewMasterError(err, 0)
		}
		notes = append(notes, n)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.NewMasterError(err, 0)
	}

	return notes, nil
}

func (s *NoteService) Delete(id shared.EntityID) *errors.MasterError {
	s.mx.Lock()
	defer s.mx.Unlock()

	_, err := s.DB().Exec(SQLDeleteNote,
		sql.Named("id", id),
	)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

// Service validation
var _ shared.Service[*shared.Note, shared.EntityID] = (*NoteService)(nil)
