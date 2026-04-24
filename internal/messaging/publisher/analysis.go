package publisher

import (
	"context"
	"encoding/json"

	"github.com/esuEdu/investment-risk-engine/internal/analysis/domain"
)

const analysisJobQueue = "risk-analysis-jobs"

type AnalysisPublisher struct {
	pub *Publisher
}

func NewAnalysisPublisher(pub *Publisher) *AnalysisPublisher {
	return &AnalysisPublisher{pub: pub}
}

func (a *AnalysisPublisher) PublishAnalysisJob(req *domain.AnalysisRequest) error {
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}
	return a.pub.Publish(context.Background(), analysisJobQueue, body)
}
