package services

import (
	"encoding/json"

	"github.com/knackwurstking/pg-press/models"
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

func ResolveTool(registry *Registry, tool *models.Tool) (*models.ResolvedTool, error) {
	var bindingTool *models.Tool
	if tool.IsBound() {
		var err error
		bindingTool, err = registry.Tools.Get(*tool.Binding)
		if err != nil {
			return nil, err
		}
	}

	notes, err := registry.Notes.GetByTool(tool.ID)
	if err != nil {
		return nil, err
	}

	return models.NewResolvedTool(tool, bindingTool, notes), nil
}

func ResolveModification[T any](registry *Registry, modification *models.Modification[any]) (*models.ResolvedModification[T], error) {
	user, err := registry.Users.Get(modification.UserID)
	if err != nil {
		return nil, err
	}

	var data T
	if err := json.Unmarshal(modification.Data, &data); err != nil {
		return nil, err
	}

	m := models.NewModification(data, user.TelegramID)
	m.ID = modification.ID
	m.CreatedAt = modification.CreatedAt

	return models.NewResolvedModification(m, user), nil
}
