// Simple proto message structures for metrics service
// This is a simplified implementation for demonstration purposes

package pb

// GaugeMetric represents a gauge-type metric with a floating-point value
type GaugeMetric struct {
	Name  string
	Value float64
}

func (x *GaugeMetric) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *GaugeMetric) GetValue() float64 {
	if x != nil {
		return x.Value
	}
	return 0
}

// CounterMetric represents a counter-type metric with an integer value
type CounterMetric struct {
	Name  string
	Value int64
}

func (x *CounterMetric) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *CounterMetric) GetValue() int64 {
	if x != nil {
		return x.Value
	}
	return 0
}

// PushGaugeRequest contains the gauge metric to be stored
type PushGaugeRequest struct {
	Metric *GaugeMetric
}

func (x *PushGaugeRequest) GetMetric() *GaugeMetric {
	if x != nil {
		return x.Metric
	}
	return nil
}

// PushGaugeResponse confirms the successful storage of a gauge metric
type PushGaugeResponse struct {
	Success bool
}

func (x *PushGaugeResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

// PushCounterRequest contains the counter metric to be stored
type PushCounterRequest struct {
	Metric *CounterMetric
}

func (x *PushCounterRequest) GetMetric() *CounterMetric {
	if x != nil {
		return x.Metric
	}
	return nil
}

// PushCounterResponse confirms the successful storage of a counter metric
type PushCounterResponse struct {
	Success bool
}

func (x *PushCounterResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

// PushMetricsRequest contains multiple metrics to be stored in a batch
type PushMetricsRequest struct {
	Gauges   []*GaugeMetric
	Counters []*CounterMetric
}

func (x *PushMetricsRequest) GetGauges() []*GaugeMetric {
	if x != nil {
		return x.Gauges
	}
	return nil
}

func (x *PushMetricsRequest) GetCounters() []*CounterMetric {
	if x != nil {
		return x.Counters
	}
	return nil
}

// PushMetricsResponse confirms the successful storage of batch metrics
type PushMetricsResponse struct {
	Success bool
}

func (x *PushMetricsResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

// GetGaugeRequest specifies the name of the gauge metric to retrieve
type GetGaugeRequest struct {
	Name string
}

func (x *GetGaugeRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

// GetGaugeResponse contains the requested gauge metric
type GetGaugeResponse struct {
	Metric *GaugeMetric
}

func (x *GetGaugeResponse) GetMetric() *GaugeMetric {
	if x != nil {
		return x.Metric
	}
	return nil
}

// GetCounterRequest specifies the name of the counter metric to retrieve
type GetCounterRequest struct {
	Name string
}

func (x *GetCounterRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

// GetCounterResponse contains the requested counter metric
type GetCounterResponse struct {
	Metric *CounterMetric
}

func (x *GetCounterResponse) GetMetric() *CounterMetric {
	if x != nil {
		return x.Metric
	}
	return nil
}

// GetAllMetricsRequest is an empty request to retrieve all metrics
type GetAllMetricsRequest struct{}

// GetAllMetricsResponse contains all stored metrics
type GetAllMetricsResponse struct {
	Gauges   []*GaugeMetric
	Counters []*CounterMetric
}

func (x *GetAllMetricsResponse) GetGauges() []*GaugeMetric {
	if x != nil {
		return x.Gauges
	}
	return nil
}

func (x *GetAllMetricsResponse) GetCounters() []*CounterMetric {
	if x != nil {
		return x.Counters
	}
	return nil
}

// PingRequest is an empty request to check service health
type PingRequest struct{}

// PingResponse confirms the service is healthy
type PingResponse struct {
	Healthy bool
}

func (x *PingResponse) GetHealthy() bool {
	if x != nil {
		return x.Healthy
	}
	return false
}

