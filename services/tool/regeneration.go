package tool

import "github.com/knackwurstking/pg-press/services/shared"

type ToolRegenerationService struct{}

var _ shared.Service[*shared.ToolRegeneration, shared.EntityID] = (*ToolRegenerationService)(nil)
