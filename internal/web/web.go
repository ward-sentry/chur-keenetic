package web

import (
	"embed"
	"io/fs"
	"strings"
)

//go:embed index.html
var indexHTML string

//go:embed icons/*.svg
var iconFiles embed.FS

func Index() *strings.Reader {
	return strings.NewReader(indexHTML)
}

func Icons() fs.FS {
	files, err := fs.Sub(iconFiles, "icons")
	if err != nil {
		return iconFiles
	}
	return files
}
