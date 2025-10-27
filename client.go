package vegadns2client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	client    *http.Client
	baseURL   string
	version   string
	User      string
	Pass      string
	APIKey    string
	APISecret string
	token     Token
}

// NewClient create a new [Client].
func NewClient(baseURL string) *Client {
	return &Client{
		client:  &http.Client{Timeout: 15 * time.Second},
		baseURL: baseURL,
		version: "1.0",
	}
}

// Send a request to VegaDNS.
func (c *Client) Send(ctx context.Context, method, endpoint string, params map[string]string) (*http.Response, error) {
	vegaURL := c.getURL(endpoint)

	p := url.Values{}
	for k, v := range params {
		p.Set(k, v)
	}

	var (
		err error
		req *http.Request
	)

	if method == http.MethodGet || method == http.MethodDelete {
		vegaURL = fmt.Sprintf("%s?%s", vegaURL, p.Encode())
		req, err = http.NewRequestWithContext(ctx, method, vegaURL, nil)
	} else {
		req, err = http.NewRequestWithContext(ctx, method, vegaURL, strings.NewReader(p.Encode()))
	}

	if err != nil {
		return nil, fmt.Errorf("preparing request: %w", err)
	}

	if c.User != "" && c.Pass != "" {
		// Basic Auth
		req.SetBasicAuth(c.User, c.Pass)
	} else if c.APIKey != "" && c.APISecret != "" {
		// OAuth
		err := c.getAuthToken(ctx)
		if err != nil {
			return nil, err
		}

		err = c.token.valid()
		if err != nil {
			return nil, err
		}

		bearer, err := c.getBearer(ctx)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Authorization", bearer)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return c.client.Do(req)
}

func (c *Client) getURL(endpoint string) string {
	return fmt.Sprintf("%s/%s/%s", c.baseURL, c.version, endpoint)
}
