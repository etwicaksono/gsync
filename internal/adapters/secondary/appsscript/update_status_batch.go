package appsscript

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// UpdateStatusBatch updates the status of multiple records in Google Sheet.
func (c *Client) UpdateStatusBatch(ctx context.Context, requests []UpdateStatusRequest) (*UpdateStatusBatchResponse, error) {
	if len(requests) == 0 {
		return &UpdateStatusBatchResponse{Success: true, Message: "No requests to process"}, nil
	}

	payload, err := json.Marshal(requests)
	if err != nil {
		return nil, fmt.Errorf("encode request: %w", err)
	}

	var lastErr error
	for attempt := 1; attempt <= updateStatusMaxRetries; attempt++ {
		if attempt > 1 {
			delay := updateStatusRetryDelay * time.Duration(1<<(attempt-2))
			fmt.Printf("    [retry %d/%d] waiting %s before retrying UpdateStatusBatch...\n", attempt, updateStatusMaxRetries, delay)
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return nil, fmt.Errorf("context cancelled during retry wait: %w", ctx.Err())
			}
		}

		attemptCtx, cancel := context.WithTimeout(ctx, updateStatusTimeout)
		result, err := c.doPostBatch(attemptCtx, payload)
		cancel()
		if err != nil {
			lastErr = fmt.Errorf("attempt %d: %w", attempt, err)
			continue
		}

		return result, nil
	}

	return nil, fmt.Errorf("UpdateStatusBatch failed after %d attempts: %w", updateStatusMaxRetries, lastErr)
}

// doPostBatch POSTs the payload and parses the batch response.
func (c *Client) doPostBatch(ctx context.Context, payload []byte) (*UpdateStatusBatchResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}

	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var result UpdateStatusBatchResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}
