 package financeapi
 
 import (
 	"bytes"
 	"context"
 	"encoding/json"
 	"fmt"
 	"io"
 	"net/http"
 	"net/http/cookiejar"
 	"time"
 )
 
 // Client is the HTTP client for Finance Manager API
 type Client struct {
 	baseURL      string
 	email        string
 	password     string
 	httpClient   *http.Client
 	authToken    string
 	accountMap   map[string]string // name -> id
 	categoryMap  map[string]string // name -> id
	accountMapping  map[string]string // sheet name -> finance name
	categoryMapping map[string]string // sheet name -> finance name
 }
 
 // NewClient creates a new Finance Manager API client
 func NewClient(baseURL, email, password string) (*Client, error) {
 	jar, err := cookiejar.New(nil)
 	if err != nil {
 		return nil, fmt.Errorf("create cookie jar: %w", err)
 	}
 
 	return &Client{
 		baseURL:  baseURL,
 		email:    email,
 		password: password,
 		httpClient: &http.Client{
 			Timeout: 30 * time.Second,
 			Jar:     jar,
 		},
		accountMap:      make(map[string]string),
		categoryMap:     make(map[string]string),
		accountMapping:  make(map[string]string),
		categoryMapping: make(map[string]string),
 	}, nil
 }
 
 // Login authenticates with the Finance Manager API
 func (c *Client) Login(ctx context.Context) error {
 	payload := LoginRequest{
		EmailOrUsername: c.email,
		Password:        c.password,
 	}
 
 	body, err := json.Marshal(payload)
 	if err != nil {
 		return fmt.Errorf("encode login request: %w", err)
 	}
 
 	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/auth/login", bytes.NewReader(body))
 	if err != nil {
 		return fmt.Errorf("create login request: %w", err)
 	}
 	req.Header.Set("Content-Type", "application/json")
 
 	resp, err := c.httpClient.Do(req)
 	if err != nil {
 		return fmt.Errorf("execute login request: %w", err)
 	}
 	defer resp.Body.Close()
 
 	respBody, err := io.ReadAll(resp.Body)
 	if err != nil {
 		return fmt.Errorf("read login response: %w", err)
 	}
 
 	if resp.StatusCode != http.StatusOK {
 		return fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(respBody))
 	}
 
 	var loginResp LoginResponse
 	if err := json.Unmarshal(respBody, &loginResp); err != nil {
 		return fmt.Errorf("decode login response: %w", err)
 	}
 
	// Try to get token from response body
	if loginResp.Token != "" {
		c.authToken = loginResp.Token
	}

	// Also check for token in response data
	var dataResp struct {
		Success bool `json:"success"`
		Data    struct {
			Token       string `json:"token"`
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &dataResp); err == nil {
		if dataResp.Data.Token != "" {
			c.authToken = dataResp.Data.Token
		}
		if dataResp.Data.AccessToken != "" {
			c.authToken = dataResp.Data.AccessToken
		}
 	}
 
 	return nil
 }
 
 // doRequest performs an authenticated HTTP request
 func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) ([]byte, int, error) {
 	var reqBody io.Reader
 	if body != nil {
 		jsonBody, err := json.Marshal(body)
 		if err != nil {
 			return nil, 0, fmt.Errorf("encode request body: %w", err)
 		}
 		reqBody = bytes.NewReader(jsonBody)
 	}
 
 	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
 	if err != nil {
 		return nil, 0, fmt.Errorf("create request: %w", err)
 	}
 
 	req.Header.Set("Content-Type", "application/json")
 	if c.authToken != "" {
 		req.Header.Set("Authorization", "Bearer "+c.authToken)
 	}
 
 	resp, err := c.httpClient.Do(req)
 	if err != nil {
 		return nil, 0, fmt.Errorf("execute request: %w", err)
 	}
 	defer resp.Body.Close()
 
 	respBody, err := io.ReadAll(resp.Body)
 	if err != nil {
 		return nil, resp.StatusCode, fmt.Errorf("read response: %w", err)
 	}
 
 	return respBody, resp.StatusCode, nil
 }
