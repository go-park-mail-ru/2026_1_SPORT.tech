package usecase

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/domain"
)

type sportTypeRepositoryStub struct {
	listSportTypesFunc func(ctx context.Context) ([]domain.SportType, error)
}

func (stub *sportTypeRepositoryStub) ListSportTypes(ctx context.Context) ([]domain.SportType, error) {
	if stub.listSportTypesFunc == nil {
		return nil, nil
	}

	return stub.listSportTypesFunc(ctx)
}

type sportTypeUseCaseTest struct {
	name       string
	repository *sportTypeRepositoryStub
	expect     []domain.SportType
	expectErr  error
}

func TestSportTypeUseCaseListSportTypesPositive(t *testing.T) {
	tests := []sportTypeUseCaseTest{
		{
			name: "Корректный список sport types",
			repository: &sportTypeRepositoryStub{
				listSportTypesFunc: func(ctx context.Context) ([]domain.SportType, error) {
					return []domain.SportType{
						{ID: 1, Name: "Бег"},
						{ID: 2, Name: "Плавание"},
					}, nil
				},
			},
			expect: []domain.SportType{
				{ID: 1, Name: "Бег"},
				{ID: 2, Name: "Плавание"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase := NewSportTypeUseCase(tt.repository)

			res, err := useCase.ListSportTypes(context.Background())
			if err != nil {
				t.Fatalf("unexpected error: got %v", err)
			}
			if !reflect.DeepEqual(res, tt.expect) {
				t.Fatalf("unexpected result: got %+v, expect %+v", res, tt.expect)
			}
		})
	}
}

func TestSportTypeUseCaseListSportTypesNegative(t *testing.T) {
	expectedErr := errors.New("list sport types")
	tests := []sportTypeUseCaseTest{
		{
			name: "Ошибка репозитория",
			repository: &sportTypeRepositoryStub{
				listSportTypesFunc: func(ctx context.Context) ([]domain.SportType, error) {
					return nil, expectedErr
				},
			},
			expectErr: expectedErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase := NewSportTypeUseCase(tt.repository)

			_, err := useCase.ListSportTypes(context.Background())
			if !errors.Is(err, tt.expectErr) {
				t.Fatalf("unexpected error: got %v, expect %v", err, tt.expectErr)
			}
		})
	}
}
