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
	checkoutSessionCompletedEvent     = "checkout.session.completed"
	checkoutSessionExpiredEvent       = "checkout.session.expired"
	checkoutSessionPaymentFailedEvent = "checkout.session.async_payment_failed"
	invoicePaidEvent                  = "invoice.paid"
	subscriptionDeletedEvent          = "customer.subscription.deleted"
	subscriptionModeValue             = "subscription"
	webhookMaxBodyBytes               = 1 << 20
	webhookSignatureTolerance         = 5 * time.Minute
	webhookConfirmTimeout             = 15 * time.Second
)

type WebhookConfirmer interface {
	ConfirmPaymentFromProvider(ctx context.Context, providerPaymentID string) (domain.DonationPayment, error)
	ConfirmSubscriptionPaymentFromProvider(ctx context.Context, providerPaymentID string, stripeSubscriptionID string) (domain.DonationPayment, error)
	ExpirePaymentFromProvider(ctx context.Context, providerPaymentID string) (domain.DonationPayment, error)
	FailPaymentFromProvider(ctx context.Context, providerPaymentID string) (domain.DonationPayment, error)
	RenewSubscriptionFromProvider(ctx context.Context, stripeSubscriptionID string, currentPeriodEnd time.Time) error
	DeactivateSubscriptionFromProvider(ctx context.Context, stripeSubscriptionID string) error
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
	Object json.RawMessage `json:"object"`
}

type stripeCheckoutSession struct {
	ID            string `json:"id"`
	Mode          string `json:"mode"`
	PaymentStatus string `json:"payment_status"`
	Subscription  string `json:"subscription"`
}

type stripeInvoice struct {
	Subscription string `json:"subscription"`
	Lines        struct {
		Data []struct {
			Period struct {
				End int64 `json:"end"`
			} `json:"period"`
		} `json:"data"`
	} `json:"lines"`
}

type stripeSubscriptionObject struct {
	ID string `json:"id"`
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

	confirmCtx, cancel := context.WithTimeout(context.WithoutCancel(request.Context()), webhookConfirmTimeout)
	defer cancel()

	if err := handler.handleEvent(confirmCtx, event); err != nil {
		if errors.Is(err, errIgnoredEvent) || errors.Is(err, domain.ErrPaymentNotFound) || errors.Is(err, domain.ErrSubscriptionNotFound) {
			responseWriter.WriteHeader(http.StatusOK)
			return
		}
		if errors.Is(err, errBadEvent) {
			handler.logger.Warn("malformed webhook event", "type", event.Type, "err", err)
			http.Error(responseWriter, "malformed event", http.StatusBadRequest)
			return
		}
		handler.logger.Error("process payment webhook", "type", event.Type, "err", err)
		http.Error(responseWriter, "webhook processing failed", http.StatusInternalServerError)
		return
	}

	responseWriter.WriteHeader(http.StatusOK)
}

var (
	errIgnoredEvent = errors.New("ignored event")
	errBadEvent     = errors.New("bad event")
)

func (handler *WebhookHandler) handleEvent(ctx context.Context, event stripeEvent) error {
	switch event.Type {
	case checkoutSessionCompletedEvent:
		return handler.handleCheckoutCompleted(ctx, event.Data.Object)
	case checkoutSessionExpiredEvent:
		return handler.handleCheckoutTerminal(ctx, event.Data.Object, handler.confirmer.ExpirePaymentFromProvider)
	case checkoutSessionPaymentFailedEvent:
		return handler.handleCheckoutTerminal(ctx, event.Data.Object, handler.confirmer.FailPaymentFromProvider)
	case invoicePaidEvent:
		return handler.handleInvoicePaid(ctx, event.Data.Object)
	case subscriptionDeletedEvent:
		return handler.handleSubscriptionDeleted(ctx, event.Data.Object)
	default:
		return errIgnoredEvent
	}
}

func (handler *WebhookHandler) handleCheckoutCompleted(ctx context.Context, raw json.RawMessage) error {
	var session stripeCheckoutSession
	if err := json.Unmarshal(raw, &session); err != nil {
		return fmt.Errorf("%w: decode checkout session: %v", errBadEvent, err)
	}
	if session.ID == "" {
		return fmt.Errorf("%w: missing session id", errBadEvent)
	}
	if session.PaymentStatus != "paid" {
		return errIgnoredEvent
	}

	if session.Mode == subscriptionModeValue {
		if session.Subscription == "" {
			return fmt.Errorf("%w: missing subscription id", errBadEvent)
		}
		_, err := handler.confirmer.ConfirmSubscriptionPaymentFromProvider(ctx, session.ID, session.Subscription)
		return err
	}

	_, err := handler.confirmer.ConfirmPaymentFromProvider(ctx, session.ID)
	return err
}

func (handler *WebhookHandler) handleCheckoutTerminal(ctx context.Context, raw json.RawMessage, action func(context.Context, string) (domain.DonationPayment, error)) error {
	var session stripeCheckoutSession
	if err := json.Unmarshal(raw, &session); err != nil {
		return fmt.Errorf("%w: decode checkout session: %v", errBadEvent, err)
	}
	if session.ID == "" {
		return fmt.Errorf("%w: missing session id", errBadEvent)
	}

	_, err := action(ctx, session.ID)
	return err
}

func (handler *WebhookHandler) handleInvoicePaid(ctx context.Context, raw json.RawMessage) error {
	var invoice stripeInvoice
	if err := json.Unmarshal(raw, &invoice); err != nil {
		return fmt.Errorf("%w: decode invoice: %v", errBadEvent, err)
	}
	if invoice.Subscription == "" || len(invoice.Lines.Data) == 0 {
		return errIgnoredEvent
	}

	periodEnd := invoice.Lines.Data[0].Period.End
	if periodEnd <= 0 {
		return errIgnoredEvent
	}

	return handler.confirmer.RenewSubscriptionFromProvider(ctx, invoice.Subscription, time.Unix(periodEnd, 0).UTC())
}

func (handler *WebhookHandler) handleSubscriptionDeleted(ctx context.Context, raw json.RawMessage) error {
	var subscription stripeSubscriptionObject
	if err := json.Unmarshal(raw, &subscription); err != nil {
		return fmt.Errorf("%w: decode subscription: %v", errBadEvent, err)
	}
	if subscription.ID == "" {
		return fmt.Errorf("%w: missing subscription id", errBadEvent)
	}

	return handler.confirmer.DeactivateSubscriptionFromProvider(ctx, subscription.ID)
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
