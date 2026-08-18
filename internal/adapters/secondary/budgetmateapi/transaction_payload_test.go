package budgetmateapi

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Regression guard. The API accepted a transaction carrying "labels" with 201
// Created and then stored none, because its schema names the field "label_ids".
// An unknown field is ignored rather than rejected, so the only protection
// against this is asserting the wire name here.
func TestCreateTransactionRequestUsesLabelIDsField(t *testing.T) {
	req := CreateTransactionRequest{
		Amount:     3000,
		AccountID:  "cml5arqkk004n2cck4isqv4i9",
		CategoryID: "cml5arqez000t2cckrqdp4oo4",
		Date:       "2026-08-18T04:07:24Z",
		Type:       "expense",
		Labels:     []string{"cmsp8rv4x000104lakbi1jvuj"},
	}

	out, err := json.Marshal(req)
	require.NoError(t, err)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal(out, &decoded))

	assert.Equal(t, []any{"cmsp8rv4x000104lakbi1jvuj"}, decoded["label_ids"],
		"API schema names this field label_ids")
	assert.NotContains(t, decoded, "labels",
		"sending labels is silently ignored by the API")
}

// Unlabelled transactions must still produce the payload they always did.
func TestCreateTransactionRequestOmitsLabelIDsWhenEmpty(t *testing.T) {
	req := CreateTransactionRequest{
		Amount:     3000,
		AccountID:  "acc",
		CategoryID: "cat",
		Date:       "2026-08-18T04:07:24Z",
		Type:       "expense",
	}

	out, err := json.Marshal(req)
	require.NoError(t, err)
	assert.NotContains(t, string(out), "label_ids")
}

// The response echoes stored labels, which is what lets a push detect that the
// API accepted the request but kept no labels.
func TestCreateTransactionResponseExposesStoredLabels(t *testing.T) {
	t.Run("labels stored", func(t *testing.T) {
		body := `{"success":true,"data":{"id":"tx1","labels":[{"id":"lbl1","name":"BUSINESS"}]}}`

		var resp CreateTransactionResponse
		require.NoError(t, json.Unmarshal([]byte(body), &resp))
		require.Len(t, resp.Data.Labels, 1)
		assert.Equal(t, "BUSINESS", resp.Data.Labels[0].Name)
	})

	// This is exactly what the failing push returned: 201, success true, and an
	// empty label list.
	t.Run("labels dropped", func(t *testing.T) {
		body := `{"success":true,"data":{"id":"tx1","labels":[]}}`

		var resp CreateTransactionResponse
		require.NoError(t, json.Unmarshal([]byte(body), &resp))
		assert.True(t, resp.Success, "the API reports success even though nothing was stored")
		assert.Empty(t, resp.Data.Labels)
	})
}
