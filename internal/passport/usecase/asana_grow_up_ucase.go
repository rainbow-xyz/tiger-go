package usecase

import (
	"context"
	"saas_service/internal/pkg/domain/passport"
)

type asanaGrowUpUsecase struct {
	asanaGrowUpRepo passport.AsanaGrowUpRepo
}

func NewAsanaGrowUpUsecase(repo passport.AsanaGrowUpRepo) passport.AsanaGrowUpUsecase {
	return &asanaGrowUpUsecase{
		asanaGrowUpRepo: repo,
	}
}

func (a *asanaGrowUpUsecase) AsanaGrowAdd(ctx context.Context, statistic passport.AsanaGrowUpStatistic) (int, error) {
	id, err := a.asanaGrowUpRepo.AsanaGrowAdd(ctx, statistic)
	return id, err
}
