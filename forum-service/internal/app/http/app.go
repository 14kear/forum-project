package http

import (
	"context"
	"fmt"
	forumHandler "github.com/14kear/forum-project/forum-service/internal/handlers/forum"
	forumRoutes "github.com/14kear/forum-project/forum-service/internal/routes"
	"github.com/14kear/forum-project/forum-service/internal/services/forum"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
)

type App struct {
	engine *gin.Engine
	server *http.Server
	log    *slog.Logger
	port   int
}

// NewApp инициализирует HTTP-сервер Gin и настраивает маршруты
func NewApp(
	log *slog.Logger,
	port int,
	forumService *forum.Forum,
	handler *forumHandler.ForumHandler,
	authMiddleware gin.HandlerFunc,
) *App {
	r := gin.Default()

	// Группировка маршрутов: /api/forum/*
	api := r.Group("/api")
	{
		forumGroup := api.Group("/forum", authMiddleware)
		forumRoutes.RegisterForumRoutes(forumGroup, handler)
	}

	// Healthcheck
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	addr := fmt.Sprintf(":%d", port)
	httpServer := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	return &App{
		engine: r,
		server: httpServer,
		log:    log,
		port:   port,
	}
}

// Run запускает HTTP-сервер
func (a *App) Run() error {
	a.log.Info("HTTP server is running", slog.String("addr", a.server.Addr))
	return a.server.ListenAndServe()
}

// Stop корректно останавливает сервер
func (a *App) Stop(ctx context.Context) error {
	a.log.Info("HTTP server is stopping...")
	return a.server.Shutdown(ctx)
}
