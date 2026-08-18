package budgetbakers_app

import (
	"testing"

	"github.com/etwicaksono/data-transaksi-service/internal/adapters/secondary/budgetbakers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newLabelAdapter() *adapter {
	return &adapter{
		labels: []budgetbakers.Label{
			{ID: "id-business", Name: "Business"},
			{ID: "id-adjustment", Name: "Adjustment"},
		},
	}
}

func TestFindLabelIDs(t *testing.T) {
	a := newLabelAdapter()

	// The sheet stores labels uppercased while the API returns them title-cased.
	t.Run("matches ignoring case", func(t *testing.T) {
		ids, err := a.findLabelIDs([]string{"BUSINESS", "ADJUSTMENT"})
		require.NoError(t, err)
		assert.Equal(t, []string{"id-business", "id-adjustment"}, ids)
	})

	// An unlabelled transaction must leave LabelIDs nil so omitempty drops the
	// field and the payload stays identical to the pre-label behaviour.
	t.Run("no labels yields nil", func(t *testing.T) {
		ids, err := a.findLabelIDs(nil)
		require.NoError(t, err)
		assert.Nil(t, ids)

		ids, err = a.findLabelIDs([]string{})
		require.NoError(t, err)
		assert.Nil(t, ids)
	})

	t.Run("blank entries are ignored", func(t *testing.T) {
		ids, err := a.findLabelIDs([]string{"BUSINESS", "  "})
		require.NoError(t, err)
		assert.Equal(t, []string{"id-business"}, ids)
	})

	t.Run("only blanks yields nil", func(t *testing.T) {
		ids, err := a.findLabelIDs([]string{"   ", ""})
		require.NoError(t, err)
		assert.Nil(t, ids)
	})

	t.Run("duplicates collapse", func(t *testing.T) {
		ids, err := a.findLabelIDs([]string{"BUSINESS", "business", "Business"})
		require.NoError(t, err)
		assert.Equal(t, []string{"id-business"}, ids)
	})

	// All-or-nothing: partially labelled data is harder to notice than a record
	// that was clearly skipped.
	t.Run("unknown label fails whole record", func(t *testing.T) {
		ids, err := a.findLabelIDs([]string{"BUSINESS", "GHOST"})
		require.Error(t, err)
		assert.Nil(t, ids)
		assert.Contains(t, err.Error(), "GHOST")
		assert.Contains(t, err.Error(), "Business", "error should list what is available")
	})

	// Labels are short words, so a substring match would too easily pick wrongly.
	t.Run("does not match on substring", func(t *testing.T) {
		_, err := a.findLabelIDs([]string{"BUSI"})
		require.Error(t, err)
	})
}

func TestFindLabelIDsWithNoLabelsLoaded(t *testing.T) {
	a := &adapter{}

	t.Run("unlabelled transaction still works", func(t *testing.T) {
		ids, err := a.findLabelIDs(nil)
		require.NoError(t, err)
		assert.Nil(t, ids)
	})

	t.Run("labelled transaction fails clearly", func(t *testing.T) {
		_, err := a.findLabelIDs([]string{"BUSINESS"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "label not found")
	})
}
