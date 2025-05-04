package grpcapp

import (
	"fmt"
	authgrpc "github.com/14kear/forum-project/auth-service/internal/grpc/auth"
	"google.golang.org/grpc"
	"log/slog"
	"net"
)

type App struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       int
}

func NewApp(log *slog.Logger, authService authgrpc.Auth, port int) *App {
	//gRPCServer := grpc.NewServer(grpc.ChainUnaryInterceptor(authInterceptor(authService)))
	gRPCServer := grpc.NewServer()

	authgrpc.Register(gRPCServer, authService)
	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		port:       port,
	}
}

// если не запустился сервер, то паникуем
func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "grpcapp.Run"

	log := a.log.With(slog.String("op", op), slog.Int("port", a.port))

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("gRPC server is running", slog.String("addr", lis.Addr().String()))

	if err := a.gRPCServer.Serve(lis); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) Stop() {
	const op = "grpcapp.Stop"

	a.log.With(slog.String("op", op)).Info("gRPC server is stopping")
	a.gRPCServer.GracefulStop() // блокирует выполнение кода пока не обработаются текущие соединения
}

//// interceptors (middleware) для проверки access tokens
//func (a *App) authInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
//	// Пропускаем методы, которые не требуют аутентификации
//	if info.FullMethod == "/auth.Auth/Login" ||
//		info.FullMethod == "/auth.Auth/Register" ||
//		info.FullMethod == "/auth.Auth/RefreshToken" {
//		return handler(ctx, req)
//	}
//
//	md, ok := metadata.FromIncomingContext(ctx)
//	if !ok {
//		return nil, status.Error(codes.Unauthenticated, "metadata not provided")
//	}
//
//	appID, err := getAppIDFromMetadata(md)
//	if err != nil {
//		return nil, status.Errorf(codes.Unauthenticated, "appID: %v", err)
//	}
//
//	secret, err := a.authSvc.GetAppSecret(ctx, appID)
//	if err != nil {
//		return nil, status.Errorf(codes.Unauthenticated, "invalid app: %v", err)
//	}
//
//	accessToken, err := getTokenFromMetadata(md, "refresh_token")
//	if err != nil {
//		return nil, status.Errorf(codes.Unauthenticated, "refresh token: %v", err)
//	}
//}
//
//func getAppIDFromMetadata(md metadata.MD) (int, error) {
//	appIDs := md.Get("x-app-id")
//	if len(appIDs) == 0 {
//		return 0, fmt.Errorf("app id not provided")
//	}
//
//	appID, err := strconv.Atoi(appIDs[0])
//	if err != nil {
//		return 0, fmt.Errorf("invalid app id")
//	}
//
//	return appID, nil
//}
//
//func getTokenFromMetadata(md metadata.MD, header string) (string, error) {
//	headers := md.Get(header)
//	if len(headers) == 0 {
//		return "", fmt.Errorf("header %s not provided", header)
//	}
//
//	token := strings.TrimPrefix(headers[0], "Bearer ")
//	if token == "" {
//		return "", fmt.Errorf("empty token")
//	}
//
//	return token, nil
//}

//func AccessLogInterceptor(log *slog.Logger) grpc.UnaryServerInterceptor {
//	return func(
//		ctx context.Context,
//		req interface{},
//		info *grpc.UnaryServerInfo,
//		handler grpc.UnaryHandler,
//	) (interface{}, error) {
//		md, ok := metadata.FromIncomingContext(ctx)
//		if ok {
//			if authHeaders := md.Get("authorization"); len(authHeaders) > 0 {
//				log.Info("Authorization header", slog.String("value", authHeaders[0]))
//			} else {
//				log.Info("Authorization header not found")
//			}
//		}
//		return handler(ctx, req)
//	}
//}
