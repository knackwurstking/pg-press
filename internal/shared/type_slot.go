package shared

type Slot int

const (
	SlotUnknown Slot = iota
	SlotUpper
	SlotLower
	SlotUpperCassette
)

func (s Slot) German() string {
	switch s {
	case SlotUpper:
		return "Oberteil"
	case SlotLower:
		return "Unterteil"
	case SlotUpperCassette:
		return "Kassette"
	default:
		return "Unbekannt"
	}
}
