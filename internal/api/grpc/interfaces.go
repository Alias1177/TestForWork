package grpc

import (
	"context"

	"github.com/alik/TestForWork/internal/client"
)

// RatesService interface for the service layer
type RatesService interface {
	GetRates(ctx context.Context, market string) (*client.RateData, error)
	HealthCheck(ctx context.Context) error
}
