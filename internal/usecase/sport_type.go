package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/domain"
)

type SportTypeUseCase struct {
	sportTypeRepository sportTypeRepository
}

func NewSportTypeUseCase(sportTypeRepository sportTypeRepository) *SportTypeUseCase {
	return &SportTypeUseCase{
		sportTypeRepository: sportTypeRepository,
	}
}

func (useCase *SportTypeUseCase) ListSportTypes(ctx context.Context) ([]domain.SportType, error) {
	return useCase.sportTypeRepository.ListSportTypes(ctx)
}
