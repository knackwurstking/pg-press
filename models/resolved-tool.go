package models

type ResolvedTool struct {
	*Tool
	bindingTool   *Tool
	notes         []*Note
	regenerations []*Regeneration
}

func NewResolvedTool(tool *Tool, bindingTool *Tool, notes []*Note) *ResolvedTool {
	return &ResolvedTool{
		Tool:        tool,
		bindingTool: bindingTool,
		notes:       notes,
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
