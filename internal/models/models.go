package models

type GaugeMetric struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

type CounterMetric struct {
	Name  string `json:"name"`
	Value int64  `json:"value"`
}
