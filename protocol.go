package herdrsock

import "encoding/json"

// EmptyParams serializes to {}.
type EmptyParams struct{}

// Request is Herdr's newline-delimited JSON request shape.
type Request struct {
	ID     string `json:"id"`
	Method string `json:"method"`
	Params any    `json:"params"`
}

// Response is Herdr's raw success/error response shape.
type Response struct {
	ID     string          `json:"id"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  *WireError      `json:"error,omitempty"`
}

type typedResult struct {
	Type string `json:"type"`
}

// ResultType returns result.type without decoding the full payload.
func (r *Response) ResultType() (string, error) {
	var out typedResult
	if err := json.Unmarshal(r.Result, &out); err != nil {
		return "", err
	}
	return out.Type, nil
}
