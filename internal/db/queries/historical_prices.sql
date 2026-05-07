-- name: UpsertHistoricalPrice :exec
INSERT INTO historical_prices (ticker, price_date, open, high, low, close, volume)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (ticker, price_date) DO UPDATE
SET open   = EXCLUDED.open,
    high   = EXCLUDED.high,
    low    = EXCLUDED.low,
    close  = EXCLUDED.close,
    volume = EXCLUDED.volume;

-- name: GetHistoricalPrices :many
SELECT ticker, price_date, open, high, low, close, volume
FROM historical_prices
WHERE ticker = $1
ORDER BY price_date ASC;

-- name: GetHistoricalPricesSince :many
SELECT ticker, price_date, open, high, low, close, volume
FROM historical_prices
WHERE ticker = $1
  AND price_date >= $2
ORDER BY price_date ASC;

-- name: GetLatestPriceDate :one
SELECT price_date
FROM historical_prices
WHERE ticker = $1
ORDER BY price_date DESC
LIMIT 1;
