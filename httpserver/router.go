package httpserver

import (
	"net/http"
	"os"
	"path/filepath"
	"wb-orders-service/service"
)

type Router struct {
	handlers *Handlers
}

func NewRouter(service *service.OrderService) *Router {
	return &Router{
		handlers: NewHandlers(service),
	}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch {
	case req.URL.Path == "/":
		r.serveIndex(w, req)
	case req.URL.Path == "/health":
		r.handlers.HealthCheckHandler(w, req)
	case len(req.URL.Path) > 7 && req.URL.Path[:7] == "/order/":
		r.handlers.GetOrderHandler(w, req)
	default:
		r.handlers.NotFoundHandler(w, req)
	}
}

func (r *Router) serveIndex(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Пытаемся найти HTML файл в нескольких возможных местах
	possiblePaths := []string{
		"web/index.html",
		"./web/index.html",
		"../web/index.html",
		"static/index.html",
	}

	var htmlPath string
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			htmlPath = path
			break
		}
	}

	// Если файл не найден, возвращаем ошибку
	if htmlPath == "" {
		http.Error(w, "HTML file not found", http.StatusInternalServerError)
		return
	}

	// Получаем абсолютный путь
	absPath, err := filepath.Abs(htmlPath)
	if err != nil {
		http.Error(w, "Failed to get absolute path", http.StatusInternalServerError)
		return
	}

	// Отдаём HTML файл
	http.ServeFile(w, req, absPath)
}
