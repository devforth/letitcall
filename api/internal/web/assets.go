package web

import (
	"embed"
	"io/fs"
)

// static is replaced with the compiled Svelte portal during the Docker build.
//
//go:embed static
var embedded embed.FS

var Assets fs.FS

func init() {
	assets, err := fs.Sub(embedded, "static")
	if err != nil {
		panic(err)
	}
	Assets = assets
}
