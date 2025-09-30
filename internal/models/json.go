package models

// Metrics represents a metric in JSON format for HTTP API communication.
// It supports both gauge and counter types of metrics through optional fields.
type Metrics struct {
	ID    string   `json:"id"`              // ID is the unique name/identifier of the metric
	MType string   `json:"type"`            // MType specifies the metric type: "gauge" or "counter"
	Delta *int64   `json:"delta,omitempty"` // Delta contains the counter value (only for counter type)
	Value *float64 `json:"value,omitempty"` // Value contains the gauge value (only for gauge type)
}
