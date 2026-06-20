package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type responseStatusCode struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseStatusCode) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func MiddlewareLog(n http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		stcd := &responseStatusCode{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}
		n.ServeHTTP(stcd, r)

		if stcd.statusCode == http.StatusMethodNotAllowed {
			Log.Warn("method not allowed for this endpoint",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Duration("duration",
					time.Since(start)))
		}

		/*Log.Info("http request handled",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.Duration("duration",
			time.Since(start)))*/
	})
}
