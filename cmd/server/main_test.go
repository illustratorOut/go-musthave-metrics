package main

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestMemStorage(t *testing.T) {
	storage := NewMemStorage()

	// Тест gauge метрики
	storage.UpdateGauge("test_gauge", 123.45)
	if storage.gauges["test_gauge"] != 123.45 {
		t.Error("Gauge значение не сохранилось правильно")
	}

	// Тест counter метрики
	storage.UpdateCounter("test_counter", 10)
	storage.UpdateCounter("test_counter", 5)
	if storage.counters["test_counter"] != 15 {
		t.Error("Counter значение не увеличилось правильно")
	}

	// Тест методов получения
	if val, _ := storage.GetGauge("test_gauge"); val != 123.45 {
		t.Error("GetGauge не работает")
	}
	if val, _ := storage.GetCounter("test_counter"); val != 15 {
		t.Error("GetCounter не работает")
	}
}

func TestUpdateHandler(t *testing.T) {
	// Настройка
	router := setupRouter()
	storage = NewMemStorage()

	// Тест успешного обновления gauge
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/update/gauge/test/95.5", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Ожидался статус 200, получили %d", w.Code)
	}
	if storage.gauges["test"] != 95.5 {
		t.Error("Gauge метрика не сохранилась")
	}

	// Тест успешного обновления counter
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/update/counter/requests/10", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Ожидался статус 200, получили %d", w.Code)
	}
	if storage.counters["requests"] != 10 {
		t.Error("Counter метрика не сохранилась")
	}
}

func TestValueHandler(t *testing.T) {
	// Настройка
	router := setupRouter()
	storage = NewMemStorage()
	storage.UpdateGauge("cpu", 75.5)
	storage.UpdateCounter("hits", 42)

	// Тест получения gauge
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/value/gauge/cpu", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Ожидался статус 200, получили %d", w.Code)
	}
	if w.Body.String() != "75.5" {
		t.Errorf("Ожидалось '75.5', получили '%s'", w.Body.String())
	}

	// Тест получения counter
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/value/counter/hits", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Ожидался статус 200, получили %d", w.Code)
	}
	if w.Body.String() != "42" {
		t.Errorf("Ожидалось '42', получили '%s'", w.Body.String())
	}

	// Тест несуществующей метрики
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/value/gauge/nonexistent", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Ожидался статус 404, получили %d", w.Code)
	}
}

func TestMainPage(t *testing.T) {
	router := setupRouter()
	storage = NewMemStorage()
	storage.UpdateGauge("memory", 1024.0)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Ожидался статус 200, получили %d", w.Code)
	}
	if w.Header().Get("Content-Type") != "text/html; charset=utf-8" {
		t.Error("Ожидался HTML контент")
	}
}

func TestErrorCases(t *testing.T) {
	router := setupRouter()

	tests := []struct {
		method string
		url    string
		status int
	}{
		{"GET", "/update/gauge/test/1.0", 404},      // Неправильный метод - Gin возвращает 404
		{"POST", "/update/invalid/test/1.0", 400},   // Неправильный тип
		{"POST", "/update/gauge/test/invalid", 400}, // Неправильное значение
		{"POST", "/update/gauge//1.0", 404},         // Пустое имя
	}

	for _, test := range tests {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(test.method, test.url, nil)
		router.ServeHTTP(w, req)

		if w.Code != test.status {
			t.Errorf("Для %s %s ожидался %d, получили %d", test.method, test.url, test.status, w.Code)
		}
	}
}

// Вспомогательная функция для создания роутера для тестов
func setupRouter() *gin.Engine {
	storage = NewMemStorage()

	router := gin.Default()
	router.SetHTMLTemplate(createTemplate())

	router.GET("/", func(c *gin.Context) {
		gauges, counters := storage.GetAllMetrics()
		c.HTML(http.StatusOK, "metrics.html", gin.H{
			"Gauges":   gauges,
			"Counters": counters,
		})
	})

	router.GET("/value/:type/:name", func(c *gin.Context) {
		metricType := c.Param("type")
		metricName := c.Param("name")

		switch metricType {
		case "gauge":
			if value, exists := storage.GetGauge(metricName); exists {
				c.String(http.StatusOK, "%v", value)
				return
			}
		case "counter":
			if value, exists := storage.GetCounter(metricName); exists {
				c.String(http.StatusOK, "%d", value)
				return
			}
		default:
			c.Status(http.StatusNotFound)
			return
		}

		c.Status(http.StatusNotFound)
	})

	router.POST("/update/:type/:name/:value", func(c *gin.Context) {
		metricType := c.Param("type")
		metricName := c.Param("name")
		metricValue := c.Param("value")

		if metricName == "" {
			c.String(http.StatusNotFound, "Metric name required")
			return
		}

		switch metricType {
		case "gauge":
			value, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				c.String(http.StatusBadRequest, "Invalid gauge value")
				return
			}
			storage.UpdateGauge(metricName, value)
			c.String(http.StatusOK, "OK")

		case "counter":
			value, err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				c.String(http.StatusBadRequest, "Invalid counter value")
				return
			}
			storage.UpdateCounter(metricName, value)
			c.String(http.StatusOK, "OK")

		default:
			c.String(http.StatusBadRequest, "Invalid metric type")
			return
		}
	})

	return router
}
