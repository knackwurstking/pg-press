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

func (b *Base) QueryCount(query string, args ...any) (int, *errors.DBError) {
	var count int
	if err := b.DB.QueryRow(query, args...).Scan(&count); err != nil {
		return 0, errors.NewDBError(err, errors.DBTypeCount)
	}
	return count, nil
}
