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

// UpdateStatus updates the status of a record in Google Sheet.
// It uses a generous per-attempt timeout and retries on transient failures
// to handle Google Apps Script cold starts and occasional network hiccups.
func (c *Client) UpdateStatus(ctx context.Context, req UpdateStatusRequest) (*UpdateStatusResponse, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("encode request: %w", err)
	}

	var lastErr error
	for attempt := 1; attempt <= updateStatusMaxRetries; attempt++ {
		if attempt > 1 {
			delay := updateStatusRetryDelay * time.Duration(1<<(attempt-2)) // 30s, 60s
			fmt.Printf("    [retry %d/%d] waiting %s before retrying UpdateStatus...\n", attempt, updateStatusMaxRetries, delay)
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return nil, fmt.Errorf("context cancelled during retry wait: %w", ctx.Err())
			}
		}

		// Each attempt gets its own independent timeout so a slow Apps Script
		// cold-start doesn't consume the parent context deadline.
		attemptCtx, cancel := context.WithTimeout(ctx, updateStatusTimeout)
		result, err := c.doPost(attemptCtx, payload)
		cancel()
		if err != nil {
			lastErr = fmt.Errorf("attempt %d: %w", attempt, err)
			continue
		}

		return result, nil
	}

	return nil, fmt.Errorf("UpdateStatus failed after %d attempts: %w", updateStatusMaxRetries, lastErr)
}

// doPost POSTs the payload to the Apps Script exec URL and reads the response.
//
// Google Apps Script redirect flow (why no custom CheckRedirect is needed):
//   1. Client POSTs to script.google.com → GAS runs doPost(), returns 302
//   2. Go's default redirect policy follows as GET to script.googleusercontent.com
//   3. That URL serves the cached JSON response from doPost()
//
// Sending another POST to step 2's URL causes a 405 because googleusercontent.com
// only serves the response over GET. The default behaviour is therefore correct.
func (c *Client) doPost(ctx context.Context, payload []byte) (*UpdateStatusResponse, error) {
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

	var result UpdateStatusResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decode response (status %d, body: %.200s): %w", resp.StatusCode, string(body), err)
	}

	return &result, nil
}
