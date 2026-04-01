package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/crowdy/lm-cli/internal/config"
	lmerrors "github.com/crowdy/lm-cli/internal/errors"
)

// UserAgent is the User-Agent header value sent with all requests.
var UserAgent = "crowdy/lm-cli/dev"

const (
	defaultBaseURL = "https://api.line.me"
	defaultTimeout = 30 * time.Second
	maxRetries     = 3
)

// Client is the base HTTP client for LINE Messaging API.
type Client struct {
	HTTP    *http.Client
	Token   string
	BaseURL string
}

// NewClient creates a new API client.
func NewClient(token string) *Client {
	baseURL := defaultBaseURL
	if ep := os.Getenv(config.EnvEndpoint); ep != "" {
		if !strings.HasPrefix(ep, "https://") && os.Getenv("LM_ALLOW_HTTP") != "1" {
			panic(fmt.Sprintf("LM_ENDPOINT must start with https://, got: %s", ep))
		}
		baseURL = ep
	}
	return &Client{
		HTTP:    &http.Client{Timeout: defaultTimeout},
		Token:   token,
		BaseURL: baseURL,
	}
}

// Do executes an HTTP request with auth headers and error handling.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", UserAgent)
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	if req.Header.Get("Content-Type") == "" && req.Body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	// Read request body for debug logging
	var reqBody []byte
	if debugLevel >= DebugAPI && req.Body != nil {
		reqBody, _ = io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewReader(reqBody))
	}
	debugLogRequest(req, reqBody)

	start := time.Now()
	var resp *http.Response
	var err error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		resp, err = c.HTTP.Do(req)
		if err != nil {
			if attempt == maxRetries || req.Body != nil {
				return nil, &lmerrors.NetworkError{Err: err}
			}
			time.Sleep(time.Duration(attempt+1) * time.Second)
			continue
		}
		// Retry on 429 or 5xx for bodyless requests only
		if (resp.StatusCode == 429 || resp.StatusCode >= 500) && attempt < maxRetries && req.Body == nil {
			retryAfter := resp.Header.Get("Retry-After")
			resp.Body.Close()
			if d := parseRetryAfter(retryAfter); d > 0 {
				time.Sleep(d)
			} else {
				time.Sleep(time.Duration(attempt+1) * time.Second)
			}
			continue
		}
		break
	}
	elapsed := time.Since(start)

	if resp == nil {
		return nil, &lmerrors.NetworkError{Err: fmt.Errorf("no response after retries")}
	}

	if debugLevel >= DebugAPI {
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		resp.Body = io.NopCloser(bytes.NewReader(respBody))
		debugLogResponse(resp, elapsed, respBody)
	} else {
		debugLogResponse(resp, elapsed, nil)
	}

	if resp.StatusCode >= 400 {
		return resp, parseAPIError(resp)
	}

	return resp, nil
}

// Request creates and executes a request, returning the response.
func (c *Client) Request(method, url string, body any) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	return c.Do(req)
}

// Get performs a GET request and decodes the response into result.
func (c *Client) Get(url string, result any) error {
	resp, err := c.Request(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}

// Post performs a POST request and decodes the response into result.
func (c *Client) Post(url string, body, result any) (*http.Response, error) {
	resp, err := c.Request(http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return resp, err
		}
	}
	return resp, nil
}

// Put performs a PUT request.
func (c *Client) Put(url string, body, result any) error {
	resp, err := c.Request(http.MethodPut, url, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}

// Delete performs a DELETE request.
func (c *Client) Delete(url string) error {
	resp, err := c.Request(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// parseAPIError reads the response body and returns a LINE API error.
func parseAPIError(resp *http.Response) error {
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	// LINE API error format: {"message": "...", "details": [...]}
	var errResp struct {
		Message string                      `json:"message"`
		Details []lmerrors.APIErrorDetail   `json:"details"`
	}

	apiErr := &lmerrors.APIError{
		StatusCode: resp.StatusCode,
		Message:    string(body),
	}

	if json.Unmarshal(body, &errResp) == nil && errResp.Message != "" {
		apiErr.Message = errResp.Message
		apiErr.Details = errResp.Details
	}

	if resp.StatusCode == 429 {
		return &lmerrors.RateLimitError{RetryAfter: resp.Header.Get("Retry-After")}
	}
	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return &lmerrors.AuthError{Message: apiErr.Message}
	}
	if resp.StatusCode == 404 {
		return &lmerrors.NotFoundError{Resource: "resource", ID: ""}
	}

	return apiErr
}

// parseRetryAfter parses an HTTP Retry-After header value.
// Supports integer seconds (RFC 7231) and HTTP-date format.
func parseRetryAfter(header string) time.Duration {
	if header == "" {
		return 0
	}
	if n, err := strconv.Atoi(header); err == nil {
		return time.Duration(n) * time.Second
	}
	if t, err := http.ParseTime(header); err == nil {
		d := time.Until(t)
		if d > 0 {
			return d
		}
	}
	return 0
}
