package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
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
	r.GET("/api/v1/analyses", h.List)
	r.GET("/api/v1/analyses/:id", h.Get)
	r.PUT("/api/v1/analyses/:id", h.Update)
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
	var body map[string]any
	json.Unmarshal(w.Body.Bytes(), &body)
	if body["message"] == nil {
		t.Errorf("want 'message' field in response, got: %s", w.Body.String())
	}
	if body["data"] == nil {
		t.Errorf("want 'data' field in response, got: %s", w.Body.String())
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
	if !strings.Contains(w.Body.String(), `"error"`) {
		t.Errorf("want 'error' field in response, got: %s", w.Body.String())
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

func TestGet_ValidID(t *testing.T) {
	wantID := uuid.New()

	repo := &mockRepo{
		getFn: func(_ context.Context, id uuid.UUID) (domain.AnalysisRequest, error) {
			return domain.AnalysisRequest{ID: id, Status: domain.StatusPending, Period: "1y"}, nil
		},
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/analyses/"+wantID.String(), nil)
	newTestRouter(repo).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("want 200 OK, got %d\nbody: %s", w.Code, w.Body.String())
	}
	var body map[string]any
	json.Unmarshal(w.Body.Bytes(), &body)
	if body["message"] == nil {
		t.Errorf("want 'message' field in response, got: %s", w.Body.String())
	}
	if body["data"] == nil {
		t.Errorf("want 'data' field in response, got: %s", w.Body.String())
	}
}

func TestGet_InvalidID(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/analyses/not-a-uuid", nil)
	newTestRouter(&mockRepo{}).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("want 400 Bad Request, got %d", w.Code)
	}
}

func TestGet_NotFound(t *testing.T) {
	repo := &mockRepo{
		getFn: func(_ context.Context, _ uuid.UUID) (domain.AnalysisRequest, error) {
			return domain.AnalysisRequest{}, errors.New("no rows in result set")
		},
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/analyses/"+uuid.New().String(), nil)
	newTestRouter(repo).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("want 404 Not Found, got %d", w.Code)
	}
}

func TestList_Defaults(t *testing.T) {
	repo := &mockRepo{
		listFn: func(_ context.Context, limit, offset int32, status *string) ([]domain.AnalysisRequest, error) {
			if limit != 20 || offset != 0 {
				t.Errorf("want default limit=20 offset=0, got limit=%d offset=%d", limit, offset)
			}
			if status != nil {
				t.Errorf("want nil status filter, got %v", *status)
			}
			return []domain.AnalysisRequest{
				{ID: uuid.New(), Status: domain.StatusPending, Period: "1y"},
			}, nil
		},
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/analyses", nil)
	newTestRouter(repo).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("want 200 OK, got %d\nbody: %s", w.Code, w.Body.String())
	}
	var body map[string]any
	json.Unmarshal(w.Body.Bytes(), &body)
	if body["data"] == nil {
		t.Errorf("want 'data' field, got: %s", w.Body.String())
	}
	if body["meta"] == nil {
		t.Errorf("want 'meta' field, got: %s", w.Body.String())
	}
	if body["message"] == nil {
		t.Errorf("want 'message' field, got: %s", w.Body.String())
	}
}

func TestList_WithQueryParams(t *testing.T) {
	repo := &mockRepo{
		listFn: func(_ context.Context, limit, offset int32, status *string) ([]domain.AnalysisRequest, error) {
			if limit != 5 {
				t.Errorf("want limit=5, got %d", limit)
			}
			if offset != 10 {
				t.Errorf("want offset=10, got %d", offset)
			}
			if status == nil || *status != "completed" {
				t.Errorf("want status=completed, got %v", status)
			}
			return []domain.AnalysisRequest{}, nil
		},
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/analyses?limit=5&offset=10&status=completed", nil)
	newTestRouter(repo).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("want 200 OK, got %d", w.Code)
	}
}

func TestList_InvalidLimit(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/analyses?limit=abc", nil)
	newTestRouter(&mockRepo{}).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("want 400 Bad Request, got %d", w.Code)
	}
}

func TestList_Empty(t *testing.T) {
	repo := &mockRepo{
		listFn: func(_ context.Context, _ int32, _ int32, _ *string) ([]domain.AnalysisRequest, error) {
			return []domain.AnalysisRequest{}, nil
		},
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/analyses", nil)
	newTestRouter(repo).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("want 200 OK, got %d", w.Code)
	}
	var body map[string]any
	json.Unmarshal(w.Body.Bytes(), &body)
	// data must be an empty array [], not null
	data, ok := body["data"].([]any)
	if !ok || len(data) != 0 {
		t.Errorf("want data=[], got: %s", w.Body.String())
	}
	// meta must be present even when the list is empty
	if body["meta"] == nil {
		t.Errorf("want 'meta' field in list response, got: %s", w.Body.String())
	}
}

func TestList_RepoError(t *testing.T) {
	repo := &mockRepo{
		listFn: func(_ context.Context, _ int32, _ int32, _ *string) ([]domain.AnalysisRequest, error) {
			return nil, errors.New("query timeout")
		},
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/analyses", nil)
	newTestRouter(repo).ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("want 500 Internal Server Error, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"error"`) {
		t.Errorf("want 'error' field in response, got: %s", w.Body.String())
	}
}

func TestUpdate_ValidRequest(t *testing.T) {
	wantID := uuid.New()

	repo := &mockRepo{
		updateStatusFn: func(_ context.Context, id uuid.UUID, status domain.Status) error {
			if id != wantID {
				t.Errorf("wrong id: want %v, got %v", wantID, id)
			}
			if status != domain.StatusCompleted {
				t.Errorf("wrong status: want %q, got %q", domain.StatusCompleted, status)
			}
			return nil
		},
	}

	body, _ := json.Marshal(map[string]string{"status": "completed"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/analyses/"+wantID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	newTestRouter(repo).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("want 200 OK, got %d\nbody: %s", w.Code, w.Body.String())
	}
}

func TestUpdate_MissingStatus(t *testing.T) {
	body, _ := json.Marshal(map[string]string{})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/analyses/"+uuid.New().String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	newTestRouter(&mockRepo{}).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("want 400 Bad Request, got %d", w.Code)
	}
}

func TestUpdate_InvalidID(t *testing.T) {
	body, _ := json.Marshal(map[string]string{"status": "completed"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/analyses/bad-id", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	newTestRouter(&mockRepo{}).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("want 400 Bad Request, got %d", w.Code)
	}
}

func TestUpdate_RepoError(t *testing.T) {
	repo := &mockRepo{
		updateStatusFn: func(_ context.Context, _ uuid.UUID, _ domain.Status) error {
			return errors.New("update failed")
		},
	}

	body, _ := json.Marshal(map[string]string{"status": "failed"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/analyses/"+uuid.New().String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	newTestRouter(repo).ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("want 500 Internal Server Error, got %d", w.Code)
	}
}
