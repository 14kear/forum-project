package auth

import (
	"context"
	"errors"
	"fmt"
	"github.com/14kear/forum-project/auth-service/internal/domain/models"
	"github.com/14kear/forum-project/auth-service/internal/lib/jwt"
	"github.com/14kear/forum-project/auth-service/internal/lib/logger/sl"
	"github.com/14kear/forum-project/auth-service/internal/storage"
	jwtGo "github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"time"
)

type Auth struct {
	log          *slog.Logger
	userSaver    UserSaver
	userProvider UserProvider
	appProvider  AppProvider
	tokenStorage TokenStorage
	accessToken  time.Duration
	refreshToken time.Duration
}

type TokenStorage interface {
	SaveToken(ctx context.Context, userID int64, appID int, token string, expiresAt time.Time) (int64, error)
	RevokeRefreshToken(ctx context.Context, userID int64, appID int, token string) error
	IsRefreshTokenValid(ctx context.Context, userID int64, appID int, token string) (bool, error)
	DeleteExpiredTokens(ctx context.Context, appID int) error
}

type UserSaver interface {
	SaveUser(ctx context.Context, email string, passHash []byte) (uid int64, err error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (user models.User, err error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

type AppProvider interface {
	App(ctx context.Context, appID int) (models.App, error)
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidAppID       = errors.New("invalid app id")
	ErrUserExists         = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
)

// NewAuth return a new instance of the Auth service
func NewAuth(
	log *slog.Logger,
	userSaver UserSaver,
	userProvider UserProvider,
	appProvider AppProvider,
	tokenStorage TokenStorage,
	accessToken time.Duration,
	refreshToken time.Duration,
) *Auth {
	return &Auth{
		log:          log,
		userSaver:    userSaver,
		userProvider: userProvider,
		appProvider:  appProvider,
		tokenStorage: tokenStorage,
		accessToken:  accessToken,
		refreshToken: refreshToken,
	}
}

// Login checks if user with given credentials exists in the system and returns access token.
// If user exists, but password is incorrect, returns error.
// If user doesn`t exist, returns error.
func (auth *Auth) Login(ctx context.Context, email, password string, appID int) (string, string, error) {
	const op = "auth.Login"

	// БЕЗОПАСНОСТЬ! Мб вообще в будущем убрать логирование email
	log := auth.log.With(slog.String("op", op), slog.String("email", email))

	log.Info("attempting to login user")

	user, err := auth.userProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			auth.log.Warn("user not found", sl.Err(err))
			return "", "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}

		auth.log.Warn("failed to get user", sl.Err(err))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		auth.log.Info("invalid credentials", sl.Err(err))
		return "", "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	app, err := auth.appProvider.App(ctx, appID)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}
	log.Info("successfully logged in")

	tokenPair, err := jwt.NewTokenPair(user, app, auth.accessToken, auth.refreshToken)
	if err != nil {
		auth.log.Error("failed to generate token pair", sl.Err(err))
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	refreshTokenSave, errTokenSave := auth.tokenStorage.SaveToken(ctx, user.ID, appID, tokenPair.RefreshToken, time.Now().Add(auth.refreshToken))
	if errTokenSave != nil {
		auth.log.Error("failed to save refresh token", sl.Err(errTokenSave))
		return "", "", fmt.Errorf("%s: failed to store refresh token with id %d : %w", op, refreshTokenSave, errTokenSave)
	}

	return tokenPair.AccessToken, tokenPair.RefreshToken, nil
}

// RegisterNewUser registers new user in the system and returns user ID.
// If user with given username already exists, returns error.
func (auth *Auth) RegisterNewUser(ctx context.Context, email string, pass string) (int64, error) {
	const op = "auth.RegisterNewUser"

	// не факт, что нужно логировать email, уточнить
	log := auth.log.With(slog.String("op", op), slog.String("email", email))

	log.Info("registering user")

	// хэш пароля + соль
	passHash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate hash password", sl.Err(err))

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := auth.userSaver.SaveUser(ctx, email, passHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserAlreadyExists) {
			log.Warn("user already exists", sl.Err(err))
			return 0, fmt.Errorf("%s: %w", op, ErrUserExists)
		}
		log.Error("failed to save user", sl.Err(err))

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user registered successfully")
	return id, nil
}

// IsAdmin checks if user is admin.
func (auth *Auth) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "auth.IsAdmin"

	log := auth.log.With(slog.String("op", op), slog.String("userID", fmt.Sprint(userID)))
	log.Info("checking if user is admin")

	isAdmin, err := auth.userProvider.IsAdmin(ctx, userID)
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			log.Warn("app not found", sl.Err(err))
			return false, fmt.Errorf("%s: %w", op, ErrInvalidAppID)
		}
		return false, fmt.Errorf("%s: %w", op, err)
	}
	log.Info("user is admin", slog.Bool("isAdmin", isAdmin))
	return isAdmin, nil
}

func (auth *Auth) RefreshTokens(ctx context.Context, refreshToken string, appID int) (string, string, error) {
	const op = "auth.RefreshToken"

	log := auth.log.With(slog.String("op", op))
	log.Info("refreshing token")

	app, err := auth.appProvider.App(ctx, appID)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	token, err := jwtGo.ParseWithClaims(refreshToken, jwtGo.MapClaims{}, func(token *jwtGo.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwtGo.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(app.Secret), nil
	})
	if err != nil {
		return "", "", fmt.Errorf("%s: invalid token: %w", op, err)
	}

	claims, ok := token.Claims.(jwtGo.MapClaims)
	if !ok || !token.Valid {
		return "", "", fmt.Errorf("%s: invalid token claims", op)
	}

	if claims["typ"] != "refresh" {
		return "", "", fmt.Errorf("%s: invalid token type: expected refresh, got %v", op, claims["typ"])
	}

	email, ok := claims["email"].(string)
	if !ok {
		return "", "", fmt.Errorf("%s: email claim missing or invalid", op)
	}

	user, err := auth.userProvider.User(ctx, email)
	if err != nil {
		return "", "", fmt.Errorf("%s: failed to get user: %w", op, err)
	}

	valid, err := auth.tokenStorage.IsRefreshTokenValid(ctx, user.ID, appID, refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("%s: failed to validate refresh token: %w", op, err)
	}
	if !valid {
		return "", "", fmt.Errorf("%s: refresh token is not valid", op)
	}

	if err := auth.tokenStorage.RevokeRefreshToken(ctx, user.ID, appID, refreshToken); err != nil {
		auth.log.Warn("failed to revoke old refresh token", sl.Err(err))
		// не фейлим, но логируем
	}

	newTokens, err := jwt.NewTokenPair(user, app, auth.accessToken, auth.refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("%s: failed to generate token pair: %w", op, err)
	}

	if _, err := auth.tokenStorage.SaveToken(ctx, user.ID, appID, newTokens.RefreshToken, time.Now().Add(auth.refreshToken)); err != nil {
		log.Error("failed to save new refresh token", sl.Err(err))
		return "", "", fmt.Errorf("%s: failed to store new refresh token: %w", op, err)
	}

	log.Info("successfully refreshed tokens")

	return newTokens.AccessToken, newTokens.RefreshToken, nil
}
