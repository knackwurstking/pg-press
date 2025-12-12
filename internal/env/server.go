package env

import "os"

var (
	Admins           = os.Getenv("ADMINS")
	ServerAddress    = os.Getenv("SERVER_ADDR")
	ServerPathPrefix = os.Getenv("SERVER_PATH_PREFIX")
	LogLevel         = os.Getenv("LOG_LEVEL")  // LogLevel would be "debug", "info", "warn", "error", or "fatal"
	LogFormat        = os.Getenv("LOG_FORMAT") // LogFormat would be "json" or "text"
)
