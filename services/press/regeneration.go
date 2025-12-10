package press

import "github.com/knackwurstking/pg-press/services/shared"

// PressRegenerationService provides methods to manage press regenerations
//
// - A press regeneration is a record which will reset a press's cycle count back to zero
// - A regeneration means that the press was broken and got renewed, so the cycle count starts fresh, but this does not matter here

type PressRegenerationService struct {
}

var _ shared.Service[*shared.PressRegeneration] = (*PressRegenerationService)(nil)
