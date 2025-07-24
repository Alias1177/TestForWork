package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// GrinexClient represents the Grinex API client
type GrinexClient struct {
	httpClient *http.Client
	baseURL    string
	market     string
	logger     *zap.Logger
}

// OrderBook represents a single order in the order book
type OrderBook struct {
	Price  string `json:"price"`
	Volume string `json:"volume"`
	Amount string `json:"amount"`
	Factor string `json:"factor"`
	Type   string `json:"type"`
}

// DepthResponse represents the response from the depth API
type DepthResponse struct {
	Timestamp int64       `json:"timestamp"`
	Asks      []OrderBook `json:"asks"`
	Bids      []OrderBook `json:"bids"`
}

// RateData represents exchange rate information
type RateData struct {
	Ask       string
	Bid       string
	Timestamp time.Time
	Market    string
}

// NewGrinexClient creates a new Grinex API client
func NewGrinexClient(baseURL, market string, timeout time.Duration, logger *zap.Logger) *GrinexClient {
	return &GrinexClient{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		baseURL: baseURL,
		market:  market,
		logger:  logger,
	}
}

// GetRates retrieves exchange rates from Grinex API
func (c *GrinexClient) GetRates(ctx context.Context, market string) (*RateData, error) {
	url := fmt.Sprintf("%s/api/v2/depth?market=%s", c.baseURL, market)

	c.logger.Debug("Making request to Grinex API",
		zap.String("url", url),
		zap.String("market", market))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		c.logger.Error("Failed to create request", zap.Error(err))
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to make request", zap.Error(err))
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Error("Unexpected status code",
			zap.Int("status_code", resp.StatusCode),
			zap.String("status", resp.Status))
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var depthResp DepthResponse
	if err := json.NewDecoder(resp.Body).Decode(&depthResp); err != nil {
		c.logger.Error("Failed to decode response", zap.Error(err))
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Validate response structure - allow empty asks/bids but log warning
	if len(depthResp.Asks) == 0 && len(depthResp.Bids) == 0 {
		c.logger.Warn("Empty asks and bids in response")
	}

	// Get first ask and bid prices
	var ask, bid string

	if len(depthResp.Asks) > 0 && depthResp.Asks[0].Price != "" {
		ask = depthResp.Asks[0].Price
	} else {
		ask = "N/A" // No ask orders available
	}

	if len(depthResp.Bids) > 0 && depthResp.Bids[0].Price != "" {
		bid = depthResp.Bids[0].Price
	} else {
		bid = "N/A" // No bid orders available
	}

	timestamp := time.Now()
	if depthResp.Timestamp > 0 {
		timestamp = time.Unix(depthResp.Timestamp/1000, 0)
	}

	rateData := &RateData{
		Ask:       ask,
		Bid:       bid,
		Timestamp: timestamp,
		Market:    market,
	}

	c.logger.Info("Successfully retrieved rates",
		zap.String("ask", ask),
		zap.String("bid", bid),
		zap.String("market", market),
		zap.Time("timestamp", timestamp))

	return rateData, nil
}
