package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
	"wb-orders-service/models"

	"github.com/nats-io/stan.go"
)

func main() {
	// Конфигурация NATS
	clusterID := "test-cluster"
	clientID := "test-publisher"
	url := "nats://localhost:4222"
	subject := "orders"

	// Подключаемся к NATS
	sc, err := stan.Connect(clusterID, clientID, stan.NatsURL(url))
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer sc.Close()

	log.Println("Connected to NATS Streaming")

	// Создаем тестовый заказ
	order := models.Order{
		OrderUID:          fmt.Sprintf("test-order-%d", time.Now().UnixNano()),
		TrackNumber:       "WBILMTESTTRACK",
		Entry:             "WBIL",
		Locale:            "en",
		InternalSignature: "",
		CustomerID:        "test",
		DeliveryService:   "meest",
		Shardkey:          "9",
		SmID:              99,
		DateCreated:       time.Now(),
		OofShard:          "1",
		Delivery: models.Delivery{
			Name:    "Test Testov",
			Phone:   "+9720000000",
			Zip:     "2639809",
			City:    "Kiryat Mozkin",
			Address: "Ploshad Mira 15",
			Region:  "Kraiot",
			Email:   "test@gmail.com",
		},
		Payment: models.Payment{
			Transaction:  "", // Намеренно оставляем пустым для теста валидации
			RequestID:    "",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       1817,
			PaymentDt:    1637907727,
			Bank:         "alpha",
			DeliveryCost: 1500,
			GoodsTotal:   317,
			CustomFee:    0,
		},
		Items: []models.Item{
			{
				ChrtID:      9934930,
				TrackNumber: "WBILMTESTTRACK",
				Price:       453,
				Rid:         "ab4219087a764ae0btest",
				Name:        "Mascaras",
				Sale:        30,
				Size:        "0",
				TotalPrice:  317,
				NmID:        2389212,
				Brand:       "Vivienne Sabo",
				Status:      202,
			},
		},
	}

	// Исправляем payment.transaction чтобы совпадало с order_uid
	order.Payment.Transaction = order.OrderUID

	// Конвертируем в JSON
	jsonData, err := json.Marshal(order)
	if err != nil {
		log.Fatalf("Failed to marshal order: %v", err)
	}

	// Публикуем сообщение
	err = sc.Publish(subject, jsonData)
	if err != nil {
		log.Fatalf("Failed to publish message: %v", err)
	}

	log.Printf("Message published to subject %s", subject)
	log.Printf("Order UID: %s", order.OrderUID)
	log.Printf("JSON data: %s", string(jsonData))
}
