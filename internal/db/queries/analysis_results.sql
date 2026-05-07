-- name: InsertAnalysisResult :exec
INSERT INTO analysis_results (
    analysis_request_id,
    annualized_volatility,
    sharpe_ratio,
    beta,
    max_drawdown,
    var_95,
    concentration_score,
    raw_metrics_json
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: GetAnalysisResult :one
SELECT
    analysis_request_id,
    annualized_volatility,
    sharpe_ratio,
    beta,
    max_drawdown,
    var_95,
    concentration_score,
    raw_metrics_json,
    created_at
FROM analysis_results
WHERE analysis_request_id = $1;
