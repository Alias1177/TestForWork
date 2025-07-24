package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	pb "github.com/alik/TestForWork/proto/rates"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/stats"
)

// Server represents the gRPC server
type Server struct {
	server       *grpc.Server
	logger       *zap.Logger
	port         int
	ratesHandler *RatesHandler
}

// NewServer creates a new gRPC server
func NewServer(
	ratesHandler *RatesHandler,
	logger *zap.Logger,
	port int,
	maxConnectionIdle time.Duration,
	enableMetrics bool,
	enableTracing bool,
) *Server {
	// Server options
	var opts []grpc.ServerOption

	// Keepalive settings
	opts = append(opts, grpc.KeepaliveParams(keepalive.ServerParameters{
		MaxConnectionIdle: maxConnectionIdle,
		Time:              2 * time.Minute,
		Timeout:           20 * time.Second,
	}))

	// Interceptors
	var unaryInterceptors []grpc.UnaryServerInterceptor
	var streamInterceptors []grpc.StreamServerInterceptor

	// Context tags (should be first)
	unaryInterceptors = append(unaryInterceptors, grpc_ctxtags.UnaryServerInterceptor())
	streamInterceptors = append(streamInterceptors, grpc_ctxtags.StreamServerInterceptor())

	// Logging
	unaryInterceptors = append(unaryInterceptors, grpc_zap.UnaryServerInterceptor(logger))
	streamInterceptors = append(streamInterceptors, grpc_zap.StreamServerInterceptor(logger))

	// Metrics
	if enableMetrics {
		unaryInterceptors = append(unaryInterceptors, grpc_prometheus.UnaryServerInterceptor)
		streamInterceptors = append(streamInterceptors, grpc_prometheus.StreamServerInterceptor)
	}

	// Tracing - using stats handler instead of interceptors in new otelgrpc version
	var statsHandler stats.Handler
	if enableTracing {
		statsHandler = otelgrpc.NewServerHandler()
	}

	// Recovery (should be last)
	unaryInterceptors = append(unaryInterceptors, grpc_recovery.UnaryServerInterceptor())
	streamInterceptors = append(streamInterceptors, grpc_recovery.StreamServerInterceptor())

	// Add interceptors to options
	opts = append(opts,
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(unaryInterceptors...)),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(streamInterceptors...)),
	)

	// Add tracing stats handler if enabled
	if statsHandler != nil {
		opts = append(opts, grpc.StatsHandler(statsHandler))
	}

	// Create server
	server := grpc.NewServer(opts...)

	// Register services
	pb.RegisterRatesServiceServer(server, ratesHandler)

	// Enable reflection for debugging
	reflection.Register(server)

	// Initialize metrics
	if enableMetrics {
		grpc_prometheus.Register(server)
	}

	return &Server{
		server:       server,
		logger:       logger,
		port:         port,
		ratesHandler: ratesHandler,
	}
}

// Start starts the gRPC server
func (s *Server) Start() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		s.logger.Error("Failed to create listener", zap.Error(err), zap.Int("port", s.port))
		return fmt.Errorf("failed to create listener: %w", err)
	}

	s.logger.Info("Starting gRPC server", zap.Int("port", s.port))

	if err := s.server.Serve(listener); err != nil {
		s.logger.Error("Failed to serve gRPC", zap.Error(err))
		return fmt.Errorf("failed to serve gRPC: %w", err)
	}

	return nil
}

// Stop gracefully stops the gRPC server
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Stopping gRPC server")

	stopped := make(chan struct{})
	go func() {
		s.server.GracefulStop()
		close(stopped)
	}()

	select {
	case <-ctx.Done():
		s.logger.Warn("Graceful shutdown timeout, forcing stop")
		s.server.Stop()
		return ctx.Err()
	case <-stopped:
		s.logger.Info("gRPC server stopped gracefully")
		return nil
	}
}
