package services

import (
	"database/sql"
	"fmt"

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

func (b *Base) CreateTable(query, tableName string) error {
	if _, err := b.DB.Exec(query); err != nil {
		return fmt.Errorf("create %s table: %v", tableName, err)
	}
	return nil
}

func (b *Base) QueryCount(query string, args ...any) (int, error) {
	var count int
	if err := b.DB.QueryRow(query, args...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (b *Base) HasTable(tableName string) bool {
	rows, err := b.DB.Query(`SELECT name FROM sqlite_master WHERE type='table' AND name='press_cycles';`)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	return rows.Next()
}

func (b *Base) GetSelectError(err error) error {
	return errors.NewDatabaseSelectError(err)
}

func (b *Base) GetInsertError(err error) error {
	return errors.NewDatabaseInsertError(err)
}

func (b *Base) GetUpdateError(err error) error {
	return errors.NewDatabaseUpdateError(err)
}

func (b *Base) GetDeleteError(err error) error {
	return errors.NewDatabaseDeleteError(err)
}
