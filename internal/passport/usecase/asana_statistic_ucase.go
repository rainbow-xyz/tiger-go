package usecase

import (
	"context"
	"saas_service/internal/pkg/domain/passport"
)

type asanaStatisticUsecase struct {
	asanaStatisticRepo passport.AsanaStatisticRepo
}

func NewAsanaStatisticUsecase(as passport.AsanaStatisticRepo) passport.AsanaStatisticUsecase {
	return &asanaStatisticUsecase{
		asanaStatisticRepo: as,
	}
}

func (a *asanaStatisticUsecase) AddAsanaStatistic(ctx context.Context, statistic passport.AsanaStatistic) (int, error) {
	id, err := a.asanaStatisticRepo.AddAsanaStatistic(ctx, statistic)
	return id, err
}
