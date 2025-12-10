package shared

// Tool represents a tool used in a press machine,
// there are upper and lower tools. Each tool can have its own regeneration history.
// And the upper tool type has an optional cassette slot.
type Tool struct {
	ID EntityID `json:"id"`
	// TODO: Continue here
}
