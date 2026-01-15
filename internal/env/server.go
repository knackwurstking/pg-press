package env

import (
	"fmt"
	"os"
)

const (
	Name = "pg-press"
)

var (
	Admins           = os.Getenv("ADMINS")
	ServerAddress    = os.Getenv("SERVER_ADDR")
	ServerPathPrefix = os.Getenv("SERVER_PATH_PREFIX")
	ServerPathImages = os.Getenv("SERVER_PATH_IMAGES")
	Verbose          = os.Getenv("VERBOSE") == "true"
)

func init() {
	if ServerPathImages == "" {
		ServerPathImages = fmt.Sprintf("/var/www/%s/images", Name)
	}

	if _, err := os.Stat(ServerPathImages); os.IsNotExist(err) {
		if err = os.MkdirAll(ServerPathImages, 0700); err != nil {
			panic(err)
		}
	}
}
