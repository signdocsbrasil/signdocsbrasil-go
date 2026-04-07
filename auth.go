package signdocsbrasil

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// authHandler manages OAuth2 token acquisition and caching.
type authHandler struct {
	clientID     string
	clientSecret string
	privateKey   *ecdsa.PrivateKey
	kid          string
	tokenURL     string
	scopes       []string
	httpClient   *http.Client

	mu          sync.Mutex
	cachedToken *cachedToken
}

type cachedToken struct {
	accessToken string
	expiresAt   time.Time
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

func newAuthHandler(cfg *Config) *authHandler {
	return &authHandler{
		clientID:     cfg.ClientID,
		clientSecret: cfg.ClientSecret,
		privateKey:   cfg.PrivateKey,
		kid:          cfg.Kid,
		tokenURL:     cfg.BaseURL + "/oauth2/token",
		scopes:       cfg.Scopes,
		httpClient:   cfg.HTTPClient,
	}
}

// getAccessToken returns a valid access token, refreshing if necessary.
// It is safe for concurrent use.
func (a *authHandler) getAccessToken() (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.cachedToken != nil && time.Now().Before(a.cachedToken.expiresAt.Add(-30*time.Second)) {
		return a.cachedToken.accessToken, nil
	}

	return a.fetchToken()
}

func (a *authHandler) fetchToken() (string, error) {
	params := url.Values{}
	params.Set("grant_type", "client_credentials")
	params.Set("client_id", a.clientID)
	params.Set("scope", strings.Join(a.scopes, " "))

	if a.clientSecret != "" {
		params.Set("client_secret", a.clientSecret)
	} else if a.privateKey != nil {
		assertion, err := a.buildJWTAssertion()
		if err != nil {
			return "", &AuthenticationError{Message: "failed to build JWT assertion", Err: err}
		}
		params.Set("client_assertion_type", "urn:ietf:params:oauth:client-assertion-type:jwt-bearer")
		params.Set("client_assertion", assertion)
	}

	req, err := http.NewRequest(http.MethodPost, a.tokenURL, strings.NewReader(params.Encode()))
	if err != nil {
		return "", &AuthenticationError{Message: "failed to create token request", Err: err}
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return "", &AuthenticationError{Message: "token request failed", Err: err}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", &AuthenticationError{Message: "failed to read token response", Err: err}
	}

	if resp.StatusCode != http.StatusOK {
		return "", &AuthenticationError{
			Message: fmt.Sprintf("token request failed (%d): %s", resp.StatusCode, string(body)),
		}
	}

	var tokenResp tokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", &AuthenticationError{Message: "failed to parse token response", Err: err}
	}

	a.cachedToken = &cachedToken{
		accessToken: tokenResp.AccessToken,
		expiresAt:   time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
	}

	return tokenResp.AccessToken, nil
}

func (a *authHandler) buildJWTAssertion() (string, error) {
	now := time.Now().Unix()

	jti, err := generateUUID()
	if err != nil {
		return "", fmt.Errorf("generate jti: %w", err)
	}

	header := map[string]string{
		"alg": "ES256",
		"typ": "JWT",
		"kid": a.kid,
	}
	payload := map[string]any{
		"iss": a.clientID,
		"sub": a.clientID,
		"aud": a.tokenURL,
		"exp": now + 300,
		"iat": now,
		"jti": jti,
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("marshal header: %w", err)
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal payload: %w", err)
	}

	encodedHeader := base64URLEncode(headerJSON)
	encodedPayload := base64URLEncode(payloadJSON)
	signingInput := encodedHeader + "." + encodedPayload

	hash := sha256.Sum256([]byte(signingInput))
	r, s, err := ecdsa.Sign(rand.Reader, a.privateKey, hash[:])
	if err != nil {
		return "", fmt.Errorf("sign JWT: %w", err)
	}

	// ES256 requires the signature to be exactly 64 bytes: r (32 bytes) || s (32 bytes)
	curveBits := a.privateKey.Curve.Params().BitSize
	keyBytes := (curveBits + 7) / 8

	rBytes := r.Bytes()
	sBytes := s.Bytes()

	sig := make([]byte, 2*keyBytes)
	copy(sig[keyBytes-len(rBytes):keyBytes], rBytes)
	copy(sig[2*keyBytes-len(sBytes):], sBytes)

	encodedSig := base64URLEncode(sig)
	return signingInput + "." + encodedSig, nil
}

func base64URLEncode(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

// generateUUID produces a v4 UUID string using crypto/rand.
func generateUUID() (string, error) {
	var uuid [16]byte
	if _, err := io.ReadFull(rand.Reader, uuid[:]); err != nil {
		return "", err
	}
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // variant 10
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16]), nil
}

// ParseES256PrivateKeyFromPEM parses a PEM-encoded ECDSA P-256 private key.
// It supports both PKCS#8 ("BEGIN PRIVATE KEY") and SEC1 ("BEGIN EC PRIVATE KEY")
// PEM formats.
//
// Users may also use crypto/x509 directly and pass the resulting
// *ecdsa.PrivateKey to WithPrivateKey.
func ParseES256PrivateKeyFromPEM(pemData []byte) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("signdocsbrasil: no PEM block found in input")
	}

	var key any
	var err error

	switch block.Type {
	case "EC PRIVATE KEY":
		key, err = x509.ParseECPrivateKey(block.Bytes)
	case "PRIVATE KEY":
		key, err = x509.ParsePKCS8PrivateKey(block.Bytes)
	default:
		return nil, fmt.Errorf("signdocsbrasil: unsupported PEM block type %q (expected \"EC PRIVATE KEY\" or \"PRIVATE KEY\")", block.Type)
	}
	if err != nil {
		return nil, fmt.Errorf("signdocsbrasil: failed to parse private key: %w", err)
	}

	ecKey, ok := key.(*ecdsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("signdocsbrasil: parsed key is not an ECDSA key")
	}
	if ecKey.Curve != elliptic.P256() {
		return nil, fmt.Errorf("signdocsbrasil: key must use P-256 curve for ES256, got %s", ecKey.Curve.Params().Name)
	}

	return ecKey, nil
}
