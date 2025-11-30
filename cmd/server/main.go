package main

import (
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
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

// UpdateGauge обновляет gauge метрику
func (m *MemStorage) UpdateGauge(name string, value float64) {
	m.gauges[name] = value
}

// UpdateCounter обновляет counter метрику
func (m *MemStorage) UpdateCounter(name string, value int64) {
	m.counters[name] += value
}

// GetGauge возвращает значение gauge метрики
func (m *MemStorage) GetGauge(name string) (float64, bool) {
	value, exists := m.gauges[name]
	return value, exists
}

// GetCounter возвращает значение counter метрики
func (m *MemStorage) GetCounter(name string) (int64, bool) {
	value, exists := m.counters[name]
	return value, exists
}

// GetAllMetrics возвращает все метрики
func (m *MemStorage) GetAllMetrics() (map[string]float64, map[string]int64) {
	return m.gauges, m.counters
}

var storage = NewMemStorage()

func main() {
	// Обработка флагов
	var addr string
	flag.StringVar(&addr, "a", "localhost:8080", "HTTP server endpoint address")
	flag.Parse()

	// Проверяем наличие неизвестных флагов
	if flag.NArg() > 0 {
		fmt.Printf("Error: unknown flag(s): %v\n", flag.Args())
		return
	}

	fmt.Printf("Starting server on %s\n", addr)

	router := gin.Default()

	// Настраиваем шаблоны
	router.SetHTMLTemplate(createTemplate())

	// Обработчик для главной страницы со списком метрик
	router.GET("/", func(c *gin.Context) {
		gauges, counters := storage.GetAllMetrics()

		c.HTML(http.StatusOK, "metrics.html", gin.H{
			"Gauges":   gauges,
			"Counters": counters,
		})
	})

	// Обработчик для получения значения метрики
	router.GET("/value/:type/:name", func(c *gin.Context) {
		metricType := c.Param("type")
		metricName := c.Param("name")

		switch metricType {
		case "gauge":
			if value, exists := storage.GetGauge(metricName); exists {
				c.String(http.StatusOK, fmt.Sprintf("%v", value))
				return
			}
		case "counter":
			if value, exists := storage.GetCounter(metricName); exists {
				c.String(http.StatusOK, fmt.Sprintf("%d", value))
				return
			}
		default:
			c.Status(http.StatusNotFound)
			return
		}

		c.Status(http.StatusNotFound)
	})

	// Обработчик для обновления метрик
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

	router.Run(addr)
}

// createTemplate создает HTML шаблон для отображения метрик
func createTemplate() *template.Template {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>Metrics</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        table { border-collapse: collapse; width: 100%; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
        .section { margin-bottom: 30px; }
    </style>
</head>
<body>
    <h1>Metrics</h1>
    
    <div class="section">
        <h2>Gauges</h2>
        {{if .Gauges}}
        <table>
            <tr><th>Name</th><th>Value</th></tr>
            {{range $name, $value := .Gauges}}
            <tr><td>{{$name}}</td><td>{{$value}}</td></tr>
            {{end}}
        </table>
        {{else}}
        <p>No gauge metrics</p>
        {{end}}
    </div>
    
    <div class="section">
        <h2>Counters</h2>
        {{if .Counters}}
        <table>
            <tr><th>Name</th><th>Value</th></tr>
            {{range $name, $value := .Counters}}
            <tr><td>{{$name}}</td><td>{{$value}}</td></tr>
            {{end}}
        </table>
        {{else}}
        <p>No counter metrics</p>
        {{end}}
    </div>
</body>
</html>`

	return template.Must(template.New("metrics.html").Parse(tmpl))
}
