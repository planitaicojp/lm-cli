package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/crowdy/lm-cli/internal/model"
	lmerrors "github.com/crowdy/lm-cli/internal/errors"
)

// RichMenuAPI provides LINE rich menu operations.
type RichMenuAPI struct {
	Client *Client
}

func (a *RichMenuAPI) Create(menu model.RichMenu) (*model.RichMenuIDResponse, error) {
	var resp model.RichMenuIDResponse
	_, err := a.Client.Post(a.Client.BaseURL+"/v2/bot/richmenu", menu, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (a *RichMenuAPI) Get(richMenuID string) (*model.RichMenu, error) {
	var menu model.RichMenu
	url := fmt.Sprintf("%s/v2/bot/richmenu/%s", a.Client.BaseURL, richMenuID)
	if err := a.Client.Get(url, &menu); err != nil {
		return nil, err
	}
	return &menu, nil
}

func (a *RichMenuAPI) List() ([]model.RichMenu, error) {
	var resp model.RichMenuListResponse
	if err := a.Client.Get(a.Client.BaseURL+"/v2/bot/richmenu/list", &resp); err != nil {
		return nil, err
	}
	return resp.Richmenus, nil
}

func (a *RichMenuAPI) Delete(richMenuID string) error {
	url := fmt.Sprintf("%s/v2/bot/richmenu/%s", a.Client.BaseURL, richMenuID)
	return a.Client.Delete(url)
}

// UploadImage uploads an image file for a rich menu.
func (a *RichMenuAPI) UploadImage(richMenuID, imagePath string) error {
	const maxImageSize = 1 * 1024 * 1024 // 1MB

	data, err := os.ReadFile(imagePath)
	if err != nil {
		return fmt.Errorf("reading image file: %w", err)
	}

	if len(data) > maxImageSize {
		return &lmerrors.ValidationError{
			Field:   "image",
			Message: fmt.Sprintf("file size %d bytes exceeds 1MB limit", len(data)),
		}
	}

	ext := strings.ToLower(filepath.Ext(imagePath))
	var contentType string
	switch ext {
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	default:
		return &lmerrors.ValidationError{Field: "image", Message: "must be .jpg or .png"}
	}

	url := fmt.Sprintf("https://api-data.line.me/v2/bot/richmenu/%s/content", richMenuID)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("creating upload request: %w", err)
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("User-Agent", UserAgent)
	if a.Client.Token != "" {
		req.Header.Set("Authorization", "Bearer "+a.Client.Token)
	}

	resp, err := a.Client.HTTP.Do(req)
	if err != nil {
		return &lmerrors.NetworkError{Err: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return parseAPIError(resp)
	}
	return nil
}

func (a *RichMenuAPI) GetDefault() (string, error) {
	var resp struct {
		RichMenuID string `json:"richMenuId"`
	}
	if err := a.Client.Get(a.Client.BaseURL+"/v2/bot/user/all/richmenu", &resp); err != nil {
		return "", err
	}
	return resp.RichMenuID, nil
}

func (a *RichMenuAPI) SetDefault(richMenuID string) error {
	url := fmt.Sprintf("%s/v2/bot/user/all/richmenu/%s", a.Client.BaseURL, richMenuID)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	_, err = a.Client.Do(req)
	return err
}

func (a *RichMenuAPI) UnsetDefault() error {
	return a.Client.Delete(a.Client.BaseURL + "/v2/bot/user/all/richmenu")
}

func (a *RichMenuAPI) CreateAlias(body any) error {
	_, err := a.Client.Post(a.Client.BaseURL+"/v2/bot/richmenu/alias", body, nil)
	return err
}

func (a *RichMenuAPI) ListAliases() ([]model.RichMenuAlias, error) {
	var resp model.RichMenuAliasListResponse
	if err := a.Client.Get(a.Client.BaseURL+"/v2/bot/richmenu/alias/list", &resp); err != nil {
		return nil, err
	}
	return resp.Aliases, nil
}

// parseJSONFile reads a JSON file and unmarshals it into dst.
func parseJSONFile(path string, dst any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading file %s: %w", path, err)
	}
	if err := json.Unmarshal(data, dst); err != nil {
		return fmt.Errorf("parsing JSON from %s: %w", path, err)
	}
	return nil
}

// ParseJSONFile is exported for use in cmd packages.
func ParseJSONFile(path string, dst any) error {
	return parseJSONFile(path, dst)
}

