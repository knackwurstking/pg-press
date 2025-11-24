package models

type ResolvedTool struct {
	*Tool
	bindingTool   *ResolvedTool
	notes         []*Note
	regenerations []*Regeneration
}

func NewResolvedTool(t *Tool, b *ResolvedTool, n []*Note, r []*Regeneration) *ResolvedTool {
	rt := &ResolvedTool{
		Tool:          t,
		bindingTool:   b,
		notes:         n,
		regenerations: r,
	}

	rt.bindingTool = rt

	return rt
}

func (rt *ResolvedTool) GetBindingTool() *ResolvedTool {
	return rt.bindingTool
}

func (rt *ResolvedTool) SetBindingTool(t *ResolvedTool) {
	rt.Binding = &t.ID
	rt.bindingTool = t
}

func (rt *ResolvedTool) GetNotes() []*Note {
	return rt.notes
}

func (rt *ResolvedTool) GetRegenerations() []*Regeneration {
	return rt.regenerations
}
