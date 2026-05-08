-- name: InsertPortfolioAsset :exec
INSERT INTO portfolio_assets (portfolio_id, ticker, weight)
VALUES ($1, $2, $3)
ON CONFLICT (portfolio_id, ticker) DO UPDATE SET weight = EXCLUDED.weight;

-- name: GetPortfolioAssets :many
SELECT portfolio_id, ticker, weight
FROM portfolio_assets
WHERE portfolio_id = $1
ORDER BY ticker;

-- name: DeletePortfolioAssets :exec
DELETE FROM portfolio_assets WHERE portfolio_id = $1;
