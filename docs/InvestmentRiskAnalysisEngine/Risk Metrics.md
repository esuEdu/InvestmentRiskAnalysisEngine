# Risk Metrics

All metrics are computed by the **Risk Worker** after consuming a job from RabbitMQ.

---

## Annualized Volatility

Measures how much portfolio returns fluctuate over a year.

$$\sigma_{annual} = \sigma_{daily} \times \sqrt{252}$$

High volatility → higher risk.

---

## Sharpe Ratio

Return earned per unit of risk taken.

$$Sharpe = \frac{E[R_p] - R_f}{\sigma_p}$$

- `E[Rp]` — expected portfolio return
- `Rf` — risk-free rate (e.g. 3-month T-bill)
- `σp` — portfolio volatility

Higher is better. > 1.0 is generally considered acceptable.

---

## Beta (β)

Sensitivity of the portfolio to benchmark (market) movements.

$$\beta = \frac{Cov(R_p, R_m)}{Var(R_m)}$$

- β = 1 → moves with the market
- β > 1 → amplifies market moves (more volatile)
- β < 1 → dampens market moves (less volatile)

---

## Maximum Drawdown

The largest peak-to-trough decline in portfolio value over the period.

$$MDD = \frac{Trough - Peak}{Peak}$$

Expressed as a negative percentage.

---

## Historical VaR (Value at Risk, 95%)

The worst expected loss over a given time horizon at a 95% confidence level, estimated from historical returns.

> e.g. VaR 95% = -3.2% means there is a 5% chance of losing more than 3.2% in a single period.

---

## Concentration Score (Herfindahl Index)

Measures how concentrated (undiversified) the portfolio is.

$$HHI = \sum_{i} w_i^2$$

- Min value (perfectly diversified) → approaches 0
- Max value (single asset) → 1.0

Higher value = lower diversification = higher concentration risk.

---

## Correlation Matrix *(planned)*

Shows pairwise return correlations between all assets.

Useful for identifying assets that move together (high correlation = lower diversification benefit).

---

## Related Notes

- [[Database Schema]]
- [[API Reference]]
