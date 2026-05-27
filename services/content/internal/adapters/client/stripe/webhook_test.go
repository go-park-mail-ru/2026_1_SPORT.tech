package stripe

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/domain"
)

type stubConfirmer struct {
	calls []string
	err   error
}

func (stub *stubConfirmer) ConfirmPaymentFromProvider(ctx context.Context, providerPaymentID string) (domain.DonationPayment, error) {
	stub.calls = append(stub.calls, providerPaymentID)
	return domain.DonationPayment{ProviderPaymentID: providerPaymentID}, stub.err
}

func signedRequest(t *testing.T, secret string, body string, timestamp int64) *http.Request {
	t.Helper()
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(strconv.FormatInt(timestamp, 10)))
	mac.Write([]byte("."))
	mac.Write([]byte(body))
	signature := hex.EncodeToString(mac.Sum(nil))

	request := httptest.NewRequest(http.MethodPost, "/webhooks/stripe", strings.NewReader(body))
	request.Header.Set("Stripe-Signature", "t="+strconv.FormatInt(timestamp, 10)+",v1="+signature)
	return request
}

func newHandler(confirmer WebhookConfirmer, now time.Time) *WebhookHandler {
	handler := NewWebhookHandler("whsec_test", confirmer, slog.New(slog.DiscardHandler))
	handler.now = func() time.Time { return now }
	return handler
}

func TestWebhookConfirmsPaidCheckoutSession(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	body := `{"type":"checkout.session.completed","data":{"object":{"id":"cs_test_123","payment_status":"paid"}}}`
	confirmer := &stubConfirmer{}
	handler := newHandler(confirmer, now)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, signedRequest(t, "whsec_test", body, now.Unix()))

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d, body=%s", recorder.Code, recorder.Body.String())
	}
	if len(confirmer.calls) != 1 || confirmer.calls[0] != "cs_test_123" {
		t.Fatalf("unexpected confirmer calls: %v", confirmer.calls)
	}
}

func TestWebhookIgnoresNonPaidEvents(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	body := `{"type":"checkout.session.completed","data":{"object":{"id":"cs_test_123","payment_status":"unpaid"}}}`
	confirmer := &stubConfirmer{}
	handler := newHandler(confirmer, now)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, signedRequest(t, "whsec_test", body, now.Unix()))

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", recorder.Code)
	}
	if len(confirmer.calls) != 0 {
		t.Fatalf("expected no confirm calls, got %v", confirmer.calls)
	}
}

func TestWebhookIgnoresUnrelatedEventType(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	body := `{"type":"payment_intent.created","data":{"object":{"id":"pi_1","payment_status":"paid"}}}`
	confirmer := &stubConfirmer{}
	handler := newHandler(confirmer, now)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, signedRequest(t, "whsec_test", body, now.Unix()))

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", recorder.Code)
	}
	if len(confirmer.calls) != 0 {
		t.Fatalf("expected no confirm calls, got %v", confirmer.calls)
	}
}

func TestWebhookRejectsBadSignature(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	body := `{"type":"checkout.session.completed","data":{"object":{"id":"cs_test_123","payment_status":"paid"}}}`
	confirmer := &stubConfirmer{}
	handler := newHandler(confirmer, now)

	request := httptest.NewRequest(http.MethodPost, "/webhooks/stripe", strings.NewReader(body))
	request.Header.Set("Stripe-Signature", "t="+strconv.FormatInt(now.Unix(), 10)+",v1=deadbeef")

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d", recorder.Code)
	}
	if len(confirmer.calls) != 0 {
		t.Fatalf("expected no confirm calls, got %v", confirmer.calls)
	}
}

func TestWebhookRejectsMissingSignatureHeader(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	body := `{"type":"checkout.session.completed","data":{"object":{"id":"cs_test_123","payment_status":"paid"}}}`
	handler := newHandler(&stubConfirmer{}, now)

	request := httptest.NewRequest(http.MethodPost, "/webhooks/stripe", strings.NewReader(body))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d", recorder.Code)
	}
}

func TestWebhookRejectsStaleTimestamp(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	body := `{"type":"checkout.session.completed","data":{"object":{"id":"cs_test_123","payment_status":"paid"}}}`
	confirmer := &stubConfirmer{}
	handler := newHandler(confirmer, now)

	staleTimestamp := now.Add(-10 * time.Minute).Unix()
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, signedRequest(t, "whsec_test", body, staleTimestamp))

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d", recorder.Code)
	}
	if len(confirmer.calls) != 0 {
		t.Fatalf("expected no confirm calls, got %v", confirmer.calls)
	}
}

func TestWebhookReturns200WhenPaymentNotFound(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	body := `{"type":"checkout.session.completed","data":{"object":{"id":"cs_unknown","payment_status":"paid"}}}`
	confirmer := &stubConfirmer{err: domain.ErrPaymentNotFound}
	handler := newHandler(confirmer, now)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, signedRequest(t, "whsec_test", body, now.Unix()))

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", recorder.Code)
	}
}

func TestWebhookReturns500OnConfirmerError(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	body := `{"type":"checkout.session.completed","data":{"object":{"id":"cs_test_123","payment_status":"paid"}}}`
	confirmer := &stubConfirmer{err: errors.New("db down")}
	handler := newHandler(confirmer, now)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, signedRequest(t, "whsec_test", body, now.Unix()))

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("unexpected status: %d", recorder.Code)
	}
}

func TestWebhookRejectsNonPost(t *testing.T) {
	handler := newHandler(&stubConfirmer{}, time.Unix(1_700_000_000, 0))
	request := httptest.NewRequest(http.MethodGet, "/webhooks/stripe", io.NopCloser(strings.NewReader(""))) //nolint:gosec

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("unexpected status: %d", recorder.Code)
	}
}
