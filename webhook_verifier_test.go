package signdocsbrasil

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
	"time"
)

func computeSignature(body, secret string, ts int64) string {
	signingInput := fmt.Sprintf("%d.%s", ts, body)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signingInput))
	return hex.EncodeToString(mac.Sum(nil))
}

func TestVerifyWebhookSignature_Valid(t *testing.T) {
	body := `{"event":"transaction.completed"}`
	secret := "whsec_test123"
	ts := time.Now().Unix()
	sig := computeSignature(body, secret, ts)

	if !VerifyWebhookSignature(body, sig, fmt.Sprintf("%d", ts), secret) {
		t.Error("expected valid signature to pass")
	}
}

func TestVerifyWebhookSignature_Invalid(t *testing.T) {
	body := `{"event":"test"}`
	ts := time.Now().Unix()

	if VerifyWebhookSignature(body, "invalid_hex", fmt.Sprintf("%d", ts), "secret") {
		t.Error("expected invalid signature to fail")
	}
}

func TestVerifyWebhookSignature_ExpiredTimestamp(t *testing.T) {
	body := `{"event":"test"}`
	secret := "whsec_test"
	ts := time.Now().Unix() - 400 // > 300s ago
	sig := computeSignature(body, secret, ts)

	if VerifyWebhookSignature(body, sig, fmt.Sprintf("%d", ts), secret) {
		t.Error("expected expired timestamp to fail")
	}
}

func TestVerifyWebhookSignature_FutureTimestamp(t *testing.T) {
	body := `{"event":"test"}`
	secret := "whsec_test"
	ts := time.Now().Unix() + 400 // > 300s in future
	sig := computeSignature(body, secret, ts)

	if VerifyWebhookSignature(body, sig, fmt.Sprintf("%d", ts), secret) {
		t.Error("expected future timestamp to fail")
	}
}

func TestVerifyWebhookSignature_CustomTolerance(t *testing.T) {
	body := `{"event":"test"}`
	secret := "whsec_test"
	ts := time.Now().Unix() - 100
	sig := computeSignature(body, secret, ts)

	if VerifyWebhookSignature(body, sig, fmt.Sprintf("%d", ts), secret, WithToleranceSeconds(50)) {
		t.Error("expected tolerance=50 to reject timestamp 100s ago")
	}
	if !VerifyWebhookSignature(body, sig, fmt.Sprintf("%d", ts), secret, WithToleranceSeconds(200)) {
		t.Error("expected tolerance=200 to accept timestamp 100s ago")
	}
}

func TestVerifyWebhookSignature_WrongSecret(t *testing.T) {
	body := `{"event":"test"}`
	ts := time.Now().Unix()
	sig := computeSignature(body, "correct_secret", ts)

	if VerifyWebhookSignature(body, sig, fmt.Sprintf("%d", ts), "wrong_secret") {
		t.Error("expected wrong secret to fail")
	}
}

func TestVerifyWebhookSignature_NonNumericTimestamp(t *testing.T) {
	if VerifyWebhookSignature("{}", "abc", "not-a-number", "secret") {
		t.Error("expected non-numeric timestamp to fail")
	}
}
