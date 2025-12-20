package shared

import (
	"github.com/knackwurstking/pg-press/internal/errors"
)

type Entity[T any] interface {
	// Validate checks if the entity has valid data
	Validate() *errors.ValidationError

	// Entity clones the entity and returns a copy
	Clone() T

	String() string
}
