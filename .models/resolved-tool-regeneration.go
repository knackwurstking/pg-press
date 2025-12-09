package models

type ResolvedToolRegeneration struct {
	*ToolRegeneration
	tool  *Tool
	cycle *Cycle
	user  *User
}

func NewResolvedRegeneration(r *ToolRegeneration, tool *Tool, cycle *Cycle, user *User) *ResolvedToolRegeneration {
	return &ResolvedToolRegeneration{
		ToolRegeneration: r,
		tool:             tool,
		cycle:            cycle,
		user:             user,
	}
}

func (r *ResolvedToolRegeneration) GetTool() *Tool {
	return r.tool
}

func (r *ResolvedToolRegeneration) GetCycle() *Cycle {
	return r.cycle
}

func (r *ResolvedToolRegeneration) GetUser() *User {
	return r.user
}
