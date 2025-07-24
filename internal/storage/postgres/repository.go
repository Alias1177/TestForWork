package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// Repository represents the PostgreSQL repository
type Repository struct {
	db     *sql.DB
	logger *zap.Logger
}

// Rate represents a rate record in the database
type Rate struct {
	ID        int64     `db:"id" json:"id"`
	Market    string    `db:"market" json:"market"`
	Ask       string    `db:"ask" json:"ask"`
	Bid       string    `db:"bid" json:"bid"`
	Timestamp time.Time `db:"timestamp" json:"timestamp"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// NewRepository creates a new PostgreSQL repository
func NewRepository(db *sql.DB, logger *zap.Logger) *Repository {
	return &Repository{
		db:     db,
		logger: logger,
	}
}

// SaveRate saves a rate to the database
func (r *Repository) SaveRate(ctx context.Context, market, ask, bid string, timestamp time.Time) error {
	query := `
		INSERT INTO rates (market, ask, bid, timestamp, created_at)
		VALUES ($1, $2, $3, $4, NOW())
	`

	r.logger.Debug("Saving rate to database",
		zap.String("market", market),
		zap.String("ask", ask),
		zap.String("bid", bid),
		zap.Time("timestamp", timestamp))

	_, err := r.db.ExecContext(ctx, query, market, ask, bid, timestamp)
	if err != nil {
		r.logger.Error("Failed to save rate",
			zap.Error(err),
			zap.String("market", market))
		return fmt.Errorf("failed to save rate: %w", err)
	}

	r.logger.Info("Rate saved successfully",
		zap.String("market", market),
		zap.String("ask", ask),
		zap.String("bid", bid))

	return nil
}

// GetRates retrieves rates from the database with pagination
func (r *Repository) GetRates(ctx context.Context, market string, limit, offset int) ([]Rate, error) {
	query := `
		SELECT id, market, ask, bid, timestamp, created_at
		FROM rates
		WHERE market = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	r.logger.Debug("Retrieving rates from database",
		zap.String("market", market),
		zap.Int("limit", limit),
		zap.Int("offset", offset))

	rows, err := r.db.QueryContext(ctx, query, market, limit, offset)
	if err != nil {
		r.logger.Error("Failed to query rates", zap.Error(err))
		return nil, fmt.Errorf("failed to query rates: %w", err)
	}
	defer rows.Close()

	var rates []Rate
	for rows.Next() {
		var rate Rate
		err := rows.Scan(&rate.ID, &rate.Market, &rate.Ask, &rate.Bid, &rate.Timestamp, &rate.CreatedAt)
		if err != nil {
			r.logger.Error("Failed to scan rate", zap.Error(err))
			return nil, fmt.Errorf("failed to scan rate: %w", err)
		}
		rates = append(rates, rate)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("Error during rows iteration", zap.Error(err))
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	r.logger.Debug("Retrieved rates from database", zap.Int("count", len(rates)))

	return rates, nil
}

// GetLatestRate retrieves the latest rate for a market
func (r *Repository) GetLatestRate(ctx context.Context, market string) (*Rate, error) {
	query := `
		SELECT id, market, ask, bid, timestamp, created_at
		FROM rates
		WHERE market = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	r.logger.Debug("Retrieving latest rate from database", zap.String("market", market))

	var rate Rate
	err := r.db.QueryRowContext(ctx, query, market).Scan(
		&rate.ID, &rate.Market, &rate.Ask, &rate.Bid, &rate.Timestamp, &rate.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Debug("No rates found", zap.String("market", market))
			return nil, nil
		}
		r.logger.Error("Failed to query latest rate", zap.Error(err))
		return nil, fmt.Errorf("failed to query latest rate: %w", err)
	}

	r.logger.Debug("Retrieved latest rate from database", zap.String("market", market))

	return &rate, nil
}

// Ping checks the database connection
func (r *Repository) Ping(ctx context.Context) error {
	if err := r.db.PingContext(ctx); err != nil {
		r.logger.Error("Database ping failed", zap.Error(err))
		return fmt.Errorf("database ping failed: %w", err)
	}
	return nil
}

// Close closes the database connection
func (r *Repository) Close() error {
	return r.db.Close()
}

// NewDB creates a new database connection
func NewDB(dsn string, maxOpenConns, maxIdleConns int, connMaxLifetime time.Duration, logger *zap.Logger) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		logger.Error("Failed to open database", zap.Error(err))
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime)

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		logger.Error("Failed to ping database", zap.Error(err))
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Database connection established successfully")

	return db, nil
}
