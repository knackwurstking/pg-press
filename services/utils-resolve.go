package services

import (
	"encoding/json"
	"net/http"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
)

func ResolveToolRegeneration(
	registry *Registry, regeneration *models.ToolRegeneration,
) (*models.ResolvedToolRegeneration, *errors.MasterError) {

	tool, merr := registry.Tools.Get(regeneration.ToolID)
	if merr != nil {
		merr.Code = http.StatusInternalServerError
		return nil, merr
	}

	cycle, merr := registry.PressCycles.Get(regeneration.CycleID)
	if merr != nil {
		merr.Code = http.StatusInternalServerError
		return nil, merr
	}

	var user *models.User
	if regeneration.PerformedBy != nil {
		var merr *errors.MasterError
		user, merr = registry.Users.Get(*regeneration.PerformedBy)
		if merr != nil {
			merr.Code = http.StatusInternalServerError
			return nil, merr
		}
	}

	return models.NewResolvedRegeneration(regeneration, tool, cycle, user), nil
}

func ResolveTool(registry *Registry, tool *models.Tool) (
	*models.ResolvedTool, *errors.MasterError,
) {

	return resolveTool(registry, tool, false)
}

func resolveTool(
	registry *Registry, tool *models.Tool, skipResolveBindingTool bool,
) (*models.ResolvedTool, *errors.MasterError) {

	var bindingTool *models.ResolvedTool
	if tool.IsBound() && !skipResolveBindingTool {
		bt, merr := registry.Tools.Get(*tool.Binding)
		if merr != nil {
			merr.Code = http.StatusInternalServerError
			return nil, merr
		}

		bindingTool, merr = resolveTool(registry, bt, true)
		if merr != nil {
			merr.Code = http.StatusInternalServerError
			return nil, merr
		}
	}

	notes, merr := registry.Notes.GetByTool(tool.ID)
	if merr != nil && merr.Code != http.StatusNotFound {
		merr.Code = http.StatusInternalServerError
		return nil, merr
	}

	regenerations, merr := registry.ToolRegenerations.GetRegenerationHistory(tool.ID)
	if merr != nil && merr.Code != http.StatusNotFound {
		merr.Code = http.StatusInternalServerError
		return nil, merr
	}

	rt := models.NewResolvedTool(tool, bindingTool, notes, regenerations)
	if bindingTool != nil {
		bindingTool.SetBindingTool(rt)
	}
	return rt, nil
}

func ResolveModification[T any](
	registry *Registry, modification *models.Modification[any],
) (*models.ResolvedModification[T], *errors.MasterError) {

	user, merr := registry.Users.Get(modification.UserID)
	if merr != nil {
		merr.Code = http.StatusInternalServerError
		return nil, merr
	}

	var data T
	if err := json.Unmarshal(modification.Data, &data); err != nil {
		return nil, errors.NewMasterError(err, http.StatusInternalServerError)
	}

	m := models.NewModification(data, user.TelegramID)
	m.ID = modification.ID
	m.CreatedAt = modification.CreatedAt

	return models.NewResolvedModification(m, user), nil
}
