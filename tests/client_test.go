package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alik/TestForWork/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestGrinexClient_GetRates(t *testing.T) {
	tests := []struct {
		name           string
		responseStatus int
		responseBody   interface{}
		expectedAsk    string
		expectedBid    string
		expectError    bool
	}{
		{
			name:           "successful response",
			responseStatus: http.StatusOK,
			responseBody: client.DepthResponse{
				Asks:      []client.OrderBook{{Price: "95.5", Amount: "1000"}},
				Bids:      []client.OrderBook{{Price: "95.3", Amount: "800"}},
				Timestamp: time.Now().Unix(),
			},
			expectedAsk: "95.5",
			expectedBid: "95.3",
			expectError: false,
		},
		{
			name:           "empty response",
			responseStatus: http.StatusOK,
			responseBody: client.DepthResponse{
				Asks:      []client.OrderBook{},
				Bids:      []client.OrderBook{},
				Timestamp: time.Now().Unix(),
			},
			expectedAsk: "N/A",
			expectedBid: "N/A",
			expectError: false,
		},
		{
			name:           "server error",
			responseStatus: http.StatusInternalServerError,
			responseBody:   "Internal Server Error",
			expectedAsk:    "",
			expectedBid:    "",
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/v2/depth", r.URL.Path)
				assert.Equal(t, "usdtrub", r.URL.Query().Get("market"))
				assert.Equal(t, "application/json", r.Header.Get("Accept"))

				w.WriteHeader(tt.responseStatus)
				if tt.responseStatus == http.StatusOK {
					if err := json.NewEncoder(w).Encode(tt.responseBody); err != nil {
						t.Errorf("Failed to encode response: %v", err)
					}
				} else {
					if _, err := w.Write([]byte(tt.responseBody.(string))); err != nil {
						t.Errorf("Failed to write response: %v", err)
					}
				}
			}))
			defer server.Close()

			// Create client
			logger := zap.NewNop()
			c := client.NewGrinexClient(server.URL, "usdtrub", 5*time.Second, logger)

			// Execute
			ctx := context.Background()
			rateData, err := c.GetRates(ctx, "usdtrub")

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, rateData)
			} else {
				require.NoError(t, err)
				require.NotNil(t, rateData)
				assert.Equal(t, tt.expectedAsk, rateData.Ask)
				assert.Equal(t, tt.expectedBid, rateData.Bid)
				assert.Equal(t, "usdtrub", rateData.Market)
				assert.NotZero(t, rateData.Timestamp)
			}
		})
	}
}

func TestGrinexClient_GetRates_Timeout(t *testing.T) {
	// Create server with delay
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(client.DepthResponse{
			Asks:      []client.OrderBook{{Price: "95.5", Amount: "1000"}},
			Bids:      []client.OrderBook{{Price: "95.3", Amount: "800"}},
			Timestamp: time.Now().Unix(),
		}); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	// Create client with short timeout
	logger := zap.NewNop()
	c := client.NewGrinexClient(server.URL, "usdtrub", 100*time.Millisecond, logger)

	// Execute
	ctx := context.Background()
	rateData, err := c.GetRates(ctx, "usdtrub")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, rateData)
	assert.Contains(t, err.Error(), "deadline exceeded")
}

func TestGrinexClient_GetRates_ContextCanceled(t *testing.T) {
	// Create server with delay
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create client
	logger := zap.NewNop()
	c := client.NewGrinexClient(server.URL, "usdtrub", 5*time.Second, logger)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Execute
	rateData, err := c.GetRates(ctx, "usdtrub")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, rateData)
}
