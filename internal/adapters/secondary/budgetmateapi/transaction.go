package budgetmateapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// CreateTransaction creates a new transaction in Finance Manager
func (c *Client) CreateTransaction(ctx context.Context, req CreateTransactionRequest) (*CreateTransactionResponse, error) {
	body, status, err := c.doRequest(ctx, http.MethodPost, "/api/v1/transactions", req)
	if err != nil {
		return nil, err
	}

	var resp CreateTransactionResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if status != http.StatusCreated && status != http.StatusOK {
		if resp.Error != nil {
			return nil, fmt.Errorf("create transaction failed: %v", resp.Error)
		}
		return nil, fmt.Errorf("create transaction failed with status %d: %s", status, string(body))
	}

	if !resp.Success && resp.Error != nil {
		return nil, fmt.Errorf("create transaction failed: %v", resp.Error)
	}

	return &resp, nil
}
