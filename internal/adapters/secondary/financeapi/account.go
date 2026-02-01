 package financeapi
 
 import (
 	"context"
 	"encoding/json"
 	"fmt"
 	"net/http"
 	"strings"
 )
 
 // FetchAccounts fetches all accounts from Finance Manager
 func (c *Client) FetchAccounts(ctx context.Context) ([]Account, error) {
 	body, status, err := c.doRequest(ctx, http.MethodGet, "/api/v1/accounts", nil)
 	if err != nil {
 		return nil, err
 	}
 
 	if status != http.StatusOK {
 		return nil, fmt.Errorf("fetch accounts failed with status %d: %s", status, string(body))
 	}
 
 	var accounts []Account
 	if err := json.Unmarshal(body, &accounts); err != nil {
 		return nil, fmt.Errorf("decode accounts: %w", err)
 	}
 
 	// Build name-to-ID map
 	for _, acc := range accounts {
 		c.accountMap[strings.ToLower(acc.Name)] = acc.ID
 	}
 
 	return accounts, nil
 }
 
 // GetAccountID returns the account ID for a given name (case-insensitive)
 func (c *Client) GetAccountID(name string) (string, bool) {
 	id, ok := c.accountMap[strings.ToLower(name)]
 	return id, ok
 }
