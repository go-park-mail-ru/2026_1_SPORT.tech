package service

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/repository"
)

type sportTypeRepository interface {
	ListSportTypes(ctx context.Context) ([]repository.SportType, error)
}

type SportType struct {
	ID   int64
	Name string
}

type SportTypeService struct {
	sportTypeRepository sportTypeRepository
}

func NewSportTypeService(sportTypeRepository sportTypeRepository) *SportTypeService {
	return &SportTypeService{
		sportTypeRepository: sportTypeRepository,
	}
}

func (service *SportTypeService) ListSportTypes(ctx context.Context) ([]SportType, error) {
	sportTypes, err := service.sportTypeRepository.ListSportTypes(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]SportType, 0, len(sportTypes))
	for _, sportType := range sportTypes {
		result = append(result, SportType{
			ID:   sportType.ID,
			Name: sportType.Name,
		})
	}

	return result, nil
}
