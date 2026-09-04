package router

import (
	"compress/gzip"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRouterCompressesHTML(t *testing.T) {
	router := NewRouter(Config{Logger: slog.New(slog.NewTextHandler(io.Discard, nil))})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if got := rec.Header().Get("Content-Encoding"); got != "gzip" {
		t.Fatalf("expected gzip content encoding, got %q", got)
	}

	reader, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	defer reader.Close()

	body, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read gzipped body: %v", err)
	}
	if !strings.Contains(string(body), "<h1>bilte co</h1>") {
		t.Fatalf("expected decompressed body to include home page content")
	}
}

func TestRouterSkipsCompressedAssets(t *testing.T) {
	staticDir := t.TempDir()
	fontPath := filepath.Join(staticDir, "font.woff2")
	if err := os.WriteFile(fontPath, []byte("font data"), 0o644); err != nil {
		t.Fatalf("failed to write font fixture: %v", err)
	}

	router := NewRouter(Config{
		StaticDir: staticDir,
		Logger:    slog.New(slog.NewTextHandler(io.Discard, nil)),
	})

	req := httptest.NewRequest(http.MethodGet, "/public/font.woff2", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if got := rec.Header().Get("Content-Encoding"); got != "" {
		t.Fatalf("expected no content encoding for woff2, got %q", got)
	}
	if got := rec.Body.String(); got != "font data" {
		t.Fatalf("expected raw font response, got %q", got)
	}
}

func TestRouterStaticCacheHeaderInProduction(t *testing.T) {
	staticDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(staticDir, "styles.css"), []byte("body{}"), 0o644); err != nil {
		t.Fatalf("failed to write css fixture: %v", err)
	}

	router := NewRouter(Config{
		Production: true,
		StaticDir:  staticDir,
		Logger:     slog.New(slog.NewTextHandler(io.Discard, nil)),
	})

	req := httptest.NewRequest(http.MethodGet, "/public/styles.css", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if got := rec.Header().Get("Cache-Control"); got != "public, max-age=31536000" {
		t.Fatalf("expected production cache header, got %q", got)
	}
}

func TestRouterHealthCheckHeader(t *testing.T) {
	router := NewRouter(Config{Logger: slog.New(slog.NewTextHandler(io.Discard, nil))})

	req := httptest.NewRequest(http.MethodGet, "/does-not-matter", nil)
	req.Header.Set("X-Health-Check", "1")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if got := rec.Body.String(); got != "ok" {
		t.Fatalf("expected health body %q, got %q", "ok", got)
	}
}
