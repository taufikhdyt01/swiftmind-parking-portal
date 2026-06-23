package violation

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"parkwatch/pkg/fine"
)

// ActiveRuleset is the subset of the rules service's response we need to price
// and snapshot a violation.
type ActiveRuleset struct {
	ID      string       `json:"id"`
	Version int          `json:"version"`
	Ruleset fine.Ruleset `json:"ruleset"`
}

// RulesClient fetches the active ruleset from the rules service over HTTP
// (service-to-service, on the internal network — not via the gateway).
type RulesClient struct {
	baseURL string
	http    *http.Client
}

// NewRulesClient builds a client for the given rules service base URL.
func NewRulesClient(baseURL string) *RulesClient {
	return &RulesClient{baseURL: baseURL, http: &http.Client{Timeout: 10 * time.Second}}
}

// Active returns the currently active ruleset.
func (c *RulesClient) Active(ctx context.Context) (*ActiveRuleset, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/active", nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("rules service returned %d", resp.StatusCode)
	}
	var out ActiveRuleset
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}
