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
 
 // Client is the HTTP client for Google Apps Script Web App
 type Client struct {
 	baseURL    string
 	httpClient *http.Client
 }
 
 // NewClient creates a new Apps Script client
 func NewClient(baseURL string) *Client {
 	return &Client{
 		baseURL: baseURL,
 		httpClient: &http.Client{
 			Timeout: 30 * time.Second,
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
 		return nil, fmt.Errorf("decode response: %w", err)
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
 
 // UpdateStatus updates the status of a record in Google Sheet
 func (c *Client) UpdateStatus(ctx context.Context, req UpdateStatusRequest) (*UpdateStatusResponse, error) {
 	payload, err := json.Marshal(req)
 	if err != nil {
 		return nil, fmt.Errorf("encode request: %w", err)
 	}
 
 	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(payload))
 	if err != nil {
 		return nil, fmt.Errorf("create request: %w", err)
 	}
 	httpReq.Header.Set("Content-Type", "application/json")
 
 	resp, err := c.httpClient.Do(httpReq)
 	if err != nil {
 		return nil, fmt.Errorf("execute request: %w", err)
 	}
 	defer resp.Body.Close()
 
 	body, err := io.ReadAll(resp.Body)
 	if err != nil {
 		return nil, fmt.Errorf("read response: %w", err)
 	}
 
 	var result UpdateStatusResponse
 	if err := json.Unmarshal(body, &result); err != nil {
 		return nil, fmt.Errorf("decode response: %w", err)
 	}
 
 	return &result, nil
 }
