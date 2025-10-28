package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
	"wb-orders-service/config"
	"wb-orders-service/models"
	"wb-orders-service/repository"
)

func main() {
	cfg := config.Load()

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.DBName, cfg.Database.SSLMode)

	repo, err := repository.NewPostgresRepository(connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer repo.Close()

	// Создаем тестовый заказ
	testOrder := &models.Order{
		OrderUID:          "test-order-123",
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
			Transaction:  "test-order-123",
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

	// Тестируем сохранение
	fmt.Println("Saving test order...")
	err = repo.SaveOrder(testOrder)
	if err != nil {
		log.Fatalf("Failed to save order: %v", err)
	}

	// Тестируем чтение
	fmt.Println("Reading test order...")
	readOrder, err := repo.GetOrderByUID("test-order-123")
	if err != nil {
		log.Fatalf("Failed to read order: %v", err)
	}

	// Выводим результат
	jsonData, _ := json.MarshalIndent(readOrder, "", "  ")
	fmt.Printf("Read order:\n%s\n", string(jsonData))

	// Тестируем получение всех заказов
	fmt.Println("Getting all orders...")
	allOrders, err := repo.GetAllOrders()
	if err != nil {
		log.Fatalf("Failed to get all orders: %v", err)
	}
	fmt.Printf("Total orders in DB: %d\n", len(allOrders))
}
