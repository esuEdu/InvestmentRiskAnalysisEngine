package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/esuEdu/investment-risk-engine/internal/analysis/domain"
	"github.com/esuEdu/investment-risk-engine/internal/analysis/usecase"
	"github.com/google/uuid"
)

type mockRepo struct {
	createFn       func(ctx context.Context, req domain.AnalysisRequest) (domain.AnalysisRequest, error)
	getFn          func(ctx context.Context, id uuid.UUID) (domain.AnalysisRequest, error)
	updateStatusFn func(ctx context.Context, id uuid.UUID, status domain.Status) error
	listFn         func(ctx context.Context, limit, offset int32, status *string) ([]domain.AnalysisRequest, error)
}

func (m *mockRepo) Create(ctx context.Context, req domain.AnalysisRequest) (domain.AnalysisRequest, error) {
	return m.createFn(ctx, req)
}
func (m *mockRepo) Get(ctx context.Context, id uuid.UUID) (domain.AnalysisRequest, error) {
	return m.getFn(ctx, id)
}
func (m *mockRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.Status) error {
	return m.updateStatusFn(ctx, id, status)
}
func (m *mockRepo) List(ctx context.Context, limit, offset int32, status *string) ([]domain.AnalysisRequest, error) {
	return m.listFn(ctx, limit, offset, status)
}

func TestExecuteCreate_Success(t *testing.T) {
	benchmark := "SPY"
	period := "1y"
	savedID := uuid.New()

	repo := &mockRepo{
		createFn: func(_ context.Context, req domain.AnalysisRequest) (domain.AnalysisRequest, error) {
			req.ID = savedID
			return req, nil
		},
	}

	uc := usecase.New(repo)
	got, err := uc.ExecuteCreate(context.Background(), uuid.New(), &benchmark, period)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Status != domain.StatusPending {
		t.Errorf("want status %q, got %q", domain.StatusPending, got.Status)
	}
	if got.Period != period {
		t.Errorf("want period %q, got %q", period, got.Period)
	}
	if got.Benchmark == nil || *got.Benchmark != benchmark {
		t.Errorf("want benchmark %q, got %v", benchmark, got.Benchmark)
	}
}

func TestExecuteCreate_RepoError(t *testing.T) {
	repo := &mockRepo{
		createFn: func(_ context.Context, _ domain.AnalysisRequest) (domain.AnalysisRequest, error) {
			return domain.AnalysisRequest{}, errors.New("db: connection refused")
		},
	}

	uc := usecase.New(repo)
	_, err := uc.ExecuteCreate(context.Background(), uuid.New(), nil, "1y")

	if err == nil {
		t.Fatal("expected an error but got nil")
	}
}

// WHY: nil benchmark is valid — the field is optional in the domain model.
// Make sure we don't crash when it is absent.
func TestExecuteCreate_NilBenchmark(t *testing.T) {
	repo := &mockRepo{
		createFn: func(_ context.Context, req domain.AnalysisRequest) (domain.AnalysisRequest, error) {
			return req, nil
		},
	}

	uc := usecase.New(repo)
	got, err := uc.ExecuteCreate(context.Background(), uuid.New(), nil, "3m")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Benchmark != nil {
		t.Errorf("expected nil benchmark, got %v", got.Benchmark)
	}
}

func TestExecuteGet_Success(t *testing.T) {
	wantID := uuid.New()
	want := domain.AnalysisRequest{
		ID:     wantID,
		Status: domain.StatusCompleted,
		Period: "6m",
	}

	repo := &mockRepo{
		getFn: func(_ context.Context, id uuid.UUID) (domain.AnalysisRequest, error) {
			if id != wantID {
				t.Errorf("wrong id forwarded to repo: want %v, got %v", wantID, id)
			}
			return want, nil
		},
	}

	uc := usecase.New(repo)
	got, err := uc.ExecuteGet(context.Background(), wantID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != want.ID {
		t.Errorf("wrong ID: want %v, got %v", want.ID, got.ID)
	}
	if got.Status != want.Status {
		t.Errorf("wrong status: want %q, got %q", want.Status, got.Status)
	}
}

func TestExecuteGet_NotFound(t *testing.T) {
	repo := &mockRepo{
		getFn: func(_ context.Context, _ uuid.UUID) (domain.AnalysisRequest, error) {
			return domain.AnalysisRequest{}, errors.New("no rows in result set")
		},
	}

	uc := usecase.New(repo)
	_, err := uc.ExecuteGet(context.Background(), uuid.New())

	if err == nil {
		t.Fatal("expected an error but got nil")
	}
}

func TestExecuteUpdate_Success(t *testing.T) {
	called := false

	repo := &mockRepo{
		updateStatusFn: func(_ context.Context, _ uuid.UUID, status domain.Status) error {
			called = true
			if status != domain.StatusCompleted {
				t.Errorf("want status %q forwarded, got %q", domain.StatusCompleted, status)
			}
			return nil
		},
	}

	uc := usecase.New(repo)
	err := uc.ExecuteUpdate(context.Background(), uuid.New(), domain.StatusCompleted)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("UpdateStatus was never called — the use case did not reach the repository")
	}
}

func TestExecuteUpdate_RepoError(t *testing.T) {
	repo := &mockRepo{
		updateStatusFn: func(_ context.Context, _ uuid.UUID, _ domain.Status) error {
			return errors.New("deadlock detected")
		},
	}

	uc := usecase.New(repo)
	err := uc.ExecuteUpdate(context.Background(), uuid.New(), domain.StatusFailed)

	if err == nil {
		t.Fatal("expected an error but got nil")
	}
}

func TestExecuteList_Success(t *testing.T) {
	want := []domain.AnalysisRequest{
		{ID: uuid.New(), Status: domain.StatusPending, Period: "1y"},
		{ID: uuid.New(), Status: domain.StatusCompleted, Period: "6m"},
	}

	repo := &mockRepo{
		listFn: func(_ context.Context, limit, offset int32, _ *string) ([]domain.AnalysisRequest, error) {
			if limit != 10 || offset != 0 {
				t.Errorf("pagination mismatch: want limit=10 offset=0, got limit=%d offset=%d", limit, offset)
			}
			return want, nil
		},
	}

	uc := usecase.New(repo)
	got, err := uc.ExecuteList(context.Background(), 10, 0, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != len(want) {
		t.Errorf("want %d items, got %d", len(want), len(got))
	}
}

func TestExecuteList_NilStatusFilter(t *testing.T) {
	repo := &mockRepo{
		listFn: func(_ context.Context, _, _ int32, status *string) ([]domain.AnalysisRequest, error) {
			if status != nil {
				t.Errorf("expected nil status filter, got %v", *status)
			}
			return []domain.AnalysisRequest{}, nil
		},
	}

	uc := usecase.New(repo)
	got, err := uc.ExecuteList(context.Background(), 5, 0, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Error("expected a slice (even empty), got nil")
	}
}

func TestExecuteList_RepoError(t *testing.T) {
	repo := &mockRepo{
		listFn: func(_ context.Context, _ int32, _ int32, _ *string) ([]domain.AnalysisRequest, error) {
			return nil, errors.New("query timeout")
		},
	}

	uc := usecase.New(repo)
	_, err := uc.ExecuteList(context.Background(), 10, 0, nil)

	if err == nil {
		t.Fatal("expected an error but got nil")
	}
}
