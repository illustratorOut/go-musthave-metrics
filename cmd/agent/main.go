package main

import (
	"fmt"
	"math/rand/v2"
	"net/http"
	"runtime"
	"strconv"
	"time"
)

// MetricsCollector собирает и отправляет метрики
type MetricsCollector struct {
	serverURL      string
	pollInterval   time.Duration
	reportInterval time.Duration
	pollCount      int64
}

// NewMetricsCollector создает новый сборщик метрик
func NewMetricsCollector(serverURL string, pollInterval, reportInterval time.Duration) *MetricsCollector {
	return &MetricsCollector{
		serverURL:      serverURL,
		pollInterval:   pollInterval,
		reportInterval: reportInterval,
		pollCount:      0,
	}
}

// collectRuntimeMetrics собирает метрики из пакета runtime
func (m *MetricsCollector) collectRuntimeMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Собираем gauge метрики из runtime
	metrics["Alloc"] = float64(memStats.Alloc)
	metrics["BuckHashSys"] = float64(memStats.BuckHashSys)
	metrics["Frees"] = float64(memStats.Frees)
	metrics["GCCPUFraction"] = memStats.GCCPUFraction
	metrics["GCSys"] = float64(memStats.GCSys)
	metrics["HeapAlloc"] = float64(memStats.HeapAlloc)
	metrics["HeapIdle"] = float64(memStats.HeapIdle)
	metrics["HeapInuse"] = float64(memStats.HeapInuse)
	metrics["HeapObjects"] = float64(memStats.HeapObjects)
	metrics["HeapReleased"] = float64(memStats.HeapReleased)
	metrics["HeapSys"] = float64(memStats.HeapSys)
	metrics["LastGC"] = float64(memStats.LastGC)
	metrics["Lookups"] = float64(memStats.Lookups)
	metrics["MCacheInuse"] = float64(memStats.MCacheInuse)
	metrics["MCacheSys"] = float64(memStats.MCacheSys)
	metrics["MSpanInuse"] = float64(memStats.MSpanInuse)
	metrics["MSpanSys"] = float64(memStats.MSpanSys)
	metrics["Mallocs"] = float64(memStats.Mallocs)
	metrics["NextGC"] = float64(memStats.NextGC)
	metrics["NumForcedGC"] = float64(memStats.NumForcedGC)
	metrics["NumGC"] = float64(memStats.NumGC)
	metrics["OtherSys"] = float64(memStats.OtherSys)
	metrics["PauseTotalNs"] = float64(memStats.PauseTotalNs)
	metrics["StackInuse"] = float64(memStats.StackInuse)
	metrics["StackSys"] = float64(memStats.StackSys)
	metrics["Sys"] = float64(memStats.Sys)
	metrics["TotalAlloc"] = float64(memStats.TotalAlloc)

	// Добавляем дополнительные метрики
	metrics["RandomValue"] = rand.Float64() // произвольное значение
	m.pollCount++                           // увеличиваем счетчик опросов

	return metrics
}

// sendMetric отправляет одну метрику на сервер
func (m *MetricsCollector) sendMetric(metricType, name string, value interface{}) error {
	var valueStr string

	switch v := value.(type) {
	case float64:
		valueStr = strconv.FormatFloat(v, 'f', -1, 64)
	case int64:
		valueStr = strconv.FormatInt(v, 10)
	default:
		return fmt.Errorf("unsupported metric value type: %T", value)
	}

	url := fmt.Sprintf("%s/update/%s/%s/%s", m.serverURL, metricType, name, valueStr)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "text/plain")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned non-200 status: %d", resp.StatusCode)
	}

	return nil
}

// sendMetrics отправляет все собранные метрики на сервер
func (m *MetricsCollector) sendMetrics() error {
	metrics := m.collectRuntimeMetrics()

	// Отправляем gauge метрики
	for name, value := range metrics {
		if err := m.sendMetric("gauge", name, value); err != nil {
			return fmt.Errorf("failed to send gauge metric %s: %v", name, err)
		}
	}

	// Отправляем counter метрику PollCount
	if err := m.sendMetric("counter", "PollCount", m.pollCount); err != nil {
		return fmt.Errorf("failed to send counter metric PollCount: %v", err)
	}

	return nil
}

// startPolling запускает периодический сбор метрик
func (m *MetricsCollector) startPolling() {
	ticker := time.NewTicker(m.pollInterval)
	defer ticker.Stop()

	for range ticker.C {
		// Просто собираем метрики, увеличивая pollCount
		m.collectRuntimeMetrics()
	}
}

// startReporting запускает периодическую отправку метрик
func (m *MetricsCollector) startReporting() {
	ticker := time.NewTicker(m.reportInterval)
	defer ticker.Stop()

	for range ticker.C {
		if err := m.sendMetrics(); err != nil {
			fmt.Printf("Error sending metrics: %v\n", err)
		} else {
			fmt.Printf("Metrics sent successfully at %s\n", time.Now().Format(time.RFC3339))
		}
	}
}

// Start запускает сбор и отправку метрик
func (m *MetricsCollector) Start() {
	fmt.Printf("Starting metrics collector:\n")
	fmt.Printf("  Server URL: %s\n", m.serverURL)
	fmt.Printf("  Poll interval: %v\n", m.pollInterval)
	fmt.Printf("  Report interval: %v\n", m.reportInterval)

	go m.startPolling()
	go m.startReporting()

	// Бесконечный цикл для поддержания работы приложения
	select {}
}

func main() {
	// Конфигурация по умолчанию
	serverURL := "http://localhost:8080"
	pollInterval := 2 * time.Second
	reportInterval := 10 * time.Second

	// Создаем и запускаем сборщик метрик
	collector := NewMetricsCollector(serverURL, pollInterval, reportInterval)
	collector.Start()
}
