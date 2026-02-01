 package financeapi
 
 import "strings"
 
 // DefaultAccountMapping maps Google Sheet account names to Finance Manager names
 var DefaultAccountMapping = map[string]string{
	// Map Google Sheet account names to Finance Manager account names (lowercase)
	"bca":          "checking account",
	"cimb syariah": "savings account",
	"jenius":       "checking account",
	"gopay":        "cash",
	"dana":         "cash",
	"cash eko":     "cash",
	"saldo pulsa":  "cash",
 }
 
 // DefaultCategoryMapping maps Google Sheet category names to Finance Manager names
 var DefaultCategoryMapping = map[string]string{
 	// Add your mappings here (lowercase keys)
 	"groceries, main meal":   "groceries",
 	"bar, cafe, snack":       "bar, cafe",
 	"internet, phone credit": "internet",
 	"drug-store, chemist":    "toiletry, health",
 	"clothes & footwear":     "clothes & shoes",
 	"movie, tv, streaming":   "tv, streaming",
 	"checks, coupons":        "others",
 	"wage, invoices":         "salary",
 }
 
 // SetAccountMapping sets the account mapping
 func (c *Client) SetAccountMapping(mapping map[string]string) {
 	c.accountMapping = make(map[string]string)
 	for k, v := range mapping {
 		c.accountMapping[strings.ToLower(k)] = strings.ToLower(v)
 	}
 }
 
 // SetCategoryMapping sets the category mapping
 func (c *Client) SetCategoryMapping(mapping map[string]string) {
 	c.categoryMapping = make(map[string]string)
 	for k, v := range mapping {
 		c.categoryMapping[strings.ToLower(k)] = strings.ToLower(v)
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
