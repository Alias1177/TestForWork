CREATE TABLE IF NOT EXISTS rates (
    id SERIAL PRIMARY KEY,
    market VARCHAR(20) NOT NULL,
    ask DECIMAL(20, 8) NOT NULL,
    bid DECIMAL(20, 8) NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_rates_market ON rates(market);
CREATE INDEX idx_rates_timestamp ON rates(timestamp);
CREATE INDEX idx_rates_created_at ON rates(created_at); 