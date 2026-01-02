package shared

// Constants for press types
const (
	MachineTypeSACMI MachineType = "SACMI"
	MachineTypeSITI  MachineType = "SITI"
)

type MachineType string

func (mt MachineType) String() string {
	return string(mt)
}
