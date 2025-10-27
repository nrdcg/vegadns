package vegadns2client

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	client    http.Client
	baseurl   string
	version   string
	User      string
	Pass      string
	APIKey    string
	APISecret string
	token     Token
}

// NewClient create a new [Client].
func NewClient(url string) Client {
	return Client{
		client:  http.Client{Timeout: 15 * time.Second},
		baseurl: url,
		version: "1.0",
		token:   Token{},
	}
}

// Send a request to VegaDNS.
func (c *Client) Send(method, endpoint string, params map[string]string) (*http.Response, error) {
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
		req, err = http.NewRequest(method, vegaURL, nil)
	} else {
		req, err = http.NewRequest(method, vegaURL, strings.NewReader(p.Encode()))
	}

	if err != nil {
		return nil, fmt.Errorf("preparing request: %w", err)
	}

	if c.User != "" && c.Pass != "" {
		// Basic Auth
		req.SetBasicAuth(c.User, c.Pass)
	} else if c.APIKey != "" && c.APISecret != "" {
		// OAuth
		c.getAuthToken()

		err = c.token.valid()
		if err != nil {
			return nil, err
		}

		req.Header.Set("Authorization", c.getBearer())
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return c.client.Do(req)
}

func (c *Client) getURL(endpoint string) string {
	return fmt.Sprintf("%s/%s/%s", c.baseurl, c.version, endpoint)
}

func (c *Client) stillAuthorized() error {
	return c.token.valid()
}
