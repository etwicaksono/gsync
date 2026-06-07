package budgetbakers

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// GetAccounts fetches all accounts
func (c *Client) GetAccounts() ([]Account, error) {
	var allAccounts []Account
	offset := 0
	limit := 200

	for {
		path := fmt.Sprintf("/v1/api/accounts?limit=%d&offset=%d", limit, offset)
		respBytes, err := c.doRequest(http.MethodGet, path, nil)
		if err != nil {
			return nil, err
		}

		var response struct {
			Accounts   []Account `json:"accounts"`
			NextOffset *int      `json:"nextOffset"`
		}
		if err := json.Unmarshal(respBytes, &response); err != nil {
			return nil, fmt.Errorf("failed to decode accounts: %w", err)
		}
		
		allAccounts = append(allAccounts, response.Accounts...)

		if response.NextOffset == nil {
			break
		}
		offset = *response.NextOffset
	}

	return allAccounts, nil
}
