package signdocsbrasil

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"math"
	"strconv"
	"time"
)

const defaultToleranceSeconds = 300

// VerifyOption is a functional option for webhook signature verification.
type VerifyOption func(*verifyOptions)

type verifyOptions struct {
	toleranceSeconds int
}

// WithToleranceSeconds sets the maximum age (in seconds) of a webhook timestamp
// that will be accepted. The default is 300 seconds (5 minutes).
func WithToleranceSeconds(seconds int) VerifyOption {
	return func(o *verifyOptions) {
		o.toleranceSeconds = seconds
	}
}

// VerifyWebhookSignature verifies the HMAC-SHA256 signature of an incoming webhook.
//
// Parameters:
//   - body: the raw request body as a string
//   - signature: the value of the X-Signature header
//   - timestamp: the value of the X-Timestamp header (Unix epoch seconds)
//   - secret: the webhook signing secret returned when registering the webhook
//
// The function checks that:
//  1. The timestamp is a valid integer
//  2. The timestamp is within the tolerance window (default 300 seconds)
//  3. The HMAC-SHA256 of "{timestamp}.{body}" matches the provided signature
//
// It uses constant-time comparison to prevent timing attacks.
func VerifyWebhookSignature(body, signature, timestamp, secret string, opts ...VerifyOption) bool {
	o := &verifyOptions{
		toleranceSeconds: defaultToleranceSeconds,
	}
	for _, opt := range opts {
		opt(o)
	}

	// Parse timestamp
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return false
	}

	// Check timestamp tolerance
	now := time.Now().Unix()
	if math.Abs(float64(now-ts)) > float64(o.toleranceSeconds) {
		return false
	}

	// Compute expected signature: HMAC-SHA256("{timestamp}.{body}", secret)
	signingInput := timestamp + "." + body
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signingInput))
	expected := hex.EncodeToString(mac.Sum(nil))

	// Constant-time comparison
	return hmac.Equal([]byte(signature), []byte(expected))
}
