package dashboard

import (
	"embed"
	"io/fs"
)

//go:embed assets
var assets embed.FS

func GetAssets() fs.FS {
	sub, err := fs.Sub(assets, "assets")
	if err != nil {
		panic("fs.Sub failed: " + err.Error())
	}
	return sub
}
