package vegadns2client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Option func(*Client) error

func WithBasicAuth(user, pass string) Option {
	return func(c *Client) error {
		c.user = user
		c.pass = pass

		return nil
	}
}

func WithOAuth(key, secret string) Option {
	return func(c *Client) error {
		c.apiKey = key
		c.apiSecret = secret

		return nil
	}
}

func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) error {
		if client == nil {
			c.client = client
		}

		return nil
	}
}

type Client struct {
	// Basic Auth
	user string
	pass string

	// OAuth
	apiKey    string
	apiSecret string

	client  *http.Client
	baseURL *url.URL
	token   Token
}

// NewClient create a new [Client].
func NewClient(baseURL string, opts ...Option) (*Client, error) {
	bu, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parsing base URL: %w", err)
	}

	c := &Client{
		client:  &http.Client{Timeout: 15 * time.Second},
		baseURL: bu.JoinPath("1.0"),
	}

	for _, opt := range opts {
		err := opt(c)
		if err != nil {
			return nil, err
		}
	}

	return c, nil
}

func (c *Client) newRequest(ctx context.Context, method string, endpoint *url.URL, params url.Values) (*http.Request, error) {
	var (
		err error
		req *http.Request
	)

	if method == http.MethodGet || method == http.MethodDelete {
		endpoint.RawQuery = params.Encode()

		req, err = http.NewRequestWithContext(ctx, method, endpoint.String(), nil)
	} else {
		req, err = http.NewRequestWithContext(ctx, method, endpoint.String(), strings.NewReader(params.Encode()))
	}

	if err != nil {
		return nil, fmt.Errorf("preparing request: %w", err)
	}

	if c.user != "" && c.pass != "" {
		// Basic Auth
		req.SetBasicAuth(c.user, c.pass)
	} else if c.apiKey != "" && c.apiSecret != "" {
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

	return req, nil
}
