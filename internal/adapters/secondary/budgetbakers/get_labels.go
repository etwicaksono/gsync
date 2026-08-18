package budgetbakers

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// GetLabels fetches all labels, skipping archived ones
func (c *Client) GetLabels() ([]Label, error) {
	var allLabels []Label
	offset := 0
	limit := 200

	for {
		path := fmt.Sprintf("/v1/api/labels?limit=%d&offset=%d", limit, offset)
		respBytes, err := c.doRequest(http.MethodGet, path, nil)
		if err != nil {
			return nil, err
		}

		var response struct {
			Labels     []Label `json:"labels"`
			NextOffset *int    `json:"nextOffset"`
		}
		if err := json.Unmarshal(respBytes, &response); err != nil {
			return nil, fmt.Errorf("failed to decode labels: %w", err)
		}

		// Archived labels are excluded so a retired label can never win a name
		// match against the one currently in use.
		for _, label := range response.Labels {
			if label.Archived {
				continue
			}
			allLabels = append(allLabels, label)
		}

		if response.NextOffset == nil {
			break
		}
		offset = *response.NextOffset
	}

	return allLabels, nil
}
