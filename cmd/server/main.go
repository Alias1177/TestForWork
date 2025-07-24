package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	migrate_postgres "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.uber.org/zap"

	"github.com/alik/TestForWork/internal/api/grpc"
	"github.com/alik/TestForWork/internal/client"
	"github.com/alik/TestForWork/internal/config"
	"github.com/alik/TestForWork/internal/service"
	"github.com/alik/TestForWork/internal/storage/postgres"
	"github.com/alik/TestForWork/pkg/logger"
)

const version = "1.0.0"

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log, err := logger.New(cfg.Logging.Level, cfg.Logging.Format)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Close()

	log.Info("Starting USDT Rates Service", zap.String("version", version))

	// Initialize tracing
	var shutdown func()
	if cfg.Tracing.Enabled {
		var err error
		shutdown, err = initTracing(cfg.Tracing.JaegerURL, cfg.Tracing.ServiceName)
		if err != nil {
			log.Error("Failed to initialize tracing", zap.Error(err))
		} else {
			defer shutdown()
			log.Info("Tracing initialized", zap.String("jaeger_url", cfg.Tracing.JaegerURL))
		}
	}

	// Initialize services
	_, grpcServer, metricsServer, err := initializeServices(cfg, log)
	if err != nil {
		log.Error("Failed to initialize services", zap.Error(err))
		os.Exit(1)
	}

	// Run the server
	runServer(grpcServer, metricsServer, log)
}

// initializeServices initializes all application services
func initializeServices(cfg *config.Config, log *logger.Logger) (*service.RatesService, *grpc.Server, *http.Server, error) {
	// Initialize database
	db, err := postgres.NewDB(
		cfg.Database.DatabaseDSN(),
		cfg.Database.MaxOpenConns,
		cfg.Database.MaxIdleConns,
		cfg.Database.ConnMaxLifetime,
		log.Logger,
	)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Run migrations
	if err := runMigrations(db, cfg.Database.DatabaseDSN(), log.Logger); err != nil {
		db.Close()
		return nil, nil, nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Initialize repository
	repo := postgres.NewRepository(db, log.Logger)

	// Initialize Grinex client
	grinexClient := client.NewGrinexClient(
		cfg.Grinex.BaseURL,
		cfg.Grinex.Market,
		cfg.Grinex.Timeout,
		log.Logger,
	)

	// Initialize service
	ratesService := service.NewRatesService(grinexClient, repo, log.Logger)

	// Initialize gRPC handler
	ratesHandler := grpc.NewRatesHandler(ratesService, log.Logger, version)

	// Initialize gRPC server
	grpcServer := grpc.NewServer(
		ratesHandler,
		log.Logger,
		cfg.Server.Port,
		cfg.Server.MaxConnectionIdle,
		cfg.Metrics.Enabled,
		cfg.Tracing.Enabled,
	)

	// Start metrics server if enabled
	var metricsServer *http.Server
	if cfg.Metrics.Enabled {
		metricsServer = startMetricsServer(cfg.Metrics.Port, cfg.Metrics.Path, log.Logger)
	}

	return ratesService, grpcServer, metricsServer, nil
}

// runServer runs the gRPC server and handles graceful shutdown
func runServer(grpcServer *grpc.Server, metricsServer *http.Server, log *logger.Logger) {
	// Start gRPC server in a goroutine
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- grpcServer.Start()
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		log.Error("Server error", zap.Error(err))
	case sig := <-sigChan:
		log.Info("Received signal, shutting down", zap.String("signal", sig.String()))
	}

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Stop gRPC server
	if err := grpcServer.Stop(shutdownCtx); err != nil {
		log.Error("Failed to stop gRPC server gracefully", zap.Error(err))
	}

	// Stop metrics server
	if metricsServer != nil {
		if err := metricsServer.Shutdown(shutdownCtx); err != nil {
			log.Error("Failed to stop metrics server gracefully", zap.Error(err))
		}
	}

	log.Info("Service stopped")
}

// initTracing initializes OpenTelemetry tracing
func initTracing(jaegerURL, serviceName string) (func(), error) {
	// Convert Jaeger URL to OTLP endpoint
	otlpEndpoint := jaegerURL
	if jaegerURL == "http://jaeger:14268/api/traces" {
		otlpEndpoint = "http://jaeger:4318"
	}

	exp, err := otlptracehttp.New(context.Background(),
		otlptracehttp.WithEndpoint(otlpEndpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String(version),
		)),
	)

	otel.SetTracerProvider(tp)

	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			// Logger might be closed at this point, use fmt
			fmt.Printf("Failed to shutdown tracer provider: %v\n", err)
		}
	}, nil
}

// startMetricsServer starts the Prometheus metrics server
func startMetricsServer(port int, path string, logger *zap.Logger) *http.Server {
	mux := http.NewServeMux()
	mux.Handle(path, promhttp.HandlerFor(
		prometheus.DefaultGatherer,
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		},
	))

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		logger.Info("Starting metrics server", zap.Int("port", port), zap.String("path", path))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Metrics server error", zap.Error(err))
		}
	}()

	return server
}

// runMigrations runs database migrations
func runMigrations(db *sql.DB, _ string, logger *zap.Logger) error {
	driver, err := migrate_postgres.WithInstance(db, &migrate_postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://internal/storage/migrations",
		"postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	logger.Info("Running database migrations")

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	logger.Info("Database migrations completed successfully")
	return nil
}
