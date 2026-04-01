package api

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/crowdy/lm-cli/internal/config"
)

func generateTestKey(t *testing.T, dir string) string {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generating RSA key: %v", err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatalf("marshaling key: %v", err)
	}
	pemBlock := &pem.Block{Type: "PRIVATE KEY", Bytes: der}
	path := filepath.Join(dir, "private.pem")
	if err := os.WriteFile(path, pem.EncodeToMemory(pemBlock), 0600); err != nil {
		t.Fatalf("writing key file: %v", err)
	}
	return path
}

func TestIssueV2Token(t *testing.T) {
	dir := t.TempDir()
	keyPath := generateTestKey(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oauth2/v2.1/token" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parsing form: %v", err)
		}
		if gt := r.FormValue("grant_type"); gt != "urn:ietf:params:oauth:grant-type:jwt-bearer" {
			t.Errorf("unexpected grant_type: %s", gt)
		}
		if r.FormValue("assertion") == "" {
			t.Error("assertion is empty")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "v2-token-abc",
			"expires_in":   2592000,
			"token_type":   "Bearer",
			"key_id":       "kid-123",
		})
	}))
	defer srv.Close()

	t.Setenv("LM_ENDPOINT", srv.URL)

	token, expiresAt, err := IssueV2Token("channel123", keyPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "v2-token-abc" {
		t.Errorf("expected v2-token-abc, got %s", token)
	}
	if time.Until(expiresAt) < 29*24*time.Hour {
		t.Errorf("expected ~30 days expiry, got %v", expiresAt)
	}
}

func TestIssueV2Token_InvalidKeyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.pem")
	if err := os.WriteFile(path, []byte("not a pem"), 0600); err != nil {
		t.Fatal(err)
	}

	_, _, err := IssueV2Token("channel123", path)
	if err == nil {
		t.Fatal("expected error for invalid key file")
	}
}

func TestIssueV2Token_MissingKeyFile(t *testing.T) {
	_, _, err := IssueV2Token("channel123", "/nonexistent/key.pem")
	if err == nil {
		t.Fatal("expected error for missing key file")
	}
}

func TestEnsureToken_V2(t *testing.T) {
	dir := t.TempDir()
	keyPath := generateTestKey(t, dir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "v2-refreshed",
			"expires_in":   2592000,
			"token_type":   "Bearer",
		})
	}))
	defer srv.Close()

	t.Setenv("LM_ENDPOINT", srv.URL)
	t.Setenv("LM_CONFIG_DIR", dir)

	cfg := &config.Config{
		Profiles: map[string]config.Profile{
			"prod": {ChannelID: "ch123", TokenType: "v2"},
		},
	}
	creds := &config.CredentialsStore{
		Profiles: map[string]config.Credentials{
			"prod": {PrivateKeyFile: keyPath},
		},
	}
	tokens := &config.TokenStore{
		Profiles: map[string]config.TokenEntry{},
	}

	token, err := EnsureToken("prod", cfg, creds, tokens)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "v2-refreshed" {
		t.Errorf("expected v2-refreshed, got %s", token)
	}

	// Token should be cached now
	entry, ok := tokens.Get("prod")
	if !ok || entry.Token != "v2-refreshed" {
		t.Error("token not cached after EnsureToken")
	}
}

func TestEnsureToken_V2_NoCreds(t *testing.T) {
	cfg := &config.Config{
		Profiles: map[string]config.Profile{
			"prod": {ChannelID: "ch123", TokenType: "v2"},
		},
	}
	creds := &config.CredentialsStore{
		Profiles: map[string]config.Credentials{},
	}
	tokens := &config.TokenStore{
		Profiles: map[string]config.TokenEntry{},
	}

	_, err := EnsureToken("prod", cfg, creds, tokens)
	if err == nil {
		t.Fatal("expected error when no credentials")
	}
}

func TestEnsureToken_EnvBypass(t *testing.T) {
	t.Setenv("LM_TOKEN", "env-token-xyz")

	cfg := &config.Config{Profiles: map[string]config.Profile{}}
	creds := &config.CredentialsStore{Profiles: map[string]config.Credentials{}}
	tokens := &config.TokenStore{Profiles: map[string]config.TokenEntry{}}

	token, err := EnsureToken("default", cfg, creds, tokens)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "env-token-xyz" {
		t.Errorf("expected env-token-xyz, got %s", token)
	}
}

func TestEnsureToken_CachedValid(t *testing.T) {
	cfg := &config.Config{
		Profiles: map[string]config.Profile{
			"test": {ChannelID: "ch1", TokenType: "stateless"},
		},
	}
	creds := &config.CredentialsStore{Profiles: map[string]config.Credentials{}}
	tokens := &config.TokenStore{
		Profiles: map[string]config.TokenEntry{
			"test": {
				Token:     "cached-token",
				ExpiresAt: time.Now().Add(1 * time.Hour),
				TokenType: "stateless",
			},
		},
	}

	token, err := EnsureToken("test", cfg, creds, tokens)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "cached-token" {
		t.Errorf("expected cached-token, got %s", token)
	}
}
