package budgetbakers

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// CreateRecords creates a batch of records
func (c *Client) CreateRecords(records []Record) (*BatchRecordResponse, error) {
	respBytes, err := c.doRequest(http.MethodPost, "/v1/api/records", records)
	if err != nil {
		// Even if HTTP status is 207, our generic doRequest might treat 207 as OK (< 400),
		// but let's check what BudgetBakers returns. 207 Multi-Status is < 400, so it won't error there.
		return nil, err
	}

	var result BatchRecordResponse
	if err := json.Unmarshal(respBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to decode batch response: %w", err)
	}
	return &result, nil
}
