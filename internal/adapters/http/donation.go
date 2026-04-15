package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/domain"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/usecase"
)

var currencyPattern = regexp.MustCompile(`^[A-Z]{3}$`)

type donationRequest struct {
	AmountValue int64   `json:"amount_value"`
	Currency    string  `json:"currency"`
	Message     *string `json:"message"`
}

type donationResponse struct {
	DonationID      int64     `json:"donation_id"`
	SenderUserID    int64     `json:"sender_user_id"`
	RecipientUserID int64     `json:"recipient_user_id"`
	AmountValue     int64     `json:"amount_value"`
	Currency        string    `json:"currency"`
	Message         *string   `json:"message"`
	CreatedAt       time.Time `json:"created_at"`
}

func (handler *Handler) handlePostProfileDonation(writer http.ResponseWriter, request *http.Request) {
	senderUserID, ok := userIDFromContext(request.Context())
	if !ok {
		writeInternalError(writer)
		return
	}

	recipientUserID, err := strconv.ParseInt(request.PathValue("user_id"), 10, 64)
	if err != nil || recipientUserID <= 0 {
		writeBadRequest(writer)
		return
	}

	var donationRequest donationRequest

	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&donationRequest); err != nil {
		writeBadRequest(writer)
		return
	}

	validationErrors := validateDonationRequest(donationRequest)
	if len(validationErrors) > 0 {
		writeValidationError(writer, validationErrors)
		return
	}

	donation, err := handler.donationUseCase.Create(request.Context(), usecase.CreateDonationCommand{
		SenderUserID:    senderUserID,
		RecipientUserID: recipientUserID,
		AmountValue:     donationRequest.AmountValue,
		Currency:        donationRequest.Currency,
		Message:         donationRequest.Message,
	})
	if err != nil {
		if errors.Is(err, usecase.ErrDonationRecipientNotFound) {
			writeNotFound(writer, "Получатель не найден")
			return
		}

		writeInternalError(writer)
		return
	}

	writeJSON(writer, http.StatusCreated, newDonationResponse(donation))
}

func validateDonationRequest(donationRequest donationRequest) []validationErrorField {
	validationErrors := make([]validationErrorField, 0)

	if donationRequest.AmountValue < 1 {
		validationErrors = append(validationErrors, validationErrorField{
			Field:   "amount_value",
			Message: "Amount value должен быть не меньше 1",
		})
	}

	if !currencyPattern.MatchString(donationRequest.Currency) {
		validationErrors = append(validationErrors, validationErrorField{
			Field:   "currency",
			Message: "Currency должен быть трехбуквенным кодом в верхнем регистре",
		})
	}

	if donationRequest.Message != nil && len(*donationRequest.Message) > 500 {
		validationErrors = append(validationErrors, validationErrorField{
			Field:   "message",
			Message: "Message должен содержать не более 500 символов",
		})
	}

	return validationErrors
}

func newDonationResponse(donation domain.Donation) donationResponse {
	return donationResponse{
		DonationID:      donation.DonationID,
		SenderUserID:    donation.SenderUserID,
		RecipientUserID: donation.RecipientUserID,
		AmountValue:     donation.AmountValue,
		Currency:        donation.Currency,
		Message:         donation.Message,
		CreatedAt:       donation.CreatedAt,
	}
}
