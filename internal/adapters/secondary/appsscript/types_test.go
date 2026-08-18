package appsscript

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The Apps Script already splits column I, so the client must accept a JSON array.
func TestSheetTransactionDecodesLabels(t *testing.T) {
	body := `{
		"data": [
			{"no":"1","date":"04/05/2025","account":"BANK","category":"FOOD","amount":"1000","note":"n","type":"expense","labels":["BUSINESS","ADJUSTMENT"]},
			{"no":"2","date":"05/05/2025","account":"BANK","category":"FOOD","amount":"2000","note":"n","type":"expense","labels":[]}
		],
		"meta": {"total_data":2,"total_filtered":2,"skip":0,"limit":0}
	}`

	var resp SheetTransactionResponse
	require.NoError(t, json.Unmarshal([]byte(body), &resp))
	require.Len(t, resp.Data, 2)

	assert.Equal(t, []string{"BUSINESS", "ADJUSTMENT"}, resp.Data[0].Labels)
	assert.Empty(t, resp.Data[1].Labels, "empty label cell must decode to no labels")
	assert.Equal(t, 2, resp.Meta.TotalData)
}

// An older Apps Script deployment omits "labels" entirely. Decoding must still
// succeed so the two components can be rolled out independently.
func TestSheetTransactionWithoutLabelsField(t *testing.T) {
	body := `{"data":[{"no":"1","account":"BANK","amount":"1000","type":"expense"}],"meta":{"total_data":1}}`

	var resp SheetTransactionResponse
	require.NoError(t, json.Unmarshal([]byte(body), &resp))
	require.Len(t, resp.Data, 1)

	assert.Nil(t, resp.Data[0].Labels, "missing field must not fail decoding")
	assert.Equal(t, "BANK", resp.Data[0].Account)
}
