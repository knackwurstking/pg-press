package shared

const (
	MachineTypeSACMI MachineType = "SACMI"
	MachineTypeSITI  MachineType = "SITI"
)

type MachineType string

type BaseMetalSheet struct {
	ID         EntityID `json:"id"`
	ToolID     EntityID `json:"tool_id"`     // Currently assigned tool
	TileHeight float64  `json:"tile_height"` // Tile height in mm
	Value      float64  `json:"value"`       // Value
}

type LowerMetalSheet struct {
	BaseMetalSheet
	MarkeHeight int         `json:"marke_height"` // Marke height
	STF         float64     `json:"stf"`          // STF value
	STFMax      float64     `json:"stf_max"`      // STF max value
	Identifier  MachineType `json:"identifier"`   // Machine type identifier ("SACMI" or "SITI")
}

type UpperMetalSheet struct {
	BaseMetalSheet
}

var _ Entity[*UpperMetalSheet] = (*UpperMetalSheet)(nil)
var _ Entity[*LowerMetalSheet] = (*LowerMetalSheet)(nil)
