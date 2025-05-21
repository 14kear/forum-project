package chat

import (
	"github.com/14kear/forum-project/forum-service/internal/services/forum"
	ssov1 "github.com/14kear/forum-project/protos/gen/go/auth"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log/slog"
	"net/http"
)

type ChatHandler struct {
	chatService *forum.Forum
	authService ssov1.AuthClient
	appID       int
	log         *slog.Logger
}

// MessageResponse представляет структуру ответа с сообщением в чате
// swagger:model
type MessageResponse struct {
	ID        int64  `json:"id"`
	Content   string `json:"content"`
	UserID    int64  `json:"userID"`
	UserEmail string `json:"userEmail"`
}

func NewChatHandler(chatService *forum.Forum, authServer ssov1.AuthClient, appID int, log *slog.Logger) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
		authService: authServer,
		appID:       appID,
		log:         log,
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// HandleWebSocket godoc
// @Summary WebSocket endpoint for chat
// @Description Establishes a WebSocket connection for exchanging chat messages. Used only for WebSocket clients. Requires `accessToken` in query parameters.
// @Tags chat
// @Param accessToken query string true "Access token for authentication"
// @Success 101 {string} string "Switching Protocols – WebSocket connection established"
// @Failure 401 {object} handlers.ErrorResponse "Unauthorized – invalid or missing token"
// @Failure 500 {object} handlers.ErrorResponse "Internal server error"
// @Router /api/forum/ws/chat [get]
// @Security ApiKeyAuth
func (h *ChatHandler) HandleWebSocket(c *gin.Context) {
	const op = "chat.HandleWebSocket"
	log := h.log.With(slog.String("op", op))
	log.Info("start")

	accessToken := c.Query("accessToken")
	ctx := c.Request.Context()

	if accessToken == "" {
		log.Warn("missing access token")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing access token"})
		return
	}

	claims, err := h.authService.ValidateToken(ctx, &ssov1.ValidateTokenRequest{
		AccessToken: accessToken,
		AppId:       int32(h.appID),
	})
	if err != nil {
		log.Warn("invalid access token", slog.Any("error", err))
		conn, _ := upgrader.Upgrade(c.Writer, c.Request, nil)
		_ = conn.WriteJSON(map[string]string{"error": "unauthorized"})
		_ = conn.Close()
		return
	}

	userID := claims.GetUserId()
	userEmail := claims.GetEmail()

	log = log.With(
		slog.Int64("userID", userID),
		slog.String("userEmail", userEmail),
	)

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Error("failed to upgrade connection", slog.Any("error", err))
		return
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Error("failed to close connection", slog.Any("error", err))
		}
	}()

	log.Info("WebSocket connection established")

	for {
		var incoming struct {
			Content string `json:"content"`
		}

		if err := conn.ReadJSON(&incoming); err != nil {
			if websocket.IsUnexpectedCloseError(err) {
				log.Info("connection closed by client")
			} else {
				log.Error("failed to read message", slog.Any("error", err))
			}
			break
		}

		log.Debug("message received", slog.String("content", incoming.Content))

		chatMessageID, err := h.chatService.CreateChatMessage(ctx, userID, incoming.Content, userEmail)
		if err != nil {
			log.Error("failed to create chat message",
				slog.Any("error", err),
				slog.String("content", incoming.Content),
			)
			_ = conn.WriteJSON(map[string]string{"error": "internal server error"})
			break
		}

		response := MessageResponse{
			ID:        chatMessageID,
			Content:   incoming.Content,
			UserID:    userID,
			UserEmail: userEmail,
		}

		if err := conn.WriteJSON(response); err != nil {
			log.Error("failed to send response", slog.Any("error", err))
			break
		}
	}
}

// GetChatMessages godoc
// @Summary Get all chat messages
// @Description Returns a list of all messages from a chat
// @Tags chat
// @Produce json
// @Success 200 {array} MessageResponse "List of chat messages"
// @Failure 500 {object} handlers.ErrorResponse "Failed to load messages"
// @Router /api/forum/ws/chat/messages [get]
func (h *ChatHandler) GetChatMessages(c *gin.Context) {
	const op = "chat.GetChatMessages"
	log := h.log.With(slog.String("op", op))

	messages, err := h.chatService.ListChatMessages(c.Request.Context())
	if err != nil {
		log.Error("failed to get messages", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load messages"})
		return
	}

	var response []MessageResponse
	for _, m := range messages {
		response = append(response, MessageResponse{
			ID:        int64(m.ID),
			Content:   m.Content,
			UserID:    m.UserID,
			UserEmail: m.UserEmail,
		})
	}

	c.JSON(http.StatusOK, response)
}
