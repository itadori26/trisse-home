// Package webui embeds the compiled single-page application into the Go
// binary. Vite writes the build output (from web/) into the dist directory
// next to this file.
//
// Ring: Frameworks & Drivers.
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
