package routes

import (
	"github.com/14kear/forum-project/forum-service/internal/handlers/forum"
	"github.com/gin-gonic/gin"
)

func RegisterPublicRoutes(rg *gin.RouterGroup, handler *forum.ForumHandler) {
	{
		rg.GET("/topics", handler.ListTopics)
		rg.GET("/topics/:id", handler.GetTopicByID)

		rg.GET("/topics/:id/comments", handler.ListCommentsByTopic)
		rg.GET("/topics/:id/comments/:commentID", handler.GetCommentByID)
	}
}

func RegisterPrivateRoutes(rg *gin.RouterGroup, handler *forum.ForumHandler) {
	{
		rg.POST("/topics", handler.CreateTopic)
		rg.DELETE("/topics/:id", handler.DeleteTopic)

		rg.POST("/topics/:id/comments", handler.CreateComment)
		rg.DELETE("topics/:id/comments/:commentID", handler.DeleteComment)

	}
}
