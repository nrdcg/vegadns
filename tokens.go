package vegadns2client

import (
	"encoding/json"
	"errors"
	"io"
	"log"
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

func (c *Client) getBearer() string {
	if c.token.valid() != nil {
		c.getAuthToken()
	}

	return c.token.formatBearer()
}

func (c *Client) getAuthToken() {
	tokenEndpoint := c.getURL("token")

	v := url.Values{}
	v.Set("grant_type", "client_credentials")

	req, err := http.NewRequest("POST", tokenEndpoint, strings.NewReader(v.Encode()))
	if err != nil {
		log.Fatalf("get AuthToken: %s", err)
	}

	req.SetBasicAuth(c.APIKey, c.APISecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	issueTime := time.Now().UTC()

	resp, err := c.client.Do(req)
	if err != nil {
		log.Fatalf("Error sending POST to getAuthToken: %s", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response from POST to getAuthToken: %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Got bad answer from VegaDNS on getAuthToken. Code: %d. Message: %s", resp.StatusCode, string(body))
	}

	if err := json.Unmarshal(body, &c.token); err != nil {
		log.Fatalf("Error unmarshalling body of POST to getAuthToken: %s", err)
	}

	if c.token.TokenType != "bearer" {
		log.Fatal("We don't support anything except bearer tokens")
	}

	c.token.ExpiresAt = issueTime.Add(time.Duration(c.token.ExpiresIn) * time.Second)
}
