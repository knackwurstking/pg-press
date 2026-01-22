package models

// PressUtilization represents the utilization of a single press
type PressUtilization struct {
	PressNumber PressNumber `json:"press_number"`
	Tools       []*Tool     `json:"tools"`
	Count       int         `json:"count"`
	Available   bool        `json:"available"`
}
