package routes

import "io/fs"

type Options struct {
	Templates fs.FS
	Data      Data
}

type Data struct {
	Version          string
	ServerPathPrefix string
}
