package jwt

import (
	"errors"
	"fmt"
	"github.com/14kear/forum-project/auth-service/internal/domain/models"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

var (
	ErrTokenExpired     = errors.New("token is expired")
	ErrInvalidToken     = errors.New("invalid token")
	ErrInvalidTokenType = errors.New("invalid token type")
)

// ValidationError wraps specific reasons for token validation failure.
type ValidationError struct {
	Reason error
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("token validation error: %v", e.Reason)
}

func NewTokenPair(user models.User, app models.App, accessTTL, refreshTTL time.Duration) (*TokenPair, error) {
	accessToken, err := newAccessToken(user, app, accessTTL)
	if err != nil {
		return nil, err
	}

	refreshToken, err := newRefreshToken(user, app, refreshTTL)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func newAccessToken(user models.User, app models.App, ttl time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["uid"] = user.ID
	claims["email"] = user.Email
	claims["app_id"] = app.ID
	claims["typ"] = "access"
	claims["exp"] = time.Now().Add(ttl).Unix()

	return token.SignedString([]byte(app.Secret))
}

func newRefreshToken(user models.User, app models.App, ttl time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["uid"] = user.ID
	claims["email"] = user.Email
	claims["typ"] = "refresh"
	claims["exp"] = time.Now().Add(ttl).Unix()

	return token.SignedString([]byte(app.Secret))
}
