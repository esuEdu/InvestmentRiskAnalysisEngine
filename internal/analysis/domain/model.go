package domain

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusPending    Status = "pending"
	StatusProcessing Status = "processing"
	StatusCompleted  Status = "completed"
	StatusFailed     Status = "failed"
)

type Asset struct {
	Ticker string  `json:"ticker"`
	Weight float64 `json:"weight"`
}

type AnalysisRequest struct {
	ID        uuid.UUID
	Status    Status
	Assets    []Asset
	Benchmark *string
	Period    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type AnalysisResult struct {
	AnnualizedVolatility float64
	SharpeRatio          float64
	Beta                 *float64
	MaxDrawdown          float64
	VaR95                float64
	ConcentrationScore   float64
	RawMetricsJSON       []byte
}

type Queue interface {
	PublishAnalysisJob(req *AnalysisRequest) error
}
