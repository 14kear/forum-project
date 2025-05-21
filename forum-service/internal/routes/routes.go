package routes

import (
	"github.com/14kear/forum-project/forum-service/internal/handlers/chat"
	"github.com/14kear/forum-project/forum-service/internal/handlers/forum"
	"github.com/gin-gonic/gin"
)

func RegisterPublicRoutes(rg *gin.RouterGroup, handler *forum.ForumHandler, chatHandler *chat.ChatHandler) {
	{
		rg.GET("/topics", handler.ListTopics)
		rg.GET("/topics/:id", handler.GetTopicByID)

		rg.GET("/topics/:id/comments", handler.ListCommentsByTopic)
		rg.GET("/topics/:id/comments/:commentID", handler.GetCommentByID)

		rg.GET("ws/chat/messages", chatHandler.GetChatMessages)
		rg.GET("/ws/chat", chatHandler.HandleWebSocket)
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
