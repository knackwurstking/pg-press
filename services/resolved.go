package services

import (
	"github.com/knackwurstking/pgpress/models"
)

func ResolveRegeneration(registry *Registry, regeneration *models.Regeneration) (*models.ResolvedRegeneration, error) {
	tool, err := registry.Tools.Get(regeneration.ToolID)
	if err != nil {
		return nil, err
	}

	cycle, err := registry.PressCycles.Get(regeneration.CycleID)
	if err != nil {
		return nil, err
	}

	var user *models.User
	if regeneration.PerformedBy != nil {
		var err error
		user, err = registry.Users.Get(*regeneration.PerformedBy)
		if err != nil {
			return nil, err
		}
	}

	return models.NewResolvedRegeneration(regeneration, tool, cycle, user), nil
}
