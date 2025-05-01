package app

import (
	grpcapp "github.com/14kear/forum-project/auth-service/internal/app/grpc"
	"github.com/14kear/forum-project/auth-service/internal/services/auth"
	"github.com/14kear/forum-project/auth-service/internal/storage/postgres"
	"log/slog"
	"time"
)

type App struct {
	GRPCServer *grpcapp.App
}

func NewApp(log *slog.Logger, grpcPort int, storagePath string, accessToken time.Duration, refreshToken time.Duration) *App {
	storage, err := postgres.New(storagePath)
	if err != nil {
		panic(err)
	}

	authService := auth.NewAuth(log, storage, storage, storage, accessToken, refreshToken)

	grpcApp := grpcapp.NewApp(log, authService, grpcPort)

	return &App{
		GRPCServer: grpcApp,
	}
}
