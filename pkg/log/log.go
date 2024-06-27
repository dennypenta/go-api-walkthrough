package log

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/google/uuid"
)

type response struct {
	http.ResponseWriter
	status int
}

func (r *response) WriteHeader(statusCode int) {
	r.status = statusCode
}

func (r *response) Flush() {
	r.ResponseWriter.WriteHeader(r.status)
}

type logContextKey struct{}

func LoggerFromContext(ctx context.Context) *slog.Logger {
	return ctx.Value(logContextKey{}).(*slog.Logger)
}

func LoggerToContext(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, logContextKey{}, l)
}

func LogFormatter(groups []string, a slog.Attr) slog.Attr {
	if a.Key != slog.TimeKey {
		return a
	}

	t := a.Value.Time()
	a.Value = slog.StringValue(t.Format(time.RFC3339))
	return a
}

func NewLogger(w io.Writer) *slog.Logger {
	logHandler := slog.NewJSONHandler(w, &slog.HandlerOptions{Level: slog.LevelError, ReplaceAttr: LogFormatter})
	return slog.New(logHandler)
}

func NewLoggingMiddleware(l *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// all the magic of opentelemetry could be here
			traceID := uuid.NewString()
			logger := l.With("service", "userService", "traceID", traceID)

			resp := &response{w, http.StatusOK}
			defer resp.Flush()

			defer func() {
				if recovered := recover(); recovered != nil {
					logger.ErrorContext(
						r.Context(), "recovered from panic",
						"uri", r.RequestURI,
						"method", r.Method,
						"recovered", recovered,
						"stack", string(debug.Stack()),
					)
					resp.WriteHeader(http.StatusInternalServerError)
				}
			}()

			start := time.Now()
			logger.DebugContext(
				r.Context(), "request received",
				"uri", r.RequestURI,
				"method", r.Method,
			)

			ctxWithLogger := LoggerToContext(r.Context(), logger)
			r = r.WithContext(ctxWithLogger)
			next.ServeHTTP(resp, r)

			end := time.Now()
			duration := end.Sub(start)
			// soon we can log the url pattern and easy to match it to our observability toolings
			// https://github.com/golang/go/issues/66405
			// but now let's enjoy RequestURI
			logger.DebugContext(
				r.Context(), "request completed",
				"duration", duration.String(),
				"time", end.Format(time.RFC3339),
				"status", resp.status,
			)
		})
	}
}
