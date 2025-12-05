package services

import (
	"encoding/json"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
)

func ResolveToolRegeneration(registry *Registry, regeneration *models.ToolRegeneration) (*models.ResolvedToolRegeneration, error) {
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
		var dberr *errors.DBError
		user, dberr = registry.Users.Get(*regeneration.PerformedBy)
		if dberr != nil {
			return nil, dberr
		}
	}

	return models.NewResolvedRegeneration(regeneration, tool, cycle, user), nil
}

func ResolveTool(registry *Registry, tool *models.Tool) (*models.ResolvedTool, error) {
	return resolveTool(registry, tool, false)
}

func resolveTool(registry *Registry, tool *models.Tool, skipResolveBindingTool bool) (*models.ResolvedTool, error) {
	var bindingTool *models.ResolvedTool
	if tool.IsBound() && !skipResolveBindingTool {
		bt, dberr := registry.Tools.Get(*tool.Binding)
		if dberr != nil {
			return nil, errors.Wrap(dberr, "get binding tool %d for %d", tool.Binding, tool.ID)
		}

		var err error
		bindingTool, err = resolveTool(registry, bt, true)
		if err != nil {
			return nil, errors.Wrap(err, "resolve binding tool %d for %d", tool.Binding, tool.ID)
		}
	}

	notes, dberr := registry.Notes.GetByTool(tool.ID)
	if dberr != nil && dberr.Typ != errors.DBTypeNotFound {
		return nil, errors.Wrap(dberr, "get notes for tool %d", tool.ID)
	}

	regenerations, dberr := registry.ToolRegenerations.GetRegenerationHistory(tool.ID)
	if dberr != nil && dberr.Typ != errors.DBTypeNotFound {
		return nil, errors.Wrap(dberr, "get regeneration for tool %d", tool.ID)
	}

	rt := models.NewResolvedTool(tool, bindingTool, notes, regenerations)
	if bindingTool != nil {
		bindingTool.SetBindingTool(rt)
	}
	return rt, nil
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
