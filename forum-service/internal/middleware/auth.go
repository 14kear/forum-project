package middleware

import (
	ssov1 "github.com/14kear/forum-project/protos/gen/go/auth"
	"strings"
)

type AuthMiddleware struct {
	authClient ssov1.AuthClient
	appID      int
}

func NewAuthMiddleware(authClient ssov1.AuthClient, appID int) *AuthMiddleware {
	return &AuthMiddleware{authClient: authClient, appID: appID}
}

//func (m *AuthMiddleware) Middleware(next http.Handler) http.Handler {
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		accessToken := extractTokenFromHeader(r.Header.Get("Authorization"))
//		refreshToken := r.Header.Get("X-Refresh-Token")
//
//		if accessToken == "" {
//			http.Error(w, "missing access token", http.StatusUnauthorized)
//			return
//		}
//
//		ctx := r.Context()
//
//		// Получим userID из токена
//		userInfo, err := m.authClient.
//	})
//}

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
