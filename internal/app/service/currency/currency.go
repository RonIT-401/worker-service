package currency

import (
	"context"
	"errors"
	"fmt"

	"github.com/RonIT-401/worker-service/internal/app/entity"
	"github.com/RonIT-401/worker-service/internal/app/repository"
	"github.com/RonIT-401/worker-service/internal/app/service"
	"github.com/rs/zerolog/log"
)

type RatesProvider interface {
	GetRates(ctx context.Context, base string) (map[string]float64, error)
}

type srv struct {
	rates RatesProvider
	repo  repository.CurrencyRate
}

func NewService(rates RatesProvider, repo repository.CurrencyRate) service.Currency {
	return &srv{rates: rates, repo: repo}
}

func (s *srv) GetRate(ctx context.Context, from, to string) (float64, error) {
	if from == to {
		return 1.0, nil
	}

	if rate, err := s.repo.GetRate(ctx, from, to); err == nil {
		return rate, nil
	} else if !errors.Is(err, entity.ErrRateNotFound) {
		log.Warn().Ctx(ctx).Err(err).Str("from", from).Str("to", to).
			Msg("Ошибка чтения курса из кэша, иду в Fixer")
	}

	rates, err := s.rates.GetRates(ctx, from)
	if err != nil {
		return 0, fmt.Errorf("get rates from %s: %w", from, err)
	}

	if err := s.repo.SetRates(ctx, from, rates); err != nil {
		log.Warn().
			Ctx(ctx).
			Err(err).
			Msg("Не удалось сохранить курсы в кэш")
	}

	rate, ok := rates[to]
	if !ok {
		return 0, fmt.Errorf("%w: %s", entity.ErrFixerCurrencyNotFound, to)
	}

	return rate, nil
}

func (s *srv) Convert(ctx context.Context, amount float64, from, to string) (float64, error) {
	rate, err := s.GetRate(ctx, from, to)
	if err != nil {
		return 0, err
	}
	return amount * rate, nil
}
