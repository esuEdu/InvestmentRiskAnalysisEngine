package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	handler "github.com/esuEdu/investment-risk-engine/internal/analysis/delivery/http"
	"github.com/esuEdu/investment-risk-engine/internal/analysis/domain"
	"github.com/esuEdu/investment-risk-engine/internal/analysis/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type mockRepo struct {
	createFn       func(ctx context.Context, req domain.AnalysisRequest) (domain.AnalysisRequest, error)
	getFn          func(ctx context.Context, id uuid.UUID) (domain.AnalysisRequest, error)
	updateStatusFn func(ctx context.Context, id uuid.UUID, status domain.Status) error
	listFn         func(ctx context.Context, limit, offset int32, status *string) ([]domain.AnalysisRequest, error)
}

func (m *mockRepo) Create(ctx context.Context, req domain.AnalysisRequest) (domain.AnalysisRequest, error) {
	if m.createFn != nil {
		return m.createFn(ctx, req)
	}
	return domain.AnalysisRequest{}, nil
}
func (m *mockRepo) Get(ctx context.Context, id uuid.UUID) (domain.AnalysisRequest, error) {
	if m.getFn != nil {
		return m.getFn(ctx, id)
	}
	return domain.AnalysisRequest{}, nil
}
func (m *mockRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.Status) error {
	if m.updateStatusFn != nil {
		return m.updateStatusFn(ctx, id, status)
	}
	return nil
}
func (m *mockRepo) List(ctx context.Context, limit, offset int32, status *string) ([]domain.AnalysisRequest, error) {
	if m.listFn != nil {
		return m.listFn(ctx, limit, offset, status)
	}
	return nil, nil
}

func newTestRouter(repo domain.Repository) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	uc := usecase.New(repo)
	h := handler.New(uc)
	r.POST("/api/v1/analyses", h.Create)
	return r
}

func post(router *gin.Engine, path string, body any) *httptest.ResponseRecorder {
	b, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	return w
}

func TestCreate_ValidRequest(t *testing.T) {
	repo := &mockRepo{
		createFn: func(_ context.Context, req domain.AnalysisRequest) (domain.AnalysisRequest, error) {
			req.ID = uuid.New()
			return req, nil
		},
	}

	w := post(newTestRouter(repo), "/api/v1/analyses", map[string]any{
		"portfolio_id": uuid.New().String(),
		"benchmark":    "SPY",
		"period":       "1y",
	})

	if w.Code != http.StatusAccepted {
		t.Errorf("want 202 Accepted, got %d\nbody: %s", w.Code, w.Body.String())
	}
}

func TestCreate_MissingPeriod(t *testing.T) {
	w := post(newTestRouter(&mockRepo{}), "/api/v1/analyses", map[string]any{
		"portfolio_id": uuid.New().String(),
	})

	if w.Code != http.StatusBadRequest {
		t.Errorf("want 400 Bad Request, got %d", w.Code)
	}
}

func TestCreate_MissingPortfolioID(t *testing.T) {
	w := post(newTestRouter(&mockRepo{}), "/api/v1/analyses", map[string]any{
		"period": "1y",
	})

	if w.Code != http.StatusBadRequest {
		t.Errorf("want 400 Bad Request, got %d", w.Code)
	}
}

func TestCreate_InvalidUUID(t *testing.T) {
	w := post(newTestRouter(&mockRepo{}), "/api/v1/analyses", map[string]any{
		"portfolio_id": "this-is-not-a-uuid",
		"period":       "1y",
	})

	if w.Code != http.StatusBadRequest {
		t.Errorf("want 400 Bad Request, got %d", w.Code)
	}
}

func TestCreate_RepoError(t *testing.T) {
	repo := &mockRepo{
		createFn: func(_ context.Context, _ domain.AnalysisRequest) (domain.AnalysisRequest, error) {
			return domain.AnalysisRequest{}, errors.New("db: connection reset by peer")
		},
	}

	w := post(newTestRouter(repo), "/api/v1/analyses", map[string]any{
		"portfolio_id": uuid.New().String(),
		"period":       "1y",
	})

	if w.Code != http.StatusInternalServerError {
		t.Errorf("want 500 Internal Server Error, got %d\nbody: %s", w.Code, w.Body.String())
	}
}

func TestCreate_NoBenchmark(t *testing.T) {
	repo := &mockRepo{
		createFn: func(_ context.Context, req domain.AnalysisRequest) (domain.AnalysisRequest, error) {
			req.ID = uuid.New()
			return req, nil
		},
	}

	w := post(newTestRouter(repo), "/api/v1/analyses", map[string]any{
		"portfolio_id": uuid.New().String(),
		"period":       "6m",
	})

	if w.Code != http.StatusAccepted {
		t.Errorf("want 202 Accepted, got %d\nbody: %s", w.Code, w.Body.String())
	}
}

func TestCreate_EmptyBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	uc := usecase.New(&mockRepo{})
	h := handler.New(uc)
	r.POST("/api/v1/analyses", h.Create)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/analyses", bytes.NewReader([]byte{}))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("want 400 Bad Request, got %d", w.Code)
	}
}
