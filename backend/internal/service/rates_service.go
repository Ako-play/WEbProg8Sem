package service

import (
	"context"
	"fmt"
	"hash/fnv"
	"sort"
	"strings"
	"time"

	"currencyparser/backend/internal/repository"
)

type RatesService struct {
	repo  *repository.RatesRepository
	teams []repository.Currency
}

func NewRatesService(repo *repository.RatesRepository, providerBaseURL string) *RatesService {
	_ = providerBaseURL
	teams := []repository.Currency{
		{Code: "NAVI", Name: "Natus Vincere"},
		{Code: "G2", Name: "G2 Esports"},
		{Code: "SPRT", Name: "Team Spirit"},
		{Code: "LQD", Name: "Team Liquid"},
		{Code: "FZE", Name: "FaZe Clan"},
		{Code: "VIT", Name: "Team Vitality"},
		{Code: "C9", Name: "Cloud9"},
		{Code: "MOUZ", Name: "MOUZ"},
		{Code: "ENCE", Name: "ENCE"},
		{Code: "FNC", Name: "Fnatic"},
	}
	return &RatesService{
		repo:  repo,
		teams: teams,
	}
}

func (s *RatesService) ListCurrencies() []repository.Currency {
	result := make([]repository.Currency, len(s.teams))
	copy(result, s.teams)
	sort.Slice(result, func(i, j int) bool { return result[i].Code < result[j].Code })
	return result
}

func (s *RatesService) Convert(ctx context.Context, base, target string, amount float64) (map[string]any, error) {
	base = normalizeCode(base)
	target = normalizeCode(target)
	if amount <= 0 {
		return nil, fmt.Errorf("weight must be greater than zero")
	}
	if base == "" || target == "" {
		return nil, fmt.Errorf("team and opponent are required")
	}
	if base == target {
		return nil, fmt.Errorf("team and opponent must be different")
	}
	rates, date, err := s.fetchLatest(ctx, base, []string{target})
	if err != nil {
		return nil, err
	}
	rate, ok := rates[target]
	if !ok {
		return nil, fmt.Errorf("opponent not found")
	}
	return map[string]any{
		"team":      base,
		"opponent":  target,
		"weight":    amount,
		"index":     rate,
		"projected": amount * rate,
		"date":      date.Format("2006-01-02"),
	}, nil
}

func (s *RatesService) Latest(ctx context.Context, base string, symbols []string) (map[string]any, error) {
	base = normalizeCode(base)
	if base == "" {
		return nil, fmt.Errorf("team is required")
	}
	rates, date, err := s.fetchLatest(ctx, base, normalizeCodes(symbols))
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"team":  base,
		"date":  date.Format("2006-01-02"),
		"live":  rates,
	}, nil
}

func (s *RatesService) History(ctx context.Context, base, target string, from, to time.Time) (map[string]any, error) {
	base = normalizeCode(base)
	target = normalizeCode(target)
	if base == "" || target == "" {
		return nil, fmt.Errorf("team and opponent are required")
	}
	if base == target {
		return nil, fmt.Errorf("team and opponent must be different")
	}
	if from.After(to) {
		return nil, fmt.Errorf("from date must be before to date")
	}

	fetchErr := s.fetchHistory(ctx, base, target, from, to)
	if fetchErr != nil {
		return nil, fetchErr
	}

	points, err := s.repo.GetHistory(ctx, base, target, from, to)
	if err != nil {
		return nil, err
	}
	if len(points) == 0 && fetchErr != nil {
		return nil, fetchErr
	}

	return map[string]any{
		"team":    base,
		"opponent": target,
		"from":    from.Format("2006-01-02"),
		"to":      to.Format("2006-01-02"),
		"history": points,
	}, nil
}

func (s *RatesService) fetchLatest(ctx context.Context, base string, symbols []string) (map[string]float64, time.Time, error) {
	targets := symbols
	if len(targets) == 0 {
		targets = s.defaultTargets(base)
	}
	day := time.Now().UTC().Truncate(24 * time.Hour)
	rates := make(map[string]float64, len(targets))
	for _, target := range targets {
		rate := matchupIndex(base, target, day)
		rates[target] = rate
		if err := s.repo.UpsertRate(ctx, base, target, rate, day, "esports-sim"); err != nil {
			return nil, time.Time{}, err
		}
	}
	return rates, day, nil
}

func (s *RatesService) fetchHistory(ctx context.Context, base, target string, from, to time.Time) error {
	start := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC)
	end := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, time.UTC)
	for day := start; !day.After(end); day = day.Add(24 * time.Hour) {
		if err := s.repo.UpsertRate(ctx, base, target, matchupIndex(base, target, day), day, "esports-sim"); err != nil {
			return err
		}
	}
	return nil
}

func normalizeCode(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func normalizeCodes(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, item := range values {
		code := normalizeCode(item)
		if code == "" {
			continue
		}
		if _, ok := seen[code]; ok {
			continue
		}
		seen[code] = struct{}{}
		result = append(result, code)
	}
	sort.Strings(result)
	return result
}

func (s *RatesService) defaultTargets(base string) []string {
	targets := make([]string, 0, len(s.teams)-1)
	for _, team := range s.teams {
		if team.Code == base {
			continue
		}
		targets = append(targets, team.Code)
	}
	return targets
}

func matchupIndex(team, opponent string, day time.Time) float64 {
	key := fmt.Sprintf("%s:%s:%s", team, opponent, day.Format("2006-01-02"))
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	base := float64(h.Sum32()%8000) / 1000.0
	return 0.85 + base
}
