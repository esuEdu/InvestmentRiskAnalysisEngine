package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}

type Portfolio struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Name      string
	Assets    []PortfolioAsset
	CreatedAt time.Time
	UpdatedAt time.Time
}

type PortfolioAsset struct {
	Ticker string  `json:"ticker"`
	Weight float64 `json:"weight"`
}
