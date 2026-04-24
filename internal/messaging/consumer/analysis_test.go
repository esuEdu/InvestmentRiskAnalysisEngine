package consumer_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/esuEdu/investment-risk-engine/internal/analysis/domain"
	"github.com/esuEdu/investment-risk-engine/internal/messaging/consumer"
	"github.com/google/uuid"
)

type mockReceiver struct {
	consumeErr  error
	lastQueue   string
	lastHandler func([]byte) error
}

func (m *mockReceiver) Consume(_ context.Context, queue string, handler func([]byte) error) error {
	m.lastQueue = queue
	m.lastHandler = handler
	return m.consumeErr
}

func TestStart_UsesCorrectQueue(t *testing.T) {
	mock := &mockReceiver{}
	ac := consumer.NewAnalysisConsumer(mock)

	_ = ac.Start(context.Background(), func(_ *domain.AnalysisRequest) error { return nil })

	if mock.lastQueue != "risk-analysis-jobs" {
		t.Errorf("want queue %q, got %q", "risk-analysis-jobs", mock.lastQueue)
	}
}

func TestStart_HandlerReceivesUnmarshaledRequest(t *testing.T) {
	mock := &mockReceiver{}
	ac := consumer.NewAnalysisConsumer(mock)

	var received *domain.AnalysisRequest
	_ = ac.Start(context.Background(), func(req *domain.AnalysisRequest) error {
		received = req
		return nil
	})

	id := uuid.New()
	benchmark := "SPY"
	body, _ := json.Marshal(domain.AnalysisRequest{ID: id, Period: "1y", Benchmark: &benchmark})

	if err := mock.lastHandler(body); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received == nil {
		t.Fatal("domain handler was never called")
	}
	if received.ID != id {
		t.Errorf("want ID %v, got %v", id, received.ID)
	}
	if received.Period != "1y" {
		t.Errorf("want period %q, got %q", "1y", received.Period)
	}
	if received.Benchmark == nil || *received.Benchmark != benchmark {
		t.Errorf("want benchmark %q, got %v", benchmark, received.Benchmark)
	}
}

func TestStart_InvalidJSONReturnsUnmarshalError(t *testing.T) {
	mock := &mockReceiver{}
	ac := consumer.NewAnalysisConsumer(mock)

	_ = ac.Start(context.Background(), func(_ *domain.AnalysisRequest) error { return nil })

	err := mock.lastHandler([]byte("not valid json {{{"))
	if err == nil {
		t.Fatal("expected unmarshal error but got nil")
	}
}

func TestStart_HandlerErrorPropagates(t *testing.T) {
	mock := &mockReceiver{}
	ac := consumer.NewAnalysisConsumer(mock)

	want := errors.New("processing failed")
	_ = ac.Start(context.Background(), func(_ *domain.AnalysisRequest) error { return want })

	body, _ := json.Marshal(domain.AnalysisRequest{ID: uuid.New(), Period: "1y"})
	err := mock.lastHandler(body)

	if err == nil {
		t.Fatal("expected error but got nil")
	}
	if err.Error() != want.Error() {
		t.Errorf("want error %q, got %q", want.Error(), err.Error())
	}
}

func TestStart_PropagatesConsumeError(t *testing.T) {
	want := errors.New("connection refused")
	mock := &mockReceiver{consumeErr: want}
	ac := consumer.NewAnalysisConsumer(mock)

	err := ac.Start(context.Background(), func(_ *domain.AnalysisRequest) error { return nil })

	if err == nil {
		t.Fatal("expected error but got nil")
	}
	if err.Error() != want.Error() {
		t.Errorf("want error %q, got %q", want.Error(), err.Error())
	}
}
