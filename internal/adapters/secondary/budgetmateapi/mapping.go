package budgetmateapi

import "strings"

// DefaultAccountMapping maps Google Sheet account names to Finance Manager names
func (c *Client) SetAccountMapping(mapping map[string]string) {
	c.accountMapping = make(map[string]string)
	for k, v := range mapping {
		c.accountMapping[k] = v
	}
}

// SetCategoryMapping sets the category mapping
func (c *Client) SetCategoryMapping(mapping map[string]string) {
	c.categoryMapping = make(map[string]string)
	for k, v := range mapping {
		c.categoryMapping[k] = v
	}
}

// SetLabelMapping sets the label mapping
func (c *Client) SetLabelMapping(mapping map[string]string) {
	c.labelMapping = make(map[string]string)
	for k, v := range mapping {
		c.labelMapping[k] = v
	}
}

// GetAccountID returns the account ID for a given name (with mapping support)
func (c *Client) GetAccountIDWithMapping(name string) (string, bool) {
	lowerName := strings.ToLower(name)

	// Check if there's a mapping
	if mappedName, ok := c.accountMapping[lowerName]; ok {
		lowerName = mappedName
	}

	id, ok := c.accountMap[lowerName]
	return id, ok
}

// GetCategoryIDWithMapping returns the category ID for a given name (with mapping support)
func (c *Client) GetCategoryIDWithMapping(name string) (string, bool) {
	lowerName := strings.ToLower(name)

	// Check if there's a mapping
	if mappedName, ok := c.categoryMapping[lowerName]; ok {
		lowerName = mappedName
	}

	id, ok := c.categoryMap[lowerName]
	return id, ok
}
