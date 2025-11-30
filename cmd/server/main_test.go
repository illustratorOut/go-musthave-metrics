package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Тестируем обновление gauge метрики
func TestUpdateGauge(t *testing.T) {
	storage := NewMemStorage()

	// Проверяем сохранение значения
	storage.UpdateGauge("test", 123.45)
	if storage.gauges["test"] != 123.45 {
		t.Error("Gauge значение не сохранилось правильно")
	}

	// Проверяем перезапись значения
	storage.UpdateGauge("test", 67.89)
	if storage.gauges["test"] != 67.89 {
		t.Error("Gauge значение не перезаписалось")
	}
}

// Тестируем обновление counter метрики
func TestUpdateCounter(t *testing.T) {
	storage := NewMemStorage()

	// Проверяем первое сохранение
	storage.UpdateCounter("test", 10)
	if storage.counters["test"] != 10 {
		t.Error("Counter значение не сохранилось правильно")
	}

	// Проверяем инкремент
	storage.UpdateCounter("test", 5)
	if storage.counters["test"] != 15 {
		t.Error("Counter значение не увеличилось")
	}
}

// Тестируем обработчик с правильными данными
func TestUpdateHandler_Success(t *testing.T) {
	storage = NewMemStorage() // Сбрасываем хранилище

	// Тест gauge метрики
	req := httptest.NewRequest("POST", "/update/gauge/cpu_usage/95.5", nil)
	rr := httptest.NewRecorder()

	updateHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Error("Ожидался статус 200")
	}

	if storage.gauges["cpu_usage"] != 95.5 {
		t.Error("Gauge метрика не сохранилась через обработчик")
	}

	// Тест counter метрики
	req = httptest.NewRequest("POST", "/update/counter/requests/10", nil)
	rr = httptest.NewRecorder()

	updateHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Error("Ожидался статус 200")
	}

	if storage.counters["requests"] != 10 {
		t.Error("Counter метрика не сохранилась через обработчик")
	}
}

// Тестируем обработчик с ошибками
func TestUpdateHandler_Errors(t *testing.T) {
	storage = NewMemStorage()

	tests := []struct {
		name string
		url  string
		want int
	}{
		{
			name: "Неправильный метод",
			url:  "/update/gauge/test/1.0",
			want: http.StatusMethodNotAllowed,
		},
		{
			name: "Неправильный путь",
			url:  "/update/gauge",
			want: http.StatusNotFound,
		},
		{
			name: "Пустое имя метрики",
			url:  "/update/gauge//1.0",
			want: http.StatusNotFound,
		},
		{
			name: "Неправильный тип метрики",
			url:  "/update/invalid/test/1.0",
			want: http.StatusBadRequest,
		},
		{
			name: "Неправильное значение gauge",
			url:  "/update/gauge/test/invalid",
			want: http.StatusBadRequest,
		},
		{
			name: "Неправильное значение counter",
			url:  "/update/counter/test/invalid",
			want: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.name == "Неправильный метод" {
				req = httptest.NewRequest("GET", tt.url, nil)
			} else {
				req = httptest.NewRequest("POST", tt.url, nil)
			}

			rr := httptest.NewRecorder()
			updateHandler(rr, req)

			if rr.Code != tt.want {
				t.Errorf("Ожидался статус %d, получили %d", tt.want, rr.Code)
			}
		})
	}
}
