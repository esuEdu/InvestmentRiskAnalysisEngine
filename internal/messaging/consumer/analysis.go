package consumer

import (
	"context"
	"encoding/json"

	"github.com/esuEdu/investment-risk-engine/internal/analysis/domain"
)

const analysisJobQueue = "risk-analysis-jobs"

type AnalysisConsumer struct {
	con *Consumer
}

func NewAnalysisConsumer(con *Consumer) *AnalysisConsumer {
	return &AnalysisConsumer{con: con}
}

func (a *AnalysisConsumer) Start(ctx context.Context, handler func(req *domain.AnalysisRequest) error) error {
	return a.con.Consume(ctx, analysisJobQueue, func(body []byte) error {
		var req domain.AnalysisRequest
		if err := json.Unmarshal(body, &req); err != nil {
			return err
		}
		return handler(&req)
	})
}
