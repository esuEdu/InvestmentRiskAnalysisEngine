package repository

import (
	"math/big"
	"strconv"

	"github.com/esuEdu/investment-risk-engine/internal/analysis/domain"
	db "github.com/esuEdu/investment-risk-engine/internal/db/generated"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func uuidToPg(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

// pgNumeric converts float64 to pgtype.Numeric by parsing its string
// representation into a big.Int mantissa and a base-10 exponent.
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

func pgNumericPtr(f *float64) pgtype.Numeric {
	if f == nil {
		return pgtype.Numeric{}
	}
	return pgNumeric(*f)
}

// numericToFloat64 converts pgtype.Numeric to float64 using exact rational
// arithmetic to avoid accumulated floating-point division errors.
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

func numericToFloat64Ptr(n pgtype.Numeric) *float64 {
	if !n.Valid {
		return nil
	}
	f := numericToFloat64(n)
	return &f
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

func resultToDomain(row db.AnalysisResult) domain.AnalysisResult {
	return domain.AnalysisResult{
		AnnualizedVolatility: numericToFloat64(row.AnnualizedVolatility),
		SharpeRatio:          numericToFloat64(row.SharpeRatio),
		Beta:                 numericToFloat64Ptr(row.Beta),
		MaxDrawdown:          numericToFloat64(row.MaxDrawdown),
		VaR95:                numericToFloat64(row.Var95),
		ConcentrationScore:   numericToFloat64(row.ConcentrationScore),
		RawMetricsJSON:       row.RawMetricsJson,
	}
}
