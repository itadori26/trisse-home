package webui

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var dist embed.FS

// Dist returns the built frontend files rooted at dist/.
func Dist() fs.FS {
	sub, _ := fs.Sub(dist, "dist") // the directory is embedded, so Sub cannot fail
	return sub
}
