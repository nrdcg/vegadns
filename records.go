package vegadns2client

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Record struct representing a Record object.
type Record struct {
	Name       string `json:"name"`
	Value      string `json:"value"`
	RecordType string `json:"record_type"`
	TTL        int    `json:"ttl"`
	RecordID   int    `json:"record_id"`
	LocationID string `json:"location_id"`
	DomainID   int    `json:"domain_id"`
}

// RecordsResponse api response list of records.
type RecordsResponse struct {
	Status  string   `json:"status"`
	Total   int      `json:"total_records"`
	Domain  Domain   `json:"domain"`
	Records []Record `json:"records"`
}

// GetRecordID helper to get the id of a record.
func (c *Client) GetRecordID(ctx context.Context, domainID int, record, recordType string) (int, error) {
	params := make(url.Values)
	params.Set("domain_id", strconv.Itoa(domainID))

	req, err := c.newRequest(ctx, http.MethodGet, c.baseURL.JoinPath("records"), params)
	if err != nil {
		return -1, err
	}

	answer := RecordsResponse{}

	err = c.do(req, http.StatusOK, &answer)
	if err != nil {
		return -1, err
	}

	for _, r := range answer.Records {
		if r.Name == record && r.RecordType == recordType {
			return r.RecordID, nil
		}
	}

	return -1, errors.New("record not found")
}

// CreateTXT creates a TXT record.
func (c *Client) CreateTXT(ctx context.Context, domainID int, fqdn, value string, ttl int) error {
	params := make(url.Values)
	params.Set("record_type", "TXT")
	params.Set("ttl", strconv.Itoa(ttl))
	params.Set("domain_id", strconv.Itoa(domainID))
	params.Set("name", strings.TrimSuffix(fqdn, "."))
	params.Set("value", value)

	req, err := c.newRequest(ctx, http.MethodPost, c.baseURL.JoinPath("records"), params)
	if err != nil {
		return err
	}

	return c.do(req, http.StatusCreated, nil)
}

// DeleteRecord deletes a record by id.
func (c *Client) DeleteRecord(ctx context.Context, recordID int) error {
	req, err := c.newRequest(ctx, http.MethodDelete, c.baseURL.JoinPath("records", strconv.Itoa(recordID)), nil)
	if err != nil {
		return err
	}

	return c.do(req, http.StatusOK, nil)
}
