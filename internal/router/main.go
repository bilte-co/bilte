package router

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/bilte-co/bilte/internal/domain"
	"github.com/bilte-co/bilte/internal/templates"
	"golang.org/x/time/rate"
)

const (
	defaultDataDir   = "data"
	defaultStaticDir = "static"
)

type Config struct {
	Production bool
	StaticDir  string
	DataDir    string
	Logger     *slog.Logger
}

func NewRouter(cfg Config) http.Handler {
	cfg = cfg.withDefaults()

	stats := newRequestStats(time.Now())

	mux := http.NewServeMux()
	mux.Handle("GET /public/", withStaticCache(cfg.Production, http.StripPrefix("/public/", http.FileServer(http.Dir(cfg.StaticDir)))))
	mux.HandleFunc("GET /public", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/public/", http.StatusMovedPermanently)
	})
	mux.HandleFunc("GET /stats", statsHandler(stats))
	mux.HandleFunc("GET /{$}", homeHandler(cfg))
	mux.HandleFunc("GET /cv", cvHandler(cfg))

	var handler http.Handler = mux
	handler = withCompression(handler)
	handler = withRateLimit(rate.NewLimiter(50, 200), handler)
	handler = withStats(stats, handler)
	handler = withHealthCheck(handler)
	handler = withRecovery(cfg.Logger, handler)
	handler = withRequestLogging(cfg.Logger, handler)

	return handler
}

func (cfg Config) withDefaults() Config {
	if cfg.StaticDir == "" {
		cfg.StaticDir = defaultStaticDir
	}
	if cfg.DataDir == "" {
		cfg.DataDir = defaultDataDir
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	return cfg
}

func homeHandler(cfg Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		title := "bilte co"
		description := "Strategy-led software engineering and consulting—delivering impactful results in high-stakes domains."

		err := templates.Render(r.Context(), w, http.StatusOK, templates.Home(&cfg.Production, &title, &description))
		if err != nil {
			cfg.Logger.Error("failed to render home page", "error", err)
		}
	}
}

func cvHandler(cfg Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		info, projects, err := loadCVData(cfg.DataDir)
		if err != nil {
			cfg.Logger.Error("failed to load cv data", "error", err)
			http.Error(w, "Error loading CV data", http.StatusInternalServerError)
			return
		}

		err = templates.Render(r.Context(), w, http.StatusOK, templates.CV(&cfg.Production, info, projects))
		if err != nil {
			cfg.Logger.Error("failed to render cv page", "error", err)
		}
	}
}

func loadCVData(dataDir string) (domain.Resume, domain.Projects, error) {
	var info domain.Resume
	if err := readJSON(filepath.Join(dataDir, "resume.json"), &info); err != nil {
		return domain.Resume{}, nil, err
	}

	var projects domain.Projects
	if err := readJSON(filepath.Join(dataDir, "projects.json"), &projects); err != nil {
		return domain.Resume{}, nil, err
	}

	return info, projects, nil
}

func readJSON(path string, v any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}
