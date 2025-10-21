package services

import (
	"database/sql"
	"fmt"

	"github.com/knackwurstking/pgpress/logger"
)

type Base struct {
	Registry *Registry
	DB       *sql.DB
	Log      *logger.Logger
}

func NewBase(r *Registry, l *logger.Logger) *Base {
	return &Base{
		Registry: r,
		DB:       r.DB,
		Log:      l,
	}
}

func (b *Base) CreateTable(query string, tableName string) error {
	if _, err := b.DB.Exec(query); err != nil {
		return fmt.Errorf("failed to create %s table: %v", tableName, err)
	}

	return nil
}

func (b *Base) GetSelectError(err error) error {
	return fmt.Errorf("database select failed: %v", err)
}

func (b *Base) GetInsertError(err error) error {
	return fmt.Errorf("database insert failed: %v", err)
}

func (b *Base) GetUpdateError(err error) error {
	return fmt.Errorf("database update failed: %v", err)
}

func (b *Base) GetDeleteError(err error) error {
	return fmt.Errorf("database delete failed: %v", err)
}
