package web

import (
	"context"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/bilte-co/bilte/internal/logging"
	"github.com/bilte-co/bilte/internal/router"
	"github.com/joho/godotenv"
)

type WebCmd struct{}

func (cmd *WebCmd) Run(ctx *context.Context) error {
	logger := logging.NewLoggerFromEnv()

	err := godotenv.Load()
	if err != nil {
		logger.Debug("🤯 failed to load environment variables")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "development"
	}

	isProduction := appEnv == "production"

	server := &http.Server{
		Addr: ":" + port,
		Handler: router.NewRouter(router.Config{
			Production: isProduction,
			Logger:     logger,
		}),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		<-(*ctx).Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("failed to shut down web server", "error", err)
		}
	}()

	logger.Info("starting web server", "addr", server.Addr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}
