package router

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/time/rate"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (r *statusRecorder) WriteHeader(status int) {
	if r.status != 0 {
		return
	}
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.WriteHeader(http.StatusOK)
	}
	n, err := r.ResponseWriter.Write(b)
	r.bytes += n
	return n, err
}

func (r *statusRecorder) Flush() {
	if r.status == 0 {
		r.WriteHeader(http.StatusOK)
	}
	if flusher, ok := r.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (r *statusRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

type requestStats struct {
	startedAt time.Time
	total     atomic.Uint64
	inFlight  atomic.Int64
	mu        sync.RWMutex
	statuses  map[int]uint64
}

func newRequestStats(startedAt time.Time) *requestStats {
	return &requestStats{
		startedAt: startedAt,
		statuses:  make(map[int]uint64),
	}
}

func (s *requestStats) record(status int) {
	s.total.Add(1)

	s.mu.Lock()
	defer s.mu.Unlock()
	s.statuses[status]++
}

func (s *requestStats) snapshot() map[string]any {
	statuses := make(map[string]uint64)

	s.mu.RLock()
	for status, count := range s.statuses {
		statuses[strconv.Itoa(status)] = count
	}
	s.mu.RUnlock()

	return map[string]any{
		"started_at": s.startedAt.UTC().Format(time.RFC3339),
		"uptime":     time.Since(s.startedAt).Round(time.Second).String(),
		"requests":   s.total.Load(),
		"in_flight":  s.inFlight.Load(),
		"statuses":   statuses,
	}
}

func statsHandler(stats *requestStats) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, stats.snapshot())
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func withHealthCheck(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Health-Check") == "1" {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
			return
		}

		next.ServeHTTP(w, r)
	})
}

func withStats(stats *requestStats, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stats.inFlight.Add(1)
		defer stats.inFlight.Add(-1)

		recorder := &statusRecorder{ResponseWriter: w}
		next.ServeHTTP(recorder, r)

		status := recorder.status
		if status == 0 {
			status = http.StatusOK
		}
		stats.record(status)
	})
}

func withRateLimit(limiter *rate.Limiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "too many requests"})
			return
		}

		next.ServeHTTP(w, r)
	})
}

func withStaticCache(production bool, next http.Handler) http.Handler {
	if !production {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=31536000")
		next.ServeHTTP(w, r)
	})
}

func withRecovery(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.Error(
					"panic while serving request",
					"method", r.Method,
					"path", r.URL.Path,
					"panic", recovered,
					"stack", string(debug.Stack()),
				)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func withRequestLogging(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &statusRecorder{ResponseWriter: w}

		next.ServeHTTP(recorder, r)

		status := recorder.status
		if status == 0 {
			status = http.StatusOK
		}

		logger.Info(
			"request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"status", status,
			"bytes", recorder.bytes,
			"duration", time.Since(start).String(),
			"remote_addr", r.RemoteAddr,
			"user_agent", r.UserAgent(),
		)
	})
}
