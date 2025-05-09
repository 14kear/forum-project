package http

import (
	"context"
	"fmt"
	forumHandler "github.com/14kear/forum-project/forum-service/internal/handlers/forum"
	forumRoutes "github.com/14kear/forum-project/forum-service/internal/routes"
	"github.com/gin-contrib/cors"
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
	handler *forumHandler.ForumHandler,
	authMiddleware gin.HandlerFunc,
) *App {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Refresh-Token"},
		ExposeHeaders:    []string{"X-New-Access-Token", "X-New-Refresh-Token"},
		AllowCredentials: true,
	}))

	// Группировка маршрутов: /api/forum/*
	api := r.Group("/api")
	{
		// Публичные маршруты
		publicForumGroup := api.Group("/forum")
		forumRoutes.RegisterPublicRoutes(publicForumGroup, handler)

		// Приватные маршруты (с авторизацией)
		privateForumGroup := api.Group("/forum", authMiddleware)
		forumRoutes.RegisterPrivateRoutes(privateForumGroup, handler)
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
