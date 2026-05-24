package httpclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	inner   *http.Client
	baseURL string
	headers map[string]string
}

func New(baseURL string, timeout time.Duration, headers map[string]string) *Client {
	return &Client{
		inner:   &http.Client{Timeout: timeout},
		baseURL: baseURL,
		headers: headers,
	}
}

type APIResponse[T any] struct {
	Success bool `json:"success"`
	Data    T    `json:"data"`
}

func Get[T any](ctx context.Context, c *Client, path string, extraHeaders map[string]string) (T, error) {
	var zero T

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return zero, fmt.Errorf("build request: %w", err)
	}

	for k, v := range c.headers {
		req.Header.Set(k, v)
	}
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}

	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt*100) * time.Millisecond)
		}

		resp, err := c.inner.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("server error %d: %s", resp.StatusCode, string(body))
			continue
		}

		if resp.StatusCode >= 400 {
			return zero, fmt.Errorf("client error %d: %s", resp.StatusCode, string(body))
		}

		var apiResp APIResponse[T]
		if err := json.Unmarshal(body, &apiResp); err != nil {
			return zero, fmt.Errorf("decode response: %w", err)
		}
		return apiResp.Data, nil
	}

	return zero, fmt.Errorf("all retries failed: %w", lastErr)
}
