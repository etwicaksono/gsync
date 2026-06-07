package budgetbakers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

const BaseURL = "https://rest.budgetbakers.com/wallet"

// Client is a REST API client for BudgetBakers Wallet
type Client struct {
	Token              string
	HTTPClient         *http.Client
	RateLimitLimit     int
	RateLimitRemaining int
}

// NewClient creates a new BudgetBakers API client
func NewClient(token string) *Client {
	return &Client{
		Token: token,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) doRequest(method, path string, body interface{}) ([]byte, error) {
	var bodyBytes []byte
	var err error
	if body != nil {
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		var bodyReader io.Reader
		if body != nil {
			bodyReader = bytes.NewReader(bodyBytes) // Re-create reader for each attempt
		}

		req, err := http.NewRequest(method, BaseURL+path, bodyReader)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+c.Token)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("request failed: %w", err)
		}
		
		// Manually decrement in case the header is missing on some endpoints
		if c.RateLimitRemaining > 0 {
			c.RateLimitRemaining--
		}

		// Check rate limits from header
		limit := resp.Header.Get("X-RateLimit-Limit")
		if limit != "" {
			if l, err := strconv.Atoi(limit); err == nil {
				c.RateLimitLimit = l
			}
		}
		remaining := resp.Header.Get("X-RateLimit-Remaining")
		if remaining != "" {
			if r, err := strconv.Atoi(remaining); err == nil {
				c.RateLimitRemaining = r
				if r < 10 && attempt == 1 {
					fmt.Printf("⚠️  Warning: API rate limit is getting low (Remaining: %d)\n", r)
				}
			}
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			if attempt < maxRetries {
				fmt.Printf("⏳ Rate limit exceeded (429). Retrying in 15 seconds (Attempt %d/%d)...\n", attempt, maxRetries)
				time.Sleep(15 * time.Second)
				continue
			}
			return respBody, fmt.Errorf("API rate limit exceeded. Please wait a while before running again")
		}

		if resp.StatusCode >= 400 {
			return respBody, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
		}

		return respBody, nil
	}
	
	return nil, fmt.Errorf("max retries exceeded")
}
