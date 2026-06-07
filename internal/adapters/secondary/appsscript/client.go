package appsscript

import (
	"net/http"
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
