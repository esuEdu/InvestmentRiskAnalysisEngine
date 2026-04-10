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

type AnalysisRequest struct {
	ID        uuid.UUID
	Status    Status
	Benchmark *string
	Period    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Queue interface {
	PublishAnalysisJob(req *AnalysisRequest) error
}
