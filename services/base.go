package services

import (
	"database/sql"

	"github.com/knackwurstking/pg-press/errors"
)

type Base struct {
	Registry *Registry
	DB       *sql.DB
}

func NewBase(r *Registry) *Base {
	return &Base{
		Registry: r,
		DB:       r.DB,
	}
}

func (b *Base) QueryCount(query string, args ...any) (int, *errors.MasterError) {
	var count int

	err := b.DB.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, errors.NewMasterError(err, 0)
	}

	return count, nil
}
