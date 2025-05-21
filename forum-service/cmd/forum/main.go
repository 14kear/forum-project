package main

import (
	"context"
	"errors"
	"github.com/14kear/forum-project/forum-service/internal/app"
	"github.com/14kear/forum-project/forum-service/internal/config"
	"github.com/14kear/forum-project/forum-service/utils"
	"log/slog"
	"net/http"
	_ "os"
	"os/signal"
	"syscall"
	"time"
)

const (
	envLocal = "local"
	//envDev   = "dev"
	//envProd  = "prod"
)

func main() {
	cfg := config.Load("forum-service/config/local.yaml")
	log := utils.New(cfg.Env)

	application := app.NewApp(log, cfg.HTTP.Port, cfg.StoragePath, cfg.GRPC.Address)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := application.HTTPServer.Run(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				log.Info("HTTP server closed gracefully")
			} else {
				log.Error("failed to run HTTP server", slog.String("error", err.Error()))
			}
		}
	}()

	log.Info("Forum service started", slog.String("env", envLocal), slog.Int("port", cfg.HTTP.Port))

	<-ctx.Done()

	log.Info("Shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := application.Stop(shutdownCtx); err != nil {
		log.Error("failed to stop application", slog.String("error", err.Error()))
	}
}
