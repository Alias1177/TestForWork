package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alik/TestForWork/internal/api/grpc"
	"github.com/alik/TestForWork/internal/client"
	pb "github.com/alik/TestForWork/proto/rates"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MockRatesService is a mock implementation of RatesService
type MockRatesService struct {
	mock.Mock
}

func (m *MockRatesService) GetRates(ctx context.Context, market string) (*client.RateData, error) {
	args := m.Called(ctx, market)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.RateData), args.Error(1)
}

func (m *MockRatesService) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestRatesHandler_GetRates(t *testing.T) {
	tests := []struct {
		name           string
		request        *pb.GetRatesRequest
		setupMocks     func(*MockRatesService)
		expectError    bool
		expectedCode   codes.Code
		expectedAsk    string
		expectedBid    string
		expectedMarket string
	}{
		{
			name: "successful request",
			request: &pb.GetRatesRequest{
				Market: "usdtrub",
			},
			setupMocks: func(service *MockRatesService) {
				rateData := &client.RateData{
					Ask:       "95.5",
					Bid:       "95.3",
					Market:    "usdtrub",
					Timestamp: time.Now(),
				}
				service.On("GetRates", mock.Anything, "usdtrub").Return(rateData, nil)
			},
			expectError:    false,
			expectedAsk:    "95.5",
			expectedBid:    "95.3",
			expectedMarket: "usdtrub",
		},
		{
			name: "empty market",
			request: &pb.GetRatesRequest{
				Market: "",
			},
			setupMocks:   func(service *MockRatesService) {},
			expectError:  true,
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "service error",
			request: &pb.GetRatesRequest{
				Market: "usdtrub",
			},
			setupMocks: func(service *MockRatesService) {
				service.On("GetRates", mock.Anything, "usdtrub").Return(nil, errors.New("service error"))
			},
			expectError:  true,
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockService := new(MockRatesService)
			tt.setupMocks(mockService)

			// Create handler
			logger := zap.NewNop()
			handler := grpc.NewRatesHandler(mockService, logger, "1.0.0")

			// Execute
			ctx := context.Background()
			response, err := handler.GetRates(ctx, tt.request)

			// Assert
			if tt.expectError {
				require.Error(t, err)
				assert.Nil(t, response)

				// Check gRPC status code
				grpcErr, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.expectedCode, grpcErr.Code())
			} else {
				require.NoError(t, err)
				require.NotNil(t, response)
				assert.Equal(t, tt.expectedAsk, response.Ask)
				assert.Equal(t, tt.expectedBid, response.Bid)
				assert.Equal(t, tt.expectedMarket, response.Market)
				assert.NotNil(t, response.Timestamp)
			}

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestRatesHandler_Healthcheck(t *testing.T) {
	tests := []struct {
		name           string
		setupMocks     func(*MockRatesService)
		expectedStatus string
	}{
		{
			name: "healthy service",
			setupMocks: func(service *MockRatesService) {
				service.On("HealthCheck", mock.Anything).Return(nil)
			},
			expectedStatus: "healthy",
		},
		{
			name: "unhealthy service",
			setupMocks: func(service *MockRatesService) {
				service.On("HealthCheck", mock.Anything).Return(errors.New("service error"))
			},
			expectedStatus: "unhealthy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockService := new(MockRatesService)
			tt.setupMocks(mockService)

			// Create handler
			logger := zap.NewNop()
			handler := grpc.NewRatesHandler(mockService, logger, "1.0.0")

			// Execute
			ctx := context.Background()
			response, err := handler.Healthcheck(ctx, &pb.HealthcheckRequest{})

			// Assert
			require.NoError(t, err)
			require.NotNil(t, response)
			assert.Equal(t, tt.expectedStatus, response.Status)
			assert.Equal(t, "1.0.0", response.Version)
			assert.NotNil(t, response.Timestamp)

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}
