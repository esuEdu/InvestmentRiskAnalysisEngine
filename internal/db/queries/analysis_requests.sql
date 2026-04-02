-- name: CreateAnalysisRequest :one
INSERT INTO analysis_requests (
    id,
    status,
    benchmark,
    period
) VALUES (
    $1,
    $2,
    $3,
    $4
)
RETURNING 
    id,
    status,
    benchmark,
    period,
    created_at,
    updated_at;


-- name: GetAnalysisRequest :one
SELECT 
    id,
    status,
    benchmark,
    period,
    created_at,
    updated_at
FROM analysis_requests
WHERE id = $1;


-- name: UpdateAnalysisRequestStatus :exec
UPDATE analysis_requests
SET 
    status = $2,
    updated_at = NOW()
WHERE id = $1;


-- name: ListAnalysisRequests :many
SELECT 
    id,
    status,
    benchmark,
    period,
    created_at,
    updated_at
FROM analysis_requests
WHERE ($3::text IS NULL OR status = $3)
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;