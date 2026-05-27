package stripe

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/domain"
)

const (
	checkoutSessionCompletedEvent = "checkout.session.completed"
	webhookMaxBodyBytes           = 1 << 20
	webhookSignatureTolerance     = 5 * time.Minute
	webhookConfirmTimeout         = 15 * time.Second
)

type WebhookConfirmer interface {
	ConfirmPaymentFromProvider(ctx context.Context, providerPaymentID string) (domain.DonationPayment, error)
}

type WebhookHandler struct {
	secret    string
	confirmer WebhookConfirmer
	logger    *slog.Logger
	now       func() time.Time
}

func NewWebhookHandler(secret string, confirmer WebhookConfirmer, logger *slog.Logger) *WebhookHandler {
	return &WebhookHandler{
		secret:    secret,
		confirmer: confirmer,
		logger:    logger,
		now:       time.Now,
	}
}

type stripeEvent struct {
	Type string          `json:"type"`
	Data stripeEventData `json:"data"`
}

type stripeEventData struct {
	Object stripeCheckoutSession `json:"object"`
}

type stripeCheckoutSession struct {
	ID            string `json:"id"`
	PaymentStatus string `json:"payment_status"`
}

func (handler *WebhookHandler) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		http.Error(responseWriter, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(io.LimitReader(request.Body, webhookMaxBodyBytes))
	if err != nil {
		handler.logger.Warn("read webhook body", "err", err)
		http.Error(responseWriter, "cannot read body", http.StatusBadRequest)
		return
	}

	if err := handler.verifySignature(request.Header.Get("Stripe-Signature"), body); err != nil {
		handler.logger.Warn("verify webhook signature", "err", err)
		http.Error(responseWriter, "invalid signature", http.StatusBadRequest)
		return
	}

	var event stripeEvent
	if err := json.Unmarshal(body, &event); err != nil {
		handler.logger.Warn("decode webhook event", "err", err)
		http.Error(responseWriter, "cannot decode event", http.StatusBadRequest)
		return
	}

	if event.Type != checkoutSessionCompletedEvent || event.Data.Object.PaymentStatus != "paid" {
		responseWriter.WriteHeader(http.StatusOK)
		return
	}

	if event.Data.Object.ID == "" {
		handler.logger.Warn("webhook event missing session id", "type", event.Type)
		http.Error(responseWriter, "missing session id", http.StatusBadRequest)
		return
	}

	confirmCtx, cancel := context.WithTimeout(context.WithoutCancel(request.Context()), webhookConfirmTimeout)
	defer cancel()

	if _, err := handler.confirmer.ConfirmPaymentFromProvider(confirmCtx, event.Data.Object.ID); err != nil {
		if errors.Is(err, domain.ErrPaymentNotFound) {
			handler.logger.Warn("webhook payment not found", "provider_payment_id", event.Data.Object.ID)
			responseWriter.WriteHeader(http.StatusOK)
			return
		}
		handler.logger.Error("confirm payment from webhook", "provider_payment_id", event.Data.Object.ID, "err", err)
		http.Error(responseWriter, "confirm failed", http.StatusInternalServerError)
		return
	}

	responseWriter.WriteHeader(http.StatusOK)
}

func (handler *WebhookHandler) verifySignature(signatureHeader string, body []byte) error {
	if signatureHeader == "" {
		return errors.New("missing Stripe-Signature header")
	}

	var timestampPart string
	var providedSignatures []string
	for _, part := range strings.Split(signatureHeader, ",") {
		key, value, ok := strings.Cut(strings.TrimSpace(part), "=")
		if !ok {
			continue
		}
		switch key {
		case "t":
			timestampPart = value
		case "v1":
			providedSignatures = append(providedSignatures, value)
		}
	}

	if timestampPart == "" {
		return errors.New("missing timestamp")
	}
	if len(providedSignatures) == 0 {
		return errors.New("missing v1 signature")
	}

	timestamp, err := strconv.ParseInt(timestampPart, 10, 64)
	if err != nil {
		return fmt.Errorf("parse timestamp: %w", err)
	}
	if drift := handler.now().Unix() - timestamp; drift > int64(webhookSignatureTolerance.Seconds()) || drift < -int64(webhookSignatureTolerance.Seconds()) {
		return fmt.Errorf("timestamp drift too large: %ds", drift)
	}

	mac := hmac.New(sha256.New, []byte(handler.secret))
	mac.Write([]byte(timestampPart))
	mac.Write([]byte("."))
	mac.Write(body)
	expected := mac.Sum(nil)
	expectedHex := hex.EncodeToString(expected)

	for _, candidate := range providedSignatures {
		if hmac.Equal([]byte(candidate), []byte(expectedHex)) {
			return nil
		}
	}

	return errors.New("no signature matched")
}
