package appsscript

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	// updateStatusTimeout is the per-attempt timeout for UpdateStatus.
	// Google Apps Script executes doPost() synchronously before redirecting,
	// so we need a generous budget for slow cold starts.
	updateStatusTimeout = 90 * time.Second

	// updateStatusMaxRetries is the number of times UpdateStatus will be retried
	// on transient failures before giving up.
	updateStatusMaxRetries = 3

	// updateStatusRetryDelay is the base delay between retries (doubles each attempt).
	updateStatusRetryDelay = 30 * time.Second

	// maxRedirectHops is the maximum number of redirects to follow.
	maxRedirectHops = 10
)

// Client is the HTTP client for Google Apps Script Web App
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new Apps Script client
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		// No custom CheckRedirect: Go's default behaviour (POST → GET on 302)
		// is exactly what Google Apps Script expects.
		//
		// GAS flow:
		//   1. Client POSTs to script.google.com/macros/s/{id}/exec
		//   2. GAS executes doPost() and returns 302 → script.googleusercontent.com
		//   3. Client GETs the googleusercontent.com URL to retrieve the response
		//
		// The timeout must be generous (≥90s) because step 2 can be slow on cold starts.
		httpClient: &http.Client{
			Timeout: 90 * time.Second,
		},
	}
}

// FetchTransactions fetches transactions from Google Sheet
func (c *Client) FetchTransactions(ctx context.Context, skip, limit int) (*SheetTransactionResponse, error) {
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	q := u.Query()
	q.Set("type", "transaction")
	if skip > 0 {
		q.Set("skip", strconv.Itoa(skip))
	}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var result SheetTransactionResponse
	if err := json.Unmarshal(body, &result); err != nil {
		if len(body) > 500 {
			return nil, fmt.Errorf("decode response: %w (body: %s...)", err, string(body[:500]))
		}
		return nil, fmt.Errorf("decode response: %w (body: %s)", err, string(body))
	}

	return &result, nil
}

// FetchTransfers fetches transfers from Google Sheet
func (c *Client) FetchTransfers(ctx context.Context, skip, limit int) (*SheetTransferResponse, error) {
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	q := u.Query()
	q.Set("type", "transfer")
	if skip > 0 {
		q.Set("skip", strconv.Itoa(skip))
	}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var result SheetTransferResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

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
