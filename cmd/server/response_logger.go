package server

import (
	"net/http"
	"strings"
	"time"

	"github.com/golang/glog"
)

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       *strings.Builder
}

func (rr *responseRecorder) WriteHeader(code int) {
	rr.statusCode = code
	rr.ResponseWriter.WriteHeader(code)
	rr.statusCode = code
}

func (rr *responseRecorder) Write(b []byte) (int, error) {
	if rr.body == nil {
		rr.body = &strings.Builder{}
	}
	rr.body.Write(b) // Capture response body
	return rr.ResponseWriter.Write(b)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rr := &responseRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(rr, r)

		duration := time.Since(start)
		glog.Infof("%s %s %d %s", r.Method, r.URL.Path, rr.statusCode, duration)

		if rr.statusCode < 200 || rr.statusCode >= 300 {
			glog.Warningf("Non-2xx response body for %s %s: %q", r.Method, r.URL.Path, rr.body.String())
		}
	})
}
