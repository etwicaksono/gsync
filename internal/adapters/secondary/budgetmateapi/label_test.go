package budgetmateapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	c, err := NewClient(srv.URL, "user", "pass")
	require.NoError(t, err)
	return c
}

// Mirrors the documented response shape of GET /api/v1/labels.
func TestFetchLabelsWrappedResponse(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/labels", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":[
			{"id":"clq789","name":"Urgent","color":"#ff0000"},
			{"id":"clq790","name":"Business"}
		]}`))
	})

	labels, err := c.FetchLabels(context.Background())
	require.NoError(t, err)
	require.Len(t, labels, 2)

	// The sheet stores labels uppercased, so lookups must ignore case.
	id, ok := c.GetLabelID("BUSINESS")
	assert.True(t, ok, "uppercase sheet value must match title-cased API name")
	assert.Equal(t, "clq790", id)

	id, ok = c.GetLabelID("  urgent  ")
	assert.True(t, ok, "surrounding whitespace must not break lookup")
	assert.Equal(t, "clq789", id)

	_, ok = c.GetLabelID("NOPE")
	assert.False(t, ok)
}

// Same fallback the categories endpoint already tolerates.
func TestFetchLabelsBareArrayResponse(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[{"id":"a1","name":"Business"}]`))
	})

	labels, err := c.FetchLabels(context.Background())
	require.NoError(t, err)
	require.Len(t, labels, 1)

	id, ok := c.GetLabelID("business")
	assert.True(t, ok)
	assert.Equal(t, "a1", id)
}

func TestFetchLabelsPropagatesHTTPError(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
	})

	_, err := c.FetchLabels(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "401")
}

func TestResolveLabelIDs(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"success":true,"data":[
			{"id":"id-business","name":"Business"},
			{"id":"id-adjustment","name":"Adjustment"}
		]}`))
	})
	_, err := c.FetchLabels(context.Background())
	require.NoError(t, err)

	t.Run("resolves every name", func(t *testing.T) {
		ids, missing := c.ResolveLabelIDs([]string{"BUSINESS", "ADJUSTMENT"})
		assert.Empty(t, missing)
		assert.Equal(t, []string{"id-business", "id-adjustment"}, ids)
	})

	t.Run("no labels yields nil", func(t *testing.T) {
		ids, missing := c.ResolveLabelIDs(nil)
		assert.Empty(t, missing)
		assert.Nil(t, ids, "absent labels must leave the payload field unset")
	})

	// All-or-nothing: a partially labelled transaction is harder to spot than one
	// that was never sent, so nothing is returned when a name is unknown.
	t.Run("unknown name returns no ids", func(t *testing.T) {
		ids, missing := c.ResolveLabelIDs([]string{"BUSINESS", "GHOST"})
		assert.Equal(t, []string{"GHOST"}, missing)
		assert.Nil(t, ids)
	})

	t.Run("reports every unknown name", func(t *testing.T) {
		_, missing := c.ResolveLabelIDs([]string{"GHOST", "PHANTOM"})
		assert.Equal(t, []string{"GHOST", "PHANTOM"}, missing)
	})

	t.Run("skips blank entries", func(t *testing.T) {
		ids, missing := c.ResolveLabelIDs([]string{"BUSINESS", "   "})
		assert.Empty(t, missing)
		assert.Equal(t, []string{"id-business"}, ids)
	})

	t.Run("deduplicates repeated labels", func(t *testing.T) {
		ids, missing := c.ResolveLabelIDs([]string{"BUSINESS", "business"})
		assert.Empty(t, missing)
		assert.Equal(t, []string{"id-business"}, ids)
	})
}

func TestResolveLabelIDsUsesMapping(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"success":true,"data":[{"id":"id-biz","name":"Business"}]}`))
	})
	_, err := c.FetchLabels(context.Background())
	require.NoError(t, err)

	c.SetLabelMapping(map[string]string{"bisnis": "business"})

	ids, missing := c.ResolveLabelIDs([]string{"BISNIS"})
	assert.Empty(t, missing)
	assert.Equal(t, []string{"id-biz"}, ids)
}
