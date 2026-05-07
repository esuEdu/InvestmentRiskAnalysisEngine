CREATE TABLE IF NOT EXISTS analysis_results (
    analysis_request_id   UUID PRIMARY KEY REFERENCES analysis_requests(id) ON DELETE CASCADE,
    annualized_volatility NUMERIC NOT NULL,
    sharpe_ratio          NUMERIC NOT NULL,
    beta                  NUMERIC,
    max_drawdown          NUMERIC NOT NULL,
    var_95                NUMERIC NOT NULL,
    concentration_score   NUMERIC NOT NULL,
    raw_metrics_json      JSONB,
    created_at            TIMESTAMP NOT NULL DEFAULT NOW()
);
