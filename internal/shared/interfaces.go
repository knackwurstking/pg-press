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

var (
	_ Entity[*Cycle]             = (*Cycle)(nil)
	_ Entity[*UpperMetalSheet]   = (*UpperMetalSheet)(nil)
	_ Entity[*LowerMetalSheet]   = (*LowerMetalSheet)(nil)
	_ Entity[*Note]              = (*Note)(nil)
	_ Entity[*PressRegeneration] = (*PressRegeneration)(nil)
	_ Entity[*Press]             = (*Press)(nil)
	_ Entity[*ToolRegeneration]  = (*ToolRegeneration)(nil)
	_ Entity[*Tool]              = (*Tool)(nil)
	_ Entity[*Cookie]            = (*Cookie)(nil)
	_ Entity[*Session]           = (*Session)(nil)
	_ Entity[*User]              = (*User)(nil)
	_ Entity[*TroubleReport]     = (*TroubleReport)(nil)
)

// Ensure Translate implementations

var (
	_ Translate = (*Tool)(nil)
	_ Translate = Slot(0)
)
