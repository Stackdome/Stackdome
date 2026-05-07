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

// fallbackHTML is served when dist/ doesn't contain a real index.html
// (i.e. the binary was built without first running the frontend build).
//
//go:embed fallback.html
var fallbackHTML []byte

// Handler serves the embedded SPA. Unknown paths fall back to index.html
// so react-router's client-side routes survive hard-refresh.
func Handler() http.Handler {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		panic(err)
	}

	hasIndex := false
	if _, err := fs.Stat(sub, "index.html"); err == nil {
		hasIndex = true
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !hasIndex {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Header().Set("Cache-Control", "no-cache")
			_, _ = w.Write(fallbackHTML)
			return
		}

		// URL.Path may be "" or "/" after upstream removeTrailingSlash.
		// Resolve to a concrete filename inside the embed.
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		if _, err := fs.Stat(sub, path); err != nil {
			// SPA fallback for unknown paths: serve index.html so
			// react-router can pick up the route on the client.
			w.Header().Set("Cache-Control", "no-cache")
			http.ServeFileFS(w, r, sub, "index.html")
			return
		}

		// Vite emits content-hashed filenames under /assets/, so they're safe
		// to cache forever. index.html and other roots must stay mutable.
		if strings.HasPrefix(r.URL.Path, "/assets/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			w.Header().Set("Cache-Control", "no-cache")
		}
		http.ServeFileFS(w, r, sub, path)
	})
}
