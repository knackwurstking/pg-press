package press

import "github.com/knackwurstking/pg-press/services/shared"

type PressRegenerationService struct {
}

var _ shared.Service[*shared.PressRegeneration] = (*PressRegenerationService)(nil)
