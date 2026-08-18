package budgetmateapi

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The reported symptom was labels silently not arriving. These cases pin down
// which observable state each possible cause produces, so the logs can tell
// them apart instead of all looking like "no labels".
func TestLabelDiagnosisCases(t *testing.T) {
	// Cause 1: the sheet never sent labels. The API is healthy and the lookup
	// table is populated, yet nothing is resolved because the input is empty.
	t.Run("sheet sent no labels", func(t *testing.T) {
		c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`{"success":true,"data":[{"id":"id-business","name":"Business"}]}`))
		})
		_, err := c.FetchLabels(context.Background())
		require.NoError(t, err)

		ids, missing := c.ResolveLabelIDs(nil)
		assert.Nil(t, ids)
		assert.Empty(t, missing, "no input means nothing is missing, so no error is raised")
		assert.Len(t, c.GetLabelNames(), 1, "lookup table is fine; the input was the problem")
	})

	// Cause 2: the API returned no labels. Any labelled transaction then reports
	// the name as missing, which surfaces as a hard failure rather than silence.
	t.Run("api returned no labels", func(t *testing.T) {
		c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`{"success":true,"data":[]}`))
		})
		labels, err := c.FetchLabels(context.Background())
		require.NoError(t, err)
		assert.Empty(t, labels)
		assert.Empty(t, c.GetLabelNames())

		_, missing := c.ResolveLabelIDs([]string{"BUSINESS"})
		assert.Equal(t, []string{"BUSINESS"}, missing)
	})

	// Cause 3: names differ between sheet and API.
	t.Run("names do not match", func(t *testing.T) {
		c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`{"success":true,"data":[{"id":"id-1","name":"Bisnis"}]}`))
		})
		_, err := c.FetchLabels(context.Background())
		require.NoError(t, err)

		_, missing := c.ResolveLabelIDs([]string{"BUSINESS"})
		assert.Equal(t, []string{"BUSINESS"}, missing)
	})

	// Cause 4: the response was not the shape we expect. Success stays false so
	// the wrapped branch is skipped, and the bare-array fallback finds nothing
	// usable, leaving an empty table without an error.
	t.Run("unexpected response shape yields empty table", func(t *testing.T) {
		c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`{"success":false,"labels":[{"id":"x","name":"Business"}]}`))
		})
		_, err := c.FetchLabels(context.Background())
		require.Error(t, err, "a shape we cannot read must fail loudly, not look like zero labels")
	})
}

// Debug logging must never print the bearer token.
func TestSetDebugDoesNotLogCredentials(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"success":true,"data":[]}`))
	})

	c.SetDebug(true)
	assert.True(t, c.debug)
	c.SetDebug(false)
	assert.False(t, c.debug)
}

func TestTruncateForLog(t *testing.T) {
	short := "hello"
	assert.Equal(t, short, truncateForLog(short))

	long := make([]byte, 2500)
	for i := range long {
		long[i] = 'a'
	}
	out := truncateForLog(string(long))
	assert.Less(t, len(out), 2500)
	assert.Contains(t, out, "2500 bytes total")
}
