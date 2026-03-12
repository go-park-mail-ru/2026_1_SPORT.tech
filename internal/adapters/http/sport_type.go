package handler

import (
	"net/http"
)

type getSportTypesResponse struct {
	SportTypes []sportTypeResponseItem `json:"sport_types"`
}

type sportTypeResponseItem struct {
	SportTypeID int64  `json:"sport_type_id"`
	Name        string `json:"name"`
}

func (handler *Handler) handleGetSportTypes(writer http.ResponseWriter, request *http.Request) {
	sportTypes, err := handler.sportTypeUseCase.ListSportTypes(request.Context())
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

	writeJSON(writer, http.StatusOK, response)
}
