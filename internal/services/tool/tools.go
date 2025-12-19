package tool

const (
	DBName = "tool"
)

const (
	SQLCreateToolTable string = `
		CREATE TABLE IF NOT EXISTS tools (
			id					INTEGER NOT NULL,
			position 			INTEGER NOT NULL,
			width 				INTEGER NOT NULL,
			height 				INTEGER NOT NULL,
			type 				TEXT NOT NULL,
			code 				TEXT NOT NULL,
			cycles_offset 		INTEGER NOT NULL DEFAULT 0,
			cycles 				INTEGER NOT NULL DEFAULT 0,
			is_dead 			INTEGER NOT NULL DEFAULT 0,
			cassette			INTEGER NOT NULL DEFAULT 0,
			min_thickness		REAL NOT NULL DEFAULT 0,
			max_thickness		REAL NOT NULL DEFAULT 0,
			model_type			TEXT NOT NULL, -- e.g.: "tool", "cassette",

			PRIMARY KEY("id" AUTOINCREMENT)
		);
	`
)
