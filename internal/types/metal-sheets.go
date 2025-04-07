package types

type MetalSheets []MetalSheet

type MetalSheet struct {
	DBKey  DBKey            `json:"db_key"`
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
	Thickness              uint8                  `json:"thickness"`
	LowerStampHeight       uint8                  `json:"lower_stamp_height"`
	LowerStampSheets       uint8                  `json:"lower_stamp_sheets"`
	UpperStampSheets       uint8                  `json:"upper_stamp_sheets"`
	ThicknessSettings      uint8                  `json:"thickness_settings"`
	ThicknessSettingsSacmi ThicknessSettingsSacmi `json:"thickness_settings_sacmi"`
}

type ThicknessSettingsSacmi struct {
	Max     uint8 `json:"max"`
	Current uint8 `json:"current"`
}
