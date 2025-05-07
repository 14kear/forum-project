package routes

import (
	"github.com/14kear/forum-project/forum-service/internal/handlers/forum"
	"github.com/gin-gonic/gin"
)

func RegisterForumRoutes(rg *gin.RouterGroup, handler *forum.ForumHandler) {
	{
		rg.POST("/topics", handler.CreateTopic)
		rg.GET("/topics", handler.ListTopics)
		rg.GET("/topics/:id", handler.GetTopicByID)
		rg.DELETE("/topics/:id", handler.DeleteTopic)

		rg.POST("/topics/:id/comments", handler.CreateComment)
		rg.GET("/topics/:id/comments", handler.ListCommentsByTopic)
		rg.GET("/topics/:id/comments/:commentID", handler.GetCommentByID)
		rg.DELETE("topics/:id/comments/:commentID", handler.DeleteComment)
	}
}
