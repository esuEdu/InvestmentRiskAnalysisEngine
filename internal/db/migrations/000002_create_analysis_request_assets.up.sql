CREATE TABLE IF NOT EXISTS analysis_request_assets (
    analysis_request_id UUID NOT NULL REFERENCES analysis_requests(id) ON DELETE CASCADE,
    ticker              TEXT NOT NULL,
    weight              NUMERIC(6,4) NOT NULL,
    PRIMARY KEY (analysis_request_id, ticker)
);
