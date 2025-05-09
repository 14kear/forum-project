package forum

import (
	"github.com/14kear/forum-project/forum-service/internal/services/forum"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type ForumHandler struct {
	forumService *forum.Forum
}

type CreateTopicRequest struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
}

type CreateCommentRequest struct {
	Content string `json:"content" binding:"required"`
}

func NewForumHandler(forumService *forum.Forum) *ForumHandler {
	return &ForumHandler{forumService: forumService}
}

func (f *ForumHandler) CreateTopic(c *gin.Context) {
	var req CreateTopicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, ok := userIDValue.(int64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id in context"})
		return
	}

	userEmailValue, exists := c.Get("userEmail")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userEmail, ok := userEmailValue.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user email in context"})
		return
	}

	topicID, err := f.forumService.CreateTopic(c.Request.Context(), req.Title, req.Content, userID, userEmail)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"topic_id": topicID})
}

func (f *ForumHandler) ListTopics(c *gin.Context) {
	topics, err := f.forumService.ListTopics(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"topics": topics})
}

func (f *ForumHandler) GetTopicByID(c *gin.Context) {
	topicIDStr := c.Param("id")
	topicID, err := strconv.Atoi(topicIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid topic id"})
		return
	}

	topic, err := f.forumService.GetTopicByID(c.Request.Context(), topicID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"topic": topic})

}

func (f *ForumHandler) DeleteTopic(c *gin.Context) {
	topicIDStr := c.Param("id")

	topicID, err := strconv.Atoi(topicIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid topic ID"})
		return
	}

	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, ok := userIDValue.(int64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id in context"})
		return
	}

	err = f.forumService.DeleteTopic(c.Request.Context(), topicID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, gin.H{})
}

func (f *ForumHandler) CreateComment(c *gin.Context) {
	var req CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, ok := userIDValue.(int64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id in context"})
		return
	}

	userEmailValue, exists := c.Get("userEmail")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userEmail, ok := userEmailValue.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user email in context"})
		return
	}

	topicIDStr := c.Param("id")
	topicID, err := strconv.Atoi(topicIDStr)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid topic ID"})
		return
	}

	commentID, err := f.forumService.CreateComment(c.Request.Context(), topicID, userID, req.Content, userEmail)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"comment_id": commentID})
}

func (f *ForumHandler) ListCommentsByTopic(c *gin.Context) {
	topicIDStr := c.Param("id")
	topicID, err := strconv.Atoi(topicIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid topic ID"})
		return
	}

	comments, err := f.forumService.CommentsByTopicID(c.Request.Context(), topicID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"comments": comments})
}

func (f *ForumHandler) GetCommentByID(c *gin.Context) {
	commentIDStr := c.Param("commentID")
	commentID, err := strconv.Atoi(commentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid comment ID"})
		return
	}

	topicIDStr := c.Param("id")
	topicID, err := strconv.Atoi(topicIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid topic ID"})
		return
	}

	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, ok := userIDValue.(int64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id in context"})
		return
	}

	comment, err := f.forumService.GetCommentByID(c.Request.Context(), commentID, topicID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"comment": comment})
}

func (f *ForumHandler) DeleteComment(c *gin.Context) {
	commentIDStr := c.Param("commentID")
	commentID, err := strconv.Atoi(commentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid comment ID"})
		return
	}

	topicIDStr := c.Param("id")
	topicID, err := strconv.Atoi(topicIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid topic ID"})
		return
	}

	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, ok := userIDValue.(int64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id in context"})
		return
	}

	err = f.forumService.DeleteComment(c.Request.Context(), commentID, topicID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, gin.H{})
}
