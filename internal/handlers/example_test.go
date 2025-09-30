package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/volchkovski/go-practicum-metrics/internal/handlers"
	"github.com/volchkovski/go-practicum-metrics/internal/models"
	"github.com/volchkovski/go-practicum-metrics/internal/routers"
	"github.com/volchkovski/go-practicum-metrics/internal/services"
	"github.com/volchkovski/go-practicum-metrics/internal/storage/mem"
)

// Example demonstrates how to collect a gauge metric via URL path parameters using a full router.
func ExampleCollectMetricHandler() {
	// Initialize storage and service
	storage := mem.NewMemStorage()
	service := services.NewMetricService(storage)

	// Create router to properly handle URL parameters
	router := routers.NewMetricRouter("", service)
	server := httptest.NewServer(router)
	defer server.Close()

	// Make request to store gauge metric
	resp, err := http.Post(server.URL+"/update/gauge/temperature/23.5", "", nil)
	if err != nil {
		log.Fatal(err)
	}
	resp.Body.Close()

	fmt.Printf("Status: %d\n", resp.StatusCode)
	// Output: Status: 200
}

// Example demonstrates how to retrieve a metric value via URL path parameters using a full router.
func ExampleMetricHandler() {
	// Initialize storage and service
	storage := mem.NewMemStorage()
	service := services.NewMetricService(storage)

	// Store a test metric first
	_ = storage.WriteGauge(context.Background(), "temperature", 23.5)

	// Create router to properly handle URL parameters
	router := routers.NewMetricRouter("", service)
	server := httptest.NewServer(router)
	defer server.Close()

	// Make request to get gauge metric
	resp, err := http.Get(server.URL + "/value/gauge/temperature")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Body: %s\n", strings.TrimSpace(string(body)))
	// Output:
	// Status: 200
	// Body: 23.5
}

// Example demonstrates how to collect a metric using JSON payload.
func ExampleCollectMetricHandlerJSON() {
	// Initialize storage and service
	storage := mem.NewMemStorage()
	service := services.NewMetricService(storage)

	// Create handler
	handler := handlers.CollectMetricHandlerJSON(service)

	// Prepare JSON payload for gauge metric
	gaugeValue := 25.7
	metric := models.Metrics{
		ID:    "cpu_usage",
		MType: "gauge",
		Value: &gaugeValue,
	}

	jsonData, err := json.Marshal(metric)
	if err != nil {
		log.Fatal(err)
	}

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.ServeHTTP(rr, req)

	fmt.Printf("Status: %d\n", rr.Code)
	// Output: Status: 200
}

// Example demonstrates how to retrieve a metric using JSON request.
func ExampleMetricHandlerJSON() {
	// Initialize storage and service
	storage := mem.NewMemStorage()
	service := services.NewMetricService(storage)

	// Store a test metric first
	_ = storage.WriteCounter(context.Background(), "requests", 42)

	// Create handler
	handler := handlers.MetricHandlerJSON(service)

	// Prepare JSON request
	metric := models.Metrics{
		ID:    "requests",
		MType: "counter",
	}

	jsonData, err := json.Marshal(metric)
	if err != nil {
		log.Fatal(err)
	}

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/value/", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.ServeHTTP(rr, req)

	fmt.Printf("Status: %d\n", rr.Code)

	var response models.Metrics
	_ = json.Unmarshal(rr.Body.Bytes(), &response)
	fmt.Printf("Metric ID: %s, Value: %d\n", response.ID, *response.Delta)
	// Output:
	// Status: 200
	// Metric ID: requests, Value: 42
}

// Example demonstrates how to collect multiple metrics in a single request.
func ExampleCollectMetricsHandlerJSON() {
	// Initialize storage and service
	storage := mem.NewMemStorage()
	service := services.NewMetricService(storage)

	// Create handler
	handler := handlers.CollectMetricsHandlerJSON(service)

	// Prepare multiple metrics
	gaugeValue := 80.5
	counterValue := int64(100)

	metrics := []models.Metrics{
		{
			ID:    "memory_usage",
			MType: "gauge",
			Value: &gaugeValue,
		},
		{
			ID:    "total_requests",
			MType: "counter",
			Delta: &counterValue,
		},
	}

	jsonData, err := json.Marshal(metrics)
	if err != nil {
		log.Fatal(err)
	}

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.ServeHTTP(rr, req)

	fmt.Printf("Status: %d\n", rr.Code)
	// Output: Status: 200
}

// Example demonstrates the complete metrics API using a router.
func ExampleNewMetricRouter() {
	// Initialize storage and service
	storage := mem.NewMemStorage()
	service := services.NewMetricService(storage)

	// Create router with no secret key (no hash middleware)
	router := routers.NewMetricRouter("", service)

	// Start a test server
	server := httptest.NewServer(router)
	defer server.Close()

	// Example 1: Store a gauge metric
	gaugeReq := fmt.Sprintf(`{"id":"temperature","type":"gauge","value":23.5}`)
	resp, err := http.Post(server.URL+"/update/", "application/json", strings.NewReader(gaugeReq))
	if err != nil {
		log.Fatal(err)
	}
	resp.Body.Close()
	fmt.Printf("Store gauge - Status: %d\n", resp.StatusCode)

	// Example 2: Store a counter metric
	counterReq := fmt.Sprintf(`{"id":"requests","type":"counter","delta":10}`)
	resp, err = http.Post(server.URL+"/update/", "application/json", strings.NewReader(counterReq))
	if err != nil {
		log.Fatal(err)
	}
	resp.Body.Close()
	fmt.Printf("Store counter - Status: %d\n", resp.StatusCode)

	// Example 3: Retrieve a metric
	valueReq := fmt.Sprintf(`{"id":"temperature","type":"gauge"}`)
	resp, err = http.Post(server.URL+"/value/", "application/json", strings.NewReader(valueReq))
	if err != nil {
		log.Fatal(err)
	}
	resp.Body.Close()
	fmt.Printf("Get metric - Status: %d\n", resp.StatusCode)

	// Example 4: Check database connection
	resp, err = http.Get(server.URL + "/ping")
	if err != nil {
		log.Fatal(err)
	}
	resp.Body.Close()
	fmt.Printf("Ping DB - Status: %d\n", resp.StatusCode)

	// Output:
	// Store gauge - Status: 200
	// Store counter - Status: 200
	// Get metric - Status: 200
	// Ping DB - Status: 200
}
