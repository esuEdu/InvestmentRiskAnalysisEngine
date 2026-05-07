CREATE TABLE IF NOT EXISTS historical_prices (
    ticker     TEXT           NOT NULL,
    price_date DATE           NOT NULL,
    open       NUMERIC,
    high       NUMERIC,
    low        NUMERIC,
    close      NUMERIC        NOT NULL,
    volume     BIGINT,
    PRIMARY KEY (ticker, price_date)
);

CREATE INDEX idx_historical_prices_ticker_date ON historical_prices(ticker, price_date DESC);
