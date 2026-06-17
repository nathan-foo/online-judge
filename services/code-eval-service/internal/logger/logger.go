package logger

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/diode"
	"github.com/rs/zerolog/log"
)

func Init() {
	zerolog.DurationFieldInteger = true

	w := diode.NewWriter(os.Stderr, 1000, 10*time.Millisecond, func(missed int) {
		fmt.Fprintf(os.Stderr, "logger dropped %d messages\n", missed)
	})

	log.Logger = zerolog.New(w).With().Timestamp().Logger()
}

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		next.ServeHTTP(ww, r)

		status := ww.Status()
		event := log.Info()
		if status >= 500 {
			event = log.Error()
		} else if status >= 400 {
			event = log.Warn()
		}

		event.
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", status).
			Int("bytes", ww.BytesWritten()).
			Dur("latency", time.Since(start)).
			Str("request_id", middleware.GetReqID(r.Context())).
			Str("remote_addr", r.RemoteAddr).
			Msg("request")
	})
}
