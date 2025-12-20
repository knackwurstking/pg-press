package tool

const (
	SQLCreateMetalSheetTable string = `
		CREATE TABLE IF NOT EXISTS metal_sheets (
			id 				INTEGER NOT NULL,
			tool_id 		INTEGER NOT NULL,
			tile_height 	REAL NOT NULL,
			value 			REAL NOT NULL,
			type 			TEXT NOT NULL,
			marke_height 	INTEGER,
			stf 			REAL,
			stf_max 		REAL,
			identifier 		TEXT,

			PRIMARY KEY("id" AUTOINCREMENT),
			FOREIGN KEY(tool_id) REFERENCES tools(id) ON DELETE CASCADE
		);
	`
)
