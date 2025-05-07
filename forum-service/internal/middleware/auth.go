package middleware

import (
	"fmt"
	ssov1 "github.com/14kear/forum-project/protos/gen/go/auth"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
	"strings"
)

type AuthMiddleware struct {
	authClient ssov1.AuthClient
	appID      int
}

func NewAuthMiddleware(authClient ssov1.AuthClient, appID int) *AuthMiddleware {
	return &AuthMiddleware{authClient: authClient, appID: appID}
}

func (m *AuthMiddleware) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Пропускаем запросы, связанные с авторизацией
		if strings.HasPrefix(c.Request.URL.Path, "/auth/") {
			c.Next()
			return
		}

		accessToken := extractTokenFromHeader(c.GetHeader("Authorization"))
		refreshToken := c.GetHeader("X-Refresh-Token")

		if accessToken == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing access token"})
			return
		}

		ctx := c.Request.Context()
		resp, err := m.authClient.ValidateToken(ctx, &ssov1.ValidateTokenRequest{
			AccessToken: accessToken,
			AppId:       int32(m.appID),
		})

		// Попытка рефреша, если access token невалиден
		if err != nil {
			st, ok := status.FromError(err)
			if !ok || st.Code() != codes.Unauthenticated || refreshToken == "" {
				fmt.Println("Error: ", err.Error())
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
				return
			}

			newTokens, err := m.authClient.RefreshTokens(ctx, &ssov1.RefreshTokenRequest{
				RefreshToken: refreshToken,
				AppId:        int32(m.appID),
			})
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token refresh failed"})
				return
			}

			// Устанавливаем новые токены в ответ
			c.Header("X-New-Access-Token", newTokens.AccessToken)
			c.Header("X-New-Refresh-Token", newTokens.RefreshToken)

			// Повторная валидация
			resp, err = m.authClient.ValidateToken(ctx, &ssov1.ValidateTokenRequest{
				AccessToken: newTokens.AccessToken,
				AppId:       int32(m.appID),
			})
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid refreshed token"})
				return
			}

			// Обновляем заголовок авторизации
			c.Request.Header.Set("Authorization", "Bearer "+newTokens.AccessToken)
		}

		// Прокидываем userID в контекст Gin
		c.Set("userID", resp.GetUserId())
		c.Next()
	}
}

func extractTokenFromHeader(header string) string {
	if header == "" {
		return ""
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}
	return parts[1]
}
