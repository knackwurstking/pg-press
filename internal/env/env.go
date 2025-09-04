package env

import "os"

var (
	ServerPathPrefix = os.Getenv("SERVER_PATH_PREFIX")
)
