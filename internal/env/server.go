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
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	if ServerPathImages == "" {
		ServerPathImages = fmt.Sprintf("%s/.%s/images", home, Name)
	}

	if _, err := os.Stat(ServerPathImages); os.IsNotExist(err) {
		if err = os.MkdirAll(ServerPathImages, 0700); err != nil {
			panic(err)
		}
	}
}
