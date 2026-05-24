package coreclient

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/quickbite/analytics-service/pkg/httpclient"
)

type Client struct {
	http *httpclient.Client
}

func New(baseURL, apiKey string) *Client {
	return &Client{
		http: httpclient.New(baseURL, 5*time.Second, map[string]string{
			"x-api-key": apiKey,
		}),
	}
}

func (c *Client) GetRolePermissions(ctx context.Context, roleName string) ([]string, error) {
	path := fmt.Sprintf("/api/internal/rbac/permissions?role=%s", url.QueryEscape(roleName))

	type permEntry struct {
		Permission string `json:"permission"`
	}
	type permResponse struct {
		Permissions []permEntry `json:"permissions"`
	}

	data, err := httpclient.Get[permResponse](ctx, c.http, path, nil)
	if err != nil {
		return nil, fmt.Errorf("get role permissions: %w", err)
	}

	perms := make([]string, len(data.Permissions))
	for i, p := range data.Permissions {
		perms[i] = p.Permission
	}
	return perms, nil
}
