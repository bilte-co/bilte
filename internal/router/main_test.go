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

func TestLoadCVDataLoadsProducts(t *testing.T) {
	dataDir := t.TempDir()
	resumeJSON := []byte(`{
		"products": [{
			"name": "socky.flights",
			"role": "Independent Product",
			"technologies": ["Go", "NATS"],
			"description": "Real-time aviation intelligence."
		}]
	}`)
	if err := os.WriteFile(filepath.Join(dataDir, "resume.json"), resumeJSON, 0o644); err != nil {
		t.Fatalf("write resume fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDir, "projects.json"), []byte(`{}`), 0o644); err != nil {
		t.Fatalf("write projects fixture: %v", err)
	}

	resume, projects, err := loadCVData(dataDir)
	if err != nil {
		t.Fatalf("load CV data: %v", err)
	}
	if len(projects) != 0 {
		t.Fatalf("expected no projects, got %d", len(projects))
	}
	if len(resume.Products) != 1 {
		t.Fatalf("expected one product, got %d", len(resume.Products))
	}

	product := resume.Products[0]
	if product.Name != "socky.flights" || product.Role != "Independent Product" {
		t.Fatalf("unexpected product identity: %#v", product)
	}
	if product.Description != "Real-time aviation intelligence." {
		t.Fatalf("unexpected product description %q", product.Description)
	}
	if len(product.Technologies) != 2 || product.Technologies[0] != "Go" || product.Technologies[1] != "NATS" {
		t.Fatalf("unexpected product technologies: %#v", product.Technologies)
	}
}
