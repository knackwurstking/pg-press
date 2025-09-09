package tool

// PressUtilization represents the utilization of a single press
type PressUtilization struct {
	PressNumber PressNumber `json:"press_number"`
	Tools       []*Tool     `json:"tools"`
	Count       int         `json:"count"`
	Available   bool        `json:"available"`
}

// PressUtilizationMap represents the complete utilization across all presses
type PressUtilizationMap struct {
	ActiveTools   map[PressNumber][]*Tool `json:"active_tools"`
	InactiveTools []*Tool                 `json:"inactive_tools"`
	Utilization   []PressUtilization      `json:"utilization"`
	TotalActive   int                     `json:"total_active"`
	TotalInactive int                     `json:"total_inactive"`
}
