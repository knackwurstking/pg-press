package env

import "os"

var (
	Admins           = os.Getenv("ADMINS")
	ServerAddress    = os.Getenv("SERVER_ADDR")
	ServerPathPrefix = os.Getenv("SERVER_PATH_PREFIX")
	Verbose          = os.Getenv("VERBOSE") == "true"
)
