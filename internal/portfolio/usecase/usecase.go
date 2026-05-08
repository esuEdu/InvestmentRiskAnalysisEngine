package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/esuEdu/investment-risk-engine/internal/portfolio/domain"
	"github.com/esuEdu/investment-risk-engine/pkg/auth"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UseCase struct {
	repo          domain.Repository
	jwtSecret     string
	apiServiceURL string
	httpClient    *http.Client
}

func New(repo domain.Repository, jwtSecret, apiServiceURL string) *UseCase {
	return &UseCase{
		repo:          repo,
		jwtSecret:     jwtSecret,
		apiServiceURL: apiServiceURL,
		httpClient:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (u *UseCase) Register(ctx context.Context, email, password string) (string, domain.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", domain.User{}, err
	}

	user, err := u.repo.CreateUser(ctx, domain.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(hash),
	})
	if err != nil {
		return "", domain.User{}, err
	}

	token, err := auth.GenerateToken(user.ID, u.jwtSecret)
	if err != nil {
		return "", domain.User{}, err
	}

	return token, user, nil
}

func (u *UseCase) Login(ctx context.Context, email, password string) (string, domain.User, error) {
	user, err := u.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", domain.User{}, fmt.Errorf("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", domain.User{}, fmt.Errorf("invalid credentials")
	}

	token, err := auth.GenerateToken(user.ID, u.jwtSecret)
	if err != nil {
		return "", domain.User{}, err
	}

	return token, user, nil
}

func (u *UseCase) CreatePortfolio(ctx context.Context, userID uuid.UUID, name string, assets []domain.PortfolioAsset) (domain.Portfolio, error) {
	if err := validateWeights(assets); err != nil {
		return domain.Portfolio{}, err
	}

	return u.repo.CreatePortfolio(ctx, domain.Portfolio{
		ID:     uuid.New(),
		UserID: userID,
		Name:   name,
		Assets: assets,
	})
}

func (u *UseCase) GetPortfolio(ctx context.Context, userID, portfolioID uuid.UUID) (domain.Portfolio, error) {
	p, err := u.repo.GetPortfolio(ctx, portfolioID)
	if err != nil {
		return domain.Portfolio{}, err
	}
	if p.UserID != userID {
		return domain.Portfolio{}, fmt.Errorf("portfolio not found")
	}
	return p, nil
}

func (u *UseCase) ListPortfolios(ctx context.Context, userID uuid.UUID) ([]domain.Portfolio, error) {
	return u.repo.ListPortfolios(ctx, userID)
}

func (u *UseCase) UpdatePortfolio(ctx context.Context, userID, portfolioID uuid.UUID, name string, assets []domain.PortfolioAsset) (domain.Portfolio, error) {
	existing, err := u.repo.GetPortfolio(ctx, portfolioID)
	if err != nil {
		return domain.Portfolio{}, err
	}
	if existing.UserID != userID {
		return domain.Portfolio{}, fmt.Errorf("portfolio not found")
	}
	if err := validateWeights(assets); err != nil {
		return domain.Portfolio{}, err
	}
	return u.repo.UpdatePortfolio(ctx, portfolioID, name, assets)
}

func (u *UseCase) DeletePortfolio(ctx context.Context, userID, portfolioID uuid.UUID) error {
	existing, err := u.repo.GetPortfolio(ctx, portfolioID)
	if err != nil {
		return err
	}
	if existing.UserID != userID {
		return fmt.Errorf("portfolio not found")
	}
	return u.repo.DeletePortfolio(ctx, portfolioID)
}

type triggerAnalysisRequest struct {
	Assets    []assetPayload `json:"assets"`
	Benchmark *string        `json:"benchmark,omitempty"`
	Period    string         `json:"period"`
}

type assetPayload struct {
	Ticker string  `json:"ticker"`
	Weight float64 `json:"weight"`
}

type triggerAnalysisResponse struct {
	Data json.RawMessage `json:"data"`
}

func (u *UseCase) TriggerAnalysis(ctx context.Context, userID, portfolioID uuid.UUID, benchmark *string, period string) (json.RawMessage, error) {
	p, err := u.GetPortfolio(ctx, userID, portfolioID)
	if err != nil {
		return nil, err
	}

	payload := triggerAnalysisRequest{Period: period, Benchmark: benchmark}
	for _, a := range p.Assets {
		payload.Assets = append(payload.Assets, assetPayload{Ticker: a.Ticker, Weight: a.Weight})
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.apiServiceURL+"/api/v1/analyses", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := u.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("api service unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("api service returned %d", resp.StatusCode)
	}

	var result triggerAnalysisResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

func validateWeights(assets []domain.PortfolioAsset) error {
	if len(assets) == 0 {
		return fmt.Errorf("portfolio must have at least one asset")
	}
	var total float64
	for _, a := range assets {
		total += a.Weight
	}
	if math.Abs(total-1.0) > 1e-4 {
		return fmt.Errorf("asset weights must sum to 1.0, got %.4f", total)
	}
	return nil
}
