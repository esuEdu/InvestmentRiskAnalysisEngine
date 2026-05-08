CREATE TABLE IF NOT EXISTS portfolio_assets (
    portfolio_id UUID         NOT NULL REFERENCES portfolios(id) ON DELETE CASCADE,
    ticker       TEXT         NOT NULL,
    weight       NUMERIC(6,4) NOT NULL CHECK (weight > 0 AND weight <= 1),
    PRIMARY KEY (portfolio_id, ticker)
);
