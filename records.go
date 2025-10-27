package vegadns2client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
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
func (c *Client) GetRecordID(domainID int, record, recordType string) (int, error) {
	params := make(map[string]string)
	params["domain_id"] = strconv.Itoa(domainID)

	resp, err := c.Send(http.MethodGet, "records", params)
	if err != nil {
		return -1, fmt.Errorf("get record ID: %w", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return -1, fmt.Errorf("get record ID: response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return -1, fmt.Errorf("get record ID: bad answer from VegaDNS (code: %d, message: %s)", resp.StatusCode, string(body))
	}

	answer := RecordsResponse{}
	if err := json.Unmarshal(body, &answer); err != nil {
		return -1, fmt.Errorf("get record ID: unmarshalling body: %w", err)
	}

	for _, r := range answer.Records {
		if r.Name == record && r.RecordType == recordType {
			return r.RecordID, nil
		}
	}

	return -1, errors.New("get record ID: record not found")
}

// CreateTXT creates a TXT record.
func (c *Client) CreateTXT(domainID int, fqdn, value string, ttl int) error {
	params := make(map[string]string)

	params["record_type"] = "TXT"
	params["ttl"] = strconv.Itoa(ttl)
	params["domain_id"] = strconv.Itoa(domainID)
	params["name"] = strings.TrimSuffix(fqdn, ".")
	params["value"] = value

	resp, err := c.Send("POST", "records", params)
	if err != nil {
		return fmt.Errorf("create TXT record: %w", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("create TXT record: response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("create TXT record: bad answer from VegaDNS (code: %d, message: %s)", resp.StatusCode, string(body))
	}

	return nil
}

// DeleteRecord deletes a record by id.
func (c *Client) DeleteRecord(recordID int) error {
	resp, err := c.Send(http.MethodDelete, fmt.Sprintf("records/%d", recordID), nil)
	if err != nil {
		return fmt.Errorf("delete record: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("delete record: response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("delete record: bad answer from VegaDNS (code: %d, message: %s)", resp.StatusCode, string(body))
	}

	return nil
}
