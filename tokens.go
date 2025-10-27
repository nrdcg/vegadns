package vegadns2client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Token struct to hold token information.
type Token struct {
	Token     string `json:"access_token"`
	TokenType string `json:"token_type"`
	ExpiresIn int    `json:"expires_in"`
	ExpiresAt time.Time
}

func (t Token) valid() error {
	if time.Now().UTC().After(t.ExpiresAt) {
		return errors.New("token expired")
	}

	return nil
}

func (t Token) formatBearer() string {
	return "Bearer " + t.Token
}

func (c *Client) getBearer() (string, error) {
	if c.token.valid() != nil {
		err := c.getAuthToken()
		if err != nil {
			return "", err
		}
	}

	return c.token.formatBearer(), nil
}

func (c *Client) getAuthToken() error {
	tokenEndpoint := c.getURL("token")

	v := url.Values{}
	v.Set("grant_type", "client_credentials")

	req, err := http.NewRequest(http.MethodPost, tokenEndpoint, strings.NewReader(v.Encode()))
	if err != nil {
		return fmt.Errorf("get auth token: %w", err)
	}

	req.SetBasicAuth(c.APIKey, c.APISecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	issueTime := time.Now().UTC()

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("get auth token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("get auth token: response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("get auth token: bad answer from VegaDNS (code: %d, message: %s)", resp.StatusCode, string(body))
	}

	if err := json.Unmarshal(body, &c.token); err != nil {
		return fmt.Errorf("get auth token: unmarshalling body: %w", err)
	}

	if c.token.TokenType != "bearer" {
		return fmt.Errorf("get auth token: don't support anything except bearer tokens (token type: %s)", c.token.TokenType)
	}

	c.token.ExpiresAt = issueTime.Add(time.Duration(c.token.ExpiresIn) * time.Second)

	return nil
}
