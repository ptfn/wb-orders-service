package httpserver

import (
	"encoding/json"
	"log"
	"net/http"
	"wb-orders-service/service"
)

type Handlers struct {
	service *service.OrderService
}

func NewHandlers(service *service.OrderService) *Handlers {
	return &Handlers{
		service: service,
	}
}

// GetOrderHandler обрабатывает запрос на получение заказа по ID
func (h *Handlers) GetOrderHandler(w http.ResponseWriter, r *http.Request) {
	// Разрешаем CORS для простоты тестирования
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Обрабатываем preflight OPTIONS запрос
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Проверяем метод запроса
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Извлекаем order_uid из URL пути
	// Ожидаем путь вида /order/12345
	if len(r.URL.Path) < 7 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	orderUID := r.URL.Path[7:] // Убираем "/order/" из пути
	if orderUID == "" {
		http.Error(w, "Order ID is required", http.StatusBadRequest)
		return
	}

	log.Printf("Received request for order: %s", orderUID)

	// Получаем заказ из сервиса
	order, err := h.service.GetOrder(orderUID)
	if err != nil {
		log.Printf("Order not found: %s, error: %v", orderUID, err)
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	// Устанавливаем заголовок Content-Type
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	// Кодируем заказ в JSON и отправляем
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ") // Для красивого форматирования JSON
	if err := encoder.Encode(order); err != nil {
		log.Printf("Failed to encode order: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("Order %s sent successfully", orderUID)
}

// HealthCheckHandler для проверки работоспособности сервиса
func (h *Handlers) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "ok",
		"cacheSize": h.service.GetCacheSize(),
	})
}

// NotFoundHandler для несуществующих маршрутов
func (h *Handlers) NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(map[string]string{
		"error": "Endpoint not found",
	})
}
