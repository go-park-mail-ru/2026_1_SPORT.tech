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
	calls            []string
	subscriptionCall [2]string
	expireCalls      []string
	failCalls        []string
	renewCalls       []string
	deactivateCalls  []string
	renewPeriodEnd   time.Time
	err              error
}

func (stub *stubConfirmer) ConfirmPaymentFromProvider(ctx context.Context, providerPaymentID string) (domain.DonationPayment, error) {
	stub.calls = append(stub.calls, providerPaymentID)
	return domain.DonationPayment{ProviderPaymentID: providerPaymentID}, stub.err
}

func (stub *stubConfirmer) ConfirmSubscriptionPaymentFromProvider(ctx context.Context, providerPaymentID string, stripeSubscriptionID string) (domain.DonationPayment, error) {
	stub.subscriptionCall = [2]string{providerPaymentID, stripeSubscriptionID}
	return domain.DonationPayment{ProviderPaymentID: providerPaymentID}, stub.err
}

func (stub *stubConfirmer) ExpirePaymentFromProvider(ctx context.Context, providerPaymentID string) (domain.DonationPayment, error) {
	stub.expireCalls = append(stub.expireCalls, providerPaymentID)
	return domain.DonationPayment{ProviderPaymentID: providerPaymentID}, stub.err
}

func (stub *stubConfirmer) FailPaymentFromProvider(ctx context.Context, providerPaymentID string) (domain.DonationPayment, error) {
	stub.failCalls = append(stub.failCalls, providerPaymentID)
	return domain.DonationPayment{ProviderPaymentID: providerPaymentID}, stub.err
}

func (stub *stubConfirmer) RenewSubscriptionFromProvider(ctx context.Context, stripeSubscriptionID string, currentPeriodEnd time.Time) error {
	stub.renewCalls = append(stub.renewCalls, stripeSubscriptionID)
	stub.renewPeriodEnd = currentPeriodEnd
	return stub.err
}

func (stub *stubConfirmer) DeactivateSubscriptionFromProvider(ctx context.Context, stripeSubscriptionID string) error {
	stub.deactivateCalls = append(stub.deactivateCalls, stripeSubscriptionID)
	return stub.err
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

func TestWebhookExpiresCheckoutSession(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	body := `{"type":"checkout.session.expired","data":{"object":{"id":"cs_test_123","payment_status":"unpaid"}}}`
	confirmer := &stubConfirmer{}
	handler := newHandler(confirmer, now)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, signedRequest(t, "whsec_test", body, now.Unix()))

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d, body=%s", recorder.Code, recorder.Body.String())
	}
	if len(confirmer.expireCalls) != 1 || confirmer.expireCalls[0] != "cs_test_123" {
		t.Fatalf("unexpected expire calls: %v", confirmer.expireCalls)
	}
	if len(confirmer.calls) != 0 {
		t.Fatalf("expected no confirm calls, got %v", confirmer.calls)
	}
}

func TestWebhookFailsCheckoutSession(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	body := `{"type":"checkout.session.async_payment_failed","data":{"object":{"id":"cs_test_123","payment_status":"unpaid"}}}`
	confirmer := &stubConfirmer{}
	handler := newHandler(confirmer, now)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, signedRequest(t, "whsec_test", body, now.Unix()))

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d, body=%s", recorder.Code, recorder.Body.String())
	}
	if len(confirmer.failCalls) != 1 || confirmer.failCalls[0] != "cs_test_123" {
		t.Fatalf("unexpected fail calls: %v", confirmer.failCalls)
	}
}

func TestWebhookConfirmsSubscriptionCheckout(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	body := `{"type":"checkout.session.completed","data":{"object":{"id":"cs_sub_1","mode":"subscription","payment_status":"paid","subscription":"sub_123"}}}`
	confirmer := &stubConfirmer{}
	handler := newHandler(confirmer, now)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, signedRequest(t, "whsec_test", body, now.Unix()))

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d, body=%s", recorder.Code, recorder.Body.String())
	}
	if confirmer.subscriptionCall != [2]string{"cs_sub_1", "sub_123"} {
		t.Fatalf("unexpected subscription confirm call: %v", confirmer.subscriptionCall)
	}
	if len(confirmer.calls) != 0 {
		t.Fatalf("expected no one-off confirm calls, got %v", confirmer.calls)
	}
}

func TestWebhookSubscriptionCheckoutMissingSubscriptionID(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	body := `{"type":"checkout.session.completed","data":{"object":{"id":"cs_sub_1","mode":"subscription","payment_status":"paid"}}}`
	handler := newHandler(&stubConfirmer{}, now)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, signedRequest(t, "whsec_test", body, now.Unix()))

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d", recorder.Code)
	}
}

func TestWebhookRenewsSubscriptionOnInvoicePaid(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	periodEnd := int64(1_700_500_000)
	body := `{"type":"invoice.paid","data":{"object":{"subscription":"sub_123","lines":{"data":[{"period":{"end":1700500000}}]}}}}`
	confirmer := &stubConfirmer{}
	handler := newHandler(confirmer, now)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, signedRequest(t, "whsec_test", body, now.Unix()))

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d, body=%s", recorder.Code, recorder.Body.String())
	}
	if len(confirmer.renewCalls) != 1 || confirmer.renewCalls[0] != "sub_123" {
		t.Fatalf("unexpected renew calls: %v", confirmer.renewCalls)
	}
	if !confirmer.renewPeriodEnd.Equal(time.Unix(periodEnd, 0).UTC()) {
		t.Fatalf("unexpected renew period end: %v", confirmer.renewPeriodEnd)
	}
}

func TestWebhookIgnoresInvoiceWithoutSubscription(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	body := `{"type":"invoice.paid","data":{"object":{"lines":{"data":[{"period":{"end":1700500000}}]}}}}`
	confirmer := &stubConfirmer{}
	handler := newHandler(confirmer, now)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, signedRequest(t, "whsec_test", body, now.Unix()))

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", recorder.Code)
	}
	if len(confirmer.renewCalls) != 0 {
		t.Fatalf("expected no renew calls, got %v", confirmer.renewCalls)
	}
}

func TestWebhookDeactivatesDeletedSubscription(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	body := `{"type":"customer.subscription.deleted","data":{"object":{"id":"sub_123","status":"canceled"}}}`
	confirmer := &stubConfirmer{}
	handler := newHandler(confirmer, now)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, signedRequest(t, "whsec_test", body, now.Unix()))

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d, body=%s", recorder.Code, recorder.Body.String())
	}
	if len(confirmer.deactivateCalls) != 1 || confirmer.deactivateCalls[0] != "sub_123" {
		t.Fatalf("unexpected deactivate calls: %v", confirmer.deactivateCalls)
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
