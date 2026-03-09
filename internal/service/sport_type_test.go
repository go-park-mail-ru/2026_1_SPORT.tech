package service

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/repository"
)

type sportTypeRepositoryStub struct {
	listSportTypesFunc func(ctx context.Context) ([]repository.SportType, error)
}

func (stub *sportTypeRepositoryStub) ListSportTypes(ctx context.Context) ([]repository.SportType, error) {
	if stub.listSportTypesFunc == nil {
		return nil, nil
	}

	return stub.listSportTypesFunc(ctx)
}

type sportTypeServiceTest struct {
	name       string
	repository *sportTypeRepositoryStub
	expect     []SportType
	expectErr  error
}

func TestSportTypeServiceListSportTypesPositive(t *testing.T) {
	tests := []sportTypeServiceTest{
		{
			name: "Корректный маппинг sport types",
			repository: &sportTypeRepositoryStub{
				listSportTypesFunc: func(ctx context.Context) ([]repository.SportType, error) {
					return []repository.SportType{
						{ID: 1, Name: "Бег"},
						{ID: 2, Name: "Плавание"},
					}, nil
				},
			},
			expect: []SportType{
				{ID: 1, Name: "Бег"},
				{ID: 2, Name: "Плавание"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewSportTypeService(tt.repository)

			res, err := service.ListSportTypes(context.Background())
			if err != nil {
				t.Fatalf("unexpected error: got %v", err)
			}
			if !reflect.DeepEqual(res, tt.expect) {
				t.Fatalf("unexpected result: got %+v, expect %+v", res, tt.expect)
			}
		})
	}
}

func TestSportTypeServiceListSportTypesNegative(t *testing.T) {
	expectedErr := errors.New("list sport types")
	tests := []sportTypeServiceTest{
		{
			name: "Ошибка репозитория",
			repository: &sportTypeRepositoryStub{
				listSportTypesFunc: func(ctx context.Context) ([]repository.SportType, error) {
					return nil, expectedErr
				},
			},
			expectErr: expectedErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewSportTypeService(tt.repository)

			_, err := service.ListSportTypes(context.Background())
			if !errors.Is(err, tt.expectErr) {
				t.Fatalf("unexpected error: got %v, expect %v", err, tt.expectErr)
			}
		})
	}
}
