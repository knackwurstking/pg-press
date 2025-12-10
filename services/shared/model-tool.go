package shared

// Tool represents a tool used in a press machine,
// there are upper and lower tools. Each tool can have its own regeneration history.
// And the upper tool type has an optional cassette slot.
type Tool struct {
	ID EntityID `json:"id"`
	// Type represents the tool type, e.g., "MASS", "FC", "GTC", etc.
	Type string `json:"type"`
	// Code is the unique tool code/identifier, "G01", "12345", etc.
	Code string `json:"code"`
	// Status represents the current state of the tool
	Status string `json:"status"`
	// Regenerating indicates if the tool is currently being regenerated
	Regenerating bool `json:"regenerating"`
	// IsDead indicates if the tool is dead/destroyed
	IsDead bool `json:"is_dead"`
}

// TODO: Add missing validate and clone methods to fit the Entity interface

var _ Entity[*Tool] = (*Tool)(nil)
