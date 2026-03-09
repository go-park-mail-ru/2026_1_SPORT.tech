package handler

import (
	"context"
	nethttp "net/http"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/service"
)

type sportTypeService interface {
	ListSportTypes(ctx context.Context) ([]service.SportType, error)
}

type getSportTypesResponse struct {
	SportTypes []sportTypeResponseItem `json:"sport_types"`
}

type sportTypeResponseItem struct {
	SportTypeID int64  `json:"sport_type_id"`
	Name        string `json:"name"`
}

func (handler *Handler) handleGetSportTypes(writer nethttp.ResponseWriter, request *nethttp.Request) {
	sportTypes, err := handler.sportTypeService.ListSportTypes(request.Context())
	if err != nil {
		writeInternalError(writer)
		return
	}

	response := getSportTypesResponse{
		SportTypes: make([]sportTypeResponseItem, 0, len(sportTypes)),
	}

	for _, sportType := range sportTypes {
		response.SportTypes = append(response.SportTypes, sportTypeResponseItem{
			SportTypeID: sportType.ID,
			Name:        sportType.Name,
		})
	}

	writeJSON(writer, nethttp.StatusOK, response)
}
