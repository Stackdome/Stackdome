// Package web embeds the built React SPA so the API server ships as a
// single binary with no external asset directory at runtime.
package web

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

// "all:" includes dotfiles like Vite's .vite/ metadata.
//
//go:embed all:dist
var distFS embed.FS

// Handler serves the embedded SPA. Unknown paths fall back to index.html
// so react-router's client-side routes survive hard-refresh.
func Handler() http.Handler {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		panic(err)
	}

	fileServer := http.FileServer(http.FS(sub))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		if _, err := fs.Stat(sub, path); err != nil {
			r2 := r.Clone(r.Context())
			r2.URL.Path = "/"
			w.Header().Set("Cache-Control", "no-cache")
			fileServer.ServeHTTP(w, r2)
			return
		}

		// Vite emits content-hashed filenames under /assets/, so they're safe
		// to cache forever. index.html and other roots must stay mutable.
		if strings.HasPrefix(r.URL.Path, "/assets/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			w.Header().Set("Cache-Control", "no-cache")
		}
		fileServer.ServeHTTP(w, r)
	})
}
