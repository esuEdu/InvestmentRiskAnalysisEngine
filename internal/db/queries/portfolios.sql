-- name: CreatePortfolio :one
INSERT INTO portfolios (id, user_id, name)
VALUES ($1, $2, $3)
RETURNING id, user_id, name, created_at, updated_at;

-- name: GetPortfolio :one
SELECT id, user_id, name, created_at, updated_at
FROM portfolios
WHERE id = $1;

-- name: ListPortfoliosByUser :many
SELECT id, user_id, name, created_at, updated_at
FROM portfolios
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: UpdatePortfolioName :one
UPDATE portfolios
SET name = $2, updated_at = NOW()
WHERE id = $1
RETURNING id, user_id, name, created_at, updated_at;

-- name: DeletePortfolio :exec
DELETE FROM portfolios WHERE id = $1;
