package service

import (
	"context"
	"fmt"
	"time"

	"github.com/alik/TestForWork/internal/client"
	"github.com/alik/TestForWork/internal/storage/postgres"
	"go.uber.org/zap"
)

// RatesService handles business logic for exchange rates
type RatesService struct {
	grinexClient GrinexClient
	repository   Repository
	logger       *zap.Logger
}

// NewRatesService creates a new rates service
func NewRatesService(grinexClient GrinexClient, repository Repository, logger *zap.Logger) *RatesService {
	return &RatesService{
		grinexClient: grinexClient,
		repository:   repository,
		logger:       logger,
	}
}

// GetRates retrieves exchange rates and saves them to the database
func (s *RatesService) GetRates(ctx context.Context, market string) (*client.RateData, error) {
	s.logger.Info("Getting rates for market", zap.String("market", market))

	// Get rates from Grinex API
	rateData, err := s.grinexClient.GetRates(ctx, market)
	if err != nil {
		s.logger.Error("Failed to get rates from Grinex", zap.Error(err))
		return nil, fmt.Errorf("failed to get rates from Grinex: %w", err)
	}

	// Save to database
	if err := s.repository.SaveRate(ctx, rateData.Market, rateData.Ask, rateData.Bid, rateData.Timestamp); err != nil {
		s.logger.Error("Failed to save rate to database", zap.Error(err))
		// Don't return error here - we still want to return the rate data
		// even if saving to DB fails
	}

	s.logger.Info("Successfully retrieved and saved rates",
		zap.String("market", rateData.Market),
		zap.String("ask", rateData.Ask),
		zap.String("bid", rateData.Bid))

	return rateData, nil
}

// GetLatestRate retrieves the latest rate from the database
func (s *RatesService) GetLatestRate(ctx context.Context, market string) (*postgres.Rate, error) {
	s.logger.Debug("Getting latest rate from database", zap.String("market", market))

	rate, err := s.repository.GetLatestRate(ctx, market)
	if err != nil {
		s.logger.Error("Failed to get latest rate from database", zap.Error(err))
		return nil, fmt.Errorf("failed to get latest rate: %w", err)
	}

	return rate, nil
}

// GetRatesHistory retrieves historical rates from the database
func (s *RatesService) GetRatesHistory(ctx context.Context, market string, limit, offset int) ([]postgres.Rate, error) {
	s.logger.Debug("Getting rates history from database",
		zap.String("market", market),
		zap.Int("limit", limit),
		zap.Int("offset", offset))

	rates, err := s.repository.GetRates(ctx, market, limit, offset)
	if err != nil {
		s.logger.Error("Failed to get rates history from database", zap.Error(err))
		return nil, fmt.Errorf("failed to get rates history: %w", err)
	}

	return rates, nil
}

// HealthCheck checks the health of the service
func (s *RatesService) HealthCheck(ctx context.Context) error {
	s.logger.Debug("Performing health check")

	// Check database connection
	if err := s.repository.Ping(ctx); err != nil {
		s.logger.Error("Database health check failed", zap.Error(err))
		return fmt.Errorf("database health check failed: %w", err)
	}

	// Try to get rates from Grinex API (with timeout)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := s.grinexClient.GetRates(ctx, "usdtrub") // Use default market for health check
	if err != nil {
		s.logger.Warn("Grinex API health check failed", zap.Error(err))
		// Don't fail the health check if external API is down
		// as this might be temporary
	}

	s.logger.Debug("Health check completed successfully")
	return nil
}
