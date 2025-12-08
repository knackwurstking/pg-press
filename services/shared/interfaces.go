package shared

import (
	"database/sql"

	"github.com/knackwurstking/pg-press/errors"
)

type Setup struct {
	// Contains configuration parameters for the service(s)
	EnableSQL bool `json:"enable_sql"`

	// db is the database connection instance if SQL is enabled
	db *sql.DB `json:"-"`
}

type Service[T any, ID comparable] interface {
	// TableName returns the (SQL) table name in use
	TableName() string

	// Setup initializes the service with the provided setup configuration
	Setup(setup *Setup) *errors.MasterError

	// Create adds a new entity to the repository
	Create(entity *T) *errors.MasterError

	// GetByID retrieves an entity by its ID
	GetByID(id ID) (*T, *errors.MasterError)

	// Update modifies an existing entity in the repository
	Update(entity *T) *errors.MasterError

	// Delete removes an entity from the repository by its ID
	Delete(id ID) *errors.MasterError

	// List retrieves all entities from the repository
	List() ([]*T, *errors.MasterError)
}
