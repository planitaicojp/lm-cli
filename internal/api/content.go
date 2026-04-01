package api

import (
	"fmt"
	"io"
	"net/http"

	lmerrors "github.com/crowdy/lm-cli/internal/errors"
)

// ContentAPI provides LINE content download operations.
type ContentAPI struct {
	Client *Client
}

// Get downloads binary content for a message ID.
// The caller is responsible for closing the returned ReadCloser.
func (a *ContentAPI) Get(messageID string) (io.ReadCloser, error) {
	url := fmt.Sprintf("%s/v2/bot/message/%s/content", a.Client.BaseURL, messageID)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", UserAgent)
	if a.Client.Token != "" {
		req.Header.Set("Authorization", "Bearer "+a.Client.Token)
	}

	resp, err := a.Client.HTTP.Do(req)
	if err != nil {
		return nil, &lmerrors.NetworkError{Err: err}
	}
	if resp.StatusCode >= 400 {
		return nil, parseAPIError(resp)
	}
	return resp.Body, nil
}
