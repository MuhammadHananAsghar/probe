package dashboard

import (
	"io/fs"
	"net/http"
)

// spaHandler serves static files from fsys and falls back to index.html for
// paths that don't match a real file (enabling client-side routing).
func spaHandler(fsys fs.FS) http.Handler {
	fileServer := http.FileServerFS(fsys)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to serve the file directly.
		f, err := fsys.Open(r.URL.Path[1:]) // strip leading "/"
		if err == nil {
			f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}
		// Fall back to index.html for client-side routing.
		r2 := r.Clone(r.Context())
		r2.URL.Path = "/"
		fileServer.ServeHTTP(w, r2)
	})
}
