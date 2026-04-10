package repository

import (
	"github.com/esuEdu/investment-risk-engine/internal/analysis/domain"
	db "github.com/esuEdu/investment-risk-engine/internal/db/generated"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func uuidToPg(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func toDomain(row db.AnalysisRequest) domain.AnalysisRequest {
	return domain.AnalysisRequest{
		ID:        uuid.UUID(row.ID.Bytes),
		Status:    domain.Status(row.Status),
		Benchmark: row.Benchmark,
		Period:    row.Period,
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}
}
