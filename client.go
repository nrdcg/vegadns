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
	User      string
	Pass      string
	APIKey    string
	APISecret string

	client  *http.Client
	baseURL *url.URL
	token   Token
}

// NewClient create a new [Client].
func NewClient(baseURL string) (*Client, error) {
	bu, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parsing base URL: %w", err)
	}

	return &Client{
		client:  &http.Client{Timeout: 15 * time.Second},
		baseURL: bu.JoinPath("1.0"),
	}, nil
}

// Send a request to VegaDNS.
func (c *Client) Send(ctx context.Context, method, urn string, params map[string]string) (*http.Response, error) {
	endpoint := c.baseURL.JoinPath(urn)

	data := url.Values{}
	for k, v := range params {
		data.Set(k, v)
	}

	var (
		err error
		req *http.Request
	)

	if method == http.MethodGet || method == http.MethodDelete {
		endpoint.RawQuery = data.Encode()

		req, err = http.NewRequestWithContext(ctx, method, endpoint.String(), nil)
	} else {
		req, err = http.NewRequestWithContext(ctx, method, endpoint.String(), strings.NewReader(data.Encode()))
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
