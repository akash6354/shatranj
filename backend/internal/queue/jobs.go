package queue

import "encoding/json"

// Job is the common envelope for background work.
type Job struct {
	ID      string          `json:"id"`
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}
