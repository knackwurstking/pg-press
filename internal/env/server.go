package env

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
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
	level := slog.LevelInfo
	if Verbose {
		level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level:      level,
			TimeFormat: time.Kitchen,
		}),
	))

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
