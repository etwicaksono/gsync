 package financeapi
 
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
 
 	var categories []Category
 	if err := json.Unmarshal(body, &categories); err != nil {
 		return nil, fmt.Errorf("decode categories: %w", err)
 	}
 
 	// Build name-to-ID map
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
