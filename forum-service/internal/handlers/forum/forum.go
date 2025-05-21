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

// CreateTopicRequest describes input for creating a topic
// swagger:model
type CreateTopicRequest struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
}

// CreateCommentRequest describes input for creating a comment
// swagger:model
type CreateCommentRequest struct {
	Content string `json:"content" binding:"required"`
}

func NewForumHandler(forumService *forum.Forum) *ForumHandler {
	return &ForumHandler{forumService: forumService}
}

// CreateTopic godoc
// @Summary Create a new forum topic
// @Description Create a new topic in the forum
// @Tags topics
// @Accept json
// @Produce json
// @Param input body CreateTopicRequest true "Topic data"
// @Success 201 {object} handlers.SuccessIDResponse "Created topic ID"
// @Failure 400 {object} handlers.ErrorResponse "Invalid input"
// @Failure 401 {object} handlers.ErrorResponse "Unauthorized"
// @Failure 500 {object} handlers.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /api/forum/topics [post]
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

// ListTopics godoc
// @Summary List all forum topics
// @Description Retrieve list of topics
// @Tags topics
// @Produce json
// @Success 200 {object} handlers.ListTopicsResponse "List of topics"
// @Failure 500 {object} handlers.ErrorResponse "Internal server error"
// @Router /api/forum/topics [get]
func (f *ForumHandler) ListTopics(c *gin.Context) {
	topics, err := f.forumService.ListTopics(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"topics": topics})
}

// GetTopicByID godoc
// @Summary Get topic by ID
// @Description Retrieve a single topic by its ID
// @Tags topics
// @Produce json
// @Param id path int true "Topic ID"
// @Success 200 {object} handlers.SingleTopicResponse "Topic data"
// @Failure 400 {object} handlers.ErrorResponse "Invalid topic ID"
// @Failure 500 {object} handlers.ErrorResponse "Internal server error"
// @Router /api/forum/topics/{id} [get]
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

// DeleteTopic godoc
// @Summary Delete a topic
// @Description Delete topic by ID (authorized user only)
// @Tags topics
// @Param id path int true "Topic ID"
// @Success 204 "No Content"
// @Failure 400 {object} handlers.ErrorResponse "Invalid topic ID"
// @Failure 401 {object} handlers.ErrorResponse "Unauthorized"
// @Failure 500 {object} handlers.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /api/forum/topics/{id} [delete]
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

// CreateComment godoc
// @Summary Create a comment on a topic
// @Description Add comment to topic by ID
// @Tags comments
// @Accept json
// @Produce json
// @Param id path int true "Topic ID"
// @Param input body CreateCommentRequest true "Comment data"
// @Success 201 {object} handlers.SuccessIDResponse "Created comment ID"
// @Failure 400 {object} handlers.ErrorResponse "Invalid input or topic ID"
// @Failure 401 {object} handlers.ErrorResponse "Unauthorized"
// @Failure 500 {object} handlers.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /api/forum/topics/{id}/comments [post]
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

// ListCommentsByTopic godoc
// @Summary List comments for a topic
// @Description Get all comments for given topic ID
// @Tags comments
// @Produce json
// @Param id path int true "Topic ID"
// @Success 200 {object} handlers.ListCommentsResponse "List of comments"
// @Failure 400 {object} handlers.ErrorResponse "Invalid topic ID"
// @Failure 500 {object} handlers.ErrorResponse "Internal server error"
// @Router /api/forum/topics/{id}/comments [get]
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

// GetCommentByID godoc
// @Summary Get comment by ID for a topic
// @Description Retrieve single comment by comment ID and topic ID
// @Tags comments
// @Produce json
// @Param id path int true "Topic ID"
// @Param commentID path int true "Comment ID"
// @Success 200 {object} handlers.SingleCommentResponse "Comment data"
// @Failure 400 {object} handlers.ErrorResponse "Invalid topic or comment ID"
// @Failure 500 {object} handlers.ErrorResponse "Internal server error"
// @Router /api/forum/topics/{id}/comments/{commentID} [get]
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

	comment, err := f.forumService.GetCommentByID(c.Request.Context(), commentID, topicID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"comment": comment})
}

// DeleteComment godoc
// @Summary Delete a comment
// @Description Delete comment by comment ID and topic ID (authorized user only)
// @Tags comments
// @Param id path int true "Topic ID"
// @Param commentID path int true "Comment ID"
// @Success 204 "No Content"
// @Failure 400 {object} handlers.ErrorResponse "Invalid topic or comment ID"
// @Failure 401 {object} handlers.ErrorResponse "Unauthorized"
// @Failure 500 {object} handlers.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /api/forum/topics/{id}/comments/{commentID} [delete]
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
