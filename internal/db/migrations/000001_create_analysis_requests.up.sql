CREATE TABLE IF NOT EXISTS analysis_requests (
    id UUID PRIMARY KEY,
    status VARCHAR(20) NOT NULL,
    benchmark VARCHAR(10),
    period VARCHAR(10) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_analysis_requests_status
ON analysis_requests(status);

CREATE INDEX idx_analysis_requests_created_at
ON analysis_requests(created_at);