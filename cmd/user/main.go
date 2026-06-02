// user service entry point.
//
// Runs BOTH a gRPC server (internal s2s on :9094) and a REST HTTP server (mini app on :8080).
// Both are wired to the same usecase so business rules stay consistent across protocols.
//
// Wiring order: configs → logger → otel → HTTP healthz first → postgres + redis →
// repository → usecase → register routes → gRPC server → graceful shutdown.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/farid/user-service/internal/user/handler/grpc"
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

	// ── HTTP server (start ASAP so Cloud Run health checks pass) ─────────────
	if cfg.AppEnv != "local" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	router.Use(otelgin.Middleware(cfg.AppName))
	router.Use(gin.Recovery(), cors.Default())

	var ready atomic.Bool
	router.GET("/healthz", func(c *gin.Context) {
		if ready.Load() {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		} else {
			c.JSON(http.StatusOK, gin.H{"status": "starting"})
		}
	})

	httpSrv := &http.Server{
		Addr:              ":" + cfg.AppPort,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		logger.Info(ctx, fmt.Sprintf("user HTTP listening on :%s", cfg.AppPort), nil)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal(ctx, "http listen failed", map[string]interface{}{logger.ErrorKey: err.Error()})
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

	// Register routes (after usecase is ready)
	userhttp.RegisterUserHandler(router.Group("/v1"), uc, cfg.SuperAppJWTPubKey)
	ready.Store(true)

	// ── gRPC server ──────────────────────────────────────────────────────────
	var grpcSrv *grpcserver.GrpcServer
	grpcSrv, err = grpcserver.NewGrpcServer(cfg.GrpcPort, grpcserver.Options{
		IdempotencyStore: idempotency.NewPostgresStore(db),
		IdempotentMethods: []string{
			model.ScopeCreateUser,
			model.ScopeUpdateUser,
			model.ScopeUpsertDriver,
			model.ScopeRegisterVehicle,
		},
	})
	if err != nil {
		logger.Warn(ctx, "grpc server init failed (HTTP-only mode)", map[string]interface{}{logger.ErrorKey: err.Error()})
	} else {
		grpc.RegisterUserHandler(grpcSrv.Server, uc)
		go func() {
			if err := grpcSrv.Start(); err != nil {
				logger.Error(ctx, "grpc serve failed", map[string]interface{}{logger.ErrorKey: err.Error()})
			}
		}()
	}

	logger.Info(ctx, "user-service fully initialized", nil)

	// ── Graceful shutdown ────────────────────────────────────────────────────
	<-ctx.Done()
	logger.Info(context.Background(), "shutdown signal received", nil)

	shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpSrv.Shutdown(shutCtx); err != nil {
		logger.Error(context.Background(), "http shutdown error", map[string]interface{}{logger.ErrorKey: err.Error()})
	}
	if grpcSrv != nil {
		grpcSrv.Shutdown()
	}
	if err := logger.Sync(); err != nil {
		fmt.Fprintln(os.Stderr, "logger sync:", err)
	}
}
