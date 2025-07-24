package grpc

import (
	"context"
	"time"

	pb "github.com/alik/TestForWork/proto/rates"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// RatesHandler implements the gRPC RatesService
type RatesHandler struct {
	pb.UnimplementedRatesServiceServer
	ratesService RatesService
	logger       *zap.Logger
	version      string
}

// NewRatesHandler creates a new gRPC rates handler
func NewRatesHandler(ratesService RatesService, logger *zap.Logger, version string) *RatesHandler {
	return &RatesHandler{
		ratesService: ratesService,
		logger:       logger,
		version:      version,
	}
}

// GetRates handles the GetRates gRPC request
func (h *RatesHandler) GetRates(ctx context.Context, req *pb.GetRatesRequest) (*pb.GetRatesResponse, error) {
	h.logger.Info("GetRates request received", zap.String("market", req.Market))

	// Validate request
	if req.Market == "" {
		h.logger.Warn("Empty market in request")
		return nil, status.Error(codes.InvalidArgument, "market is required")
	}

	// Get rates from service
	rateData, err := h.ratesService.GetRates(ctx, req.Market)
	if err != nil {
		h.logger.Error("Failed to get rates", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get rates")
	}

	// Convert to protobuf response
	response := &pb.GetRatesResponse{
		Ask:       rateData.Ask,
		Bid:       rateData.Bid,
		Timestamp: timestamppb.New(rateData.Timestamp),
		Market:    rateData.Market,
	}

	h.logger.Info("GetRates request completed successfully",
		zap.String("market", req.Market),
		zap.String("ask", rateData.Ask),
		zap.String("bid", rateData.Bid))

	return response, nil
}

// Healthcheck handles the Healthcheck gRPC request
func (h *RatesHandler) Healthcheck(ctx context.Context, req *pb.HealthcheckRequest) (*pb.HealthcheckResponse, error) {
	h.logger.Debug("Healthcheck request received")

	// Perform health check
	err := h.ratesService.HealthCheck(ctx)
	serviceStatus := "healthy"
	if err != nil {
		h.logger.Warn("Health check failed", zap.Error(err))
		serviceStatus = "unhealthy"
	}

	response := &pb.HealthcheckResponse{
		Status:    serviceStatus,
		Version:   h.version,
		Timestamp: timestamppb.New(time.Now()),
	}

	h.logger.Debug("Healthcheck request completed", zap.String("status", serviceStatus))

	return response, nil
}
