-- name: InsertAnalysisRequestAsset :exec
INSERT INTO analysis_request_assets (analysis_request_id, ticker, weight)
VALUES ($1, $2, $3);

-- name: GetAnalysisRequestAssets :many
SELECT ticker, weight
FROM analysis_request_assets
WHERE analysis_request_id = $1
ORDER BY ticker;
