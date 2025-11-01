package env

import "os"

var (
	ServerPathPrefix = os.Getenv("SERVER_PATH_PREFIX")
	Debug            = os.Getenv("DEBUG") == "true"
	LogFormat        = os.Getenv("LOG_FORMAT")
)
