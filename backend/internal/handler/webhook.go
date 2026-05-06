package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/stripe/stripe-go/v76"
)

type webhookHandler struct {
	secret string
}

func (h *webhookHandler) stripe(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, "read error", http.StatusBadRequest)
		return
	}

	if h.secret != "" {
		if err := verifyStripeSignature(r.Header.Get("Stripe-Signature"), body, h.secret); err != nil {
			http.Error(w, "invalid signature", http.StatusBadRequest)
			return
		}
	}

	var event stripe.Event
	if err := json.Unmarshal(body, &event); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	switch string(event.Type) {
	case "checkout.session.completed":
		slog.Info("stripe webhook: checkout.session.completed")
	case "customer.subscription.deleted":
		slog.Info("stripe webhook: customer.subscription.deleted")
	case "invoice.payment_failed":
		slog.Info("stripe webhook: invoice.payment_failed")
	default:
		slog.Info("stripe webhook: unhandled event", "type", event.Type)
	}

	w.WriteHeader(http.StatusOK)
}

func verifyStripeSignature(header string, body []byte, secret string) error {
	var ts int64
	var sig string

	for _, part := range strings.Split(header, ",") {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "t":
			parsed, err := strconv.ParseInt(kv[1], 10, 64)
			if err != nil {
				return err
			}
			ts = parsed
		case "v1":
			sig = kv[1]
		}
	}

	if ts == 0 || sig == "" {
		return fmt.Errorf("missing timestamp or signature")
	}
	if delta := time.Now().Unix() - ts; delta > 300 || delta < -300 {
		return fmt.Errorf("timestamp too old")
	}

	msg := fmt.Sprintf("%d.%s", ts, string(body))
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(msg))
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(sig)) {
		return fmt.Errorf("signature mismatch")
	}
	return nil
}
