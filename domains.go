package vegadns2client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// Domain struct containing a domain object.
type Domain struct {
	Status   string `json:"status"`
	Domain   string `json:"domain"`
	DomainID int    `json:"domain_id"`
	OwnerID  int    `json:"owner_id"`
}

// DomainResponse api response of a domain list.
type DomainResponse struct {
	Status  string   `json:"status"`
	Total   int      `json:"total_domains"`
	Domains []Domain `json:"domains"`
}

// GetDomainID returns the id for a domain.
func (c *Client) GetDomainID(ctx context.Context, domain string) (int, error) {
	params := make(url.Values)
	params.Set("search", domain)

	req, err := c.newRequest(ctx, http.MethodGet, c.baseURL.JoinPath("domains"), params)
	if err != nil {
		return -1, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return -1, fmt.Errorf("get domain ID: %w", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return -1, fmt.Errorf("get domain ID: response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return -1, fmt.Errorf("get domain ID: bad answer from VegaDNS (code: %d, message: %s)", resp.StatusCode, string(body))
	}

	answer := DomainResponse{}
	if err := json.Unmarshal(body, &answer); err != nil {
		return -1, fmt.Errorf("get domain ID: unmarshalling body: %w", err)
	}

	for _, d := range answer.Domains {
		if d.Domain == domain {
			return d.DomainID, nil
		}
	}

	return -1, fmt.Errorf("get domain ID: domain %s not found", domain)
}

// GetAuthZone retrieves the closest match to a given domain.
// Example: Given an argument "a.b.c.d.e", and a VegaDNS hosted domain of "c.d.e",
// GetClosestMatchingDomain will return "c.d.e".
func (c *Client) GetAuthZone(ctx context.Context, fqdn string) (string, int, error) {
	fqdn = strings.TrimSuffix(fqdn, ".")

	numComponents := len(strings.Split(fqdn, "."))
	for i := 1; i < numComponents; i++ {
		tmpHostname := strings.SplitN(fqdn, ".", i)[i-1]
		log.Printf("tmpHostname for i = %d: %s\n", i, tmpHostname)

		if domainID, err := c.GetDomainID(ctx, tmpHostname); err == nil {
			log.Printf("Found zone: %s\n\tShortened to %s\n", tmpHostname, strings.TrimSuffix(tmpHostname, "."))

			return strings.TrimSuffix(tmpHostname, "."), domainID, nil
		}
	}

	return "", -1, fmt.Errorf("unable to find auth zone for fqdn %s", fqdn)
}
