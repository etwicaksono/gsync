 package financeapi
 
 import (
 	"context"
 	"encoding/json"
 	"fmt"
 	"net/http"
 )
 
 // CreateTransfer creates a new transfer in Finance Manager
 func (c *Client) CreateTransfer(ctx context.Context, req CreateTransferRequest) (*CreateTransferResponse, error) {
 	body, status, err := c.doRequest(ctx, http.MethodPost, "/api/v1/transfers", req)
 	if err != nil {
 		return nil, err
 	}
 
 	var resp CreateTransferResponse
 	if err := json.Unmarshal(body, &resp); err != nil {
 		return nil, fmt.Errorf("decode response: %w", err)
 	}
 
 	if status != http.StatusCreated && status != http.StatusOK {
 		if resp.Error != "" {
 			return nil, fmt.Errorf("create transfer failed: %s", resp.Error)
 		}
 		return nil, fmt.Errorf("create transfer failed with status %d: %s", status, string(body))
 	}
 
 	return &resp, nil
 }
