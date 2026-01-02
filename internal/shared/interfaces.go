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

type Translate interface {
	German() string
}

// Ensure Entity implementations

var _ Entity[*Cycle] = (*Cycle)(nil)
var _ Entity[*UpperMetalSheet] = (*UpperMetalSheet)(nil)
var _ Entity[*LowerMetalSheet] = (*LowerMetalSheet)(nil)
var _ Entity[*Note] = (*Note)(nil)
var _ Entity[*PressRegeneration] = (*PressRegeneration)(nil)
var _ Entity[*Press] = (*Press)(nil)
var _ Entity[*ToolRegeneration] = (*ToolRegeneration)(nil)
var _ Entity[*Tool] = (*Tool)(nil)
var _ Entity[*Cookie] = (*Cookie)(nil)
var _ Entity[*Session] = (*Session)(nil)
var _ Entity[*User] = (*User)(nil)

// Ensure Translate implementations

var _ Translate = (*Tool)(nil)
var _ Translate = Slot(0)
