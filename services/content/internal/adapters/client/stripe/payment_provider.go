package stripe

//go:generate go run github.com/mailru/easyjson/easyjson $GOFILE

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/infrastructure/config"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/usecase"
	"github.com/mailru/easyjson"
)

const providerName = "stripe"

type PaymentProvider struct {
	client    *http.Client
	baseURL   string
	secretKey string
	returnURL string
	cancelURL string
}

//easyjson:json
type checkoutSessionResponse struct {
	ID            string `json:"id"`
	Status        string `json:"status"`
	PaymentStatus string `json:"payment_status"`
	URL           string `json:"url"`
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
		baseURL:   strings.TrimRight(cfg.StripeAPIBaseURL, "/"),
		secretKey: cfg.StripeSecretKey,
		returnURL: cfg.StripeReturnURL,
		cancelURL: cfg.StripeCancelURL,
	}, nil
}

func (provider *PaymentProvider) ProviderName() string {
	return providerName
}

func (provider *PaymentProvider) CreatePayment(ctx context.Context, request usecase.PaymentProviderCreateRequest) (usecase.PaymentProviderPayment, error) {
	form := url.Values{}
	form.Set("mode", "payment")
	form.Set("success_url", absoluteURLOrFallback(request.ReturnURL, provider.returnURL))
	form.Set("cancel_url", absoluteURLOrFallback(request.CancelURL, provider.cancelURL))
	form.Set("line_items[0][quantity]", "1")
	form.Set("line_items[0][price_data][currency]", strings.ToLower(request.Currency))
	form.Set("line_items[0][price_data][unit_amount]", strconv.FormatInt(stripeMinorUnits(request.AmountValue, request.Currency), 10))
	form.Set("line_items[0][price_data][product_data][name]", request.Description)

	var response checkoutSessionResponse
	if err := provider.do(ctx, http.MethodPost, "/checkout/sessions", request.IdempotenceKey, form, &response); err != nil {
		return usecase.PaymentProviderPayment{}, err
	}

	return usecase.PaymentProviderPayment{
		ProviderPaymentID: response.ID,
		Status:            stripeStatus(response),
		ConfirmationURL:   response.URL,
	}, nil
}

func (provider *PaymentProvider) GetPayment(ctx context.Context, providerPaymentID string) (usecase.PaymentProviderPayment, error) {
	var response checkoutSessionResponse
	if err := provider.do(ctx, http.MethodGet, "/checkout/sessions/"+url.PathEscape(providerPaymentID), "", nil, &response); err != nil {
		return usecase.PaymentProviderPayment{}, err
	}

	return usecase.PaymentProviderPayment{
		ProviderPaymentID: response.ID,
		Status:            stripeStatus(response),
		ConfirmationURL:   response.URL,
	}, nil
}

func (provider *PaymentProvider) do(ctx context.Context, method string, path string, idempotencyKey string, form url.Values, target easyjson.Unmarshaler) error {
	var body *bytes.Reader
	if form == nil {
		body = bytes.NewReader(nil)
	} else {
		body = bytes.NewReader([]byte(form.Encode()))
	}

	request, err := http.NewRequestWithContext(ctx, method, provider.baseURL+path, body)
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	request.SetBasicAuth(provider.secretKey, "")
	if form != nil {
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	if idempotencyKey != "" {
		request.Header.Set("Idempotency-Key", idempotencyKey)
	}

	response, err := provider.client.Do(request)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		responseBody, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return fmt.Errorf("unexpected status: %s: %s", response.Status, strings.TrimSpace(string(responseBody)))
	}
	if err := easyjson.UnmarshalFromReader(response.Body, target); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}

func stripeStatus(response checkoutSessionResponse) string {
	if response.PaymentStatus == "paid" {
		return "succeeded"
	}
	if response.Status == "expired" {
		return "canceled"
	}
	return "pending"
}

func stripeMinorUnits(amountValue int32, currency string) int64 {
	switch strings.ToUpper(currency) {
	case "JPY", "KRW":
		return int64(amountValue)
	default:
		return int64(amountValue) * 100
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func absoluteURLOrFallback(value string, fallback string) string {
	candidate := firstNonEmpty(strings.TrimSpace(value), strings.TrimSpace(fallback))
	parsedURL, err := url.Parse(candidate)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return strings.TrimSpace(fallback)
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return strings.TrimSpace(fallback)
	}

	return candidate
}

var _ usecase.PaymentProvider = (*PaymentProvider)(nil)
