package models

type ResolvedTool struct {
	*Tool
	bindingTool   *Tool
	notes         []*Note
	regenerations []*Regeneration
}

func NewResolvedTool(t *Tool, b *Tool, n []*Note, r []*Regeneration) *ResolvedTool {
	return &ResolvedTool{
		Tool:          t,
		bindingTool:   b,
		notes:         n,
		regenerations: r,
	}
}

func (rt *ResolvedTool) GetBindingTool() *Tool {
	return rt.bindingTool
}

func (rt *ResolvedTool) GetNotes() []*Note {
	return rt.notes
}

func (rt *ResolvedTool) GetRegenerations() []*Regeneration {
	return rt.regenerations
}
