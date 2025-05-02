package main

import (
	"github.com/14kear/forum-project/auth-service/internal/app"
	"github.com/14kear/forum-project/auth-service/internal/config"
	"github.com/14kear/forum-project/auth-service/internal/lib/logger/handlers/slogpretty"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

const (
	envLocal = "local"
	envProd  = "prod"
	envDev   = "dev"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("Starting up auth service", slog.Any("config", cfg)) // временно, УБРАТЬ В БУДУЩЕМ

	application := app.NewApp(log, cfg.GRPC.Port, cfg.StoragePath, cfg.AccessToken, cfg.RefreshToken)

	go application.GRPCServer.MustRun()

	// получаем сигнал от системы, сама занимается завершением себя
	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	// висим на этой строчке, блокирующая операция
	signl := <-stop
	log.Info("Shutting down auth service", slog.String("signal", signl.String()))
	application.GRPCServer.Stop()
	log.Info("Application stopped")

}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
