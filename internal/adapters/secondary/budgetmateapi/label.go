package budgetmateapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// FetchLabels fetches all labels from Finance Manager
func (c *Client) FetchLabels(ctx context.Context) ([]Label, error) {
	body, status, err := c.doRequest(ctx, http.MethodGet, "/api/v1/labels", nil)
	if err != nil {
		return nil, err
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("fetch labels failed with status %d: %s", status, string(body))
	}

	// Try wrapped response first
	var resp LabelsResponse
	if err := json.Unmarshal(body, &resp); err == nil && resp.Success {
		for _, label := range resp.Data {
			c.labelMap[strings.ToLower(label.Name)] = label.ID
		}
		c.logLoadedLabels("wrapped", resp.Data)
		return resp.Data, nil
	}

	// Fallback to direct array
	var labels []Label
	if err := json.Unmarshal(body, &labels); err != nil {
		return nil, fmt.Errorf("decode labels: %w (body: %s)", err, string(body))
	}
	for _, label := range labels {
		c.labelMap[strings.ToLower(label.Name)] = label.ID
	}
	c.logLoadedLabels("bare array", labels)
	return labels, nil
}

// logLoadedLabels reports what the lookup table ended up containing. An empty
// table is the difference between "this label is unknown" and "labels were never
// loaded", which is otherwise invisible at the call site.
func (c *Client) logLoadedLabels(shape string, labels []Label) {
	if !c.debug {
		return
	}

	fmt.Printf("  [labels] parsed %d label(s) from %s response\n", len(labels), shape)
	for _, label := range labels {
		fmt.Printf("           - %q -> %s\n", label.Name, label.ID)
	}
	if len(labels) == 0 {
		fmt.Printf("           WARNING: no labels loaded, every labelled transaction will fail\n")
	}
}

// GetLabelID returns the label ID for a given name (case-insensitive)
func (c *Client) GetLabelID(name string) (string, bool) {
	id, ok := c.labelMap[strings.ToLower(strings.TrimSpace(name))]
	return id, ok
}

// GetLabelIDWithMapping returns the label ID for a given name (with mapping support)
func (c *Client) GetLabelIDWithMapping(name string) (string, bool) {
	lowerName := strings.ToLower(strings.TrimSpace(name))

	// Check if there's a mapping
	if mappedName, ok := c.labelMapping[lowerName]; ok {
		lowerName = mappedName
	}

	id, ok := c.labelMap[lowerName]
	return id, ok
}

// GetLabelNames returns all available label names
func (c *Client) GetLabelNames() []string {
	names := make([]string, 0, len(c.labelMap))
	for name := range c.labelMap {
		names = append(names, name)
	}
	return names
}

// ResolveLabelIDs converts sheet label names into Finance Manager label IDs.
// Resolution is all-or-nothing: if any name is unknown the caller gets the list
// of missing names, so a transaction is never sent with only part of its labels.
func (c *Client) ResolveLabelIDs(names []string) (ids []string, missing []string) {
	if len(names) == 0 {
		return nil, nil
	}

	ids = make([]string, 0, len(names))
	seen := make(map[string]struct{}, len(names))

	for _, name := range names {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			continue
		}

		id, ok := c.GetLabelIDWithMapping(trimmed)
		if !ok {
			missing = append(missing, trimmed)
			continue
		}

		if _, duplicate := seen[id]; duplicate {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}

	if len(missing) > 0 {
		return nil, missing
	}
	if len(ids) == 0 {
		return nil, nil
	}
	return ids, nil
}
