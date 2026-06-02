// user service entry point.
//
// Serves gRPC and REST on a single port (8080) via h2c multiplexing.
// Cloud Run routes gRPC (HTTP/2, content-type: application/grpc) and REST (HTTP/1.1)
// to the same container port — load balancing handled by Cloud Run.
//
// Wiring order: configs → logger → otel → HTTP health first → postgres + redis →
// repository → usecase → register routes → gRPC on same port → graceful shutdown.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	usergrpc "github.com/farid/user-service/internal/user/handler/grpc"
	userhttp "github.com/farid/user-service/internal/user/handler/http"
	"github.com/farid/user-service/internal/user/model"
	userpg "github.com/farid/user-service/internal/user/repository/postgres"
	useruc "github.com/farid/user-service/internal/user/usecase"

	"github.com/farid/user-service/pkg/configs"
	pgdb "github.com/farid/user-service/pkg/db/postgres"
	"github.com/farid/user-service/pkg/grpcserver"
	"github.com/farid/user-service/pkg/idempotency"
	"github.com/farid/user-service/pkg/logger"
	pkgOtel "github.com/farid/user-service/pkg/otel"
	"github.com/farid/user-service/pkg/redis"
)

func main() {
	cfg := configs.NewConfig(configs.ConfigLoader{Env: os.Getenv("PROJECT_ENV")})
	if err := logger.NewLogger(cfg.AppName+"-user", cfg.AppEnv); err != nil {
		panic(err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	otel := pkgOtel.NewOpenTelemetry(cfg.OTLPEndpoint, "user", cfg.AppEnv)
	defer func() {
		if err := otel.EndAPM(); err != nil {
			fmt.Fprintln(os.Stderr, "otel shutdown:", err)
		}
	}()
	_ = otel.RegisterRuntimeMetrics()

	// ── HTTP router (start ASAP so Cloud Run health checks pass) ─────────────
	if cfg.AppEnv != "local" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	router.Use(otelgin.Middleware(cfg.AppName))
	router.Use(gin.Recovery(), cors.Default())

	var ready atomic.Bool
	router.GET("/health", func(c *gin.Context) {
		if ready.Load() {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		} else {
			c.JSON(http.StatusOK, gin.H{"status": "starting"})
		}
	})

	// ── gRPC server (created early, registered after usecase ready) ──────────
	grpcSrv, grpcErr := grpcserver.NewGrpcServerNoListen(grpcserver.Options{
		IdempotentMethods: []string{
			model.ScopeCreateUser,
			model.ScopeUpdateUser,
			model.ScopeUpsertDriver,
			model.ScopeRegisterVehicle,
		},
	})

	// ── Multiplexed HTTP server (gRPC + REST on same port) ───────────────────
	mux := grpcHTTPMux(grpcSrv, router)
	httpSrv := &http.Server{
		Addr:              ":" + cfg.AppPort,
		Handler:           h2c.NewHandler(mux, &http2.Server{}),
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		logger.Info(ctx, fmt.Sprintf("user-service listening on :%s (gRPC+HTTP)", cfg.AppPort), nil)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal(ctx, "listen failed", map[string]interface{}{logger.ErrorKey: err.Error()})
		}
	}()

	// ── Infra (after HTTP is serving) ────────────────────────────────────────
	db, err := pgdb.NewPostgresDB(pgdb.PostgresDsn{
		Host: cfg.DbHost, Port: cfg.DbPort, User: cfg.DbUsername, Password: cfg.DbPassword, Db: cfg.DbName,
		MaxOpen: cfg.DbMaxOpen, MaxIdle: cfg.DbMaxIdle,
	})
	if err != nil {
		logger.Error(ctx, "postgres init failed (DB calls will fail)", map[string]interface{}{logger.ErrorKey: err.Error()})
	}
	defer func() {
		if db != nil {
			_ = db.Close()
		}
	}()

	cache := redis.InitConnection(cfg.RedisDB, cfg.RedisHost, cfg.RedisPort, cfg.RedisPassword, cfg.RedisAppConfig)
	if pingErr := cache.Ping(ctx); pingErr != nil {
		logger.Warn(ctx, "redis ping failed (continuing without cache)", map[string]interface{}{logger.ErrorKey: pingErr.Error()})
	}

	// ── Domain wiring ────────────────────────────────────────────────────────
	repo := userpg.NewUserRepository(db, cfg.PgCryptoKey)
	vehicleRepo := userpg.NewVehicleRepository(db)
	uc := useruc.NewUserUsecase(repo, vehicleRepo, cache)

	// Register HTTP routes
	userhttp.RegisterUserHandler(router.Group("/v1"), uc, cfg.SuperAppJWTPubKey)

	// Register gRPC handlers (idempotency store needs DB)
	if grpcErr == nil && grpcSrv != nil {
		if db != nil {
			grpcSrv.SetIdempotencyStore(idempotency.NewPostgresStore(db))
		}
		usergrpc.RegisterUserHandler(grpcSrv.Server, uc)
	}

	ready.Store(true)
	logger.Info(ctx, "user-service fully initialized", nil)

	// ── Graceful shutdown ────────────────────────────────────────────────────
	<-ctx.Done()
	logger.Info(context.Background(), "shutdown signal received", nil)

	shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if grpcSrv != nil {
		grpcSrv.Server.GracefulStop()
	}
	if err := httpSrv.Shutdown(shutCtx); err != nil {
		logger.Error(context.Background(), "http shutdown error", map[string]interface{}{logger.ErrorKey: err.Error()})
	}
	if err := logger.Sync(); err != nil {
		fmt.Fprintln(os.Stderr, "logger sync:", err)
	}
}

// grpcHTTPMux routes gRPC requests to grpcServer, everything else to httpHandler.
func grpcHTTPMux(grpcSrv *grpcserver.GrpcServer, httpHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if grpcSrv != nil && r.ProtoMajor == 2 && strings.HasPrefix(r.Header.Get("Content-Type"), "application/grpc") {
			grpcSrv.Server.ServeHTTP(w, r)
			return
		}
		httpHandler.ServeHTTP(w, r)
	})
}
