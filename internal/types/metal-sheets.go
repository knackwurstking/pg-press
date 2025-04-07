package types

type MetalSheets []MetalSheet

type MetalSheet struct {
	DBKey  DBKey                 `json:"db_key"`
	Format string                `json:"format"`
	ToolID string                `json:"tool_id"`
	Press  MetalSheetPress       `json:"press"`
	Data   []MetalSheetDataEntry `json:"data"`
}

type MetalSheetPress struct {
	DBKey DBKey `json:"db_key"`
	Nr    Press `json:"nr"`
}

type MetalSheetDataEntry struct {
	DBKey                  DBKey                  `json:"db_key"`
	Thickness              float32                `json:"thickness"`
	LowerStampHeight       int8                   `json:"lower_stamp_height"`
	LowerStampSheets       float32                `json:"lower_stamp_sheets"`
	UpperStampSheets       float32                `json:"upper_stamp_sheets"`
	ThicknessSettings      float32                `json:"thickness_settings"`
	ThicknessSettingsSacmi ThicknessSettingsSacmi `json:"thickness_settings_sacmi"`
}

type ThicknessSettingsSacmi struct {
	Max     float32 `json:"max"`
	Current float32 `json:"current"`
}
