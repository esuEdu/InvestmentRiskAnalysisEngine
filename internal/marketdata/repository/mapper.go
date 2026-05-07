package repository

import (
	"math"
	"math/big"
	"strconv"
	"time"

	db "github.com/esuEdu/investment-risk-engine/internal/db/generated"
	"github.com/esuEdu/investment-risk-engine/internal/marketdata/domain"
	"github.com/jackc/pgx/v5/pgtype"
)

func toPricePoint(row db.HistoricalPrice) domain.PricePoint {
	var volume int64
	if row.Volume != nil {
		volume = *row.Volume
	}
	return domain.PricePoint{
		Date:   row.PriceDate.Time,
		Open:   numericToFloat64(row.Open),
		High:   numericToFloat64(row.High),
		Low:    numericToFloat64(row.Low),
		Close:  numericToFloat64(row.Close),
		Volume: volume,
	}
}

func toPgDate(t time.Time) pgtype.Date {
	return pgtype.Date{Time: t, Valid: true}
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

// numericToFloat64 converts pgtype.Numeric to float64 using exact rational
// arithmetic to avoid accumulated floating-point division errors, then rounds
// to 4 decimal places (sufficient precision for stock prices).
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
	return math.Round(f*10000) / 10000
}
