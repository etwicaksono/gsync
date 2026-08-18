package budgetbakers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// roundTripperFunc lets a test intercept requests without touching the
// package-level BaseURL used by doRequest.
type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func newStubClient(t *testing.T, handler func(*http.Request) (int, string)) *Client {
	t.Helper()

	c := NewClient("test-token")
	c.HTTPClient = &http.Client{
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			status, body := handler(r)
			return &http.Response{
				StatusCode: status,
				Body:       io.NopCloser(bytes.NewReader([]byte(body))),
				Header:     make(http.Header),
				Request:    r,
			}, nil
		}),
	}
	return c
}

// Mirrors the documented response shape of GET /v1/api/labels.
func TestGetLabels(t *testing.T) {
	var requested []string
	c := newStubClient(t, func(r *http.Request) (int, string) {
		requested = append(requested, r.URL.RequestURI())
		return http.StatusOK, `{"labels":[
			{"id":"lbl-1","name":"Business","color":"#123456"},
			{"id":"lbl-2","name":"Adjustment"}
		],"nextOffset":null}`
	})

	labels, err := c.GetLabels()
	require.NoError(t, err)
	require.Len(t, labels, 2)

	assert.Equal(t, "lbl-1", labels[0].ID)
	assert.Equal(t, "Business", labels[0].Name)
	// BaseURL already ends in /wallet, so that prefix is expected here.
	assert.Equal(t, []string{"/wallet/v1/api/labels?limit=200&offset=0"}, requested)
}

// A stale label must not be able to win a name match against the live one.
func TestGetLabelsSkipsArchived(t *testing.T) {
	c := newStubClient(t, func(r *http.Request) (int, string) {
		return http.StatusOK, `{"labels":[
			{"id":"old","name":"Business","archived":true},
			{"id":"new","name":"Business"}
		],"nextOffset":null}`
	})

	labels, err := c.GetLabels()
	require.NoError(t, err)
	require.Len(t, labels, 1, "archived label must be dropped")
	assert.Equal(t, "new", labels[0].ID)
}

func TestGetLabelsFollowsPagination(t *testing.T) {
	var offsets []string
	c := newStubClient(t, func(r *http.Request) (int, string) {
		offset := r.URL.Query().Get("offset")
		offsets = append(offsets, offset)

		if offset == "0" {
			return http.StatusOK, `{"labels":[{"id":"a","name":"A"}],"nextOffset":200}`
		}
		return http.StatusOK, `{"labels":[{"id":"b","name":"B"}],"nextOffset":null}`
	})

	labels, err := c.GetLabels()
	require.NoError(t, err)
	require.Len(t, labels, 2)
	assert.Equal(t, []string{"0", "200"}, offsets)
	assert.Equal(t, "b", labels[1].ID)
}

func TestGetLabelsPropagatesAPIError(t *testing.T) {
	c := newStubClient(t, func(r *http.Request) (int, string) {
		return http.StatusUnauthorized, `{"error":"invalid token"}`
	})

	_, err := c.GetLabels()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "401")
}

// Guards the wire format: labelIds must be the field name, and it must vanish
// entirely when a transaction has no labels so existing payloads are unchanged.
func TestRecordLabelIDsSerialization(t *testing.T) {
	base := Record{
		AccountID:   "acc-1",
		CategoryID:  "cat-1",
		Amount:      AmountWithCurrency{Value: -1000},
		PaymentType: "cash",
		RecordDate:  time.Date(2025, 5, 4, 10, 0, 0, 0, time.UTC),
	}

	t.Run("omitted when empty", func(t *testing.T) {
		out, err := json.Marshal(base)
		require.NoError(t, err)
		assert.NotContains(t, string(out), "labelIds")
	})

	t.Run("present when set", func(t *testing.T) {
		withLabels := base
		withLabels.LabelIDs = []string{"lbl-1", "lbl-2"}

		out, err := json.Marshal(withLabels)
		require.NoError(t, err)

		var decoded map[string]any
		require.NoError(t, json.Unmarshal(out, &decoded))
		assert.Equal(t, []any{"lbl-1", "lbl-2"}, decoded["labelIds"])
	})
}

// Keeps the stub honest: the paths under test are the real ones.
func TestLabelsEndpointPathIsValid(t *testing.T) {
	u, err := url.Parse(BaseURL + fmt.Sprintf("/v1/api/labels?limit=%d&offset=%s", 200, strconv.Itoa(0)))
	require.NoError(t, err)
	assert.Equal(t, "/wallet/v1/api/labels", u.Path)
}
