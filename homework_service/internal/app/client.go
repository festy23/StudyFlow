package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type UserClient struct {
	baseURL    string
	httpClient *http.Client
}

type File struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	Size      int64     `json:"size"`
	MimeType  string    `json:"mime_type"`
	CreatedAt time.Time `json:"created_at"`
	OwnerID   string    `json:"owner_id"`
}

func NewUserClient(baseURL string) *UserClient {
	return &UserClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *UserClient) UserExists(ctx context.Context, userID string) bool {
	url := fmt.Sprintf("%s/users/%s/exists", c.baseURL, userID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

func (c *UserClient) GetUserRole(ctx context.Context, userID string) (string, error) {
	url := fmt.Sprintf("%s/users/%s/role", c.baseURL, userID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get user role")
	}

	var result struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.Role, nil
}

func (c *UserClient) IsPair(ctx context.Context, tutorID, studentID string) bool {
	url := fmt.Sprintf("%s/pairs/check?tutor_id=%s&student_id=%s", c.baseURL, tutorID, studentID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

type FileClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewFileClient(baseURL string) *FileClient {
	return &FileClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *FileClient) FileExists(ctx context.Context, fileID string) bool {
	url := fmt.Sprintf("%s/files/%s/exists", c.baseURL, fileID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

func (c *FileClient) GetFileURL(ctx context.Context, fileID string, userID string) (string, error) {
	url := fmt.Sprintf("%s/files/%s/url?user_id=%s", c.baseURL, fileID, userID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get file URL")
	}

	var result struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.URL, nil
}

func (c *FileClient) GetFile(ctx context.Context, id string) (*File, error) {
	url := fmt.Sprintf("%s/files/%s", c.baseURL, id)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var file File
	if err := json.NewDecoder(resp.Body).Decode(&file); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &file, nil
}
