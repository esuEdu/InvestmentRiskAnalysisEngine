package publisher_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/esuEdu/investment-risk-engine/internal/analysis/domain"
	"github.com/esuEdu/investment-risk-engine/internal/messaging/publisher"
	"github.com/google/uuid"
)

type mockSender struct {
	publishFn func(ctx context.Context, queue string, body []byte) error
	lastQueue string
	lastBody  []byte
}

func (m *mockSender) Publish(ctx context.Context, queue string, body []byte) error {
	m.lastQueue = queue
	m.lastBody = body
	if m.publishFn != nil {
		return m.publishFn(ctx, queue, body)
	}
	return nil
}

func TestPublishAnalysisJob_SendsToCorrectQueue(t *testing.T) {
	mock := &mockSender{}
	ap := publisher.NewAnalysisPublisher(mock)

	if err := ap.PublishAnalysisJob(&domain.AnalysisRequest{ID: uuid.New(), Period: "1y"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mock.lastQueue != "risk-analysis-jobs" {
		t.Errorf("want queue %q, got %q", "risk-analysis-jobs", mock.lastQueue)
	}
}

func TestPublishAnalysisJob_BodyIsValidJSON(t *testing.T) {
	mock := &mockSender{}
	ap := publisher.NewAnalysisPublisher(mock)

	benchmark := "SPY"
	id := uuid.New()
	req := &domain.AnalysisRequest{ID: id, Period: "6m", Benchmark: &benchmark}

	if err := ap.PublishAnalysisJob(req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got domain.AnalysisRequest
	if err := json.Unmarshal(mock.lastBody, &got); err != nil {
		t.Fatalf("body is not valid JSON: %v", err)
	}
	if got.ID != id {
		t.Errorf("want ID %v, got %v", id, got.ID)
	}
	if got.Period != req.Period {
		t.Errorf("want period %q, got %q", req.Period, got.Period)
	}
	if got.Benchmark == nil || *got.Benchmark != benchmark {
		t.Errorf("want benchmark %q, got %v", benchmark, got.Benchmark)
	}
}

func TestPublishAnalysisJob_NilBenchmark(t *testing.T) {
	mock := &mockSender{}
	ap := publisher.NewAnalysisPublisher(mock)

	if err := ap.PublishAnalysisJob(&domain.AnalysisRequest{ID: uuid.New(), Period: "3m"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got domain.AnalysisRequest
	if err := json.Unmarshal(mock.lastBody, &got); err != nil {
		t.Fatalf("body is not valid JSON: %v", err)
	}
	if got.Benchmark != nil {
		t.Errorf("expected nil benchmark in payload, got %v", got.Benchmark)
	}
}

func TestPublishAnalysisJob_PropagatesPublishError(t *testing.T) {
	want := errors.New("broker unavailable")
	mock := &mockSender{
		publishFn: func(_ context.Context, _ string, _ []byte) error { return want },
	}

	ap := publisher.NewAnalysisPublisher(mock)
	err := ap.PublishAnalysisJob(&domain.AnalysisRequest{ID: uuid.New(), Period: "1y"})

	if err == nil {
		t.Fatal("expected error but got nil")
	}
	if err.Error() != want.Error() {
		t.Errorf("want error %q, got %q", want.Error(), err.Error())
	}
}
