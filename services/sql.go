package services

const (
	SQLCreateUserTables = `

	-- Create table for cookies

	CREATE TABLE IF NOT EXISTS cookies (
		user_agent TEXT NOT NULL,
		value TEXT NOT NULL,
		api_key TEXT NOT NULL,
		last_login INTEGER NOT NULL,
		PRIMARY KEY("value")
	);

	-- Create table for users

	CREATE TABLE IF NOT EXISTS users (
		telegram_id INTEGER NOT NULL,
		user_name TEXT NOT NULL,
		api_key TEXT NOT NULL UNIQUE,
		last_feed TEXT NOT NULL,
		PRIMARY KEY("telegram_id")
	);

	-- TODO: Create indexes here

`

	SQLCreateDataTables = `

	-- Create table for modifications

	CREATE TABLE IF NOT EXISTS modifications (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		entity_type TEXT NOT NULL,
		entity_id INTEGER NOT NULL,
		data BLOB NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	);

	-- Create table for feeds 

	CREATE TABLE IF NOT EXISTS feeds (
		id INTEGER NOT NULL,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		user_id INTEGER NOT NULL,
		created_at INTEGER NOT NULL,
		PRIMARY KEY("id" AUTOINCREMENT)
	);

	-- Create table for attachments

	CREATE TABLE IF NOT EXISTS attachments (
		id INTEGER NOT NULL,
		mime_type TEXT NOT NULL,
		data BLOB NOT NULL,
		PRIMARY KEY("id" AUTOINCREMENT)
	);

	-- Create table for metal sheets

	CREATE TABLE IF NOT EXISTS metal_sheets (
		id INTEGER NOT NULL,
		tile_height REAL NOT NULL,
		value REAL NOT NULL,
		marke_height INTEGER NOT NULL,
		stf REAL NOT NULL,
		stf_max REAL NOT NULL,
		identifier TEXT NOT NULL,
		tool_id INTEGER NOT NULL,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY("id" AUTOINCREMENT),
	);

	-- Create table for notes

	CREATE TABLE IF NOT EXISTS notes (
		id INTEGER NOT NULL,
		level INTEGER NOT NULL,
		content TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		linked TEXT,
		PRIMARY KEY("id" AUTOINCREMENT)
	);

	-- Create table for press cycles

	CREATE TABLE IF NOT EXISTS press_cycles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		press_number INTEGER NOT NULL,
		tool_id INTEGER NOT NULL,
		tool_position TEXT NOT NULL,
		total_cycles INTEGER NOT NULL DEFAULT 0,
		date DATETIME NOT NULL,
		performed_by INTEGER NOT NULL
	);

	-- Create table for press regenerations

	CREATE TABLE IF NOT EXISTS press_regenerations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		press_number INTEGER NOT NULL,
		started_at DATETIME NOT NULL,
		completed_at DATETIME,
		reason TEXT
	);

	-- Create table for tool regenerations

	CREATE TABLE IF NOT EXISTS tool_regenerations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		tool_id INTEGER NOT NULL,
		cycle_id INTEGER NOT NULL,
		reason TEXT,
		performed_by INTEGER NOT NULL
	);

	-- Create table for tools

	CREATE TABLE IF NOT EXISTS tools (
		id INTEGER NOT NULL,
		position TEXT NOT NULL,
		format BLOB NOT NULL,
		type TEXT NOT NULL,
		code TEXT NOT NULL,
		regenerating INTEGER NOT NULL DEFAULT 0,
		is_dead INTEGER NOT NULL DEFAULT 0,
		press INTEGER,
		binding INTEGER,
		PRIMARY KEY("id" AUTOINCREMENT)
	);

	-- Create table for trouble reports

	CREATE TABLE IF NOT EXISTS trouble_reports (
		id INTEGER NOT NULL,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		linked_attachments TEXT NOT NULL,
		use_markdown BOOLEAN DEFAULT 0,
		PRIMARY KEY("id" AUTOINCREMENT)
	);

	-- TODO: Create indexes here

	`
)
