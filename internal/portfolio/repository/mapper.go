package repository

import (
	"math/big"
	"strconv"
	"time"

	"github.com/esuEdu/investment-risk-engine/internal/portfolio/domain"
	db "github.com/esuEdu/investment-risk-engine/internal/db/generated"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func uuidToPg(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func pgToUUID(u pgtype.UUID) uuid.UUID {
	return uuid.UUID(u.Bytes)
}

func pgToTime(t pgtype.Timestamptz) time.Time {
	return t.Time
}

func pgNumeric(f float64) pgtype.Numeric {
	s := strconv.FormatFloat(f, 'f', 10, 64)
	neg := false
	if len(s) > 0 && s[0] == '-' {
		neg = true
		s = s[1:]
	}
	dotIdx := -1
	for i, c := range s {
		if c == '.' {
			dotIdx = i
			break
		}
	}
	var intStr string
	var fracLen int
	if dotIdx >= 0 {
		fracLen = len(s) - dotIdx - 1
		intStr = s[:dotIdx] + s[dotIdx+1:]
	} else {
		intStr = s
	}
	bigInt := new(big.Int)
	bigInt.SetString(intStr, 10)
	if neg {
		bigInt.Neg(bigInt)
	}
	return pgtype.Numeric{Int: bigInt, Exp: -int32(fracLen), Valid: true}
}

func numericToFloat64(n pgtype.Numeric) float64 {
	if !n.Valid || n.NaN || n.Int == nil {
		return 0
	}
	rat := new(big.Rat).SetInt(n.Int)
	if n.Exp >= 0 {
		mul := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n.Exp)), nil)
		rat.Mul(rat, new(big.Rat).SetInt(mul))
	} else {
		div := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(-n.Exp)), nil)
		rat.Quo(rat, new(big.Rat).SetInt(div))
	}
	f, _ := rat.Float64()
	return f
}

func userToDomain(row db.User) domain.User {
	return domain.User{
		ID:           pgToUUID(row.ID),
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		CreatedAt:    pgToTime(row.CreatedAt),
	}
}

func portfolioToDomain(row db.Portfolio) domain.Portfolio {
	return domain.Portfolio{
		ID:        pgToUUID(row.ID),
		UserID:    pgToUUID(row.UserID),
		Name:      row.Name,
		CreatedAt: pgToTime(row.CreatedAt),
		UpdatedAt: pgToTime(row.UpdatedAt),
	}
}

func assetsToDomain(rows []db.PortfolioAsset) []domain.PortfolioAsset {
	out := make([]domain.PortfolioAsset, len(rows))
	for i, r := range rows {
		out[i] = domain.PortfolioAsset{
			Ticker: r.Ticker,
			Weight: numericToFloat64(r.Weight),
		}
	}
	return out
}
