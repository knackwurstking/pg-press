package types

type MetalSheets []MetalSheet

type MetalSheet struct {
	DBKey DBKey `json:"db_key"`

	Format string           `json:"format"`
	ToolID string           `json:"tool_id"`
	Data   []MetalSheetData `json:"data"`
}

type MetalSheetData struct {
	Press  Press                 `json:"press"`
	Filter MetalSheetDataFilter  `json:"filter"`
	Data   []MetalSheetDataEntry `json:"data"`
}

type MetalSheetDataFilter []uint8

type MetalSheetDataEntry struct {
	Thickness float32 `json:"thickness"`

	LowerStampHeight int8 `json:"lower_stamp_height"`

	LowerStampSheets float32 `json:"lower_stamp_sheets"`
	UpperStampSheets float32 `json:"upper_stamp_sheets"`

	ThicknessSettings      float32                `json:"thickness_settings"`
	ThicknessSettingsSacmi ThicknessSettingsSacmi `json:"thickness_settings_sacmi"`
}

type ThicknessSettingsSacmi struct {
	Max     float32 `json:"max"`
	Current float32 `json:"current"`
}
