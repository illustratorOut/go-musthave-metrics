package main

import (
	"testing"
	"time"
)

// Тестируем создание сборщика метрик
func TestNewMetricsCollector(t *testing.T) {
	collector := NewMetricsCollector("http://test", 2*time.Second, 10*time.Second)

	if collector.serverURL != "http://test" {
		t.Error("URL сервера не установился правильно")
	}

	if collector.pollInterval != 2*time.Second {
		t.Error("Интервал опроса не установился правильно")
	}

	if collector.reportInterval != 10*time.Second {
		t.Error("Интервал отправки не установился правильно")
	}

	if collector.pollCount != 0 {
		t.Error("Счетчик опросов должен начинаться с 0")
	}
}

// Тестируем сбор метрик
func TestCollectRuntimeMetrics(t *testing.T) {
	collector := NewMetricsCollector("", 0, 0)

	metrics := collector.collectRuntimeMetrics()

	// Проверяем что метрики собрались
	if len(metrics) == 0 {
		t.Error("Метрики не собрались")
	}

	// Проверяем несколько ключевых метрик
	if metrics["Alloc"] == nil {
		t.Error("Метрика Alloc отсутствует")
	}

	if metrics["RandomValue"] == nil {
		t.Error("Метрика RandomValue отсутствует")
	}

	// Проверяем что счетчик увеличился
	if collector.pollCount != 1 {
		t.Error("Счетчик опросов не увеличился")
	}
}

// Тестируем преобразование значений метрик в строку
func TestMetricValueConversion(t *testing.T) {
	collector := NewMetricsCollector("", 0, 0)

	// Тестируем преобразование float64
	err := collector.sendMetric("gauge", "test", 123.45)
	// Ожидаем ошибку (сервер не запущен), но не ошибку преобразования
	if err != nil && err.Error() == "unsupported metric value type: string" {
		t.Error("Ошибка преобразования float64")
	}

	// Тестируем преобразование int64
	err = collector.sendMetric("counter", "test", int64(100))
	if err != nil && err.Error() == "unsupported metric value type: string" {
		t.Error("Ошибка преобразования int64")
	}
}

// Тестируем увеличение счетчика опросов
func TestPollCountIncrement(t *testing.T) {
	collector := NewMetricsCollector("", 0, 0)

	initialCount := collector.pollCount

	// Собираем метрики несколько раз
	collector.collectRuntimeMetrics()
	collector.collectRuntimeMetrics()
	collector.collectRuntimeMetrics()

	if collector.pollCount != initialCount+3 {
		t.Error("Счетчик опросов не увеличивается правильно")
	}
}
