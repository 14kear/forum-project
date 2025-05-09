package app

import (
	"context"
	httpapp "github.com/14kear/forum-project/forum-service/internal/app/http"
	"github.com/14kear/forum-project/forum-service/internal/grpcclient"
	forumHandler "github.com/14kear/forum-project/forum-service/internal/handlers/forum"
	"github.com/14kear/forum-project/forum-service/internal/middleware"
	"github.com/14kear/forum-project/forum-service/internal/services/forum"
	"github.com/14kear/forum-project/forum-service/internal/storage/postgres"
	"google.golang.org/grpc"
	"log/slog"
)

type App struct {
	HTTPServer *httpapp.App
	conn       *grpc.ClientConn
}

func NewApp(log *slog.Logger, httpPort int, storagePath string, authGRPCAddr string) *App {
	storage, err := postgres.New(storagePath)
	if err != nil {
		panic(err)
	}

	forumService := forum.NewForum(log, storage, storage, storage)
	forumServer := forumHandler.NewForumHandler(forumService)

	// gRPC client to auth-service
	conn, err := grpc.Dial(authGRPCAddr, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	authClient := grpcclient.NewClient(conn)
	authMiddleware := middleware.NewAuthMiddleware(authClient.AuthClient, 1)

	httpApp := httpapp.NewApp(log, httpPort, forumServer, authMiddleware.Middleware())

	return &App{
		HTTPServer: httpApp,
		conn:       conn,
	}
}

func (a *App) Stop(ctx context.Context) error {
	if err := a.HTTPServer.Stop(ctx); err != nil {
		return err
	}
	return a.conn.Close()
}
