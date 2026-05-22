package yookassa

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/infrastructure/config"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/usecase"
	"github.com/mailru/easyjson"
)

type PaymentProvider struct {
	client    *http.Client
	baseURL   string
	shopID    string
	secretKey string
	returnURL string
}

//easyjson:json
type createPaymentRequest struct {
	Amount       paymentAmount       `json:"amount"`
	Capture      bool                `json:"capture"`
	Confirmation paymentConfirmation `json:"confirmation"`
	Description  string              `json:"description"`
}

//easyjson:json
type paymentAmount struct {
	Value    string `json:"value"`
	Currency string `json:"currency"`
}

//easyjson:json
type paymentConfirmation struct {
	Type            string `json:"type"`
	ReturnURL       string `json:"return_url,omitempty"`
	ConfirmationURL string `json:"confirmation_url,omitempty"`
}

//easyjson:json
type paymentResponse struct {
	ID           string              `json:"id"`
	Status       string              `json:"status"`
	Confirmation paymentConfirmation `json:"confirmation"`
}

func NewPaymentProvider(cfg config.PaymentConfig) (*PaymentProvider, error) {
	timeout, err := cfg.HTTPTimeoutDuration()
	if err != nil {
		return nil, err
	}

	return &PaymentProvider{
		client: &http.Client{
			Timeout: timeout,
		},
		baseURL:   strings.TrimRight(cfg.YooKassaAPIBaseURL, "/"),
		shopID:    cfg.YooKassaShopID,
		secretKey: cfg.YooKassaSecret,
		returnURL: cfg.YooKassaReturnURL,
	}, nil
}

func (provider *PaymentProvider) CreatePayment(ctx context.Context, request usecase.PaymentProviderCreateRequest) (usecase.PaymentProviderPayment, error) {
	payload := createPaymentRequest{
		Amount: paymentAmount{
			Value:    fmt.Sprintf("%d.00", request.AmountValue),
			Currency: request.Currency,
		},
		Capture: true,
		Confirmation: paymentConfirmation{
			Type:      "redirect",
			ReturnURL: provider.returnURL,
		},
		Description: request.Description,
	}

	var response paymentResponse
	if err := provider.do(ctx, http.MethodPost, "/payments", request.IdempotenceKey, payload, &response); err != nil {
		return usecase.PaymentProviderPayment{}, err
	}

	return usecase.PaymentProviderPayment{
		ProviderPaymentID: response.ID,
		Status:            response.Status,
		ConfirmationURL:   response.Confirmation.ConfirmationURL,
	}, nil
}

func (provider *PaymentProvider) GetPayment(ctx context.Context, providerPaymentID string) (usecase.PaymentProviderPayment, error) {
	var response paymentResponse
	if err := provider.do(ctx, http.MethodGet, "/payments/"+providerPaymentID, "", nil, &response); err != nil {
		return usecase.PaymentProviderPayment{}, err
	}

	return usecase.PaymentProviderPayment{
		ProviderPaymentID: response.ID,
		Status:            response.Status,
		ConfirmationURL:   response.Confirmation.ConfirmationURL,
	}, nil
}

func (provider *PaymentProvider) do(ctx context.Context, method string, path string, idempotenceKey string, payload any, target any) error {
	var body *bytes.Reader
	if payload == nil {
		body = bytes.NewReader(nil)
	} else {
		marshaler, ok := payload.(easyjson.Marshaler)
		if !ok {
			return fmt.Errorf("payload %T does not implement easyjson.Marshaler", payload)
		}

		data, err := easyjson.Marshal(marshaler)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		body = bytes.NewReader(data)
	}

	request, err := http.NewRequestWithContext(ctx, method, provider.baseURL+path, body)
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	request.SetBasicAuth(provider.shopID, provider.secretKey)
	request.Header.Set("Content-Type", "application/json")
	if idempotenceKey != "" {
		request.Header.Set("Idempotence-Key", idempotenceKey)
	}

	response, err := provider.client.Do(request)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("unexpected status: %s", response.Status)
	}
	unmarshaler, ok := target.(easyjson.Unmarshaler)
	if !ok {
		return fmt.Errorf("target %T does not implement easyjson.Unmarshaler", target)
	}
	if err := easyjson.UnmarshalFromReader(response.Body, unmarshaler); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}

var _ usecase.PaymentProvider = (*PaymentProvider)(nil)
