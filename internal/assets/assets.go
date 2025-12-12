package assets

import (
	"embed"
	"io/fs"

	"github.com/labstack/echo/v4"
)

var (
	//go:embed assets
	Assets embed.FS
)

func GetAssets() fs.FS {
	return echo.MustSubFS(Assets, "assets")
}
