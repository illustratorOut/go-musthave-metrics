package main

import (
	"net/http"
	"strconv"
	"strings"
)

// MemStorage - хранилище метрик
type MemStorage struct {
	gauges   map[string]float64
	counters map[string]int64
}

// NewMemStorage создает новое хранилище
func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

// UpdateGauge обновляет gauge метрику - новое значение должно замещать предыдущее.
func (m *MemStorage) UpdateGauge(name string, value float64) {
	m.gauges[name] = value
}

// UpdateCounter обновляет counter метрику - новое значение должно добавляться к предыдущему, если какое-то значение уже было известно серверу
func (m *MemStorage) UpdateCounter(name string, value int64) {
	m.counters[name] += value
}

var storage = NewMemStorage()

func updateHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	// Разбираем путь
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 5 {
		http.Error(w, "Invalid path", http.StatusNotFound)
		return
	}

	metricType := parts[2]
	metricName := parts[3]
	metricValue := parts[4]

	// Проверяем имя метрики
	if metricName == "" {
		http.Error(w, "Metric name required", http.StatusNotFound)
		return
	}

	// Обрабатываем метрику
	switch metricType {
	case "gauge":
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			http.Error(w, "Invalid gauge value", http.StatusBadRequest)
			return
		}
		storage.UpdateGauge(metricName, value)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
		// fmt.Println(storage.gauges[metricName])

	case "counter":
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			http.Error(w, "Invalid counter value", http.StatusBadRequest)
			return
		}
		storage.UpdateCounter(metricName, value)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
		// fmt.Println(storage.counters[metricName])

	default:
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
		return
	}
}

func main() {
	http.HandleFunc("/update/", updateHandler)
	http.ListenAndServe(":8080", nil)
}
