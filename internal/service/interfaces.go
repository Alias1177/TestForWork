package service

import (
	"context"
	"time"

	"github.com/alik/TestForWork/internal/client"
	"github.com/alik/TestForWork/internal/storage/postgres"
)

// GrinexClient interface for Grinex API client
type GrinexClient interface {
	GetRates(ctx context.Context, market string) (*client.RateData, error)
}

// Repository interface for data storage
type Repository interface {
	SaveRate(ctx context.Context, market, ask, bid string, timestamp time.Time) error
	GetRates(ctx context.Context, market string, limit, offset int) ([]postgres.Rate, error)
	GetLatestRate(ctx context.Context, market string) (*postgres.Rate, error)
	Ping(ctx context.Context) error
}
