package services

import (
	"database/sql"
	"fmt"
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

func (b *Base) GetSelectError(err error) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("database select: %v", err)
}

func (b *Base) GetInsertError(err error) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("database insert: %v", err)
}

func (b *Base) GetUpdateError(err error) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("database update: %v", err)
}

func (b *Base) GetDeleteError(err error) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("database delete: %v", err)
}
