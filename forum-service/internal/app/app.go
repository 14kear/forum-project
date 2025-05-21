package app

import (
	"context"
	httpapp "github.com/14kear/forum-project/forum-service/internal/app/http"
	"github.com/14kear/forum-project/forum-service/internal/grpcclient"
	"github.com/14kear/forum-project/forum-service/internal/handlers/chat"
	forumHandler "github.com/14kear/forum-project/forum-service/internal/handlers/forum"
	"github.com/14kear/forum-project/forum-service/internal/middleware"
	"github.com/14kear/forum-project/forum-service/internal/services/forum"
	"github.com/14kear/forum-project/forum-service/internal/storage/postgres"
	"google.golang.org/grpc"
	"log/slog"
	"time"
)

type App struct {
	HTTPServer *httpapp.App
	Forum      *forum.Forum
	conn       *grpc.ClientConn
	cancel     context.CancelFunc
}

func NewApp(log *slog.Logger, httpPort int, storagePath string, authGRPCAddr string) *App {
	storage, err := postgres.New(storagePath)
	if err != nil {
		panic(err)
	}

	// gRPC client to auth-service
	conn, err := grpc.Dial(authGRPCAddr, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	authClient := grpcclient.NewClient(conn)
	authMiddleware := middleware.NewAuthMiddleware(authClient.AuthClient, 1)

	forumService := forum.NewForum(log, storage, storage, storage, authClient.AuthClient)
	forumServer := forumHandler.NewForumHandler(forumService)

	chatServer := chat.NewChatHandler(forumService, authClient.AuthClient, 1, log)

	httpApp := httpapp.NewApp(log, httpPort, forumServer, chatServer, authMiddleware.Middleware())

	ctx, cancel := context.WithCancel(context.Background())

	app := &App{
		HTTPServer: httpApp,
		Forum:      forumService,
		conn:       conn,
		cancel:     cancel,
	}

	// фоновая задача очистки
	go func() {
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Info("cleanup goroutine stopped")
				return
			case <-ticker.C:
				ctxTimeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				err := forumService.CleanupOldMessages(ctxTimeout, 24*time.Hour)
				cancel()

				if err != nil {
					log.Error("failed to cleanup old messages", slog.Any("error", err))
				}
			}
		}
	}()

	return app
}

func (a *App) Stop(ctx context.Context) error {
	a.cancel()
	if err := a.HTTPServer.Stop(ctx); err != nil {
		return err
	}
	return a.conn.Close()
}
