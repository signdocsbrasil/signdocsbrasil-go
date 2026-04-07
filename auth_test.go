package signdocsbrasil

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func setupTokenServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *authHandler) {
	t.Helper()
	server := httptest.NewServer(handler)
	auth := &authHandler{
		clientID:     "test-client",
		clientSecret: "test-secret",
		tokenURL:     server.URL + "/oauth2/token",
		scopes:       []string{"transactions:read"},
		httpClient:   server.Client(),
	}
	return server, auth
}

func TestAuthHandler_ClientSecretFlow(t *testing.T) {
	server, auth := setupTokenServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		ct := r.Header.Get("Content-Type")
		if !strings.Contains(ct, "application/x-www-form-urlencoded") {
			t.Errorf("expected form-urlencoded, got %s", ct)
		}
		r.ParseForm()
		if r.FormValue("grant_type") != "client_credentials" {
			t.Errorf("expected client_credentials, got %s", r.FormValue("grant_type"))
		}
		if r.FormValue("client_id") != "test-client" {
			t.Errorf("expected test-client, got %s", r.FormValue("client_id"))
		}
		if r.FormValue("client_secret") != "test-secret" {
			t.Errorf("expected test-secret, got %s", r.FormValue("client_secret"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tokenResponse{
			AccessToken: "tok_123",
			ExpiresIn:   3600,
		})
	})
	defer server.Close()

	token, err := auth.getAccessToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "tok_123" {
		t.Errorf("expected tok_123, got %s", token)
	}
}

func TestAuthHandler_TokenCaching(t *testing.T) {
	calls := 0
	server, auth := setupTokenServer(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tokenResponse{AccessToken: "tok_cached", ExpiresIn: 3600})
	})
	defer server.Close()

	t1, _ := auth.getAccessToken()
	t2, _ := auth.getAccessToken()

	if t1 != "tok_cached" || t2 != "tok_cached" {
		t.Error("expected cached token")
	}
	if calls != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}

func TestAuthHandler_RefreshWithin30sBuffer(t *testing.T) {
	calls := 0
	server, auth := setupTokenServer(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		if calls == 1 {
			json.NewEncoder(w).Encode(tokenResponse{AccessToken: "tok_1", ExpiresIn: 20})
		} else {
			json.NewEncoder(w).Encode(tokenResponse{AccessToken: "tok_2", ExpiresIn: 3600})
		}
	})
	defer server.Close()

	t1, _ := auth.getAccessToken()
	if t1 != "tok_1" {
		t.Errorf("expected tok_1, got %s", t1)
	}

	// 20s < 30s buffer, next call should refresh
	t2, _ := auth.getAccessToken()
	if t2 != "tok_2" {
		t.Errorf("expected tok_2, got %s", t2)
	}
	if calls != 2 {
		t.Errorf("expected 2 calls, got %d", calls)
	}
}

func TestAuthHandler_ErrorOnNon200(t *testing.T) {
	server, auth := setupTokenServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		w.Write([]byte("unauthorized"))
	})
	defer server.Close()

	_, err := auth.getAccessToken()
	if err == nil {
		t.Fatal("expected error")
	}
	var authErr *AuthenticationError
	if ok := isAuthErr(err, &authErr); !ok {
		t.Errorf("expected AuthenticationError, got %T", err)
	}
}

func isAuthErr(err error, target **AuthenticationError) bool {
	ae, ok := err.(*AuthenticationError)
	if ok {
		*target = ae
	}
	return ok
}

func TestAuthHandler_PrivateKeyJwtFlow(t *testing.T) {
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		if r.FormValue("client_assertion") == "" {
			t.Error("expected client_assertion")
		}
		if r.FormValue("client_assertion_type") == "" {
			t.Error("expected client_assertion_type")
		}
		if r.FormValue("client_secret") != "" {
			t.Error("should not have client_secret")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tokenResponse{AccessToken: "tok_jwt", ExpiresIn: 3600})
	}))
	defer server.Close()

	auth := &authHandler{
		clientID:   "test-client",
		privateKey: key,
		kid:        "key-001",
		tokenURL:   server.URL + "/oauth2/token",
		scopes:     []string{"transactions:read"},
		httpClient: server.Client(),
	}

	token, err := auth.getAccessToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "tok_jwt" {
		t.Errorf("expected tok_jwt, got %s", token)
	}
}

func TestParseES256PrivateKeyFromPEM_Valid(t *testing.T) {
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	der, _ := x509.MarshalECPrivateKey(key)
	pemBlock := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: der})

	parsed, err := ParseES256PrivateKeyFromPEM(pemBlock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.Curve != elliptic.P256() {
		t.Error("expected P-256 curve")
	}
}

func TestParseES256PrivateKeyFromPEM_InvalidPEM(t *testing.T) {
	_, err := ParseES256PrivateKeyFromPEM([]byte("not pem data"))
	if err == nil {
		t.Fatal("expected error for invalid PEM")
	}
}
