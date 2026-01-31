package assets

import (
	"embed"
	"io/fs"

	"github.com/labstack/echo/v4"
)

var (
	//go:embed public
	Public embed.FS
)

func GetPublic() fs.FS {
	return echo.MustSubFS(Public, "public")
}
