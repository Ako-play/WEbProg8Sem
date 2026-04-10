package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RatesRepository struct {
	pool *pgxpool.Pool
}

func NewRatesRepository(pool *pgxpool.Pool) *RatesRepository {
	return &RatesRepository{pool: pool}
}

func (r *RatesRepository) UpsertRate(ctx context.Context, base, target string, rate float64, rateDate time.Time, source string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO exchange_rates (id, base_code, target_code, rate, rate_date, source)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (base_code, target_code, rate_date)
		DO UPDATE SET rate = EXCLUDED.rate, source = EXCLUDED.source, updated_at = NOW()
	`, uuid.NewString(), base, target, rate, rateDate, source)
	if err != nil {
		return fmt.Errorf("upsert rate: %w", err)
	}
	return nil
}

func (r *RatesRepository) GetHistory(ctx context.Context, base, target string, from, to time.Time) ([]RatePoint, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT TO_CHAR(rate_date, 'YYYY-MM-DD') AS rate_date, rate
		FROM exchange_rates
		WHERE base_code = $1
		  AND target_code = $2
		  AND rate_date BETWEEN $3 AND $4
		ORDER BY rate_date ASC
	`, base, target, from, to)
	if err != nil {
		return nil, fmt.Errorf("query history: %w", err)
	}
	defer rows.Close()

	result := make([]RatePoint, 0)
	for rows.Next() {
		var item RatePoint
		if err := rows.Scan(&item.Date, &item.Rate); err != nil {
			return nil, fmt.Errorf("scan history row: %w", err)
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate history rows: %w", err)
	}
	return result, nil
}

func (r *RatesRepository) GetLatestRates(ctx context.Context, base string, targets []string) (map[string]float64, time.Time, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT DISTINCT ON (target_code)
			target_code,
			rate,
			rate_date
		FROM exchange_rates
		WHERE base_code = $1
		ORDER BY target_code, rate_date DESC
	`, base)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("query latest rates: %w", err)
	}
	defer rows.Close()

	allowedTargets := make(map[string]struct{}, len(targets))
	for _, target := range targets {
		allowedTargets[target] = struct{}{}
	}

	rates := make(map[string]float64, len(targets))
	latestDate := time.Time{}
	for rows.Next() {
		var targetCode string
		var rate float64
		var rateDate time.Time
		if err := rows.Scan(&targetCode, &rate, &rateDate); err != nil {
			return nil, time.Time{}, fmt.Errorf("scan latest rate row: %w", err)
		}
		if len(allowedTargets) > 0 {
			if _, ok := allowedTargets[targetCode]; !ok {
				continue
			}
		}
		rates[targetCode] = rate
		if rateDate.After(latestDate) {
			latestDate = rateDate
		}
	}
	if err := rows.Err(); err != nil {
		return nil, time.Time{}, fmt.Errorf("iterate latest rate rows: %w", err)
	}
	return rates, latestDate, nil
}
