package budgetmateapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// FetchCategories fetches all categories from Finance Manager
 func (c *Client) FetchCategories(ctx context.Context) ([]Category, error) {
 	body, status, err := c.doRequest(ctx, http.MethodGet, "/api/v1/categories", nil)
 	if err != nil {
 		return nil, err
 	}
 
 	if status != http.StatusOK {
 		return nil, fmt.Errorf("fetch categories failed with status %d: %s", status, string(body))
 	}
 
	// Try wrapped response first
	var resp CategoriesResponse
	if err := json.Unmarshal(body, &resp); err == nil && resp.Success {
		for _, cat := range resp.Data {
			c.categoryMap[strings.ToLower(cat.Name)] = cat.ID
		}
		return resp.Data, nil
 	}
 
	// Fallback to direct array
	var categories []Category
	if err := json.Unmarshal(body, &categories); err != nil {
		return nil, fmt.Errorf("decode categories: %w (body: %s)", err, string(body))
	}
 	for _, cat := range categories {
 		c.categoryMap[strings.ToLower(cat.Name)] = cat.ID
 	}
 	return categories, nil
 }
 
 // GetCategoryID returns the category ID for a given name (case-insensitive)
 func (c *Client) GetCategoryID(name string) (string, bool) {
 	id, ok := c.categoryMap[strings.ToLower(name)]
 	return id, ok
 }

// GetCategoryNames returns all available category names
func (c *Client) GetCategoryNames() []string {
	names := make([]string, 0, len(c.categoryMap))
	for name := range c.categoryMap {
		names = append(names, name)
	}
	return names
}
