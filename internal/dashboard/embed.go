package dashboard

import (
	"embed"
	"io/fs"
)

//go:embed ui/dist
var uiFiles embed.FS

// DistFS returns the embedded ui/dist filesystem.
func DistFS() fs.FS {
	sub, err := fs.Sub(uiFiles, "ui/dist")
	if err != nil {
		panic("dashboard: sub fs: " + err.Error())
	}
	return sub
}
