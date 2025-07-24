package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alik/TestForWork/internal/client"
	"github.com/alik/TestForWork/internal/service"
	"github.com/alik/TestForWork/internal/storage/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// MockGrinexClient is a mock implementation of GrinexClient
type MockGrinexClient struct {
	mock.Mock
}

func (m *MockGrinexClient) GetRates(ctx context.Context, market string) (*client.RateData, error) {
	args := m.Called(ctx, market)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.RateData), args.Error(1)
}

// MockRepository is a mock implementation of Repository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) SaveRate(ctx context.Context, market, ask, bid string, timestamp time.Time) error {
	args := m.Called(ctx, market, ask, bid, timestamp)
	return args.Error(0)
}

func (m *MockRepository) GetRates(ctx context.Context, market string, limit, offset int) ([]postgres.Rate, error) {
	args := m.Called(ctx, market, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]postgres.Rate), args.Error(1)
}

func (m *MockRepository) GetLatestRate(ctx context.Context, market string) (*postgres.Rate, error) {
	args := m.Called(ctx, market)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*postgres.Rate), args.Error(1)
}

func (m *MockRepository) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestRatesService_GetRates(t *testing.T) {
	tests := []struct {
		name           string
		setupMocks     func(*MockGrinexClient, *MockRepository)
		market         string
		expectError    bool
		expectedAsk    string
		expectedBid    string
		expectedMarket string
	}{
		{
			name: "successful get rates",
			setupMocks: func(grinex *MockGrinexClient, repo *MockRepository) {
				rateData := &client.RateData{
					Ask:       "95.5",
					Bid:       "95.3",
					Market:    "usdtrub",
					Timestamp: time.Now(),
				}
				grinex.On("GetRates", mock.Anything, "usdtrub").Return(rateData, nil)
				repo.On("SaveRate", mock.Anything, "usdtrub", "95.5", "95.3", mock.Anything).Return(nil)
			},
			market:         "usdtrub",
			expectError:    false,
			expectedAsk:    "95.5",
			expectedBid:    "95.3",
			expectedMarket: "usdtrub",
		},
		{
			name: "grinex client error",
			setupMocks: func(grinex *MockGrinexClient, repo *MockRepository) {
				grinex.On("GetRates", mock.Anything, "usdtrub").Return(nil, errors.New("API error"))
			},
			market:      "usdtrub",
			expectError: true,
		},
		{
			name: "save rate error (should not fail request)",
			setupMocks: func(grinex *MockGrinexClient, repo *MockRepository) {
				rateData := &client.RateData{
					Ask:       "95.5",
					Bid:       "95.3",
					Market:    "usdtrub",
					Timestamp: time.Now(),
				}
				grinex.On("GetRates", mock.Anything, "usdtrub").Return(rateData, nil)
				repo.On("SaveRate", mock.Anything, "usdtrub", "95.5", "95.3", mock.Anything).Return(errors.New("DB error"))
			},
			market:         "usdtrub",
			expectError:    false, // Should not fail even if DB save fails
			expectedAsk:    "95.5",
			expectedBid:    "95.3",
			expectedMarket: "usdtrub",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockGrinex := new(MockGrinexClient)
			mockRepo := new(MockRepository)
			tt.setupMocks(mockGrinex, mockRepo)

			// Create service
			logger := zap.NewNop()
			s := service.NewRatesService(mockGrinex, mockRepo, logger)

			// Execute
			ctx := context.Background()
			rateData, err := s.GetRates(ctx, tt.market)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, rateData)
			} else {
				require.NoError(t, err)
				require.NotNil(t, rateData)
				assert.Equal(t, tt.expectedAsk, rateData.Ask)
				assert.Equal(t, tt.expectedBid, rateData.Bid)
				assert.Equal(t, tt.expectedMarket, rateData.Market)
			}

			// Verify mock expectations
			mockGrinex.AssertExpectations(t)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestRatesService_GetLatestRate(t *testing.T) {
	tests := []struct {
		name        string
		setupMocks  func(*MockRepository)
		market      string
		expectError bool
	}{
		{
			name: "successful get latest rate",
			setupMocks: func(repo *MockRepository) {
				rate := &postgres.Rate{
					ID:        1,
					Market:    "usdtrub",
					Ask:       "95.5",
					Bid:       "95.3",
					Timestamp: time.Now(),
					CreatedAt: time.Now(),
				}
				repo.On("GetLatestRate", mock.Anything, "usdtrub").Return(rate, nil)
			},
			market:      "usdtrub",
			expectError: false,
		},
		{
			name: "repository error",
			setupMocks: func(repo *MockRepository) {
				repo.On("GetLatestRate", mock.Anything, "usdtrub").Return(nil, errors.New("DB error"))
			},
			market:      "usdtrub",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockGrinex := new(MockGrinexClient)
			mockRepo := new(MockRepository)
			tt.setupMocks(mockRepo)

			// Create service
			logger := zap.NewNop()
			s := service.NewRatesService(mockGrinex, mockRepo, logger)

			// Execute
			ctx := context.Background()
			rate, err := s.GetLatestRate(ctx, tt.market)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, rate)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, rate)
			}

			// Verify mock expectations
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestRatesService_HealthCheck(t *testing.T) {
	tests := []struct {
		name        string
		setupMocks  func(*MockGrinexClient, *MockRepository)
		expectError bool
	}{
		{
			name: "healthy service",
			setupMocks: func(grinex *MockGrinexClient, repo *MockRepository) {
				repo.On("Ping", mock.Anything).Return(nil)
				grinex.On("GetRates", mock.Anything, "usdtrub").Return(&client.RateData{}, nil)
			},
			expectError: false,
		},
		{
			name: "database unhealthy",
			setupMocks: func(grinex *MockGrinexClient, repo *MockRepository) {
				repo.On("Ping", mock.Anything).Return(errors.New("DB error"))
			},
			expectError: true,
		},
		{
			name: "api unhealthy but service still healthy",
			setupMocks: func(grinex *MockGrinexClient, repo *MockRepository) {
				repo.On("Ping", mock.Anything).Return(nil)
				grinex.On("GetRates", mock.Anything, "usdtrub").Return(nil, errors.New("API error"))
			},
			expectError: false, // API error should not fail health check
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockGrinex := new(MockGrinexClient)
			mockRepo := new(MockRepository)
			tt.setupMocks(mockGrinex, mockRepo)

			// Create service
			logger := zap.NewNop()
			s := service.NewRatesService(mockGrinex, mockRepo, logger)

			// Execute
			ctx := context.Background()
			err := s.HealthCheck(ctx)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify mock expectations
			mockGrinex.AssertExpectations(t)
			mockRepo.AssertExpectations(t)
		})
	}
}
