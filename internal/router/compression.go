package router

import (
	"compress/gzip"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

var gzipWriterPool = sync.Pool{
	New: func() any {
		w, err := gzip.NewWriterLevel(io.Discard, gzip.DefaultCompression)
		if err != nil {
			panic(err)
		}
		return w
	},
}

type gzipResponseWriter struct {
	http.ResponseWriter
	writer      *gzip.Writer
	status      int
	wroteHeader bool
}

func (w *gzipResponseWriter) WriteHeader(status int) {
	if w.status != 0 {
		return
	}
	w.status = status

	if !compressibleStatus(status) {
		w.wroteHeader = true
		w.ResponseWriter.WriteHeader(status)
	}
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}

	if !compressibleStatus(w.status) {
		if !w.wroteHeader {
			w.wroteHeader = true
			w.ResponseWriter.WriteHeader(w.status)
		}
		return w.ResponseWriter.Write(b)
	}

	if w.writer == nil {
		w.Header().Del("Content-Length")
		w.Header().Set("Content-Encoding", "gzip")

		gzipWriter := gzipWriterPool.Get().(*gzip.Writer)
		gzipWriter.Reset(w.ResponseWriter)
		w.writer = gzipWriter

		w.wroteHeader = true
		w.ResponseWriter.WriteHeader(w.status)
	}

	return w.writer.Write(b)
}

func (w *gzipResponseWriter) Flush() {
	if w.writer != nil {
		_ = w.writer.Flush()
	}
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (w *gzipResponseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func (w *gzipResponseWriter) finish() {
	if w.writer != nil {
		_ = w.writer.Close()
		w.writer.Reset(io.Discard)
		gzipWriterPool.Put(w.writer)
		return
	}

	if w.status != 0 && !w.wroteHeader {
		w.ResponseWriter.WriteHeader(w.status)
	}
}

func withCompression(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !shouldCompress(r) {
			next.ServeHTTP(w, r)
			return
		}

		appendVary(w.Header(), "Accept-Encoding")
		gzipWriter := &gzipResponseWriter{ResponseWriter: w}
		defer gzipWriter.finish()

		next.ServeHTTP(gzipWriter, r)
	})
}

func shouldCompress(r *http.Request) bool {
	if r.Method == http.MethodHead {
		return false
	}
	if r.Header.Get("Range") != "" {
		return false
	}
	if strings.Contains(r.Header.Get("Accept"), "text/event-stream") {
		return false
	}
	if !acceptsGzip(r.Header.Get("Accept-Encoding")) {
		return false
	}

	switch strings.ToLower(filepath.Ext(r.URL.Path)) {
	case ".avif", ".br", ".gif", ".gz", ".ico", ".jpeg", ".jpg", ".mp3", ".mp4", ".pdf", ".png", ".webp", ".woff", ".woff2", ".zip":
		return false
	default:
		return true
	}
}

func acceptsGzip(header string) bool {
	for _, part := range strings.Split(header, ",") {
		token, params, _ := strings.Cut(strings.TrimSpace(part), ";")
		if !strings.EqualFold(token, "gzip") && token != "*" {
			continue
		}
		return encodingWeight(params) > 0
	}

	return false
}

func encodingWeight(params string) float64 {
	for _, param := range strings.Split(params, ";") {
		key, value, ok := strings.Cut(strings.TrimSpace(param), "=")
		if !ok || !strings.EqualFold(key, "q") {
			continue
		}

		weight, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return 1
		}
		return weight
	}

	return 1
}

func compressibleStatus(status int) bool {
	switch status {
	case http.StatusNoContent, http.StatusNotModified:
		return false
	default:
		return status >= 200
	}
}

func appendVary(header http.Header, value string) {
	for _, existing := range header.Values("Vary") {
		for _, part := range strings.Split(existing, ",") {
			if strings.EqualFold(strings.TrimSpace(part), value) {
				return
			}
		}
	}

	header.Add("Vary", value)
}
