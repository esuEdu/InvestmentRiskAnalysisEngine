package publisher

import (
	"context"
	"encoding/json"

	"github.com/esuEdu/investment-risk-engine/internal/analysis/domain"
)

const analysisJobQueue = "risk-analysis-jobs"

type Sender interface {
	Publish(ctx context.Context, queue string, body []byte) error
}

type AnalysisPublisher struct {
	pub Sender
}

func NewAnalysisPublisher(pub Sender) *AnalysisPublisher {
	return &AnalysisPublisher{pub: pub}
}

func (a *AnalysisPublisher) PublishAnalysisJob(req *domain.AnalysisRequest) error {
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}
	return a.pub.Publish(context.Background(), analysisJobQueue, body)
}
