package models

type ResolvedTool struct {
	*Tool
	bindingTool   *ResolvedTool
	notes         []*Note
	regenerations []*ToolRegeneration
}

func NewResolvedTool(t *Tool, b *ResolvedTool, n []*Note, r []*ToolRegeneration) *ResolvedTool {
	rt := &ResolvedTool{
		Tool:          t,
		bindingTool:   b,
		notes:         n,
		regenerations: r,
	}

	return rt
}

func (rt *ResolvedTool) GetBindingTool() *ResolvedTool {
	return rt.bindingTool
}

func (rt *ResolvedTool) SetBindingTool(t *ResolvedTool) {
	rt.bindingTool = t
	rt.Binding = &rt.bindingTool.ID
}

func (rt *ResolvedTool) GetNotes() []*Note {
	return rt.notes
}

func (rt *ResolvedTool) GetRegenerations() []*ToolRegeneration {
	return rt.regenerations
}
