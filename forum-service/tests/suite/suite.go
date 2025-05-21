package suite

import (
	"context"
	"github.com/14kear/forum-project/forum-service/internal/services/forum"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	authTestsuite "github.com/14kear/forum-project/auth-service/pkg/testsuite"
	"github.com/14kear/forum-project/forum-service/internal/app"
	"github.com/14kear/forum-project/forum-service/internal/config"
	"github.com/14kear/forum-project/forum-service/utils"
	ssov1 "github.com/14kear/forum-project/protos/gen/go/auth"
)

type Suite struct {
	*testing.T
	Cfg          *config.Config
	App          *app.App
	Server       *httptest.Server
	HTTPClient   *http.Client
	BaseURL      string
	AuthClient   ssov1.AuthClient
	ForumService *forum.Forum

	AuthSuite *authTestsuite.Suite
}

func New(t *testing.T) (context.Context, *Suite) {
	t.Helper()
	t.Parallel()

	// Поднимаем auth-service grpc-сервер через общий testsuite
	authSuite := authTestsuite.New(t)
	addr := authSuite.GRPCaddr

	cfg := config.Load("../config/local.yaml")
	log := utils.New(cfg.Env)
	application := app.NewApp(log, cfg.HTTP.Port, cfg.StoragePath, addr)

	engine := application.HTTPServer.Engine()
	testServer := httptest.NewServer(engine)

	t.Cleanup(func() {
		testServer.Close()
		application.Stop(context.Background())
		authSuite.Cancel()              // Отмена контекста из auth suite
		authSuite.App.GRPCServer.Stop() // Остановка grpc сервера auth-service
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	return ctx, &Suite{
		T:            t,
		Cfg:          cfg,
		App:          application,
		Server:       testServer,
		HTTPClient:   testServer.Client(),
		BaseURL:      testServer.URL,
		AuthClient:   authSuite.AuthClient,
		AuthSuite:    authSuite,
		ForumService: application.Forum,
	}
}
