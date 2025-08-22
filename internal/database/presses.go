package database

import "database/sql"

const (
	createPressTableQuery = `
		DROP TABLE IF EXISTS tools;
		CREATE TABLE IF NOT EXISTS tools (
			number INTEGER NOT NULL,
			from DATETIME NOT NULL,
			to DATETIME,
			upper_tool_id INTEGER NOT NULL,
			lower_tool_id INTEGER NOT NULL,
			total_cycles INTEGER NOT NULL,
			partial_cycles INTEGER NOT NULL,
			PRIMARY KEY("number")
		);
		INSERT INTO tools (number, from, to, upper_tool_id, lower_tool_id, total_cycles, partial_cycles)
		VALUES
			(0, '2022-01-01 00:00:00', '2022-02-01 00:00:00', 0, 1, 0, 0),
			(0, '2022-02-01 00:00:00', '2022-03-01 00:00:00', 0, 1, 0, 0),
			(0, '2022-03-01 00:00:00', '2022-04-01 00:00:00', 0, 1, 0, 0),
			(0, '2022-04-01 00:00:00', NULL, 0, 1, 0, 0),
	`
)

type Presses struct {
	db    *sql.DB
	feeds *Feeds
}
