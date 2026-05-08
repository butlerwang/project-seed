package handler_test

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/butlerwang/project-seed/backend/internal/handler"
	"github.com/butlerwang/project-seed/backend/internal/storage"
	"github.com/stretchr/testify/assert"
)

func stripeSignature(secret, payload string, ts int64) string {
	msg := fmt.Sprintf("%d.%s", ts, payload)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(msg))
	sig := hex.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("t=%d,v1=%s", ts, sig)
}

func TestStripeWebhookRejectsInvalidSignature(t *testing.T) {
	cfg := testConfig()
	cfg.StripeWebhookSecret = "whsec_testsecret"
	r := handler.NewRouter(cfg, nil, nil, storage.NewMemoryStorage())

	body := []byte(`{"type":"checkout.session.completed"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/stripe", bytes.NewReader(body))
	req.Header.Set("Stripe-Signature", "t=0,v1=badsig")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestStripeWebhookDisabledWithoutSecret(t *testing.T) {
	r := handler.NewRouter(testConfig(), nil, nil, storage.NewMemoryStorage())

	body := []byte(`{"type":"checkout.session.completed"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/stripe", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotImplemented, rr.Code)
}

func TestStripeWebhookAcceptsValidSignature(t *testing.T) {
	cfg := testConfig()
	cfg.StripeWebhookSecret = "whsec_testsecret"
	r := handler.NewRouter(cfg, nil, nil, storage.NewMemoryStorage())

	body := []byte(`{"type":"checkout.session.completed","data":{"object":{}}}`)
	ts := time.Now().Unix()
	sig := stripeSignature("whsec_testsecret", string(body), ts)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/stripe", bytes.NewReader(body))
	req.Header.Set("Stripe-Signature", sig)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}
