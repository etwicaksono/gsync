package budgetbakers

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// GetCategories fetches all categories
func (c *Client) GetCategories() ([]Category, error) {
	var allCategories []Category
	offset := 0
	limit := 200

	for {
		path := fmt.Sprintf("/v1/api/categories?limit=%d&offset=%d", limit, offset)
		respBytes, err := c.doRequest(http.MethodGet, path, nil)
		if err != nil {
			return nil, err
		}

		var response struct {
			Categories []Category `json:"categories"`
			NextOffset *int       `json:"nextOffset"`
		}
		if err := json.Unmarshal(respBytes, &response); err != nil {
			return nil, fmt.Errorf("failed to decode categories: %w", err)
		}

		allCategories = append(allCategories, response.Categories...)

		if response.NextOffset == nil {
			break
		}
		offset = *response.NextOffset
	}

	return allCategories, nil
}
