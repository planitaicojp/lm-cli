package api

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/crowdy/lm-cli/internal/config"
	lmerrors "github.com/crowdy/lm-cli/internal/errors"
)

// IssueStatelessToken obtains a stateless channel access token via
// POST https://api.line.me/oauth2/v2.1/token
func IssueStatelessToken(channelID, channelSecret string) (string, time.Time, error) {
	baseURL := config.EnvOr(config.EnvEndpoint, defaultBaseURL)
	endpoint := baseURL + "/oauth2/v2.1/token"

	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", channelID)
	form.Set("client_secret", channelSecret)

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBufferString(form.Encode()))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("creating token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", UserAgent)

	debugLogRequest(req, []byte(form.Encode()))

	client := &http.Client{Timeout: 30 * time.Second}
	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return "", time.Time{}, &lmerrors.NetworkError{Err: err}
	}
	defer resp.Body.Close()

	elapsed := time.Since(start)
	if debugLevel >= DebugAPI {
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body = io.NopCloser(bytes.NewReader(respBody))
		debugLogResponse(resp, elapsed, respBody)
	} else {
		debugLogResponse(resp, elapsed, nil)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", time.Time{}, &lmerrors.AuthError{
			Message: fmt.Sprintf("token issuance failed (HTTP %d): %s", resp.StatusCode, string(body)),
		}
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		TokenType   string `json:"token_type"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", time.Time{}, fmt.Errorf("decoding token response: %w", err)
	}

	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	return tokenResp.AccessToken, expiresAt, nil
}

// IssueV2Token obtains a v2.1 channel access token using JWT assertion.
// It reads the RSA private key from privateKeyPath, creates a signed JWT,
// and exchanges it for an access token via POST /oauth2/v2.1/token.
func IssueV2Token(channelID, privateKeyPath string) (string, time.Time, error) {
	keyData, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("reading private key: %w", err)
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return "", time.Time{}, &lmerrors.ValidationError{
			Field:   "private_key_file",
			Message: "failed to decode PEM block from private key file",
		}
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("parsing private key: %w", err)
	}
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return "", time.Time{}, &lmerrors.ValidationError{
			Field:   "private_key_file",
			Message: "private key is not RSA",
		}
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"iss":       channelID,
		"sub":       channelID,
		"aud":       "https://api.line.biz/",
		"exp":       now.Add(30 * time.Minute).Unix(),
		"token_exp": 2592000,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	assertion, err := token.SignedString(rsaKey)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("signing JWT assertion: %w", err)
	}

	baseURL := config.EnvOr(config.EnvEndpoint, defaultBaseURL)
	endpoint := baseURL + "/oauth2/v2.1/token"

	form := url.Values{}
	form.Set("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
	form.Set("assertion", assertion)

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBufferString(form.Encode()))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("creating token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", UserAgent)

	debugLogRequest(req, []byte(form.Encode()))

	client := &http.Client{Timeout: 30 * time.Second}
	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return "", time.Time{}, &lmerrors.NetworkError{Err: err}
	}
	defer resp.Body.Close()

	elapsed := time.Since(start)
	if debugLevel >= DebugAPI {
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body = io.NopCloser(bytes.NewReader(respBody))
		debugLogResponse(resp, elapsed, respBody)
	} else {
		debugLogResponse(resp, elapsed, nil)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", time.Time{}, &lmerrors.AuthError{
			Message: fmt.Sprintf("v2 token issuance failed (HTTP %d): %s", resp.StatusCode, string(body)),
		}
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		TokenType   string `json:"token_type"`
		KeyID       string `json:"key_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", time.Time{}, fmt.Errorf("decoding token response: %w", err)
	}

	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	return tokenResp.AccessToken, expiresAt, nil
}

// EnsureToken returns a valid token for the given profile.
// Priority: LM_TOKEN env var > cached valid token > re-issue (stateless) > error.
func EnsureToken(profile string, cfg *config.Config, creds *config.CredentialsStore, tokens *config.TokenStore) (string, error) {
	// 1. Environment variable bypass
	if t := config.EnvOr(config.EnvToken, ""); t != "" {
		return t, nil
	}

	// 2. Cached token
	if tokens.IsValid(profile) {
		entry, _ := tokens.Get(profile)
		return entry.Token, nil
	}

	// 3. Need to (re-)issue
	p, ok := cfg.Profiles[profile]
	if !ok {
		return "", &lmerrors.ConfigError{
			Message: fmt.Sprintf("profile %q not found, run 'lm auth login'", profile),
		}
	}

	tokenType := p.TokenType
	if tokenType == "" {
		tokenType = "longterm"
	}

	switch tokenType {
	case "longterm":
		// longterm tokens don't auto-refresh — require manual login
		if entry, ok := tokens.Get(profile); ok && entry.Token != "" {
			return entry.Token, nil
		}
		return "", &lmerrors.AuthError{
			Message: "no token found, run 'lm auth login'",
		}

	case "stateless":
		cred, ok := creds.Get(profile)
		if !ok || cred.ChannelSecret == "" {
			return "", &lmerrors.AuthError{
				Message: fmt.Sprintf("no credentials for profile %q, run 'lm auth login'", profile),
			}
		}
		channelID := config.EnvOr(config.EnvChannelID, p.ChannelID)
		secret := config.EnvOr(config.EnvSecret, cred.ChannelSecret)

		token, expiresAt, err := IssueStatelessToken(channelID, secret)
		if err != nil {
			return "", err
		}

		tokens.Set(profile, config.TokenEntry{
			Token:     token,
			ExpiresAt: expiresAt,
			TokenType: "stateless",
		})
		_ = tokens.Save()
		return token, nil

	case "v2":
		cred, ok := creds.Get(profile)
		if !ok || cred.PrivateKeyFile == "" {
			return "", &lmerrors.AuthError{
				Message: fmt.Sprintf("no private key configured for profile %q, run 'lm auth login --type v2'", profile),
			}
		}
		channelID := config.EnvOr(config.EnvChannelID, p.ChannelID)

		token, expiresAt, err := IssueV2Token(channelID, cred.PrivateKeyFile)
		if err != nil {
			return "", err
		}

		tokens.Set(profile, config.TokenEntry{
			Token:     token,
			ExpiresAt: expiresAt,
			TokenType: "v2",
		})
		_ = tokens.Save()
		return token, nil

	default:
		return "", &lmerrors.ConfigError{
			Message: fmt.Sprintf("unsupported token type %q, run 'lm auth login'", tokenType),
		}
	}
}
