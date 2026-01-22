package models

type ResolvedRegeneration struct {
	*Regeneration
	tool  *Tool
	cycle *Cycle
	user  *User
}

func NewResolvedRegeneration(r *Regeneration, tool *Tool, cycle *Cycle, user *User) *ResolvedRegeneration {
	return &ResolvedRegeneration{
		Regeneration: r,
		tool:         tool,
		cycle:        cycle,
		user:         user,
	}
}

func (rr *ResolvedRegeneration) GetTool() *Tool {
	return rr.tool
}

func (rr *ResolvedRegeneration) GetCycle() *Cycle {
	return rr.cycle
}

func (rr *ResolvedRegeneration) GetUser() *User {
	return rr.user
}
